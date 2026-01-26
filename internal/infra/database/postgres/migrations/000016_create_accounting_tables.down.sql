-- Rollback migration: Drop accounting and cash management tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_reconciliations_updated_at ON reconciliations;
DROP TRIGGER IF EXISTS update_cash_sessions_updated_at ON cash_sessions;
DROP TRIGGER IF EXISTS update_appointment_accounts_updated_at ON appointment_accounts;

-- Drop indexes for reconciliations
DROP INDEX IF EXISTS idx_recon_status_clinic;
DROP INDEX IF EXISTS idx_recon_user_date;
DROP INDEX IF EXISTS idx_recon_clinic_payment_currency_date;
DROP INDEX IF EXISTS idx_recon_session_payment_currency;

-- Drop indexes for appointment_account_entries
DROP INDEX IF EXISTS idx_entries_service;
DROP INDEX IF EXISTS idx_entries_corrects;
DROP INDEX IF EXISTS idx_entries_creator_created;
DROP INDEX IF EXISTS idx_entries_session_type_payment;
DROP INDEX IF EXISTS idx_entries_doctor_created;
DROP INDEX IF EXISTS idx_entries_account_created;

-- Drop indexes for cash_sessions
DROP INDEX IF EXISTS idx_sessions_clinic_closed;
DROP INDEX IF EXISTS idx_sessions_clinic_status_opened;
DROP INDEX IF EXISTS idx_sessions_user_clinic_status;

-- Drop indexes for appointment_accounts
DROP INDEX IF EXISTS idx_appointment_accounts_appointment;
DROP INDEX IF EXISTS idx_appointment_accounts_org_created;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS reconciliations CASCADE;
DROP TABLE IF EXISTS appointment_account_entries CASCADE;
DROP TABLE IF EXISTS cash_sessions CASCADE;
DROP TABLE IF EXISTS appointment_accounts CASCADE;
