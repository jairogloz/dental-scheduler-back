# Accounting System Requirements - Refined (Step 1)

## Core Accounting Principles

### Multi-Currency Support

- **Primary currencies**: MXN and USD
- Each transaction can be in either currency
- **Exchange rate management**:
  - Default exchange rate (USD → MXN) configured per clinic in organization settings
  - Rate can be overridden per transaction by receptionist
  - Store `exchange_rate_used` on each entry for historical accuracy
  - Example: Clinic A default rate = 20.50 MXN per USD

### Appointment-Level Accounting

**Definition**: Each appointment has its own financial ledger where all related transactions are recorded.

**What this means**:

- Appointments and financial transactions are separate entities
- Transactions are **linked to** appointments, not embedded in them
- One appointment can have:
  - Multiple service charges (different treatments)
  - Multiple discounts
  - Multiple payments (partial payments over time)
  - Mix of payment methods (cash + card)
  - Multiple currencies in the same appointment

**Benefits**:

- Clean separation of scheduling and accounting concerns
- Easy to query financial data without touching appointment records
- Historical audit trail preserved independently
- Cancelled/no-show appointments don't pollute financial reports

### Payment Methods

Support for multiple payment methods:

- **Cash**: Physical currency
- **Card**: Credit/debit (grouped together for now)
- **Transfer**: Bank transfer (future)
- **Other**: Flexible for regional payment methods

Each `payment` entry must include:

- `payment_method` field (required)
- `currency` (MXN or USD)
- `amount_cents` (in the specified currency)
- `exchange_rate_used` (if currency is USD)

### Financial Reconciliation (not just "cash")

**Source of truth**: Transaction entries (ledger), not appointments

**Reconciliation reports based on**:

- **Date range**: When payment was recorded
- **Clinic**: Which location received the payment
- **Payment method**: Cash separate from card
- **Currency**: MXN and USD tracked separately, with conversions when needed

**What counts in reconciliation**:

- ✅ All `payment` entries for the period
- ✅ Filtered by clinic and payment method
- ❌ Cancelled appointments (only if payment was actually received)
- ❌ No-show appointments (unless payment was received)

**Example queries**:

- "Cash collected on Jan 23, 2026 at Clinic Norte"
- "Card payments in USD at all clinics this week"
- "All payments by Dr. Martinez's patients in December"

### Transaction Entries Separated from Appointments

**Core architecture decision**:

- Appointments table remains focused on scheduling
- Financial data lives in separate ledger tables
- Linked by `appointment_id` foreign key

**Advantages**:

1. **Multiple entries per appointment**: Add charges/payments/discounts over time
2. **Direct reconciliation queries**: Query ledger directly without joining appointments
3. **Audit trail**: Every change is a new entry, nothing gets deleted
4. **Performance**: Financial reports don't scan appointment tables
5. **Cancelled appointments**: Don't affect financial data unless explicitly recorded

## Data Model Requirements

### appointment_accounts

**Decision: Separate table (Option A)**

- 1:1 relationship with appointment
- Created **only when first charge/payment is recorded** ("Cobrar" workflow)
- Not every appointment has an account (e.g., no-shows, cancellations without charges)
- Exists only for appointments with actual financial activity

**Rationale**:

- Keeps appointment table focused on scheduling
- Cleaner separation of concerns
- Future-proof for account-level metadata (status, locked_at, closed_by, etc.)
- Easier to query "all active accounts" without scanning all appointments

### appointment_entries (ledger)

**Decision: Immutable correction entries (Option B)**

- Many entries per account
- Each entry is **immutable** (append-only, never updated or deleted)
- Errors are corrected by adding new `correction` entries
- Types:
  - `service_charge`: A treatment/service provided
  - `discount`: Reduction in price
  - `payment`: Money received
  - `refund`: Money returned
  - `correction`: Reverses a previous entry (links to original via `corrects_entry_id`)

**Signed Amounts**:

- `service_charge`, `payment` → positive amounts (+300000 = 3000 MXN)
- `discount`, `refund` → negative amounts (-50000 = -500 MXN)
- `correction` → opposite sign of corrected entry
- Totals computed with simple `SUM(amount_cents)`

**Required fields per entry**:

- `type` (enum: service_charge, discount, payment, refund, correction)
- `currency` (enum: MXN, USD)
- `amount_cents` (signed integer, in the specified currency)
- `description` (human-readable)
- `created_by_user_id` (who recorded this)
- `created_at` (when it was recorded)

