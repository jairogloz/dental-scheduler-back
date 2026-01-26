# Money System Implementation Plan (Dental SaaS)

## Terminology

- **Corte de caja** (English): _cash reconciliation_ / _end-of-day cash reconciliation_

---

## Step 1 — Define accounting scope and invariants

**Goal:** Avoid future rewrites by fixing core assumptions early.

- **Multi-currency support**: MXN and USD per transaction
- **Exchange rates**: Configurable per clinic, overridable per transaction
- **Source of truth**: appointment-level accounting (ledger-based)
- **Payment methods**: Cash, card, transfer (hardcoded enums)
- **Cash reconciliation**: Based on _paid/charged entries_, not calendar events
- **Cash sessions**: Multiple sessions per day allowed (for mid-shift cash drops)
- **Cancelled / no-show appointments**: Do not affect reconciliation unless payment recorded
- **Immutable ledger**: Corrections via reversal entries, not edits/deletes
- **Commission tracking**: Internal (percentage) vs external (flat fee) doctors

**Deliverable:** `accounting_requirements.md` ✅

---

## Step 2 — Choose minimal but future-proof data model

**Goal:** Support multiple charges, doctors, discounts, commissions, and cash session management.

**Decision:**  
Do NOT link money transactions directly 1:1 to appointments.  
Instead, create a ledger per appointment + cash session management.

**Entities:**

- `appointment_accounts` (1:1 with appointment, created on first charge)
- `appointment_entries` (many per account, immutable ledger)
- `cash_sessions` (apertura de caja, multiple per day allowed)
- `reconciliations` (one per payment_method + currency per session)

---

## Step 3 — Database schema (Postgres)

### 3.1 appointment_accounts

**Purpose**: 1:1 with appointment, created only when first charge/payment recorded

**Fields:**

- id (uuid, pk)
- organization_id (uuid, fk)
- appointment_id (uuid, unique, fk)
- created_at (timestamp)
- updated_at (timestamp)

**Indexes:**

- (organization_id, created_at)
- (appointment_id) — unique

---

### 3.2 appointment_entries (ledger)

**Purpose**: Immutable ledger of all financial transactions per appointment

**Fields:**

- id (uuid, pk)
- appointment_account_id (uuid, fk)
- type (enum: service_charge, discount, payment, refund, correction)
- currency (enum: MXN, USD)
- amount_cents (bigint, signed) — positive for charges/payments, negative for discounts/refunds
- description (text, not null)
- created_by_user_id (uuid, fk, not null)
- created_at (timestamp, not null)

**Conditional fields:**

- payment_method (enum: cash, card, transfer) — required if type = payment
- exchange_rate_used (numeric(10,4)) — required if currency = USD
- doctor_id (uuid, fk) — required if type = service_charge
- corrects_entry_id (uuid, fk) — required if type = correction

**Doctor commission fields** (for service_charge):

- doctor_type (enum: internal, external) — required if type = service_charge
- commission_pct (numeric(5,2)) — required if doctor_type = internal
- external_doctor_fee_cents (bigint) — required if doctor_type = external
- is_sensitive (boolean, default false) — true if contains sensitive data

**Optional fields:**

- service_id (uuid, fk) — link to service catalog
- quantity (integer, default 1)
- unit_price_cents (bigint)
- notes (text)
- cash_session_id (uuid, fk) — links to cash session when created

**Constraints:**

- Foreign keys: appointment_account_id, doctor_id, created_by_user_id, corrects_entry_id, cash_session_id, service_id
- NOT NULL: id, appointment_account_id, type, currency, amount_cents, description, created_by_user_id, created_at
- amount_cents != 0 (business rule, not DB constraint)

**Indexes:**

- (appointment_account_id, created_at)
- (doctor_id, created_at)
- (cash_session_id, type, payment_method, currency) — for reconciliation queries
- (created_by_user_id, created_at)
- (corrects_entry_id) — for correction tracking

---

### 3.3 cash_sessions (apertura de caja)

**Purpose**: Track cash handling periods between reconciliations

**Fields:**

