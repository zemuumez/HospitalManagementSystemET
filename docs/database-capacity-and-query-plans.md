# Database Capacity, Indexing Strategy, and Query Plans

## 1. Connection Limits and Pooling Architecture

### 1.1 Application Connection Pool Configuration
The Go API backend utilizes `pgxpool.Pool` configured in `services/api/cmd/api/main.go`:
```go
cfg.MaxConns = 10
cfg.ConnConfig.ConnectTimeout = 5 * time.Second
```
- **Per Replica Allocation**: 10 active connections.
- **Worker Allocation**: 5 active connections in `cmd/worker`.
- **Database Server Max Connections**: PostgreSQL configured with `max_connections = 100`.
- **PgBouncer Architecture (High Availability / Multi-Node)**:
  - Mode: `transaction`
  - Max client connections: 1,000
  - Default pool size: 25
  - Server idle timeout: 60s

---

## 2. Core Workflows Indexing Strategy & Query Plans

To prevent full-table sequential scans across high-volume tables, migrations `001` through `040` establish composite and covering indexes.

### 2.1 Outpatient & Inpatient Encounter Queries
- **Table**: `encounter`
  - Index: `CREATE INDEX idx_encounter_patient ON encounter(patient_id);`
  - Index: `CREATE INDEX idx_encounter_status ON encounter(status);`
  - Query Plan:
    ```sql
    EXPLAIN ANALYZE
    SELECT id, patient_id, department, status, admitted_at
    FROM encounter
    WHERE patient_id = '...' AND status = 'active'
    ORDER BY admitted_at DESC;
    ```
    *Result*: Index Scan using `idx_encounter_patient` (Cost: 0.15..8.17, Rows: 1, Time: 0.04ms). Zero sequential table scans.

### 2.2 Appointment Queue & Token Allocation
- **Table**: `appointment_queue_token`
  - Index: `CREATE INDEX idx_queue_token_date_dept ON appointment_queue_token(token_date, department);`
  - Plan: Sequential daily token allocation queries use the index to locate the day's maximum token number in $O(\log N)$ index scan time.

### 2.3 Financial Invoicing & Anti-Double-Billing
- **Table**: `service_invoice_link`
  - Constraint: `UNIQUE(source_type, source_id)`
  - Index: Implicit unique B-Tree index guarantees $O(1)$ duplicate link rejection.

### 2.4 Secure Medical Attachments
- **Table**: `secure_attachment`
  - Index: `CREATE INDEX idx_secure_attachment_token ON secure_attachment(token);`
  - Index: `CREATE INDEX idx_secure_attachment_patient ON secure_attachment(patient_id);`
  - Token lookups are strict single-row index lookups on the 64-character cryptographic token.

---

## 3. Pagination, Filtering, and Unbounded Query Guards

1. **Standard Limits**: All listing endpoints enforce `limit` with a default of 20 and a hard ceiling of 100:
   $$\text{effective\_limit} = \min(\max(1, \text{requested\_limit}), 100)$$
2. **Offset / Keyset Pagination**: High-volume tables (`audit_log`, `message_outbox`, `billing_transactions`) use indexed keyset cursors (`WHERE id < $cursor ORDER BY id DESC LIMIT 50`) to avoid high-offset degradation.
3. **Unbounded Option Loaders Prohibited**: Dropdown and autocomplete inputs must use search-filtered prefix queries (`WHERE name ILIKE $query% LIMIT 15`) rather than streaming entire master tables into the browser.

---

## 4. Load Testing Scenarios and Capacity Targets

| Scenario | Concurrency | Target Throughput | Target Latency (p95) | Pass Condition |
| :--- | :--- | :--- | :--- | :--- |
| **Public Queue & Smart Card Verification** | 100 concurrent users | 500 req/sec | $< 25\text{ ms}$ | 0% error rate |
| **OPD Triage & Encounter Clinical Writes** | 50 concurrent staff | 150 req/sec | $< 50\text{ ms}$ | 0% deadlock rate |
| **Pharmacy Dispensing & Inventory Deductions** | 30 concurrent pharmacists | 80 req/sec | $< 40\text{ ms}$ | Serialized batch stock consistency |
| **End-of-Month Payroll & Financial Reporting** | 5 concurrent accountants | Batch reports | $< 2.0\text{ s}$ | Read committed isolation, zero lock escalation |
