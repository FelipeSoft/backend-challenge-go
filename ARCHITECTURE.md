## Architecture & Design Decisions

This document bridges the gap between the requirements of the Distributed Wager Processing Challenge and the current state of implementation, detailing technical trade-offs, design patterns, and known limitations.

---

### 1. Overview & Technology Stack

The system is engineered in **Go**, orchestrated via **Uber Fx** (`go.uber.org/fx`) for modular dependency injection and lifecycle management, backed by **PostgreSQL** (`pgx/v5`) and **AWS SQS FIFO** (emulated locally via **LocalStack**). Security is enforced through external OAuth 2.0 / OIDC authentication (**Keycloak**).

---

### 2. Implementation Status: Delivered vs. Pending

| Component | Status | Key Details |
| --- | --- | --- |
| **Core Architecture** | **Delivered** | Uber Fx modules, strict dependency mapping, and structured graceful shutdown hooks (`fx.Lifecycle`). |
| **Persistence & Ledger** | **Delivered** | Raw SQL via `pgx/v5`, append-only immutable `wallet_ledger_entries`, and strict decimal-safe balance tracking using `int64` minor units. |
| **Concurrency Control** | **Delivered** | Hybrid strategy combining Pessimistic Locking (`SELECT ... FOR UPDATE`) on wallets and Optimistic Locking (`version` checks) to completely eliminate lost updates. |
| **Idempotency & Inbox** | **Delivered** | Persistent request tracking via the `inbox` pattern and unique constraints on `(consumer_name, message_id)` and idempotency keys. |
| **Transactional Outbox** | **Delivered** | Guarantees atomic writes of domain state and integration events within the same database transaction, processed asynchronously by a background publisher worker. |
| **Dual Ingress (HTTP + SQS)** | **Delivered** | Unified business use-cases supporting both REST endpoints and SQS message consumers with shared idempotency guarantees. |
| **Reference Resolution (`PENDING_REFERENCE`)** | **Pending** | Basic state handling exists, but automated background worker resolution for orphaned rollbacks/refunds requires full completion. |
| **Observability & Tracing** | **Pending** | Structured JSON logging is in place, but Prometheus metrics, OpenTelemetry distributed tracing spans, and `causationId` tracking are slated for post-MVP. |
| **Automated Test Suite** | **Partial** | Unit and basic integration flows are established; comprehensive multi-instance split-brain simulation tests and reproducible load-testing scripts remain in progress. |

---

### 3. Architecture Defense & Design Trade-offs

#### A. HTTP API & Error Mapping

* **Design Choice:** Write-side commands return custom-tailored response payloads aggregated from isolated sub-domains, whereas read-side queries pull straight from optimized projections.
* **Trade-off:** Generic HTTP error status codes were initially prioritized to accelerate feature delivery. Production hardening requires a formalized **Error Mapping layer** to translate domain rejections (e.g., insufficient funds, missing references) into precise HTTP semantics for effective error telemetry and client consumption.

#### B. Concurrency Management

* **Design Choice:** To safeguard player wallets from race conditions without resorting to risky global table locks, a hybrid concurrency model is implemented:
1. **Pessimistic Locking:** `FOR UPDATE` serializes concurrent transactions targeting the exact same wallet row.
2. **Optimistic Locking:** Row versioning (`version = version + 1`) acts as a secondary circuit breaker to instantly reject concurrent conflicts via `RowsAffected() == 0`.


* **Future Scaling Consideration:** Under ultra-high throughput environments where concurrent database I/O on the `wallet_ledger_entries` and `wallets` tables creates a bottleneck, a **Cache-Aside balance pattern** with scheduled write-behind batching and periodic reconciliation windows could be introduced.

---

### 4. Roadmap & Outstanding Items

1. **Event Tracing:** Inject full propagation of `correlationId` and `causationId` headers across the SQS envelope boundary.
2. **Asynchronous Reference Worker:** Complete the background worker loop that handles `PENDING_REFERENCE` states with exponential backoff and terminal dead-letter handling.
3. **Load & Stress Testing:** Implement a reproducible benchmarking script (e.g., using k6) to document throughput, error rates, and p95/p99 latency metrics.