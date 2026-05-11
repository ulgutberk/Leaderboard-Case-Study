#!/bin/bash
set -e

# Allow replicator user to connect for streaming replication
echo "host replication replicator all md5" >> /var/lib/postgresql/data/pg_hba.conf

# Reload config so the new rule takes effect immediately
psql -U postgres -c "SELECT pg_reload_conf();"