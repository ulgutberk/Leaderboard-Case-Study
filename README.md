# Leaderboard Service

A production-ready REST API service that provides real-time leaderboard management for high-traffic gaming systems.

Features multi-board support, automatic score reset scheduling, Redis-based fast ranking, PostgreSQL primary/replica replication, automatic failover via Redis Sentinel, and PgBouncer connection pooling.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| API | Go 1.23, gorilla/mux |
| Cache | Redis 7 (ZSET) + Redis Sentinel (1 master, 2 replicas, 3 sentinels) |
| Persistence | PostgreSQL 15 (primary + streaming replica) |
| Connection Pool | PgBouncer |
| Container | Docker, Docker Compose |
| Docs | Swagger (swaggo) |

---

## Architecture

```
                        ┌─────────────────────────────────────┐
                        │         Leaderboard API (Go)         │
                        │  handlers → services → repositories  │
                        └────────────┬────────────┬─────────┬──┘
                                     │            │         │
                   ┌─────────────────▼──┐    ┌────▼──────┐  └──────────────────────┐
                   │     PgBouncer      │    │  Redis    │   │  Monitor Service      │
                   │  (connection pool) │    │  Sentinel │   │  (health check /      │
                   └────────┬───────────┘    │  ×3       │   │   email alerts)       │
                            │               └────┬──────┘   └───────────────────────┘
                            │                    │
               ┌────────────▼──────┐    ┌────────▼───────────────────┐
               │  PostgreSQL       │    │  Redis Master               │
               │  Primary          │    │  (ZSET scores + TS hashes)  │
               └────────┬──────────┘    └──────────┬─────────────────┘
                        │ streaming repl            │ async repl
               ┌────────▼──────────┐    ┌──────────▼─────────────────┐
               │  PostgreSQL       │    │  Redis Slave ×2             │
               │  Replica          │    │                             │
               └───────────────────┘    └────────────────────────────┘

               ┌──────────────────────────────────────────────────────────────────┐
               │  Failover Watcher                                                 │
               │  monitors  → postgres-primary          (pg_isready polling)       │
               │  controls  → postgres-replica          (promote on failure)        │
               │  controls  → pgbouncer                 (config update + reload)    │
               └──────────────────────────────────────────────────────────────────┘
```

### Data Flow

**Write path:** Postgres UPSERT → Redis ZADD + HSET (atomic pipeline)  
**Read path:** Redis only (O(log N + k) ranking via `ZRANGEBYSCORE`)  
**Crash recovery:** On startup, `WarmCache()` rebuilds Redis from Postgres

### Project Structure

```
cmd/
  app/main.go                     # Entry point: router, middleware, DI
  scheduler/                      # Board reset scheduler

internal/
  handlers/                       # HTTP layer (board, score, user)
  services/                       # Business logic (board, score, user)
  repositories/                   # Data access layer (Postgres + Redis)
  models/                         # Domain models
  
configs/config.go                 # Env-based configuration

pkg/middleware/
  rate_limiter.go                 # Redis-backed IP rate limiter
  logger.go                       # Request logger

migrations/                       # Sequential SQL migration files
failover/failover.sh              # PostgreSQL auto-failover + failback script
monitor/monitor.sh                # Health check + email alert service

pgbouncer/                        # PgBouncer configuration

docs/                             # Swagger JSON/YAML
```

---

## Setup & Running

### Requirements

- Docker >= 24
- Docker Compose >= 2.20

### Start All Services

```bash
docker compose up --build
```

Once all services are healthy, the API is available at `http://localhost:8080`.

### Health Check

```bash
curl http://localhost:8080/health
```

### Reset Database and Restart

```bash
docker compose down -v
docker compose up --build
```

### Swagger UI

```
http://localhost:8080/swagger/index.html
```

### Run Tests

