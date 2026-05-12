FROM docker:24-cli

RUN apk add --no-cache \
    bash \
    curl \
    postgresql-client \
    msmtp

COPY failover.sh /failover.sh
RUN chmod +x /failover.sh

ENTRYPOINT ["/failover.sh"]