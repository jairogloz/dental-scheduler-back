-- Migration: Create accounting and cash management tables
-- Description: Adds appointment_accounts, appointment_account_entries, cash_sessions, and reconciliations tables

-- ============================================================================
-- Table 1: appointment_accounts
-- Purpose: Links appointments to their financial ledger (created on-demand)
-- ============================================================================

CREATE TABLE appointment_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    appointment_id UUID NOT NULL UNIQUE REFERENCES appointments(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for appointment_accounts
CREATE INDEX idx_appointment_accounts_org_created ON appointment_accounts(organization_id, created_at);
CREATE UNIQUE INDEX idx_appointment_accounts_appointment ON appointment_accounts(appointment_id);

-- Comments for appointment_accounts
COMMENT ON TABLE appointment_accounts IS 'Financial account for appointments with actual financial activity';
COMMENT ON COLUMN appointment_accounts.appointment_id IS 'One-to-one relationship with appointment';
COMMENT ON COLUMN appointment_accounts.organization_id IS 'Organization that owns this account';

-- ============================================================================
-- Table 2: cash_sessions
-- Purpose: Track cash handling periods (multiple sessions per day allowed)
-- ============================================================================

CREATE TABLE cash_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id),
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    starting_float_cents BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL,
    opening_type VARCHAR(20) NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for cash_sessions
CREATE INDEX idx_sessions_user_clinic_status ON cash_sessions(user_id, clinic_id, status);
CREATE INDEX idx_sessions_clinic_status_opened ON cash_sessions(clinic_id, status, opened_at);
CREATE INDEX idx_sessions_clinic_closed ON cash_sessions(clinic_id, closed_at);

-- Comments for cash_sessions
COMMENT ON TABLE cash_sessions IS 'Cash handling periods between reconciliations (multiple per day allowed)';
COMMENT ON COLUMN cash_sessions.user_id IS 'The receptionist/admin who opened the session';
COMMENT ON COLUMN cash_sessions.closed_at IS 'NULL indicates session is currently open';
COMMENT ON COLUMN cash_sessions.starting_float_cents IS 'Cash in drawer at start of session for making change';
COMMENT ON COLUMN cash_sessions.status IS 'Session status: open, closed (validated in app layer)';
COMMENT ON COLUMN cash_sessions.opening_type IS 'How opened: manual, auto (validated in app layer)';

-- ============================================================================
-- Table 3: appointment_account_entries
-- Purpose: Immutable ledger of all financial transactions
-- ============================================================================

CREATE TABLE appointment_account_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    appointment_account_id UUID NOT NULL REFERENCES appointment_accounts(id) ON DELETE CASCADE,
    
    -- Core transaction fields
    type VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    amount_cents BIGINT NOT NULL,
    description TEXT NOT NULL,
    created_by_user_id UUID NOT NULL REFERENCES auth.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Conditional fields (required based on type)
    payment_method VARCHAR(20),
    exchange_rate_used NUMERIC(10,4),
    doctor_id UUID REFERENCES doctors(id),
    corrects_entry_id UUID REFERENCES appointment_account_entries(id),
    
    -- Doctor commission fields (for service_charge only)
    doctor_type VARCHAR(20),
    commission_pct NUMERIC(5,2),
    external_doctor_fee_cents BIGINT,
    is_sensitive BOOLEAN DEFAULT FALSE,
    
    -- Optional fields
    service_id VARCHAR(255) REFERENCES services(id),
    quantity INTEGER DEFAULT 1,
    unit_price_cents BIGINT,
    notes TEXT,
    cash_session_id UUID REFERENCES cash_sessions(id),
    
    -- Constraints
    CONSTRAINT amount_not_zero CHECK (amount_cents != 0)
);

-- Indexes for appointment_account_entries
CREATE INDEX idx_entries_account_created ON appointment_account_entries(appointment_account_id, created_at);
CREATE INDEX idx_entries_doctor_created ON appointment_account_entries(doctor_id, created_at);
CREATE INDEX idx_entries_session_type_payment ON appointment_account_entries(cash_session_id, type, payment_method, currency);
CREATE INDEX idx_entries_creator_created ON appointment_account_entries(created_by_user_id, created_at);
CREATE INDEX idx_entries_corrects ON appointment_account_entries(corrects_entry_id);
CREATE INDEX idx_entries_service ON appointment_account_entries(service_id);

