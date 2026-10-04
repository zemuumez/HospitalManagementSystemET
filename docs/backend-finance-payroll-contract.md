# Backend Billing, Finance, and Payroll Contract

## 1. Overview

This module delivers backend capabilities for hospital operational accounting, non-patient revenue & operational expenses, employee payroll lifecycle, anti-double-billing linkage for clinical services, and financial executive reporting:
- **Hospital Expense Heads & Expenses**: Structured expense classification heads (e.g. Facility Rent, Equipment & Maintenance, Electricity & Utilities, Fuel & Transport, Cleaning & Hygiene, Hospital Supplies) with immutable expense records in integer ETB minor units, invoice tracking, and recorded-by user associations.
- **Hospital Income Heads & Incomes**: Ancillary revenue heads (e.g. Special Consultations, Canteen & Facility Lease, Transport & Ambulance Receipts, Donations & Grants, Equipment Rental, Miscellaneous Income) with receipt tracking.
- **Employee Payroll Lifecycle**: Staff payroll generation with basic salary, allowances, deductions, and server-calculated net salary (`basic + allowance - deductions >= 0`), status transitions (`0`=Unpaid, `1`=Paid), payment timestamping, duplicate payroll period conflict guards (`UNIQUE(user_id, month, year)`), and self-service payroll slip access for staff.
- **Anti-Double-Billing Service Invoice Links**: Strict billing linkage (`service_invoice_link`) enforcing `UNIQUE(source_type, source_id)` across ambulances, operations, clinical services, and blood issues to guarantee no charge can be billed to an invoice more than once.
- **Financial Summary Reports**: Period-based aggregation (`GET /v1/finance-reports/summary?from=...&to=...`) computing total income, total operating expenses, total paid employee payroll, net balance in minor units, and categorized head breakdowns.

---

## 2. Authorization & Role Matrix

| Resource | Read Endpoints | Manage Endpoints | Permitted Roles |
|---|---|---|---|
| **Expense Heads** | `GET /v1/expense-heads` | `POST /v1/expense-heads` | Read: `admin`, `accountant` (`finance.read`). Manage: `admin`, `accountant` (`finance.manage`). |
| **Expenses** | `GET /v1/expenses`, `GET /v1/expenses/{id}` | `POST /v1/expenses` | Read: `admin`, `accountant` (`finance.read`). Manage: `admin`, `accountant` (`finance.manage`). |
| **Income Heads** | `GET /v1/income-heads` | `POST /v1/income-heads` | Read: `admin`, `accountant` (`finance.read`). Manage: `admin`, `accountant` (`finance.manage`). |
| **Incomes** | `GET /v1/incomes`, `GET /v1/incomes/{id}` | `POST /v1/incomes` | Read: `admin`, `accountant` (`finance.read`). Manage: `admin`, `accountant` (`finance.manage`). |
| **Employee Payroll** | `GET /v1/payrolls`, `GET /v1/payrolls/{id}` | `POST /v1/payrolls`, `POST /v1/payrolls/{id}/pay` | Read: `admin`, `accountant` (`payroll.read`); staff members can only read their own payroll slips (`payroll.read_own`). Manage: `admin`, `accountant` (`payroll.manage`). |
| **Service Invoice Links** | `GET /v1/service-invoice-links` | `POST /v1/service-invoice-links` | Read: `admin`, `accountant` (`billing.read`). Manage: `admin`, `accountant` (`billing.manage`). |
| **Finance Summary** | `GET /v1/finance-reports/summary` | *N/A (Read-only aggregation)* | Read: `admin`, `accountant` (`finance.read`). |

---

## 3. Endpoints

### 3.1 Expense Heads & Expenses

