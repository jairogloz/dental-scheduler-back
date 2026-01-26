# Database Schema - Accounting System

## Overview

This document defines the database schema for the accounting and cash management system. The schema consists of 4 new tables that integrate with the existing dental scheduler system.

**Design Principles:**

- **Immutable Ledger**: Entries are never updated or deleted, only added
- **Signed Amounts**: Positive for income (charges, payments), negative for expenses (discounts, refunds)
- **Multi-Currency**: Support for MXN and USD with exchange rate tracking
- **Cash Sessions**: Multiple sessions per day to support mid-shift cash drops
- **Application-Layer Validation**: Valid enum values enforced in Go code, not database constraints

---

## Tables

### 1. appointment_accounts

**Purpose**: Links appointments to their financial ledger. Created on-demand when first charge/payment is recorded.

```sql
CREATE TABLE appointment_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    appointment_id UUID NOT NULL UNIQUE REFERENCES appointments(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_appointment_accounts_org_created ON appointment_accounts(organization_id, created_at);
CREATE UNIQUE INDEX idx_appointment_accounts_appointment ON appointment_accounts(appointment_id);

-- Comments
COMMENT ON TABLE appointment_accounts IS 'Financial account for appointments with actual financial activity';
COMMENT ON COLUMN appointment_accounts.appointment_id IS 'One-to-one relationship with appointment';
```

---

### 2. appointment_account_entries

**Purpose**: Immutable ledger of all financial transactions related to appointments.

```sql
CREATE TABLE appointment_account_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    appointment_account_id UUID NOT NULL REFERENCES appointment_accounts(id) ON DELETE CASCADE,

    -- Core transaction fields
    type VARCHAR(20) NOT NULL,
    -- Valid values: 'service_charge', 'discount', 'payment', 'refund', 'correction'
    -- Validated in application layer

    currency VARCHAR(3) NOT NULL,
    -- Valid values: 'MXN', 'USD'
    -- Validated in application layer

    amount_cents BIGINT NOT NULL,
    -- Signed amount: positive for charges/payments, negative for discounts/refunds
    -- Amount in smallest currency unit (cents/centavos)

    description TEXT NOT NULL,
    -- Human-readable description of the transaction

    created_by_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Conditional fields (required based on type)
    payment_method VARCHAR(20),
    -- Valid values: 'cash', 'card', 'transfer'
    -- Required if type = 'payment'
    -- Validated in application layer

    exchange_rate_used NUMERIC(10,4),
    -- Required if currency = 'USD'
    -- Stores the USD to MXN exchange rate used for this transaction

    doctor_id UUID REFERENCES doctors(id),
    -- Required if type = 'service_charge'

    corrects_entry_id UUID REFERENCES appointment_account_entries(id),
    -- Required if type = 'correction'
    -- Links to the original entry being corrected

    -- Doctor commission fields (for service_charge only)
    doctor_type VARCHAR(20),
    -- Valid values: 'internal', 'external'
    -- Required if type = 'service_charge'
    -- Validated in application layer

    commission_pct NUMERIC(5,2),
    -- Commission percentage for internal doctors (e.g., 40.00 = 40%)
    -- Required if doctor_type = 'internal'

    external_doctor_fee_cents BIGINT,
    -- Flat fee for external doctors in cents
    -- Required if doctor_type = 'external'

    is_sensitive BOOLEAN DEFAULT FALSE,
    -- True if entry contains sensitive data not visible to patients
    -- Used to hide external_doctor_fee_cents from patient views

    -- Optional fields
    service_id UUID REFERENCES services(id),
    -- Link to service catalog

    quantity INTEGER DEFAULT 1,
    -- Number of units for service_charge entries

    unit_price_cents BIGINT,
    -- Price per unit for service_charge entries

    notes TEXT,
    -- Additional context, especially for corrections

    cash_session_id UUID REFERENCES cash_sessions(id),
    -- Links entry to the cash session when it was created
    -- Nullable for backward compatibility

    CONSTRAINT amount_not_zero CHECK (amount_cents != 0)
);

-- Indexes
CREATE INDEX idx_entries_account_created ON appointment_account_entries(appointment_account_id, created_at);
CREATE INDEX idx_entries_doctor_created ON appointment_account_entries(doctor_id, created_at);
CREATE INDEX idx_entries_session_type_payment ON appointment_account_entries(cash_session_id, type, payment_method, currency);
CREATE INDEX idx_entries_creator_created ON appointment_account_entries(created_by_user_id, created_at);
CREATE INDEX idx_entries_corrects ON appointment_account_entries(corrects_entry_id);
CREATE INDEX idx_entries_service ON appointment_account_entries(service_id);

-- Comments
COMMENT ON TABLE appointment_account_entries IS 'Immutable ledger of all financial transactions for appointments';
COMMENT ON COLUMN appointment_account_entries.type IS 'Entry type: service_charge, discount, payment, refund, correction (validated in app)';
COMMENT ON COLUMN appointment_account_entries.currency IS 'Currency code: MXN, USD (validated in app)';
COMMENT ON COLUMN appointment_account_entries.amount_cents IS 'Signed amount in cents: positive for income, negative for expenses';
COMMENT ON COLUMN appointment_account_entries.payment_method IS 'Payment method: cash, card, transfer (required if type=payment, validated in app)';
COMMENT ON COLUMN appointment_account_entries.doctor_type IS 'Doctor type: internal, external (required if type=service_charge, validated in app)';
COMMENT ON COLUMN appointment_account_entries.commission_pct IS 'Commission % for internal doctors (required if doctor_type=internal)';
COMMENT ON COLUMN appointment_account_entries.external_doctor_fee_cents IS 'Flat fee for external doctors (required if doctor_type=external)';
COMMENT ON COLUMN appointment_account_entries.is_sensitive IS 'Hides external doctor fees from patient-facing views';
COMMENT ON COLUMN appointment_account_entries.corrects_entry_id IS 'References original entry if this is a correction';
COMMENT ON COLUMN appointment_account_entries.cash_session_id IS 'Cash session when entry was created (for reconciliation)';
```

