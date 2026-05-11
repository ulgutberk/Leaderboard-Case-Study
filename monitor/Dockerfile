FROM alpine:3.18
RUN apk add --no-cache curl msmtp
COPY monitor.sh /monitor.sh
RUN chmod +x /monitor.sh
ENTRYPOINT ["/monitor.sh"]