- id (uuid, pk)
- organization_id (uuid, fk, not null)
- clinic_id (uuid, fk, not null)
- user_id (uuid, fk, not null) — receptionist who opened session
- opened_at (timestamp, not null)
- closed_at (timestamp, nullable) — null = currently open
- starting_float_cents (bigint, not null) — amount in drawer at start
- status (enum: open, closed)
- opening_type (enum: manual, auto)
- notes (text, nullable)
- created_at (timestamp, not null)
- updated_at (timestamp, not null)

**Constraints:**

- Foreign keys: organization_id, clinic_id, user_id
- One open session per user per clinic (business rule)

**Indexes:**

- (user_id, clinic_id, status) — find user's open session
- (clinic_id, status, opened_at) — active sessions per clinic
- (clinic_id, closed_at) — historical lookup

---

### 3.4 reconciliations (corte de caja)

**Purpose**: Track cash reconciliations when closing sessions

**Fields:**

- id (uuid, pk)
- cash_session_id (uuid, fk, not null)
- organization_id (uuid, fk, not null)
- clinic_id (uuid, fk, not null)
- payment_method (enum: cash, card, transfer)
- currency (enum: MXN, USD)
- reconciled_at (timestamp, not null)
- reconciled_by_user_id (uuid, fk, not null)
- expected_amount_cents (bigint, not null) — calculated from entries
- actual_amount_cents (bigint, not null) — what receptionist counted
- float_left_cents (bigint, not null) — amount left in drawer
- deposited_cents (bigint, not null) — amount moved to safe (actual - float)
- discrepancy_cents (bigint, not null) — actual - expected
- envelope_id (text, nullable) — physical envelope identifier
- status (enum: pending, closed, disputed)
- notes (text, nullable)
- created_at (timestamp, not null)
- updated_at (timestamp, not null)

**Constraints:**

- Foreign keys: cash_session_id, organization_id, clinic_id, reconciled_by_user_id
- deposited_cents = actual_amount_cents - float_left_cents (computed)
- discrepancy_cents = actual_amount_cents - expected_amount_cents (computed)

**Indexes:**

- (cash_session_id, payment_method, currency)
- (clinic_id, payment_method, currency, reconciled_at) — reports
- (reconciled_by_user_id, reconciled_at)
- (status, clinic_id)
- (envelope_id) — track physical envelopes

---

**Note**: Business validations (e.g., "if type=payment then payment_method required") are enforced in Go domain layer, NOT as database constraints.

---

## Step 4 — Backfill and compatibility

**Decision:** Create appointment account **on-demand** when first charge/payment recorded ("Cobrar" workflow).

- Migration to add cash_sessions and reconciliations tables
- Backfill existing appointment data (optional, for historical appointments with payments)
- appointment_entries.cash_session_id nullable for backward compatibility

---

## Step 5 — Backend domain layer (Go)

Create `internal/accounting` package:

**Entities:**

- AppointmentAccount
- AppointmentEntry (with validation for conditional fields)
- CashSession
- Reconciliation

**Business rules:**

- Entry immutability (correction via reversal entries)
- Cash session management (one open per user per clinic)
- Commission calculation logic (internal vs external doctors)
- Multi-currency handling with exchange rates
- Signed amounts (positive for charges/payments, negative for discounts/refunds)

**Value objects:**

- Money (amount_cents + currency)
- ExchangeRate
- CommissionInfo (doctor_type + commission_pct OR external_fee)

**Services:**

- AppointmentAccountingService (create charges, payments, corrections)
- CashSessionService (open/close sessions)
- ReconciliationService (calculate expected amounts, create reconciliations)

---

## Step 6 — Backend API endpoints (MVP)

### Appointment Accounting

- GET /appointments/{id}/account — get account with all entries
- POST /appointments/{id}/account/entries — create charge/payment/discount
- POST /appointments/{id}/account/entries/{entry_id}/correct — create correction entry
- GET /appointments/{id}/account/balance — get total charged vs paid

### Cash Session Management

- POST /cash-sessions/open — open new cash session
- GET /cash-sessions/current — get user's current open session
- GET /cash-sessions/{id} — get session details with entries
- POST /cash-sessions/{id}/close — close session (triggers reconciliation)

