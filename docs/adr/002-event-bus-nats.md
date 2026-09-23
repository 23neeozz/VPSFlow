# ADR-002: Event Bus — NATS JetStream

## Status

Accepted

## Context

BossCloud requires an event-driven architecture for domain events (VM lifecycle, auth events, billing metering, audit). The bus must support at-least-once delivery, consumer groups, replay, and low operational overhead for early deployment.

## Decision

Use **NATS JetStream** as the primary event bus.

- Subjects follow: `bosscloud.<context>.<entity>.<event>.v1`
- Transactional outbox pattern per service for reliable publication
- Idempotent consumers with deduplication via `event_id`
- Dead Letter Queue (DLQ) per subject stream

## Consequences

### Positive
- Low latency, simple operations
- Built-in persistence and consumer groups
- Lightweight for single-node to multi-node deployment
- Go client ecosystem mature

### Negative
- Less ecosystem than Kafka for very high throughput analytics
- May require migration to Kafka at extreme scale (millions events/sec)

### Neutral
- Event schema versioning managed independently via AsyncAPI contracts

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Apache Kafka | Industry standard, massive throughput | Heavy ops overhead for MVP | Premature for initial scale |
| RabbitMQ | Mature, flexible routing | Less suited for event sourcing replay | Weaker replay semantics |
| Redis Streams | Already in stack | Limited durability guarantees | Not sufficient for critical events |

## References

- [Event Contracts](../../contracts/asyncapi/events-v1.yaml)
