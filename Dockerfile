FROM golang:1.26.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Install swag CLI for Swagger doc generation
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.1

COPY . .

# Regenerate Swagger docs so the embedded spec is always up-to-date
RUN swag init -g cmd/app/main.go -o docs

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/leaderboard ./cmd/app/

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/leaderboard .

EXPOSE 8080

CMD ["./leaderboard"]