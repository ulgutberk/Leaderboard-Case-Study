#!/bin/bash

export PGPASSWORD=postgres
FAILED=false
PROMOTED=false
PGBOUNCER_CFG=/etc/pgbouncer/pgbouncer.ini

echo "[FAILOVER] $(date) - Failover watcher started."

while true; do
  if pg_isready -h postgres-primary -U postgres -d leaderboard -t 3 > /dev/null 2>&1; then
    if [ "$FAILED" = "true" ]; then
      echo "[INFO] $(date) - Primary is back online. Starting failback..."

      # Get container and network info
      PRIMARY_CTR=$(docker ps -a --filter "name=postgres-primary" --format "{{.Names}}" | head -1)
      NETWORK=$(docker inspect "$PRIMARY_CTR" \
        --format '{{range $k, $v := .NetworkSettings.Networks}}{{$k}}{{end}}' | head -1)
      VOLUME=$(docker inspect "$PRIMARY_CTR" \
        --format '{{range .Mounts}}{{if eq .Destination "/var/lib/postgresql/data"}}{{.Name}}{{end}}{{end}}')

      echo "[INFO] Container=$PRIMARY_CTR Network=$NETWORK Volume=$VOLUME"

      # 1. Stop primary container (postgres process owns the data dir)
      docker stop "$PRIMARY_CTR"
      echo "[INFO] $(date) - Primary container stopped."

      # 2. Run pg_rewind in temporary container mounting the same volume
      docker run --rm \
        --network "$NETWORK" \
        -e PGPASSWORD=postgres \
        -v "${VOLUME}:/var/lib/postgresql/data" \
        postgres:15 \
        bash -c "
          chown -R postgres:postgres /var/lib/postgresql/data &&
          gosu postgres pg_rewind \
            --target-pgdata=/var/lib/postgresql/data \
            --source-server='host=postgres-replica port=5432 user=postgres password=postgres dbname=postgres' &&
          touch /var/lib/postgresql/data/standby.signal &&
          echo \"primary_conninfo = 'host=postgres-replica port=5432 user=replicator password=replicator_pass'\" \
            >> /var/lib/postgresql/data/postgresql.auto.conf
        " \
        && echo "[INFO] $(date) - pg_rewind completed, standby.signal created." \
        || { echo "[FAILBACK ERROR] pg_rewind failed."; docker start "$PRIMARY_CTR"; FAILED=false; PROMOTED=false; sleep 5; continue; }

      # 3. Start primary as standby (will replay WAL from replica)
      docker start "$PRIMARY_CTR"
      echo "[INFO] $(date) - Primary started as standby, waiting for WAL sync..."

      # 4. Wait for WAL lag = 0 (max 60s)
      SYNCED=false
      for i in $(seq 1 30); do
        sleep 2
        LAG=$(psql -h postgres-primary -U postgres -d leaderboard -t \
          -c "SELECT pg_wal_lsn_diff(pg_last_wal_receive_lsn(), pg_last_wal_replay_lsn());" \
          2>/dev/null | tr -d ' \n')
        echo "[INFO] $(date) - WAL lag: ${LAG:-connecting...} bytes"
        [ "$LAG" = "0" ] && SYNCED=true && break
      done

      if [ "$SYNCED" = "false" ]; then
        echo "[FAILBACK] WARNING: Sync timeout. Will promote anyway."
      fi

      # 5. Promote primary back (removes standby.signal)
      psql -h postgres-primary -U postgres -d leaderboard \
        -c "SELECT pg_promote();" > /dev/null 2>&1 \
        && echo "[INFO] $(date) - Primary promoted back." \
        || echo "[FAILBACK] WARNING: pg_promote failed."

      # 6. Point PgBouncer back to primary
      cp "$PGBOUNCER_CFG" /tmp/pgbouncer_new.ini
      sed "s/host=postgres-replica/host=postgres-primary/" \
        /tmp/pgbouncer_new.ini > "$PGBOUNCER_CFG"
      rm -f /tmp/pgbouncer_new.ini

      PGPASSWORD=pgbounceradmin psql -h pgbouncer -p 5432 \
        -U pgbounceradmin pgbouncer -c "RELOAD;" > /dev/null 2>&1 \
        && echo "[INFO] $(date) - PgBouncer reloaded — traffic → primary." \
        || echo "[INFO] WARNING: PgBouncer reload failed."

      # 7. Send email
      export PGPASSWORD=postgres
      printf "Subject: [LEADERBOARD] FAILBACK Complete\nTo: %s\nFrom: %s\n\nTime: %s\nFailback complete.\nPrimary is back and serving traffic.\nReplica is now following primary again.\n" \
        "$ALERT_EMAIL_TO" "$ALERT_EMAIL_FROM" "$(date)" \
        | msmtp --file=/etc/msmtp/msmtprc "$ALERT_EMAIL_TO" \
        && echo "[INFO] Failback email sent." \
        || echo "[INFO] WARNING: Failback email failed."
    fi
    FAILED=false
    PROMOTED=false
  else
    if [ "$PROMOTED" = "false" ]; then
      echo "[FAILOVER] $(date) - Primary DOWN. Promoting replica..."

      # Promote replica
      if psql -h postgres-replica -U postgres -d leaderboard \
        -c "SELECT pg_promote();" > /dev/null 2>&1; then
        echo "[FAILOVER] $(date) - Replica promoted."
      else
        echo "[FAILOVER] $(date) - WARNING: pg_promote() failed (may already be primary)."
      fi

      # Update PgBouncer config to point to replica
      cp "$PGBOUNCER_CFG" /tmp/pgbouncer_new.ini
      sed "s/host=postgres-primary/host=postgres-replica/" \
        /tmp/pgbouncer_new.ini > "$PGBOUNCER_CFG"
      rm -f /tmp/pgbouncer_new.ini

      # Reload PgBouncer (once)
      PGPASSWORD=pgbounceradmin psql -h pgbouncer -p 5432 \
        -U pgbounceradmin pgbouncer -c "RELOAD;" > /dev/null 2>&1 \
        && echo "[FAILOVER] $(date) - PgBouncer reloaded — traffic → replica." \
        || echo "[FAILOVER] $(date) - WARNING: PgBouncer reload failed."

      # Send alert email
      export PGPASSWORD=postgres
      printf "Subject: [LEADERBOARD] FAILOVER - Primary DOWN\nTo: %s\nFrom: %s\n\nTime: %s\nPostgres primary is DOWN.\nReplica has been promoted.\nPgBouncer now routes traffic to replica.\nManual intervention required to restore primary.\n" \
        "$ALERT_EMAIL_TO" "$ALERT_EMAIL_FROM" "$(date)" \
        | msmtp --file=/etc/msmtp/msmtprc "$ALERT_EMAIL_TO" \
        && echo "[FAILOVER] Alert email sent." \
        || echo "[FAILOVER] WARNING: Alert email failed."

      FAILED=true
      PROMOTED=true
    else
      echo "[INFO] $(date) - Replica already promoted. Waiting for manual intervention."
    fi
  fi

  sleep 5
done