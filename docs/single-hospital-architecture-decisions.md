# Architectural Decisions: Single-Hospital vs. Multi-Tenant SaaS Scope

## 1. Context and Objective

The legacy specification (`ULSHMS Saas.pdf`) references "SaaS" in its document title, but its underlying database schema and controller implementations lack tenant isolation keys (e.g. `hospital_id` or `tenant_id` columns are absent across 95% of tables, present only as isolated experiments in medicine tables). 

Following explicit user instructions, the HMS ET system is intentionally architected as a **dedicated, high-performance, single-hospital platform**, optimized for Ethiopian healthcare institutions rather than a generic multi-tenant cloud subscription product.

---

## 2. Intentional Architectural Decisions & Divergences

### 2.1 Elimination of Multi-Tenant SaaS Complexities
- **No SaaS Subscription Engine**: Excludes multi-tenant billing, plan tiers, subscription gates, credit card billing for software licenses, and tenant switching menus.
- **Dedicated Data Isolation**: The database schema directly represents the hospital's operations without requiring row-level tenant filtering (`WHERE hospital_id = ...`) on every transaction, removing risk of cross-tenant data leakage.
- **Resource Optimization**: Connection pooling (`pgxpool.Pool`), background workers, and memory allocation are dedicated 100% to the hospital's clinical and administrative throughput.

### 2.2 Localization for Ethiopian Healthcare Operations
- **Currency Standard**: All monetary values are strictly denominated in Ethiopian Birr (ETB) and stored as integer cents (1 ETB = 100 cents), eliminating floating-point rounding inaccuracies.
- **Bilingual English and Amharic Support**: User interface and report templates support bilingual rendering, with Amharic translations for clinical departments, diagnoses, vitals, and front-desk workflows.
- **East Africa Time (EAT / UTC+3)**: Patient appointments, attendance logs, nursing vitals, and surgical operations are recorded in UTC `timestamptz` and projected to EAT calendar-day boundaries.
- **Local Identity & Payments**: 
  - National ID / Kebele card / MRN tracking.
  - Payment rails prioritize Cash, CBE Birr, and Telebirr alongside optional Stripe integration.

### 2.3 Single-Hospital Administrative Hierarchy
- The `admin` role represents the Hospital Chief Executive Officer / Systems Administrator with full authority over hospital settings, staff provisioning, department revisions, and financial audits.
- Departmental workflows (OPD, IPD, Emergency, Laboratory, Pharmacy, Blood Bank, Billing) operate as coordinated units within the single institutional facility.