### Reconciliation

- GET /cash-sessions/{id}/reconciliation-preview — calculate expected amounts
- POST /cash-sessions/{id}/reconcile — create reconciliation records
- GET /reconciliations — list reconciliations (with filters)
- GET /reconciliations/{id} — get reconciliation details

### Reports

- GET /reports/cash-reconciliation?date=&clinic_id= — daily reconciliation report
- GET /reports/doctor-collections?start_date=&end_date=&doctor_id= — doctor earnings
- GET /reports/discrepancies?clinic_id= — reconciliation discrepancies

**Validations:**

- Prevent entry creation without open cash session
- Validate currency/exchange rate combinations
- Enforce commission fields based on doctor_type
- Conflict-safe entry creation (idempotency)

---

## Step 7 — Server-side aggregation & reports

**Per appointment:**

- Total charges (service_charge entries)
- Total discounts (discount entries)
- Total payments (payment entries by method and currency)
- Balance due (charges - discounts - payments)
- Commission breakdown (internal vs external doctors)

**Per cash session:**

- Expected amounts by payment_method and currency
- Entry count and totals
- Session duration
- Entries created during session

**Daily reconciliation:**

- By clinic, payment method, currency
- Expected vs actual amounts
- Discrepancies
- Deposits to safe (envelope tracking)
- Multiple sessions per user per day

**Doctor collections:**

- Total charged per doctor
- Commission calculations (future)
- By date range and clinic
- Internal vs external doctor breakdown

---

## Step 8 — Frontend data layer

**Recommendation:** React Query (or equivalent server-state library)

**Hooks:**

- useAppointmentAccount(appointmentId) — get account and entries
- useAppointmentBalance(appointmentId) — get balance summary
- useCurrentCashSession() — get user's open session
- useCashSessionDetails(sessionId) — get session with entries
- useReconciliationPreview(sessionId) — calculate expected amounts
- useCashReconciliation(filters) — list reconciliations
- useDiscrepancyReport(clinicId) — get discrepancies

**Mutations:**

- useCreateEntry() — create charge/payment/discount
- useCreateCorrection() — correct existing entry
- useOpenCashSession() — open new session
- useCloseCashSession() — close and reconcile session

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

**Receptionist:**

- Open/close cash sessions
- Create entries (charges, payments, discounts)
- Create corrections
- Override exchange rates
- View own reconciliations

**Doctor:**

- Read-only access to their service_charge entries
- Cannot see sensitive fields (external_doctor_fee_cents)
- Cannot see other doctors' commissions

**Admin:**

- Full access including sensitive fields
- View all reconciliations and discrepancies
- Review cross-session corrections
- Access historical reconciliation adjustments (informational)

**Business Rules:**

- Only receptionists/admins can open cash sessions
- Cannot create entries without open cash session
- Cannot close session without reconciliation
- Entries never deleted or edited (corrections only)

---

## Step 12 — Commission tracking & future computation

**Data captured** (already in schema):

- doctor_id (who performed service)
- doctor_type (internal or external)
- commission_pct (for internal doctors)
- external_doctor_fee_cents (for external doctors)
- is_sensitive (flag for hiding external fees from patients)

**Internal doctor example:**

```
Service charge: 3000 MXN
Doctor type: internal
Commission: 40%
→ Doctor receives: 1200 MXN (computed later)
→ Clinic keeps: 1800 MXN
```

**External doctor example:**

```
Service charge: 3000 MXN (what patient pays)
Doctor type: external
External fee: 2000 MXN (what doctor receives)
→ Doctor receives: 2000 MXN flat
→ Clinic keeps: 1000 MXN
→ Patient NEVER sees the 2000 breakdown (is_sensitive=true)
```

**Future: Commission computation reports**

- Calculate doctor earnings by period
- Support different commission structures
- Track payments to doctors (separate from patient payments)
- Commission adjustments and corrections

**Note:** Commission _payment_ tracking is deferred. For now, just capture commission data.

---

## Final note

Decide early whether **charges** or **payments received** represent cash reconciliation.
Receptionist workflows often mix both — be explicit in your system.
