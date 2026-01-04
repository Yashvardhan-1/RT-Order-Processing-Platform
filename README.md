## Real-Time Order Processing Platform (Event-Driven, Kafka-First)

This document outlines the architecture and phased plan to build a fully event-driven order processing platform. All services communicate via Kafka; no inter-service HTTP calls are used for business flows.

### Objectives
- Accept client orders, process payments, reserve inventory, notify customers, and emit analytics with strict per-order ordering guarantees.
- Scale horizontally via Kafka partitioning; maintain resilience to node failures (Kafka and services).
- Provide observability and compliance trails (audit topic).

### Repository Layout (planned)
```
infra/
 ├── docker-compose.yml
 └── kafka/
services/
 ├── order/
 ├── payment/
 ├── inventory/
 └── notification/
proto/
```

### Core Architecture
- **Event flow:** Client → Order Service → Kafka → Downstream Services (Payment, Inventory, Notification, Analytics). No synchronous HTTP between services; coordination is via Kafka topics.
- **Services (independent deployables, Kafka consumers/producers):**
  - **Order Service (Go):** Entry point. Exposes HTTP/REST (chi/gin) for clients and gRPC for internal admin/testing. Publishes `orders.created` (and `orders.audit`), enforces idempotency and trace propagation.
  - **Payment Service (Go):** Kafka-only. Consumes `orders.created`; uses transactional/idempotent producer to emit `payments.completed`. Persists idempotency and payment state in PostgreSQL.
  - **Inventory Service (Go):** Consumes `payments.completed`; manages reservations/commits/releases. Emits `inventory.updated`. Provides gRPC for admin/debug. Backed by PostgreSQL/Redis for stock state.
  - **Notification Service (Python):** Consumes `payments.completed` and `orders.created`; calls external email/SMS providers; emits `notifications.events` for delivery status.
  - **Analytics Service (Python):** Consumes across topics for real-time and batch metrics; may write aggregates back to warehouse/OLAP or derived topics.
- **Communication:** Kafka topics only; consumers use partition-aware processing to preserve per-key ordering.

### Kafka Cluster Setup
- **Topology:** 5 Kafka nodes, each runs both **Broker** and **KRaft controller** (quorum-based).
- **Rationale:** Majority quorum = 3; cluster tolerates 2 node failures while maintaining availability.
- **Replication targets:** Minimum replication factor 3 for durability; min in-sync replicas (min ISR) = 2 for availability under failure.

### Topics and Partitioning
| Topic | Partitions | Key | Purpose |
| --- | --- | --- | --- |
| `orders.created` | 48 | `order_id` | Order intake and orchestration root |
| `payments.completed` | 24 | `order_id` | Payment completion outcomes per order |
| `inventory.updated` | 24 | `sku` | Stock reservation/commit/release events |
| `notifications.events` | 12 | `user_id` | Customer messaging events and outcomes |
| `orders.audit` | 6 | `order_id` | Compliance and traceability |

**Partitioning rationale:** Keying by `order_id` guarantees ordering per order across dependent streams; enables horizontal scale by adding partitions. `inventory.updated` uses `sku` to keep per-item sequencing for stock changes. `notifications.events` uses `user_id` to sequence user-facing comms.

### High-Level Design (HLD)
- **Ordering guarantees:** Consumers process messages per partition sequentially; avoid cross-partition coordination. Exactly-once semantics optional via idempotent producers + transactional writes where required.
- **State storage:** Each service owns its state store (e.g., relational DB or compacted Kafka state store) and persists consumer offsets transactionally when necessary to ensure at-least-once with idempotency.
- **Contracts (examples):**
  - `orders.created`: `{ order_id, user_id, line_items[{sku, qty}], amount, currency, created_at, trace_id }`
  - `payments.completed`: `{ order_id, payment_id, state, reason?, amount, currency, processed_at, trace_id }`
  - `inventory.updated`: `{ sku, order_id, delta, reservation_id, state, processed_at, trace_id }`
  - `notifications.events`: `{ notification_id, user_id, channel, template, state, order_id?, sent_at?, trace_id }`
  - `orders.audit`: `{ order_id, event, actor, payload_ref, occurred_at, trace_id }`
- **Delivery semantics:**
  - Producers: idempotent enabled; required acks=all.
  - Consumers: at-least-once with idempotent handlers; retry with backoff; poison-message DLQ per topic.
- **Observability:** Emit metrics (lag, throughput, error rate), structured logs with trace_id, and traces via OpenTelemetry. Add audit topic for compliance.
- **Failure handling:** DLQs per domain topic; circuit-breaking for downstream dependencies (e.g., payment gateway) within the service, not across services.

### Phased Plan
1) **Foundation**
   - Provision 5-node Kafka with KRaft; set replication factor=3, min ISR=2, appropriate retention and cleanup policies.
   - Create topics and ACLs; enforce schema registry and compatibility policies.
   - Set up CI/CD scaffolding, linting, formatting, and container base images.
2) **Contracts and Schemas**
   - Define Avro/Protobuf/JSON schemas for all topics; include trace_id and versioning.
   - Establish producer/consumer libraries with idempotent and transactional support where applicable.
   - Add contract tests against a local Kafka test container.
3) **Order Service**
   - Implement order intake API (client-facing), validation, idempotency keys, persistence.
   - Publish `orders.created` and `orders.audit`; ensure deduplication and trace propagation.
   - Add consumer-side tests for replays and ordering guarantees.
4) **Payment Service**
   - Consume `orders.created`; integrate with payment gateway adapter.
   - Emit `payments.events`; handle retries, timeouts, and DLQ for poison payments.
   - Add sagas/state machine tests for payment transitions.
5) **Inventory Service**
   - Consume `orders.created` and `payments.events`; manage reservations and releases.
   - Emit `inventory.events`; ensure per-sku ordering and idempotent stock updates.
   - Load test reservation throughput and contention scenarios.
6) **Notification Service**
   - Consume `payments.events` and `inventory.events`; send email/SMS via provider adapters.
   - Emit `notifications.events`; add templating and rate limiting per user_id.
   - Add end-to-end tests with mocked providers.
7) **Analytics Service**
   - Consume across topics; build materialized views or push to warehouse.
   - Publish derived aggregates if needed; validate ordering on joins per key.
   - Add freshness and completeness checks.
8) **Hardening and Operations**
   - Add DLQs, replay tooling, and backfill playbooks.
   - Implement autoscaling rules for consumers based on lag.
   - Add chaos testing for broker and service failures; verify quorum and ISR behavior.
9) **Launch Readiness**
   - Performance test under peak load; validate partition utilization and hotspotting.
   - Run disaster recovery drills; document runbooks and on-call procedures.
   - Finalize dashboards/alerts for lag, errors, and key business SLIs.

### Next Steps
- Stand up Kafka locally (Docker) and create topics with desired partitions and RF.
- Begin with Foundation and Contracts phases before service implementation.

