# Leaderboard Case Study

## Architecture

All client requests flow through the Leaderboard Service (Go REST API), which connects to two data stores:

- **Postgres** — persistent storage for board and user metadata.
- **Redis ZSET** — fast score operations (read/write/ranking).

```
Client
  └── Leaderboard Service (Go REST API)
        ├── Postgres  (boards, users, scores — persistent)
        └── Redis     (leaderboard ZSET — fast ranked access)
```

### Project Structure

```
cmd/app/main.go                  # Entry point
internal/
  handlers/board_handler.go      # HTTP layer — routes requests to service
  services/board_service.go      # Business logic
  repositories/board_repo.go     # Data access (Postgres + Redis)
  models/                        # Domain models (Board, Score, User)
configs/config.go                # Environment-based configuration
migrations/001_init.sql          # Postgres schema (users, boards, scores)
Dockerfile
docker-compose.yml
```

## Running

Start all services (Postgres, Redis, Leaderboard API) with Docker:

```sh
docker compose up --build
```

To reset the database and re-run migrations:

```sh
docker compose down -v
docker compose up
```

Health check:

```sh
curl http://localhost:8080/health
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/boards` | Create a new board |
| GET | `/boards/{id}` | Get board by ID |
| POST | `/boards/{id}/scores` | Set a user's score |
| GET | `/boards/{id}/scores?limit=10` | Get top scores |
| POST | `/boards/{id}/reset` | Reset all scores on a board |

## Requirements

- Docker & Docker Compose
- Go 1.21+ (only needed for local development outside Docker)