**Conditional fields**:

- `payment_method` (enum: cash, card, transfer) — **required if type = payment**
- `exchange_rate_used` (decimal) — **required if currency = USD**
- `doctor_id` (uuid) — **required if type = service_charge**
- `corrects_entry_id` (uuid) — **required if type = correction**

**Doctor commission fields** (for service_charge only):

- `doctor_type` (enum: internal, external) — **required if type = service_charge**
- `commission_pct` (decimal) — **required if doctor_type = internal**
  - Example: 40.0 means doctor gets 40% of service_charge amount
- `external_doctor_fee_cents` (integer) — **required if doctor_type = external**
  - Example: 200000 = 2000 MXN flat fee, regardless of what patient was charged
  - **Sensitive field**: Never shown in patient-facing views
- `is_sensitive` (boolean) — true if entry contains data not visible to patients

**Optional fields**:

- `service_id` (uuid) — Link to existing service catalog (already available in appointments)
- `quantity` (integer, default 1) — For multiple units of same service
- `unit_price_cents` (integer) — Price per unit for service_charge
- `notes` (text) — Additional context for corrections or special cases
- `cash_session_id` (uuid, fk to cash_sessions) — Links entry to the cash session when it was created (nullable for backward compatibility)

### cash_sessions (apertura de caja / sesión de caja)

**Purpose**: Track cash handling periods between reconciliations. Decouples cash management from work shifts for flexibility.

**Key insight**: A receptionist can have multiple cash sessions per day. Each session ends with a reconciliation (deposit to safe), allowing mid-day cash drops when thresholds are reached.

**Key principles**:

- One user can only have ONE open cash session at a time per clinic
- Entries can only be created when a cash session is open
- Cash sessions can be opened manually or auto-opened with warning
- After closing (reconciling), user can immediately open new session to continue working
- Each session closure = one deposit to safe (with envelope and reconciliation ID)

**Fields**:

- `id` (uuid, pk)
- `organization_id` (uuid, fk)
- `clinic_id` (uuid, fk)
- `user_id` (uuid, fk) — The receptionist/admin who opened the session
- `opened_at` (timestamp) — When the cash session was opened
- `closed_at` (timestamp, nullable) — When the session was closed (null = currently open)
- `starting_float_cents` (integer) — Amount of cash left in drawer at session start
- `status` (enum: open, closed)
- `opening_type` (enum: manual, auto) — How the session was opened
- `notes` (text, nullable) — Optional notes about the session
- `created_at` (timestamp)
- `updated_at` (timestamp)

**Indexes**:

- (user_id, clinic_id, status) — Find user's open session at clinic
- (clinic_id, status, opened_at) — Find active sessions per clinic
- (clinic_id, closed_at) — Historical session lookup

**Business rules**:

- Only one open cash session per user per clinic at a time
- Only users with receptionist or admin roles can open cash sessions
- Cash session must be open to create appointment_entries
- Closing a cash session requires reconciliation
- Multiple sessions allowed per user per day (for mid-shift cash drops)

### reconciliations (cash closeouts / corte de caja)

**Purpose**: Track cash reconciliations when closing cash sessions. Each reconciliation represents one deposit to safe.

**Key insight**: One reconciliation per payment method per currency per cash session. A cash session might have:

- Reconciliation #1: Cash MXN
- Reconciliation #2: Cash USD
- Reconciliation #3: Card MXN
- Reconciliation #4: Card USD

**Physical process**: Each reconciliation generates an envelope for safe deposit labeled with reconciliation ID.

**Fields**:

- `id` (uuid, pk)
- `cash_session_id` (uuid, fk to cash_sessions) — Links to the cash session being reconciled
- `organization_id` (uuid, fk)
- `clinic_id` (uuid, fk)
- `payment_method` (enum: cash, card, transfer)
- `currency` (enum: MXN, USD)
- `reconciled_at` (timestamp) — When reconciliation was performed
- `reconciled_by_user_id` (uuid, fk) — Who performed the reconciliation
- `expected_amount_cents` (integer) — Calculated from `appointment_entries.payment` WHERE cash_session_id = this.cash_session_id
- `actual_amount_cents` (integer) — What receptionist counted/verified in drawer
- `float_left_cents` (integer) — Amount left in drawer for next session (for change)
- `deposited_cents` (integer) — Amount moved to safe (actual - float_left)
- `discrepancy_cents` (integer) — actual - expected (can be positive or negative)
- `envelope_id` (string, nullable) — Physical envelope identifier for safe deposit
- `status` (enum: pending, closed, disputed)
- `notes` (text) — Receptionist notes about discrepancies or special situations
- `created_at` (timestamp)
- `updated_at` (timestamp)

