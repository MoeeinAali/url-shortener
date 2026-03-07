# URL Shortener

This project is a URL shortener implemented in Go using a **CQRS-style architecture** with separate models for writes and reads.

- **Write side** persists source-of-truth link data in **PostgreSQL**
- **Read side** serves fast queries from **Redis**
- **Event bus** uses **NATS** to propagate read-model updates asynchronously

## Tech Stack

- Go (Gin, GORM)
- PostgreSQL (write model)
- Redis (read model)
- NATS (event messaging)
- Docker Compose (local orchestration)

## CQRS in This Project

CQRS (Command Query Responsibility Segregation) splits operations into:

- **Commands (Write Model):** change system state (create/disable links)
- **Queries (Read Model):** read optimized views (redirect/stats)

In this codebase:

- Command API handlers call `CommandService` and write to PostgreSQL via `LinkRepository`
- Query API handlers call `QueryService` and read from Redis via `ReadRepository`
- Query operations can emit events (for example `link.clicked`) to NATS
- A projector service subscribes to events and updates Redis projections

## Why Separate Write DB and Read DB?

Using distinct write/read databases is useful in CQRS systems for these reasons:

- **Independent optimization:** relational integrity and transactions in PostgreSQL for writes, low-latency key/value reads in Redis
- **Scalability:** read traffic can scale independently from write traffic
- **Performance isolation:** heavy read load does not directly impact write throughput
- **Flexible read models:** Redis structures can be shaped for API needs, without changing write schema
- **Event-driven evolution:** projector can build additional projections without touching command logic

Tradeoff: read models are usually **eventually consistent** (updates may appear with slight delay).

## Architecture Diagram (Mermaid)

```mermaid
flowchart LR
  Client[Client] -->|POST /links\nPOST /links/:short/disable| API[API Service]
  Client -->|GET /:short\nGET /links/:short/stats| API

  subgraph Command Side
    API --> CH[Command Handler]
    CH --> CS[Command Service]
    CS --> WR[Write Repository]
    WR --> PG[(PostgreSQL - Write DB)]
  end

  subgraph Query Side
    API --> QH[Query Handler]
    QH --> QS[Query Service]
    QS --> RR[Read Repository]
    RR --> REDIS[(Redis - Read DB)]
  end

  QS -->|Publish link.clicked| NATS[(NATS)]

  subgraph Projection Worker
    NATS --> PRJ[Projector Service]
    PRJ --> RR2[Read Repository]
    RR2 --> REDIS
  end
```

## Repository Structure

```text
cmd/
  api/                 # HTTP API entrypoint
  projector/           # Background projector entrypoint
internal/
  handlers/            # HTTP handlers (command/query)
  service/             # Command and query services
  repository/          # PostgreSQL and Redis repositories
  projector/           # Event projector logic
  messaging/           # NATS client wrapper
  db/                  # DB clients and migrations
  domain/              # Domain entities and events
pkg/shortener/         # Short code generator
docker-compose.yml
Dockerfile
```

## API Endpoints

Base URL: `http://localhost:8080`

### 1) Create Short Link (Command)

- **Method:** `POST`
- **Path:** `/links`
- **Body:**

```json
{
  "url": "https://example.com/some/long/path"
}
```

- **Success response:** `201 Created`

```json
{
  "short": "Ab12Cd",
  "url": "https://example.com/some/long/path"
}
```

### 2) Disable Short Link (Command)

- **Method:** `POST`
- **Path:** `/links/:short/disable`
- **Success response:** `200 OK`

```json
{
  "status": "disabled"
}
```

### 3) Redirect by Short Code (Query)

- **Method:** `GET`
- **Path:** `/:short`
- **Success response:** `302 Found` (redirect to original URL)
- **Failure response:** `404 Not Found`

```json
{
  "error": "link not found"
}
```

### 4) Link Stats (Query)

- **Method:** `GET`
- **Path:** `/links/:short/stats`
- **Success response:** `200 OK`

```json
{
  "short": "Ab12Cd",
  "clicks": 42
}
```

## Event Flow

- **Published subject:** `link.clicked`
- **Payload:**

```json
{
  "short_code": "Ab12Cd"
}
```

- **Subscriber:** projector service
- **Projection update:** increment `clicks:<short_code>` in Redis

## Running Locally

### Option A: Docker Compose (recommended)

```bash
docker compose up --build
```

Services:

- API: `localhost:8080`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- NATS: `localhost:4222`

### Option B: Run binaries manually

1. Start PostgreSQL, Redis, NATS
2. Run API:

```bash
go run ./cmd/api
```

3. Run projector (separate terminal):

```bash
go run ./cmd/projector
```

## Configuration

Current config is loaded from `internal/config/config.go`:

- Postgres DSN: `postgres://admin:admin@postgres:5432/shortener?sslmode=disable`
- Redis address: `redis:6379`
- NATS URL: `nats://nats:4222`

For production, move these values to environment variables or a secret manager.

## Notes on Current Implementation

- The projector currently subscribes to `link.clicked` and updates click counters in Redis.
- Query endpoints read from Redis, so data must exist in the read model for redirects/stats.
- If you extend this project, consider publishing and projecting additional events (for example link-created / link-disabled) to keep Redis fully synchronized with PostgreSQL.

## CQRS Summary

This project demonstrates CQRS by separating **state-changing commands** from **read-optimized queries**, with asynchronous projection updates through NATS. The main benefit is clear separation of concerns, better read/write scalability, and a foundation for event-driven growth.