```bash
# All tests
go test ./...

# Handler tests only
go test ./tests/handlers/...

# Service tests only
go test ./tests/services/...

# Repository tests only
go test ./tests/repositories/...

# Verbose output
go test -v ./...
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDR` | `:8080` | HTTP server address |
| `POSTGRES_URL` | `postgres://...@localhost:5432/leaderboard` | Primary Postgres connection |
| `POSTGRES_READ_URL` | `postgres://...@localhost:5432/leaderboard` | Read replica connection |
| `REDIS_ADDR` | `localhost:6379` | Direct Redis address (if no Sentinel) |
| `REDIS_SENTINEL_ADDR` | — | Comma-separated Sentinel addresses |
| `REDIS_MASTER_NAME` | `mymaster` | Sentinel master set name |
| `RESET_SCAN_INTERVAL` | `1m` | Board reset scan interval |
| `RATE_LIMIT_REQUESTS` | `100` | Allowed requests per window |
| `RATE_LIMIT_WINDOW` | `1m` | Rate limit window size |

---

## API Reference

All endpoints return JSON. Error responses use the format `{"error": "message"}`.

### Users

#### `POST /users` — Create User

```json
// Request
{
  "id": "user_abc",
  "username": "player1"
}

// Response 201
{
  "id": "user_abc",
  "username": "player1"
}
```

#### `GET /users/{id}` — Get User

```json
// Response 200
{
  "id": "user_abc",
  "username": "player1"
}
```

---

### Boards

#### `POST /boards` — Create Board

```json
// Request
{
  "name": "Weekly Champions",
  "description": "Weekly leaderboard",
  "schedule": {
    "type": "interval",
    "intervalSeconds": 604800
  }
}

// Response 201
{
  "boardId": "board_1",
  "name": "Weekly Champions",
  "description": "Weekly leaderboard",
  "schedule": { "type": "interval", "intervalSeconds": 604800 },
  "nextResetAt": "2026-05-21T10:00:00Z"
}
```

#### `GET /boards` — List All Boards

```json
// Response 200
[
  { "boardId": "board_1", "name": "Weekly Champions" },
  { "boardId": "board_2", "name": "All-Time Rankings" }
]
```

#### `GET /boards/{boardId}` — Get Board Detail

```json
// Response 200
{
  "boardId": "board_1",
  "name": "Weekly Champions",
  "description": "Weekly leaderboard",
  "schedule": { "type": "interval", "intervalSeconds": 604800 },
  "nextResetAt": "2026-05-21T10:00:00Z"
}
```

---

### Scores

#### `POST /boards/{boardId}/scores` — Set / Update Score

```json
// Request
{
  "userId": "user_abc",
  "score": 4200
}

// Response 200
{
  "boardId": "board_1",
  "userId": "user_abc",
  "score": 4200
}
```

**Note:** When a score is updated, `created_at` is preserved — on equal scores, the earlier submission ranks higher.

#### `GET /boards/{boardId}/scores?limit=10` — Get Top N Scores

```json
// Response 200
[
  { "userId": "user_abc", "score": 9500 },
  { "userId": "user_xyz", "score": 8200 }
]
```

#### `GET /boards/{boardId}/scores/{userId}/surroundings?n=3` — Get Surrounding Ranks

```json
// Response 200
{
  "user": { "userId": "user_abc", "score": 4200 },
  "above": [
    { "userId": "user_top1", "score": 4800 },
    { "userId": "user_top2", "score": 4600 }
  ],
  "below": [
    { "userId": "user_low1", "score": 3900 },
    { "userId": "user_low2", "score": 3500 }
  ]
}
```

#### `POST /boards/{boardId}/reset` — Reset All Scores

```json
// Response 200
{ "message": "scores reset" }
```

#### `POST /boards/{boardId}/mock-scores` — Generate Mock Data

```json
// Request
{ "count": 100 }

// Response 200
{
  "boardId": "board_1",
  "count": 100,
  "scores": [ ... ]
}
```

---

## Engineering Decisions

### Ranking with Redis ZSET

Ranking operations run in O(log N + k) complexity. The alternative — a Postgres `ORDER BY score DESC` query — would introduce significant latency under high read load. Redis absorbs that pressure entirely.

### Write-Through Cache Strategy

Every score write is first UPSERTed into Postgres (for data integrity and preserving `created_at`), then mirrored to Redis atomically via a pipeline. Redis is used exclusively for reads. On service restart, `WarmCache()` rebuilds Redis from Postgres — consistency is never broken.