---

### 3. cash_sessions

**Purpose**: Track cash handling periods (apertura de caja / sesión de caja). A receptionist can have multiple sessions per day for mid-shift cash drops.

```sql
CREATE TABLE cash_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    -- The receptionist/admin who opened the session

    opened_at TIMESTAMP NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP,
    -- NULL = session is currently open

    starting_float_cents BIGINT NOT NULL,
    -- Amount of cash left in drawer at session start (for making change)

    status VARCHAR(20) NOT NULL,
    -- Valid values: 'open', 'closed'
    -- Validated in application layer

    opening_type VARCHAR(20) NOT NULL,
    -- Valid values: 'manual', 'auto'
    -- Indicates how the session was opened
    -- Validated in application layer

    notes TEXT,
    -- Optional notes about the session

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_sessions_user_clinic_status ON cash_sessions(user_id, clinic_id, status);
CREATE INDEX idx_sessions_clinic_status_opened ON cash_sessions(clinic_id, status, opened_at);
CREATE INDEX idx_sessions_clinic_closed ON cash_sessions(clinic_id, closed_at);

-- Comments
COMMENT ON TABLE cash_sessions IS 'Cash handling periods between reconciliations (multiple per day allowed)';
COMMENT ON COLUMN cash_sessions.status IS 'Session status: open, closed (validated in app)';
COMMENT ON COLUMN cash_sessions.opening_type IS 'How opened: manual, auto (validated in app)';
COMMENT ON COLUMN cash_sessions.starting_float_cents IS 'Cash in drawer at start of session for making change';
COMMENT ON COLUMN cash_sessions.closed_at IS 'NULL indicates session is currently open';
```

---

### 4. reconciliations

**Purpose**: Record cash reconciliations when closing cash sessions. Each reconciliation represents one deposit to safe with physical envelope.

