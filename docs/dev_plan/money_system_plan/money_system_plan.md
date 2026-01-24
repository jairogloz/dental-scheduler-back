# Money System Implementation Plan (Dental SaaS)

## Terminology
- **Corte de caja** (English): *cash reconciliation* / *end-of-day cash reconciliation*

---

## Step 1 — Define accounting scope and invariants
**Goal:** Avoid future rewrites by fixing core assumptions early.

- Single currency per organization (MXN)
- Source of truth: appointment-level accounting
- Cash reconciliation based on *paid/charged entries*, not calendar events
- Cancelled / no-show appointments should not affect cash unless explicitly overridden

**Deliverable:** `accounting_invariants.md`

---

## Step 2 — Choose minimal but future-proof data model
**Goal:** Support multiple charges, doctors, discounts, commissions.

**Decision:**  
Do NOT link money transactions directly 1:1 to appointments.  
Instead, create a ledger per appointment.

Entities:
- `appointment_accounts` (1:1 with appointment)
- `appointment_entries` (many per account)

---

## Step 3 — Database schema (Postgres)

### 3.1 appointment_accounts
- id (uuid, pk)
- organization_id
- appointment_id (unique, fk)
- currency (default MXN)
- status (open/closed, optional)
- created_at
- updated_at

### 3.2 appointment_entries (ledger)
- id (uuid, pk)
- appointment_account_id (fk)
- type (enum):
  - service_charge
  - discount
  - adjustment
  - payment (future)
  - refund (future)
- description
- doctor_id (required for service_charge)
- service_id (future)
- quantity (default 1)
- unit_price_cents
- amount_cents (signed)
- commission_pct (nullable)
- created_by_user_id
- created_at

**Constraints**
- amount_cents != 0
- if type = service_charge → doctor_id NOT NULL

**Indexes**
- (organization_id, created_at)
- (appointment_account_id, created_at)
- (doctor_id, created_at)

---

## Step 4 — Backfill and compatibility
**Decision:** Create appointment account automatically when appointment is created.

- Migration to backfill existing appointments
- Optional background script for historical data

---

## Step 5 — Backend domain layer (Go)
Create `internal/accounting` package:
- AppointmentAccount
- AppointmentEntry
- Business rules for adding/removing entries
- Amounts stored in cents

---

## Step 6 — Backend API endpoints (MLP)

### Endpoints
- GET /appointments/{id}/account
- POST /appointments/{id}/account/entries
- DELETE /account/entries/{entry_id}
- GET /reports/cash-reconciliation?date=&clinic_id=

**Validations**
- Prevent invalid transitions
- Conflict-safe entry creation

---

## Step 7 — Server-side aggregation & reports
- Totals per appointment:
  - charges
  - discounts
  - net total
- Daily totals:
  - by clinic
  - by doctor

---

## Step 8 — Frontend data layer
**Recommendation:** React Query (or equivalent server-state library)

Hooks:
- useAppointmentAccount(appointmentId)
- useCashReconciliation(date, clinicId)

---

## Step 9 — Appointment modal UI changes
Add tabs:
- Detalles
- Cobro
- Historial (future)

Cobro tab:
- Line items list
- Add charge / discount / adjustment
- Totals summary

---

## Step 10 — Integrations
- Calendar: show total charged
- Patient history: show appointment totals

---

## Step 11 — Permissions & guardrails
- Receptionist: add/edit entries
- Doctor: read-only
- Admin: override
- Optional locking after closeout

---

## Step 12 — Commission future-proofing
Store:
- doctor_id
- commission_pct per service_charge

No commission computation yet — only data capture.

---

## Final note
Decide early whether **charges** or **payments received** represent cash reconciliation.
Receptionist workflows often mix both — be explicit in your system.
