FROM golang:1.26.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/leaderboard ./cmd/app/

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/leaderboard .

EXPOSE 8080

CMD ["./leaderboard"]