```sql
CREATE TABLE reconciliations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cash_session_id UUID NOT NULL REFERENCES cash_sessions(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,

    payment_method VARCHAR(20) NOT NULL,
    -- Valid values: 'cash', 'card', 'transfer'
    -- Validated in application layer

    currency VARCHAR(3) NOT NULL,
    -- Valid values: 'MXN', 'USD'
    -- Validated in application layer

    reconciled_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reconciled_by_user_id UUID NOT NULL REFERENCES users(id),

    -- Reconciliation amounts (all in cents)
    expected_amount_cents BIGINT NOT NULL,
    -- Calculated from appointment_account_entries WHERE cash_session_id = this.cash_session_id

    actual_amount_cents BIGINT NOT NULL,
    -- What receptionist actually counted in the drawer

    float_left_cents BIGINT NOT NULL,
    -- Amount left in drawer for next session (for making change)

    deposited_cents BIGINT NOT NULL,
    -- Amount moved to safe in envelope (= actual_amount_cents - float_left_cents)

    discrepancy_cents BIGINT NOT NULL,
    -- Difference between expected and actual (= actual_amount_cents - expected_amount_cents)

    envelope_id TEXT,
    -- Physical envelope identifier for tracking safe deposits
    -- Format suggestion: R-{id}-{date}-{user}

    status VARCHAR(20) NOT NULL,
    -- Valid values: 'pending', 'closed', 'disputed'
    -- Validated in application layer

    notes TEXT,
    -- Receptionist notes about discrepancies or special situations

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_recon_session_payment_currency ON reconciliations(cash_session_id, payment_method, currency);
CREATE INDEX idx_recon_clinic_payment_currency_date ON reconciliations(clinic_id, payment_method, currency, reconciled_at);
CREATE INDEX idx_recon_user_date ON reconciliations(reconciled_by_user_id, reconciled_at);
CREATE INDEX idx_recon_status_clinic ON reconciliations(status, clinic_id);
CREATE INDEX idx_recon_envelope ON reconciliations(envelope_id);

-- Comments
COMMENT ON TABLE reconciliations IS 'Cash reconciliations when closing sessions (one per payment_method + currency)';
COMMENT ON COLUMN reconciliations.payment_method IS 'Payment method: cash, card, transfer (validated in app)';
COMMENT ON COLUMN reconciliations.currency IS 'Currency code: MXN, USD (validated in app)';
COMMENT ON COLUMN reconciliations.expected_amount_cents IS 'Calculated from entries in this cash session';
COMMENT ON COLUMN reconciliations.actual_amount_cents IS 'What receptionist counted in drawer';
COMMENT ON COLUMN reconciliations.float_left_cents IS 'Amount left for next session (for change)';
COMMENT ON COLUMN reconciliations.deposited_cents IS 'Amount to safe = actual - float_left';
COMMENT ON COLUMN reconciliations.discrepancy_cents IS 'Difference = actual - expected';
COMMENT ON COLUMN reconciliations.envelope_id IS 'Physical envelope identifier for safe deposit';
COMMENT ON COLUMN reconciliations.status IS 'Reconciliation status: pending, closed, disputed (validated in app)';
```

---

## Enum Values Reference

These values are **validated in the application layer (Go)**, not enforced by database constraints.

### appointment_account_entries.type

- `service_charge` - A treatment/service provided to patient
- `discount` - Reduction in price
- `payment` - Money received from patient
- `refund` - Money returned to patient
- `correction` - Reverses a previous entry (links via corrects_entry_id)

### appointment_account_entries.currency

- `MXN` - Mexican Peso
- `USD` - US Dollar

### appointment_account_entries.payment_method

- `cash` - Physical currency
- `card` - Credit/debit card
- `transfer` - Bank transfer

### appointment_account_entries.doctor_type

- `internal` - Clinic employee, receives commission percentage
- `external` - External provider, receives flat fee

### cash_sessions.status

- `open` - Session is currently active
- `closed` - Session has been reconciled and closed

### cash_sessions.opening_type

- `manual` - User explicitly clicked "Abrir Caja"
- `auto` - System auto-opened when user created first entry

### reconciliations.status

- `pending` - Reconciliation in progress
- `closed` - Reconciliation completed
- `disputed` - Discrepancy under review

---

## Business Rules

These rules are enforced in the **application layer** (Go domain/usecase), not database constraints:

### appointment_account_entries

1. **Currency validation**: If currency = 'USD', then exchange_rate_used must be set
2. **Payment method validation**: If type = 'payment', then payment_method must be set
3. **Amount validation**: amount_cents != 0
4. **Service charge validation**: If type = 'service_charge', then:
   - doctor_id must be set
   - doctor_type must be set ('internal' or 'external')
5. **Commission validation**:
   - If doctor_type = 'internal', then commission_pct must be set
   - If doctor_type = 'external', then external_doctor_fee_cents must be set
6. **Correction validation**: If type = 'correction', then corrects_entry_id must reference a valid entry
7. **Signed amount validation**:
   - service_charge, payment → amount_cents must be positive
   - discount, refund → amount_cents must be negative
   - correction → amount_cents must be opposite sign of corrected entry
8. **Immutability**: Entries are never updated or deleted, only created (append-only ledger)
9. **Cash session requirement**: Entries should only be created when user has an open cash_session

### cash_sessions