**Indexes**:

- (cash_session_id, payment_method, currency)
- (clinic_id, payment_method, currency, reconciled_at)
- (reconciled_by_user_id, reconciled_at)
- (status, clinic_id)
- (envelope_id) — For tracking physical envelopes

**Note**: Each cash session can have multiple reconciliations (one per payment_method + currency combination). All reconciliations for a session should be completed before closing the session.

## Invariants & Business Rules

**Note**: These validations will be implemented at the **business layer** (Go domain/usecase), NOT as database constraints. This keeps the database flexible and allows for better error messages and business logic evolution.

1. **Currency consistency**: All entries for an appointment can have different currencies
2. **Exchange rate required**: If entry.currency = USD, then exchange_rate_used must be set
3. **Payment method required**: If entry.type = payment, then payment_method must be set
4. **Amount cannot be zero**: amount_cents != 0
5. **Service charges need doctor**: If type = service_charge, then doctor_id NOT NULL AND doctor_type NOT NULL
6. **Commission requirements**:
   - If doctor_type = internal, then commission_pct must be set
   - If doctor_type = external, then external_doctor_fee_cents must be set
7. **Correction links**: If type = correction, then corrects_entry_id must reference valid entry
8. **Signed amount validation**:
   - service_charge, payment → amount_cents must be positive
   - discount, refund → amount_cents must be negative
   - correction → amount_cents must be opposite sign of corrected entry
9. **Immutability**: Entries are never updated or deleted, only created (append-only ledger)
10. **Audit trail**: Every entry records who created it and when
11. **Sensitive data**: external_doctor_fee_cents must never appear in patient-facing APIs/views
12. **Cash session requirement**: appointment_entries can only be created when user has an open cash_session
13. **One cash session per user**: A user can only have one open cash_session at a time per clinic
14. **Cash session permissions**: Only receptionists and admins can open cash sessions
15. **Multiple sessions per day**: A user can have multiple cash sessions per day (for mid-shift cash drops)

**Database Constraints** (minimal, only for data integrity):

- Foreign keys (appointment_id, doctor_id, created_by_user_id, corrects_entry_id, cash_session_id)
- NOT NULL on truly required fields (id, type, currency, amount_cents, created_at, created_by_user_id)
- Unique constraints where needed (appointment_id in appointment_accounts)
- Indexes for performance

**Important**: No locks or constraints prevent adding entries after reconciliation. See "Corrections After Reconciliation" workflow below.

## Workflows

### Cash Session Management Workflow

**Opening a cash session (Manual)**:

1. Receptionist clicks "Abrir Caja" button
2. System checks: Does user already have an open cash session at this clinic?
   - If yes → Show error: "Ya tienes una caja abierta en Clínica X desde HH:MM"
   - If no → System asks for starting float amount (suggested from clinic config, e.g., 500 MXN)
3. System creates cash_session record with `status = open`, `opening_type = manual`, `starting_float_cents`
4. Display shows: "Caja abierta: 09:00 - ahora (Ana, fondo inicial: $500)"
5. User can now create entries

**Opening a cash session (Automatic)**:

1. Receptionist tries to create first entry (charge/payment) without open cash session
2. System shows dialog:

   ```
   ⚠️ No tienes una caja abierta

   Para registrar pagos necesitas abrir una sesión de caja.
   Al abrir una sesión, serás responsable de todos los
   movimientos hasta que cierres el corte.

   Fondo inicial en caja: [500] MXN

   [Cancelar]  [Abrir Caja y Continuar]
   ```

3. If user confirms:
   - Create cash_session with `status = open`, `opening_type = auto`, `starting_float_cents`
   - Link entry to this session
   - Continue with original operation
4. If user cancels → Operation cancelled

**During cash session**:

- All entries created are automatically linked to current open session via `cash_session_id`
- Display always shows: "Caja: 09:00 - ahora (activa, Ana)"
- User can view current session totals in real-time
- When cash accumulates to threshold (e.g., 5,000 MXN), user can choose to close and reopen

