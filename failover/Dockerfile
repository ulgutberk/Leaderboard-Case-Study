FROM postgres:15
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    msmtp \
    ca-certificates \
    docker.io \
    && rm -rf /var/lib/apt/lists/*
COPY failover.sh /failover.sh
RUN chmod +x /failover.sh
ENTRYPOINT ["/failover.sh"]