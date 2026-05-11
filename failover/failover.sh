#!/bin/bash

export PGPASSWORD=postgres
FAILED=false
PROMOTED=false
PGBOUNCER_CFG=/etc/pgbouncer/pgbouncer.ini

echo "[FAILOVER] $(date) - Failover watcher started."

while true; do
  if pg_isready -h postgres-primary -U postgres -d leaderboard -t 3 > /dev/null 2>&1; then
    if [ "$FAILED" = "true" ]; then
      echo "[INFO] $(date) - Primary is back online. Manual re-sync required before switching back."
      export PGPASSWORD=postgres
      printf "Subject: [LEADERBOARD] Primary Back Online\nTo: %s\nFrom: %s\n\nTime: %s\nPostgres primary is back online.\nNote: Replica is still acting as primary.\nManual re-sync and failback required before switching traffic back.\n" \
        "$ALERT_EMAIL_TO" "$ALERT_EMAIL_FROM" "$(date)" \
        | msmtp --file=/etc/msmtp/msmtprc "$ALERT_EMAIL_TO" \
        && echo "[INFO] Recovery email sent." \
        || echo "[INFO] WARNING: Recovery email failed."
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