**Closing a cash session (Mid-shift cash drop)**:

1. User clicks "Cerrar Caja" (cash has reached threshold or end of day)
2. System shows reconciliation modal (see "Reconciliation Workflow" below)
3. After successful reconciliation:
   - Cash_session status changes to `closed`
   - Money deposited to safe in envelope labeled with reconciliation ID
   - User CAN immediately open new cash session to continue working
   - New session starts with the `float_left_cents` from previous reconciliation

**Multiple sessions per day (typical flow)**:

```
09:00 - Ana opens Cash Session #1 (starting float: 500 MXN)
... patients come, cash builds to 5,500 MXN ...

14:00 - Ana closes Cash Session #1
        Reconciles: Expected 5,500, Actual 5,500 ✓
        Float left: 500 MXN
        Deposited: 5,000 MXN (Envelope #R001)

14:05 - Ana opens Cash Session #2 (starting float: 500 MXN)
... more patients ...

18:00 - Ana closes Cash Session #2 (end of day)
        Reconciles: Expected 3,200, Actual 3,200 ✓
        Float left: 500 MXN (for tomorrow)
        Deposited: 2,700 MXN (Envelope #R002)
```

### "Cobrar" (Charge) Workflow

**Prerequisites**: User must have an open cash session

1. User opens appointment details
2. User clicks "Cobrar" button
3. **System checks**: Does user have open cash session?
   - If no → Show "Abrir Caja" dialog (see above)
   - If yes → Continue