-- Comments for appointment_account_entries
COMMENT ON TABLE appointment_account_entries IS 'Immutable ledger of all financial transactions for appointments';
COMMENT ON COLUMN appointment_account_entries.type IS 'Entry type: service_charge, discount, payment, refund, correction (validated in app layer)';
COMMENT ON COLUMN appointment_account_entries.currency IS 'Currency code: MXN, USD (validated in app layer)';
COMMENT ON COLUMN appointment_account_entries.amount_cents IS 'Signed amount in cents: positive for income, negative for expenses';
COMMENT ON COLUMN appointment_account_entries.payment_method IS 'Payment method: cash, card, transfer (required if type=payment, validated in app layer)';
COMMENT ON COLUMN appointment_account_entries.exchange_rate_used IS 'Required if currency=USD, stores USD to MXN exchange rate';
COMMENT ON COLUMN appointment_account_entries.doctor_type IS 'Doctor type: internal, external (required if type=service_charge, validated in app layer)';
COMMENT ON COLUMN appointment_account_entries.commission_pct IS 'Commission % for internal doctors (required if doctor_type=internal)';
COMMENT ON COLUMN appointment_account_entries.external_doctor_fee_cents IS 'Flat fee for external doctors (required if doctor_type=external)';
COMMENT ON COLUMN appointment_account_entries.is_sensitive IS 'Hides external doctor fees from patient-facing views';
COMMENT ON COLUMN appointment_account_entries.corrects_entry_id IS 'References original entry if this is a correction';
COMMENT ON COLUMN appointment_account_entries.cash_session_id IS 'Cash session when entry was created (for reconciliation)';

-- ============================================================================
-- Table 4: reconciliations
-- Purpose: Record cash reconciliations when closing cash sessions
-- ============================================================================

CREATE TABLE reconciliations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cash_session_id UUID NOT NULL REFERENCES cash_sessions(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    clinic_id UUID NOT NULL REFERENCES clinics(id) ON DELETE CASCADE,
    payment_method VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    reconciled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reconciled_by_user_id UUID NOT NULL REFERENCES auth.users(id),
    
    -- Reconciliation amounts (all in cents)
    expected_amount_cents BIGINT NOT NULL,
    actual_amount_cents BIGINT NOT NULL,
    float_left_cents BIGINT NOT NULL,
    deposited_cents BIGINT NOT NULL,
    discrepancy_cents BIGINT NOT NULL,
    
    status VARCHAR(20) NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for reconciliations
CREATE INDEX idx_recon_session_payment_currency ON reconciliations(cash_session_id, payment_method, currency);
CREATE INDEX idx_recon_clinic_payment_currency_date ON reconciliations(clinic_id, payment_method, currency, reconciled_at);
CREATE INDEX idx_recon_user_date ON reconciliations(reconciled_by_user_id, reconciled_at);
CREATE INDEX idx_recon_status_clinic ON reconciliations(status, clinic_id);

-- Comments for reconciliations
COMMENT ON TABLE reconciliations IS 'Cash reconciliations when closing sessions (one per payment_method + currency)';
COMMENT ON COLUMN reconciliations.payment_method IS 'Payment method: cash, card, transfer (validated in app layer)';
COMMENT ON COLUMN reconciliations.currency IS 'Currency code: MXN, USD (validated in app layer)';
COMMENT ON COLUMN reconciliations.expected_amount_cents IS 'Calculated from entries in this cash session';
COMMENT ON COLUMN reconciliations.actual_amount_cents IS 'What receptionist counted in drawer';
COMMENT ON COLUMN reconciliations.float_left_cents IS 'Amount left for next session (for change)';
COMMENT ON COLUMN reconciliations.deposited_cents IS 'Amount to safe = actual - float_left';
COMMENT ON COLUMN reconciliations.discrepancy_cents IS 'Difference = actual - expected';
COMMENT ON COLUMN reconciliations.status IS 'Reconciliation status: pending, closed, disputed (validated in app layer)';

-- ============================================================================
-- Updated_at triggers for new tables
-- ============================================================================

CREATE TRIGGER update_appointment_accounts_updated_at
    BEFORE UPDATE ON appointment_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cash_sessions_updated_at
    BEFORE UPDATE ON cash_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reconciliations_updated_at
    BEFORE UPDATE ON reconciliations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Note: appointment_account_entries doesn't have updated_at trigger
-- because it's an immutable ledger (entries are never updated)
