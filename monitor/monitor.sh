#!/bin/sh

PREV_STATUS=""

send_email() {
  SUBJECT="$1"
  BODY="$2"
  printf "Subject: %s\nTo: %s\nFrom: %s\n\n%s\n" \
    "$SUBJECT" "$ALERT_EMAIL_TO" "$ALERT_EMAIL_FROM" "$BODY" \
    | msmtp --file=/etc/msmtp/msmtprc "$ALERT_EMAIL_TO"
}

# Startup test email
echo "[STARTUP] $(date) - Monitor started, sending test email..."
send_email \
  "[LEADERBOARD] Monitor Started" \
  "Time: $(date)\nLeaderboard monitor service has started successfully.\nYou will receive alerts if the service goes down." \
  && echo "[STARTUP] Test email sent successfully." \
  || echo "[STARTUP] WARNING: Test email failed — check msmtprc config."

while true; do
  RESPONSE=$(curl -s -o /tmp/health.json -w '%{http_code}' http://leaderboard:8080/health || echo "000")
  BODY=$(cat /tmp/health.json 2>/dev/null || echo "{}")

  if [ "$RESPONSE" != "200" ]; then
    if [ "$PREV_STATUS" != "down" ]; then
      send_email \
        "[LEADERBOARD] Service DOWN - HTTP $RESPONSE" \
        "Time: $(date)\nHTTP Status: $RESPONSE\nHealth Response: $BODY" \
        && echo "[ALERT] $(date) - Service DOWN! HTTP=$RESPONSE Body=$BODY | Email sent." \
        || echo "[ALERT] $(date) - Service DOWN! HTTP=$RESPONSE Body=$BODY | WARNING: Email failed."
      PREV_STATUS="down"
    fi
  else
    if [ "$PREV_STATUS" = "down" ]; then
      echo "[RECOVERY] $(date) - Service recovered."
      send_email \
        "[LEADERBOARD] Service RECOVERED" \
        "Time: $(date)\nService is back online.\nHealth Response: $BODY" \
        && echo "[RECOVERY] Email sent." \
        || echo "[RECOVERY] WARNING: Email failed."
    fi
    PREV_STATUS="up"
  fi

  sleep 3
done