4. System creates `appointment_account` record (if doesn't exist)
5. System shows charging modal with:
   - Default service from appointment (pre-filled, includes `service_id`)
   - Doctor from appointment (pre-filled, includes `doctor_id`)
   - Calculated price from service catalog
   - Currency selector (MXN/USD)
   - Exchange rate (if USD selected)
6. User confirms → `service_charge` entry created with `service_id` from appointment
7. User can add payment → `payment` entry created
8. Account balance shown (charged - paid)

**Note**: `service_id` is already available from the appointment, no need to create service catalog from scratch

### Reconciliation (Corte de Caja) Workflow

**Prerequisites**: User must have an open cash session to close

**The flow**:

1. Receptionist clicks "Cerrar Caja"
2. System validates:
   - User has an open cash session
   - Gets current session details (cash_session_id, opened_at, starting_float)
3. System displays last session/closeout info:
   - "Última sesión: Hoy 14:00 por Ana (depositó $5,000)" (if same day, previous session)
   - "Última sesión: Ayer 18:15 por María" (if previous day)
   - "Primera sesión del día" (if no previous session today)
4. System shows reconciliation modal with payment methods/currencies
5. For each combination (Cash MXN, Cash USD, Card MXN, Card USD):
   - System calculates `expected_amount` from entries WHERE `cash_session_id = current_session.id`
   - Receptionist enters `actual_amount` (what they counted/verified in drawer)
   - Receptionist enters `float_to_leave` (amount to leave for next session, suggested: 500 MXN)
   - System calculates `to_deposit = actual - float_to_leave`
   - System shows `discrepancy` (if any)
6. Receptionist adds notes (optional, especially if discrepancies exist)
7. System generates `envelope_id` (e.g., format: R-{reconciliation_id}-{date})
8. Receptionist confirms → System:
   - Creates `reconciliation` record(s) linked to cash_session
   - Sets cash_session `status = closed`, `closed_at = NOW()`
   - Displays: "Depositar en sobre: Reconciliación #{envelope_id} - Ana - $5,000 MXN"
   - User's cash session is now closed

**What gets calculated**:

```sql
-- Expected cash in MXN for current session
SELECT SUM(amount_cents)
FROM appointment_entries
WHERE type = 'payment'
  AND payment_method = 'cash'
  AND currency = 'MXN'
  AND cash_session_id = ?
```

**Display to receptionist**:

```
Cerrar Caja - Clínica Norte
Sesión iniciada: 24/01/2026 09:00 por Ana (fondo inicial: $500)
Última sesión: 23/01/2026 18:15 por María

Efectivo MXN:
  Esperado: $5,500.00
  Contado en caja: [ ___________ ] ← receptionist enters
  Fondo a dejar: [500] ← receptionist enters (suggested)
  A depositar: $5,000.00 (auto-calculated: contado - fondo)
  Diferencia: $0.00 (auto-calculated: contado - esperado)

Efectivo USD:
  Esperado: $250.00
  Contado: [ ___________ ]
  Fondo a dejar: [20]
  A depositar: $230.00
  Diferencia: $0.00

Tarjeta MXN:
  Esperado: $8,500.00
  Real: [ ___________ ]
  ...

Notas (opcional): [ _________________ ]

📋 Sobre para caja fuerte: R-001-240126-Ana
   Contenido: $5,000 MXN + $230 USD

[Cancelar]  [Cerrar Caja y Depositar]
```

**After reconciliation**:

- Reconciliation records created linked to cash_session with envelope_id
- Cash_session marked as closed
- Receptionist can immediately open new cash session to continue (optional)
- Physical envelope labeled with reconciliation ID goes to safe
- Receptionist can see history of all cash sessions and reconciliations
- Admin can review discrepancies

### Corrections After Reconciliation Workflow

**Scenario**: Cash session was closed at 14:00, but at 14:30 receptionist realizes a payment from 11:00 was recorded with wrong amount.

**Design principle**: Entries are NEVER locked, reconciliations are historical snapshots.

**What happens**:

1. Receptionist must open NEW cash session first (auto-open dialog appears if not already open)
2. Receptionist creates correction entry (as usual)
3. Correction entry has `created_at = 14:30`, `cash_session_id = new_session_id` (different session!)
4. Original reconciliation record remains unchanged
5. System flags: "Entries were modified after this cash session was closed"

**Scenario**: Shift was closed at 18:30, but at 19:00 receptionist realizes a payment from 15:00 was recorded with wrong amount.

**Design principle**: Entries are NEVER locked, reconciliations are historical snapshots.

**What happens**:

1. Receptionist must open NEW shift first (auto-open dialog appears)
2. Receptionist creates correction entry (as usual)
3. Correction entry has `created_at = 19:00`, `shift_id = new_shift_id` (different shift!)
4. Original reconciliation record remains unchanged
5. System flags: "Entries were modified after last reconciliation"
6. Two options for handling:

**Option A: Re-run historical reconciliation** (recommended for large discrepancies)

- Admin can view "What should reconciliation have been with corrections applied"
- System recalculates reconciliation period with new data
- Shows comparison: Original vs Corrected amounts
- No database changes to original reconciliation
- Purely informational view

**Option B: Adjust next reconciliation** (simpler, for small discrepancies)

- Correction appears in NEXT reconciliation period
- Next reconciliation's `expected_amount` will reflect the correction
- Discrepancy naturally appears in next closeout
- Receptionist notes: "Includes correction from previous shift"

**Example**:

```
Shift 1 (09:00-18:30):
- Expected: 15,000 MXN
- Actual: 15,000 MXN
- ✅ Reconciliation closed

19:00 - Correction made: entry from 15:00 changed from 3000 to 2000
**Example with cash sessions**:
```

Cash Session #1 (Ana, 09:00-14:00):
- Reconciliation at 14:00:
  - Cash MXN Expected: 5,500 MXN
  - Cash MXN Actual: 5,500 MXN
  - Float left: 500 MXN
  - Deposited: 5,000 MXN (Envelope R-001) ✅
  - Session closed

14:30 - Ana opens Cash Session #2
14:35 - Ana creates correction: entry from session #1 at 11:00 changed from 3000 to 2000
        Correction is in session #2, references entry from session #1
        Effect: -1000 MXN for session #1 calculations (informational)

When Ana closes Session #2 at 18:00:
- Expected includes only session #2 entries (NOT corrections from session #1)
- Physical cash is correct (customer got charged 3000, we deposited it)

Admin view for Session #1:
- Original reconciliation: 5,500 expected, 5,500 actual, 5,000 deposited ✅
- ⚠️ Post-closure correction: -1000 MXN made in Session #2 by Ana
- Adjusted calculation (informational): Should have been 4,500 expected
- Note: Physical deposit was correct at time of reconciliation
```

**System notifications**:

- When viewing closed session: "⚠️ 1 correction was made after this session closed (Session #2 by Ana)"
- Link to view affected entries and correction
- Option to see adjusted reconciliation (informational only)

**Why this works**:

- Cash session-based accountability (can't create entries without session)
- Complete audit trail maintained (corrections linked to original session entries)
- Errors are visible and traceable (cross-session corrections flagged)
- Physical cash reality is preserved (deposits matched physical count)
- Admin can always review and understand discrepancies
- Multiple sessions per day support mid-shift cash drops

### Error Correction Workflow

1. User realizes mistake (e.g., charged 3000 instead of 2000)
2. User selects erroneous entry and clicks "Corregir"
3. System creates two entries:
   - `correction` entry: links to original, reverses amount (-3000)
   - New correct entry: the intended amount (+2000)
4. Original entry remains visible in audit history
5. Balance computed correctly: 3000 - 3000 + 2000 = 2000

**Why this is better than delete/edit**:

- Complete audit trail of all mistakes
- Shows who made error and who corrected it
- Can review correction patterns for training
- Complies with financial/medical audit requirements

### Commission Examples

**Internal Doctor (commission_pct)**:

```

Service: Root Canal
Charged to patient: 3000 MXN
Dr. Martinez (internal): 40% commission
→ Entry: service_charge, amount_cents=300000, doctor_type=internal, commission_pct=40.0
→ Dr. Martinez will receive: 1200 MXN (computed later)

```

**External Doctor (flat fee)**:

```

Service: Orthodontics Consultation
Charged to patient: 3000 MXN
Dr. Lopez (external) fixed fee: 2000 MXN
→ Entry: service_charge, amount_cents=300000, doctor_type=external, external_doctor_fee_cents=200000, is_sensitive=true
→ Dr. Lopez receives: 2000 MXN
→ Clinic keeps: 1000 MXN
→ Patient NEVER sees the 2000 MXN breakdown

````

## Reports Required

### Daily Cash Reconciliation

- By clinic
- By payment method (cash vs. card separately)
- By currency (MXN and USD shown separately)
- Date range filter
- Include: payments received that day
- Exclude: cancelled appointments (unless payment recorded)

### Doctor Collections

- Total charged per doctor
- Total collected per doctor
- By date range
- By clinic
- Future: Commission calculations

### Patient Account Balance

- Total charged for appointment
- Total paid for appointment
- Balance due
- Payment history

## Configuration Needed

### Hardcoded Application Constants

**Payment methods** (enum in backend):

- `cash`
- `card` (credit/debit grouped)
- `transfer`

**Currencies** (enum in backend):

- `MXN`
- `USD`

**Doctor types** (enum in backend):

- `internal` (receives commission_pct)
- `external` (receives flat fee)

### Organization/Clinic Settings

Add to clinic configuration:

```json
{
  "default_exchange_rate_usd_to_mxn": 20.5
}
````

**Note**: If a clinic needs to disable a payment method or add new ones in the future, this can be made configurable. Start simple.

### User Permissions

- **Receptionist**: Create entries, create corrections, override exchange rate per transaction
- **Doctor**: Read-only access to their entries (cannot see sensitive fields)
- **Admin**: Full access including sensitive fields, can lock account after closeout
- **Optional**: Lock accounts after daily closeout (prevents further entries)

## Next Steps

1. ✅ Define requirements (this document)
2. ⏭️ Design database schema with multi-currency support
3. ⏭️ Implement domain layer (Go entities and business logic)
4. ⏭️ Create repository interfaces and implementations
5. ⏭️ Build API endpoints
6. ⏭️ Create frontend UI for charge entry
7. ⏭️ Build reconciliation reports
8. ⏭️ Implement commission calculation reports

## Pending / Future Items

### Referral Commissions

**Not in immediate scope, but keep in mind**:

- Doctors/staff can earn commissions for referring patients
- Different from service commissions
- Possible approaches:
  - Separate `referral_commissions` table
  - New entry type: `referral_commission` in appointment_entries
  - Link to `referring_doctor_id` or `referring_source_id`
- May have different commission structures:
  - Flat fee per referral
  - Percentage of first appointment
  - Percentage of patient lifetime value
- Need to track:
  - Who referred
  - Which patient was referred
  - When commission is paid
  - How much

**Decision needed**: Separate subsystem or integrate into appointment_entries?

### Enhanced Service Catalog

**Note**: Service catalog already exists and `service_id` is available from appointments.

Future enhancements:

- Service price history tracking
- Price variations by clinic/doctor
- Service bundles/packages
- Automatic price suggestions based on historical data

### Advanced Features

- Insurance claims integration
- Treatment plans with multiple appointments
- Payment plans (installments)
- Automated reminders for unpaid balances
- Credit notes / gift cards

---

**Date**: January 24, 2026  
**Status**: Requirements refined with decisions on architecture, commissions, and corrections
