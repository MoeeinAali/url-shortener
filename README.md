# URL Shortener — CQRS + DDD (Hexagonal)

A URL shortener + click analytics built with **CQRS** and tactical **Domain-Driven
Design** in a Hexagonal/Onion layout. Writes and reads use **separate databases**,
and events flow through a **Transactional Outbox + NATS JetStream** so nothing is
ever lost and the read model is fully rebuildable.

> The full design write-up (in Persian) lives in [`docs/REPORT.md`](docs/REPORT.md).
> The approved design spec is in [`docs/superpowers/specs/`](docs/superpowers/specs/).

## Stack

| Concern | Tech |
|---|---|
| Write DB (source of truth) | PostgreSQL (links, outbox, analytics, processed events) |
| Read DB (fast model) | Redis (`link:*`, `clicks:*`) |
| Durable event bus | NATS JetStream |
| Language / HTTP | Go 1.25 / Gin |
| Processes | `api`, `relay`, `projector` (one image) |

## Architecture

```
   POST /links, /disable ─▶ api (command) ─┬─▶ Postgres: links + outbox (one TX)
   GET  /:short          ─▶ api (query)  ──┤     ▲
   GET  /:short/stats                      │     │ poll
        │                                  ▼     │
        └─▶ Redis (read model) ◀─ projector ◀─ JetStream ◀─ relay ◀─ outbox
                                     │
                                     ├─▶ Postgres link_analytics (durable count)
                                     └─ Bootstrap: rebuild Redis from Postgres on boot
```

- **Command side** writes the aggregate **and** its domain events in one DB
  transaction (Transactional Outbox → solves dual-write).
- **relay** forwards unpublished outbox rows to JetStream (at-least-once).
- **projector** consumes events **idempotently** (dedup via `processed_events`),
  maintains the durable count in Postgres and the fast counters in Redis, and
  **rebuilds Redis from Postgres on startup**.

See [`docs/REPORT.md`](docs/REPORT.md) for component, sequence, and ASCII diagrams.

## Layout

```
internal/
  domain/         aggregate (Link), value objects, domain events, repository port
  application/    command/ query/ port/  (CQRS use cases + ports)
  infrastructure/ persistence/{postgres,redis}, messaging/jetstream, relay, projector, event, config, logger
  interfaces/http gin handlers, router, domain-error → HTTP mapping
cmd/              api/ relay/ projector/
```

## Run

```bash
docker compose up -d --build
```

Brings up `postgres`, `redis`, `nats`, `api`, `relay`, `projector`. Data persists
across restarts via named volumes (Postgres + Redis AOF + JetStream file storage).

### Example

```bash
# create
curl -s -XPOST localhost:8080/links -H 'Content-Type: application/json' \
     -d '{"url":"https://example.com"}'
# → {"short_code":"Ab3X9kZ","long_url":"https://example.com","short_url":"http://localhost:8080/Ab3X9kZ"}

curl -s -i localhost:8080/Ab3X9kZ            # 302 → Location: https://example.com
curl -s localhost:8080/links/Ab3X9kZ/stats   # {"short_code":"Ab3X9kZ","clicks":1}
curl -s -XPOST localhost:8080/links/Ab3X9kZ/disable
```

## API

| Method | Path | Description | Success |
|---|---|---|---|
| POST | `/links` | create (`{"url":"..."}`) | `201` |
| POST | `/links/:short/disable` | disable | `200` |
| GET | `/:short` | redirect (records a click) | `302` |
| GET | `/links/:short/stats` | click count | `200` |
| GET | `/healthz` | health | `200` |

## Configuration (env vars)

`HTTP_ADDR`, `POSTGRES_DSN`, `REDIS_ADDR`, `REDIS_PASSWORD`, `NATS_URL`,
`BASE_URL`, `RELAY_POLL_INTERVAL`, `RELAY_BATCH_SIZE` — all have docker-friendly
defaults in `internal/infrastructure/config`.

## Tests

```bash
go test ./...
```

Covers value-object validation, aggregate behavior/invariants, application
handlers (with in-memory fakes), and the HTTP router (route registration +
redirect/stats behavior).