### Tie-Breaking

Redis ZSET accepts only a single float score per member. To break ties, a Redis HASH stores each user's first submission timestamp (unix nanoseconds) per board. When reading rankings, users with equal scores are sorted by this timestamp — **the earlier submission ranks higher.**

### Redis Sentinel (High Availability)

Deployed as 1 master + 2 replicas + 3 sentinels with quorum = 2. When the master fails, Sentinel elects a new master automatically; the `go-redis` sentinel client reconnects transparently. No application-level intervention is needed.

### PostgreSQL Streaming Replication + Failover Watcher

The `failover-watcher` service periodically checks the health of the primary:

- **On primary failure:** The replica is promoted, the PgBouncer config is updated to `host=postgres-replica`, and `RELOAD` is issued. Traffic is redirected to the new primary with no downtime.
- **On primary recovery:** `pg_rewind` closes the WAL gap, a `standby.signal` is added, WAL sync is awaited, then the primary is re-promoted and PgBouncer is updated back to `host=postgres-primary`.

### PgBouncer Connection Pooling

PostgreSQL has a hard connection limit. PgBouncer with `MAX_CLIENT_CONN=1000` eliminates this bottleneck. During failover, a config reload instantly redirects all pooled connections to the new primary.

### Redis-Backed Rate Limiter

A fixed-window counter per IP is maintained in Redis using an atomic Lua script. `INCR` and `EXPIRE` are executed in a single Lua call, eliminating race conditions. The rate limit is consistent across multiple service instances. On Redis failure, the middleware **fails open** — the service is never blocked.

### Automatic Board Reset

Each board carries a `schedule.intervalSeconds` that defines its reset period. A background scheduler calls `ResetDueBoards()` every minute; `activePeriodStart` computes whether a board is due by comparing `lastResetAt` against the current period boundary. Resets are applied to both Redis and Postgres atomically.

---

## Error Handling & Edge Cases

| Scenario | Behavior |
|----------|----------|
| Redis down | Write fails (Postgres was written, Redis mirror returns error). On restart, `WarmCache` restores Redis. Rate limiter fails open. |
| PostgreSQL primary down | Failover watcher promotes the replica; PgBouncer routing is updated automatically. |
| Primary comes back online | `pg_rewind` closes the WAL gap; failback is fully automated. |
| Duplicate score submission | Postgres `ON CONFLICT DO UPDATE` makes it idempotent — `created_at` is preserved. |
| Invalid boardId | Returns 404; no score operation is performed. |
| Invalid userId (FK violation) | Postgres `23503` FK constraint error → 404 "User not found". |
| Duplicate board name | Postgres `23505` unique constraint → 400 with a descriptive error message. |
| Rate limit exceeded | 429 `{"error":"rate limit exceeded"}` + `Retry-After: 60` header. |
| Service down (monitor) | An email alert is sent; a second email is sent on recovery. |

---

## Test Strategy

```
tests/
  handlers/        # HTTP handler tests (via httptest)
  services/        # Business logic unit tests (mock repository)
  repositories/    # Repository tests
internal/
  repositories/    # board_helpers_test, score_helpers_test
  services/        # score_helpers_test
```

- **Handler tests:** The HTTP layer is tested in isolation using `httptest.NewRecorder`.
- **Service tests:** Repository interfaces are mocked; business logic is tested independently.
- **Repository tests:** Integration tests against real Postgres + Redis connections.
- **Mocking:** Interface-based dependency injection means no real dependencies are required in unit tests.

```bash
go test ./...
```

---

## Monitoring

The `monitor` service polls the `/health` endpoint every 3 seconds. If the service goes down, an email alert is sent; a recovery email follows when the service comes back online.

```
ALERT_EMAIL_TO=your@email.com
ALERT_EMAIL_FROM=sender@email.com
```

Email delivery is handled by `msmtp`. Configuration file: `monitor/msmtprc` (added .gitignore).

---

## Future Improvements

- Kubernetes deployment (for horizontal scaling)
- Prometheus + Grafana metrics (request rate, latency, error rate)
- JWT authentication middleware
- Audit log (score change history)