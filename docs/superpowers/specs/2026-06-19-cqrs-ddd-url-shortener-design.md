# Design Spec — URL Shortener (CQRS + DDD, Hexagonal)

Date: 2026-06-19
Status: Approved

## Goal

Refactor the existing URL-shortener into a clean, production-grade CQRS system that
faithfully applies tactical DDD (Hexagonal/Onion architecture) with durable,
loss-free event delivery via the Transactional Outbox pattern + NATS JetStream.

## Functional requirements

1. Create a short link.
2. Disable a link.
3. Redirect short → long URL.
4. View click analytics per link.

## Architectural requirements

- **Command side** writes to a **Write DB** (Postgres): aggregate state + outbox.
- **Query side** serves from a separate **Read DB** (Redis): `short→long` map + click counters.
- Every redirect emits a `LinkClicked` event; a Projector updates the read model
  (eventual consistency).
- All comms over REST/HTTP.
- Data must persist across restarts.

## Process topology (3 binaries, 1 image)

- **api** — Command (POST) + Query (GET) HTTP, separated internally into two layers.
- **relay** — polls `outbox_events`, publishes to JetStream (solves dual-write).
- **projector** — durable JetStream consumer; idempotently updates Postgres
  `link_analytics` (durable count) and Redis (fast read model); rebuilds Redis on boot.

## Layering (Onion / Hexagonal)

```
internal/
  domain/link/        aggregate, value objects, domain events, errors, repository port
  domain/shared/      DomainEvent contract
  application/        command/ query/ port/  (use cases + ports)
  infrastructure/     persistence/postgres, persistence/redis, messaging/jetstream, config, logger
  interfaces/http/    gin handlers, router, domain-error → HTTP mapping
cmd/                  api/ relay/ projector/
```

Domain has **zero** infra dependencies. Persistence models (GORM tags) live only in
`infrastructure/persistence/postgres/models.go` and are mapped to/from the aggregate
(persistence ignorance).

## Domain model

- **Aggregate root `Link`**: private fields; `NewLink(code, url)` raises `LinkCreated`;
  `Reconstitute(...)` loads without events; `Disable()` enforces invariant
  (`ErrLinkAlreadyDisabled`) and raises `LinkDisabled`; `PullEvents()`.
- **Value objects**: `LinkID` (uuid), `ShortCode` (base62, len 7, crypto/rand),
  `URL` (scheme http/https + host, normalized), `LinkStatus` (Active/Disabled).
- **Domain events**: `LinkCreated`, `LinkDisabled`, `LinkClicked` — each with
  `EventID`, `OccurredAt`, `AggregateID`, `ShortCode`.

## Postgres tables (Write DB)

- `links` — aggregate state (+ `version` for optimistic concurrency).
- `outbox_events` — id, aggregate_id, type, payload(jsonb), occurred_at, published_at.
- `link_analytics` — short_code, click_count, last_clicked_at (durable count = source of truth).
- `processed_events` — event_id (projector idempotency / dedup).

## Redis (Read DB)

- `link:{short}` → long URL.
- `clicks:{short}` → counter.

## Flows & guarantees

| Op | Path | Guarantee |
|---|---|---|
| Create | aggregate → `links`+outbox in one TX → relay → JS → projector → Redis | atomic, no loss |
| Disable | load → `Disable()` → `links`+outbox in one TX → projector → DEL Redis | redirect → 404/410 after |
| Redirect | read `link:{short}` from Redis → synchronous INSERT `LinkClicked` into outbox → 302 | click durable from outbox |
| Stats | read `clicks:{short}` from Redis | eventual consistency |

- **Idempotency**: projector checks `processed_events` before applying.
- **Rebuildability**: on boot, projector rebuilds Redis from `links` + `link_analytics`.

## Infra hardening

- docker-compose: named volumes (Postgres, Redis AOF), JetStream (`-js`), healthchecks,
  `depends_on: service_healthy`.
- Dockerfile: multi-stage, Go 1.25, builds all three binaries.
- Config from env vars with defaults; graceful shutdown (signal/context) in all binaries.

## Testing

- Domain unit tests (value-object validation, aggregate behavior + invariants).
- Application handler tests with in-memory fakes.

## Deliverable

- Persian final report (`docs/REPORT.md`): CQRS concepts & why split DBs; DDD tactical
  patterns; Transactional Outbox & dual-write; eventual consistency & idempotency;
  architecture diagram (Mermaid + ASCII); per-operation sequences; run/test guide.
