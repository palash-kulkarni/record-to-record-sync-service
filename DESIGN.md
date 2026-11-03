# Record-to-Record Sync Service

## Sub-problem chosen

I picked the outbound sync pipeline: an internal record change comes in, gets validated, transformed into the target CRM's schema, checked against the rate limit, and dispatched. This slice touches every meaningful concern in the system (validation, schema mapping, provider abstraction, rate limiting, CRUD routing) without trying to boil the ocean.

The inbound direction (external to internal) is structurally the same with transforms reversed, so solving outbound gives you the pattern for both.

---

## Assumptions

Internal changes arrive as `SyncEvent` structs from an event bus (Kafka, CDC, or similar). The source is mocked in `cmd/main.go`. Each event carries exactly one record; no batching. Events may be replayed, so every event has a `CorrelationID` for tracing and deduplication.

Rate limiting is scoped per `(providerID, accountID)` pair. Salesforce and HubSpot have different API quotas, and one account exhausting its bucket should not affect others. I used a token bucket rather than a fixed window because fixed windows cause burst spikes at every reset boundary. Token bucket smooths that out while still capping the long-term rate.

Schema transformation is stateless: same input always produces same output, no external lookups. This keeps transforms independently testable and easy to reason about.

Three failure classes with different retry semantics: transform/validation errors are bad data and go to dead-letter (no retry); rate-limited requests back-pressure the caller to retry with backoff; provider API errors are transient and also retry with backoff.

---

## High-level flow

```
SyncEvent (action + internal record)
    |
InputValidator      rejects malformed events early
    |
SchemaTransformer   maps internal model to provider-specific external model
    |
OutputValidator     validates transformed payload against provider schema
    |
RateLimiter         token bucket per (providerID, accountID)
    |
CRMProvider         executes CRUD via external API (mocked)
    |
SyncResult          SUCCESS | TRANSFORM_ERROR | VALIDATION_ERROR |
                    RATE_LIMITED | PROVIDER_ERROR
```

---

## Key design decisions

### Provider and transformer interfaces

Both `CRMProvider` and `SchemaTransformer` are interfaces, one implementation per CRM. The pipeline depends only on the interfaces, so adding a new provider means a new struct and nothing else changes. This also makes each transformer independently testable without touching the pipeline.

### Result type for expected failures

`SyncResult` carries a typed `Status` rather than using Go errors for expected failure modes. Rate limiting and validation failures are not bugs, they are normal outcomes. Returning them as results rather than errors keeps the pipeline readable and avoids wrapping errors across multiple call levels.

### Token bucket rate limiter

One bucket per `(providerID:accountID)` key, lazy-initialised in a `sync.Map`. The refill is calculated from elapsed time on each `TryAcquire` call, so there is no background goroutine needed. The current implementation is in-memory and scoped to a single process. See the scaling section for what changes in production.

---

## Package layout

```
record-to-record-sync-service/
  model/        SyncEvent, InternalRecord, ExternalRecord, SyncResult, SyncAction
  validation/   input and output validators
  transform/    SchemaTransformer interface, SalesforceTransformer, HubSpotTransformer
  ratelimit/    TokenBucket, RateLimiter, ProviderRateConfig
  provider/     CRMProvider interface, mock Salesforce and HubSpot
  pipeline/     SyncPipeline orchestrator
  cmd/          main.go demo
```

---

## Scaling to 300M requests/day

300M/day is roughly 3,500 RPS sustained with bursts to 7,000. A single process can handle around 500 RPS, so the path to scale is horizontal: multiple instances consuming from a partitioned Kafka topic. Partitioning by `accountID` preserves per-account ordering and means no cross-instance coordination in the hot path.

**Distributed rate limiting.** The in-memory token bucket breaks under horizontal scaling because each instance has its own bucket, so N instances effectively multiply the allowed rate by N against the external API. The fix is to move bucket state to Redis and use a Lua script for atomic check-and-decrement. The `TryAcquire` interface stays the same; only the implementation changes. A single Redis `EVAL` round-trip is around 0.5ms, acceptable given the CRM API call that follows it dominates latency anyway.

**Retry and dead-letter.** Right now `RATE_LIMITED` and `PROVIDER_ERROR` results are returned to the caller with no automatic retry. At scale, the caller needs to publish these to a retry topic with exponential backoff metadata in the message header. After N attempts the event moves to a dead-letter topic for manual inspection. Kafka consumer group offset commit gives at-least-once delivery; the `CorrelationID` already in the model handles deduplication on the receiving end.

**Circuit breaker per provider.** Without one, a CRM API outage causes every in-flight request to fail, fills the retry queue, and wastes rate limit tokens on requests that will not succeed. A circuit breaker tracks consecutive `PROVIDER_ERROR` counts per provider and opens after a threshold, failing fast during a cooldown window and probing with a single request before closing again. This is the main mechanism for keeping 99.9% availability when an external dependency goes down.

**Latency budget.** The CRM API call dominates at typically 100-500ms. Transform, validation, and a Redis rate limit check together add under 5ms. The practical lever is concurrency: running multiple goroutines per Kafka partition consumer to overlap in-flight HTTP calls. The pipeline is goroutine-safe; nothing in `SyncPipeline` is shared mutable state.

---

## How to run

```bash
go test ./...       # run all tests
go run cmd/main.go  # run demo
```