1. **One open session per user per clinic**: A user can only have one open cash_session at a time per clinic
2. **Multiple sessions per day allowed**: User can have multiple closed sessions per day (for mid-shift cash drops)
3. **Permissions**: Only receptionists and admins can open cash sessions

### reconciliations

1. **Computed fields**:
   - deposited_cents = actual_amount_cents - float_left_cents
   - discrepancy_cents = actual_amount_cents - expected_amount_cents
2. **Session relationship**: Each cash_session can have multiple reconciliations (one per payment_method + currency combination)
3. **All reconciliations required**: All payment methods and currencies used in session must be reconciled before closing session

---

## Entity Relationships

```
organizations
    ├── appointment_accounts
    ├── cash_sessions
    └── reconciliations

clinics
    ├── cash_sessions
    └── reconciliations

appointments
    └── appointment_accounts (1:1, created on-demand)
        └── appointment_account_entries (1:many, immutable ledger)

cash_sessions
    ├── appointment_account_entries (many entries created during session)
    └── reconciliations (many, one per payment_method + currency)

users (receptionists/admins)
    ├── cash_sessions (many, user opens sessions)
    ├── reconciliations (many, user performs reconciliations)
    └── appointment_account_entries (many, user creates entries)

doctors
    └── appointment_account_entries (many, receives commissions)

services
    └── appointment_account_entries (many, service catalog link)
```

---

## Example Data Flow

### Scenario: Patient Appointment with Payment

```sql
-- 1. Receptionist opens cash session (manual)
INSERT INTO cash_sessions (
    organization_id, clinic_id, user_id,
    starting_float_cents, status, opening_type
) VALUES (
    'org-uuid', 'clinic-uuid', 'receptionist-uuid',
    50000, 'open', 'manual'  -- 500 MXN starting float
);

-- 2. Patient arrives, service is charged
-- First, create appointment account (if doesn't exist)
INSERT INTO appointment_accounts (
    organization_id, appointment_id
) VALUES (
    'org-uuid', 'appointment-uuid'
);

-- Then, create service charge entry
INSERT INTO appointment_account_entries (
    appointment_account_id, type, currency, amount_cents,
    description, created_by_user_id,
    doctor_id, doctor_type, commission_pct,
    service_id, cash_session_id
) VALUES (
    'account-uuid', 'service_charge', 'MXN', 300000,
    'Root Canal Treatment', 'receptionist-uuid',
    'doctor-uuid', 'internal', 40.00,
    'service-uuid', 'session-uuid'
);

-- 3. Patient pays cash
INSERT INTO appointment_account_entries (
    appointment_account_id, type, currency, amount_cents,
    description, created_by_user_id,
    payment_method, cash_session_id
) VALUES (
    'account-uuid', 'payment', 'MXN', 300000,
    'Cash payment', 'receptionist-uuid',
    'cash', 'session-uuid'
);

-- 4. Receptionist closes session and reconciles
-- Calculate expected: SELECT SUM(amount_cents) FROM appointment_account_entries
--   WHERE cash_session_id = 'session-uuid' AND type = 'payment' AND payment_method = 'cash'
-- Result: 300000 cents (3000 MXN)

INSERT INTO reconciliations (
    cash_session_id, organization_id, clinic_id,
    payment_method, currency, reconciled_by_user_id,
    expected_amount_cents, actual_amount_cents,
    float_left_cents, deposited_cents, discrepancy_cents,
    envelope_id, status
) VALUES (
    'session-uuid', 'org-uuid', 'clinic-uuid',
    'cash', 'MXN', 'receptionist-uuid',
    300000, 300000,  -- expected and actual match
    50000, 250000, 0,  -- float 500, deposit 2500, no discrepancy
    'R-001-250126-Ana', 'closed'
);

-- Update session as closed
UPDATE cash_sessions
SET status = 'closed', closed_at = NOW()
WHERE id = 'session-uuid';
```

---

## Migration Considerations

1. **Add tables in order**:
   - appointment_accounts (depends on appointments, organizations)
   - cash_sessions (depends on organizations, clinics, users)
   - appointment_account_entries (depends on appointment_accounts, cash_sessions, doctors, services)
   - reconciliations (depends on cash_sessions)

2. **Backward compatibility**:
   - appointment_account_entries.cash_session_id is nullable
   - Existing appointments without accounts are fine (created on-demand)

3. **Data backfill** (optional):
   - Backfill appointment_accounts for historical appointments with payments
   - Mark historical entries with NULL cash_session_id

---

**Status**: Ready for implementation  
**Last Updated**: January 25, 2026