#### `GET /v1/expense-heads`
- **Response `200 OK`**:
  ```json
  {
    "heads": [
      {
        "id": "e1111111-1111-1111-1111-111111111111",
        "name": "Building Rent",
        "description": "Premises lease and facility rental costs",
        "createdAt": "2026-10-05T00:00:00Z",
        "updatedAt": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `POST /v1/expense-heads`
- **Request Body**:
  ```json
  {
    "name": "Laboratory Consumables",
    "description": "Reagents, test kits, and diagnostic consumables"
  }
  ```
- **Response `201 Created`**: Returns created `ExpenseHead`.

#### `GET /v1/expenses`
- **Query Params**: `headId` (optional), `from` (optional date), `to` (optional date), `page` (default 1)
- **Response `200 OK`**:
  ```json
  {
    "expenses": [
      {
        "id": "e2222222-2222-2222-2222-222222222222",
        "expenseHeadId": "e1111111-1111-1111-1111-111111111111",
        "expenseHeadName": "Building Rent",
        "name": "October 2026 Facility Lease",
        "invoiceNumber": "LEASE-OCT-2026",
        "date": "2026-10-01T00:00:00Z",
        "amountMinor": 50000000,
        "description": "Hospital main building monthly rental",
        "recordedBy": "user-accountant-01",
        "recordedByName": "Accountant User",
        "createdAt": "2026-10-05T00:00:00Z",
        "updatedAt": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/expenses`
- **Request Body**:
  ```json
  {
    "expenseHeadId": "e1111111-1111-1111-1111-111111111111",
    "name": "Backup Generator Diesel",
    "invoiceNumber": "FUEL-2026-89",
    "date": "2026-10-03",
    "amountMinor": 350000,
    "description": "500 Liters fuel for main generator"
  }
  ```
- **Response `201 Created`**: Returns created `Expense`.

---

### 3.2 Income Heads & Incomes

#### `GET /v1/income-heads`
- **Response `200 OK`**:
  ```json
  {
    "heads": [
      {
        "id": "i1111111-1111-1111-1111-111111111111",
        "name": "Special Consultations",
        "description": "Executive clinical and specialist consultation receipts",
        "createdAt": "2026-10-05T00:00:00Z",
        "updatedAt": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `POST /v1/income-heads`
- **Request Body**:
  ```json
  {
    "name": "Pharmacy Wholesale Concessions",
    "description": "Rebates from pharmaceutical distributors"
  }
  ```
- **Response `201 Created`**: Returns created `IncomeHead`.

#### `GET /v1/incomes`
- **Query Params**: `headId` (optional), `from` (optional), `to` (optional), `page` (default 1)
- **Response `200 OK`**:
  ```json
  {
    "incomes": [
      {
        "id": "i2222222-2222-2222-2222-222222222222",
        "incomeHeadId": "i1111111-1111-1111-1111-111111111111",
        "incomeHeadName": "Special Consultations",
        "name": "Executive Specialist Package",
        "invoiceNumber": "REC-2026-101",
        "date": "2026-10-02T00:00:00Z",
        "amountMinor": 950000,
        "description": "Special cardiology consultation package",
        "recordedBy": "user-accountant-01",
        "recordedByName": "Accountant User",
        "createdAt": "2026-10-05T00:00:00Z",
        "updatedAt": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/incomes`
- **Request Body**:
  ```json
  {
    "incomeHeadId": "i1111111-1111-1111-1111-111111111111",
    "name": "Canteen Concession Monthly Fee",
    "invoiceNumber": "CANTEEN-OCT-26",
    "date": "2026-10-04",
    "amountMinor": 2500000,
    "description": "Hospital cafeteria operator lease fee"
  }
  ```
- **Response `201 Created`**: Returns created `Income`.

---

### 3.3 Employee Payroll

#### `GET /v1/payrolls`
- **Query Params**: `userId` (optional), `month` (optional), `year` (optional), `status` (optional `0` or `1`), `page` (default 1)
- **Authorization Scoping**:
  - Staff without `payroll.read` can only query their own payroll (`userId` defaults to their authenticated ID; specifying another user returns `403 Forbidden`).
  - Admin/Accountant can query across all staff.
- **Response `200 OK`**:
  ```json
  {
    "payrolls": [
      {
        "id": "p1111111-1111-1111-1111-111111111111",
        "payrollNumber": "PAY-849201-4921",
        "userId": "doc-01",
        "userName": "Dr. Abebe Bekele",
        "userEmail": "abebe@hospital.et",
        "role": "doctor",
        "month": "October",
        "year": 2026,
        "basicSalaryMinor": 3000000,
        "allowanceMinor": 500000,
        "deductionsMinor": 200000,
        "netSalaryMinor": 3300000,
        "status": 0,
        "paymentDate": null,
        "createdBy": "user-accountant-01",
        "createdByName": "Accountant User",
        "createdAt": "2026-10-05T00:00:00Z",
        "updatedAt": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/payrolls`
- **Request Body**:
  ```json
  {
    "userId": "doc-01",
    "month": "October",
    "year": 2026,
    "basicSalaryMinor": 3000000,
    "allowanceMinor": 500000,
    "deductionsMinor": 200000
  }
  ```
- **Validation**:
  - `netSalaryMinor = basicSalaryMinor + allowanceMinor - deductionsMinor >= 0`.
  - Duplicate `(userId, month, year)` returns `409 Conflict`.
- **Response `201 Created`**: Returns created `EmployeePayroll` with initial status `0` (Unpaid).

#### `POST /v1/payrolls/{id}/pay`
- **Action**: Marks the payroll record as paid (`status = 1`) and records the current timestamp into `paymentDate`.
- **Validation**: Cannot pay an already paid payroll (returns `409 Conflict`).
- **Response `200 OK`**: Returns updated `EmployeePayroll`.

---

### 3.4 Service Invoice Links (Anti-Double-Billing)

#### `POST /v1/service-invoice-links`
- **Request Body**:
  ```json
  {
    "invoiceId": "inv-00000000-0000-0000-0000-000000000001",
    "sourceType": "ambulance",
    "sourceId": "amb-call-2026-10-01",
    "amountMinor": 50000
  }
  ```
- **Supported `sourceType` values**: `'service'`, `'operation'`, `'ambulance'`, `'blood_issue'`.
- **Anti-Double-Billing Guard**:
  - Enforced by database constraint `UNIQUE(source_type, source_id)`.
  - Attempting to link the same source record to another invoice (or multiple times) returns `409 Conflict`.
- **Response `201 Created`**: Returns created `ServiceInvoiceLink`.

#### `GET /v1/service-invoice-links?invoiceId={invoiceId}`
- **Response `200 OK`**:
  ```json
  {
    "links": [
      {
        "id": "l1111111-1111-1111-1111-111111111111",
        "invoiceId": "inv-00000000-0000-0000-0000-000000000001",
        "sourceType": "ambulance",
        "sourceId": "amb-call-2026-10-01",
        "amountMinor": 50000,
        "createdAt": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

---

### 3.5 Finance Summary Report

#### `GET /v1/finance-reports/summary`
- **Query Params**: `from` (ISO date string, e.g. `2026-10-01`), `to` (ISO date string, e.g. `2026-10-31`)
- **Response `200 OK`**:
  ```json
  {
    "fromDate": "2026-10-01T00:00:00Z",
    "toDate": "2026-10-31T00:00:00Z",
    "totalIncomeMinor": 950000,
    "totalExpenseMinor": 450000,
    "totalPayrollMinor": 3300000,
    "netBalanceMinor": -2800000,
    "expensesByHead": [
      {
        "headId": "e1111111-1111-1111-1111-111111111111",
        "headName": "Building Rent",
        "totalMinor": 450000
      }
    ],
    "incomesByHead": [
      {
        "headId": "i1111111-1111-1111-1111-111111111111",
        "headName": "Special Consultations",
        "totalMinor": 950000
      }
    ]
  }
  ```

---

## 4. Database Schema (Migration 035)

- `hospital_expense_head`: Master catalog of hospital expense categories.
- `hospital_expense`: Ledger of operational expenditures.
- `hospital_income_head`: Master catalog of hospital revenue categories.
- `hospital_income`: Ledger of ancillary revenues.
- `employee_payroll`: Staff payroll entries with status tracking and `UNIQUE(user_id, month, year)`.
- `service_invoice_link`: Bridge table with `UNIQUE(source_type, source_id)` preventing duplicate charges on invoices.
