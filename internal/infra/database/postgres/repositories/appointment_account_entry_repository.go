package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// AppointmentAccountEntryPostgresRepository implements the AppointmentAccountEntryRepository interface
type AppointmentAccountEntryPostgresRepository struct {
	db *sql.DB
}

// NewAppointmentAccountEntryPostgresRepository creates a new instance
func NewAppointmentAccountEntryPostgresRepository(db *sql.DB) repositories.AppointmentAccountEntryRepository {
	return &AppointmentAccountEntryPostgresRepository{db: db}
}

// Create creates a new entry (immutable - no update or delete)
func (r *AppointmentAccountEntryPostgresRepository) Create(ctx context.Context, entry *entities.AppointmentAccountEntry) error {
	query := `
		INSERT INTO appointment_account_entries (
			id, appointment_account_id, type, currency, amount_cents, description,
			created_by_user_id, created_at, payment_method, exchange_rate_used,
			doctor_id, corrects_entry_id, doctor_type, commission_pct,
			external_doctor_fee_cents, is_sensitive, service_id, quantity,
			unit_price_cents, notes, cash_session_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err := r.db.ExecContext(ctx, query,
		entry.ID,
		entry.AppointmentAccountID,
		entry.Type,
		entry.Currency,
		entry.AmountCents,
		entry.Description,
		entry.CreatedByUserID,
		entry.CreatedAt,
		entry.PaymentMethod,
		entry.ExchangeRateUsed,
		entry.DoctorID,
		entry.CorrectsEntryID,
		entry.DoctorType,
		entry.CommissionPct,
		entry.ExternalDoctorFeeCents,
		entry.IsSensitive,
		entry.ServiceID,
		entry.Quantity,
		entry.UnitPriceCents,
		entry.Notes,
		entry.CashSessionID,
	)

	if err != nil {
		return fmt.Errorf("failed to create appointment account entry: %w", err)
	}

	return nil
}

// GetByID retrieves an entry by its ID
func (r *AppointmentAccountEntryPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.AppointmentAccountEntry, error) {
	query := `
		SELECT id, appointment_account_id, type, currency, amount_cents, description,
			created_by_user_id, created_at, payment_method, exchange_rate_used,
			doctor_id, corrects_entry_id, doctor_type, commission_pct,
			external_doctor_fee_cents, is_sensitive, service_id, quantity,
			unit_price_cents, notes, cash_session_id
		FROM appointment_account_entries
		WHERE id = $1`

	var entry entities.AppointmentAccountEntry
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&entry.ID,
		&entry.AppointmentAccountID,
		&entry.Type,
		&entry.Currency,
		&entry.AmountCents,
		&entry.Description,
		&entry.CreatedByUserID,
		&entry.CreatedAt,
		&entry.PaymentMethod,
		&entry.ExchangeRateUsed,
		&entry.DoctorID,
		&entry.CorrectsEntryID,
		&entry.DoctorType,
		&entry.CommissionPct,
		&entry.ExternalDoctorFeeCents,
		&entry.IsSensitive,
		&entry.ServiceID,
		&entry.Quantity,
		&entry.UnitPriceCents,
		&entry.Notes,
		&entry.CashSessionID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get appointment account entry: %w", err)
	}

	return &entry, nil
}

// List retrieves entries with optional filters
func (r *AppointmentAccountEntryPostgresRepository) List(ctx context.Context, filters repositories.AppointmentAccountEntryFilters) ([]*entities.AppointmentAccountEntry, error) {
	query := `
		SELECT id, appointment_account_id, type, currency, amount_cents, description,
			created_by_user_id, created_at, payment_method, exchange_rate_used,
			doctor_id, corrects_entry_id, doctor_type, commission_pct,
			external_doctor_fee_cents, is_sensitive, service_id, quantity,
			unit_price_cents, notes, cash_session_id
		FROM appointment_account_entries
		WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	if filters.AppointmentAccountID != nil {
		query += fmt.Sprintf(" AND appointment_account_id = $%d", argPos)
		args = append(args, *filters.AppointmentAccountID)
		argPos++
	}

	if filters.CashSessionID != nil {
		query += fmt.Sprintf(" AND cash_session_id = $%d", argPos)
		args = append(args, *filters.CashSessionID)
		argPos++
	}

	if filters.DoctorID != nil {
		query += fmt.Sprintf(" AND doctor_id = $%d", argPos)
		args = append(args, *filters.DoctorID)
		argPos++
	}

	if filters.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", argPos)
		args = append(args, *filters.Type)
		argPos++
	}

	if filters.PaymentMethod != nil {
		query += fmt.Sprintf(" AND payment_method = $%d", argPos)
		args = append(args, *filters.PaymentMethod)
		argPos++
	}

	if filters.Currency != nil {
		query += fmt.Sprintf(" AND currency = $%d", argPos)
		args = append(args, *filters.Currency)
		argPos++
	}

	if filters.CreatedByUserID != nil {
		query += fmt.Sprintf(" AND created_by_user_id = $%d", argPos)
		args = append(args, *filters.CreatedByUserID)
		argPos++
	}

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, *filters.StartDate)
		argPos++
	}

	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, *filters.EndDate)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filters.Limit)
		argPos++

		if filters.Page > 0 {
			offset := (filters.Page - 1) * filters.Limit
			query += fmt.Sprintf(" OFFSET $%d", argPos)
			args = append(args, offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list appointment account entries: %w", err)
	}
	defer rows.Close()

	var entries []*entities.AppointmentAccountEntry
	for rows.Next() {
		var entry entities.AppointmentAccountEntry
		err := rows.Scan(
			&entry.ID,
			&entry.AppointmentAccountID,
			&entry.Type,
			&entry.Currency,
			&entry.AmountCents,
			&entry.Description,
			&entry.CreatedByUserID,
			&entry.CreatedAt,
			&entry.PaymentMethod,
			&entry.ExchangeRateUsed,
			&entry.DoctorID,
			&entry.CorrectsEntryID,
			&entry.DoctorType,
			&entry.CommissionPct,
			&entry.ExternalDoctorFeeCents,
			&entry.IsSensitive,
			&entry.ServiceID,
			&entry.Quantity,
			&entry.UnitPriceCents,
			&entry.Notes,
			&entry.CashSessionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan appointment account entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetByAccountID retrieves all entries for an appointment account
func (r *AppointmentAccountEntryPostgresRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entities.AppointmentAccountEntry, error) {
	filters := repositories.AppointmentAccountEntryFilters{
		AppointmentAccountID: &accountID,
	}
	return r.List(ctx, filters)
}

// GetByCashSessionID retrieves all entries for a cash session
func (r *AppointmentAccountEntryPostgresRepository) GetByCashSessionID(ctx context.Context, cashSessionID uuid.UUID) ([]*entities.AppointmentAccountEntry, error) {
	filters := repositories.AppointmentAccountEntryFilters{
		CashSessionID: &cashSessionID,
	}
	return r.List(ctx, filters)
}

// GetBalance calculates the balance for an appointment account
func (r *AppointmentAccountEntryPostgresRepository) GetBalance(ctx context.Context, accountID uuid.UUID) (*repositories.AccountBalance, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN type = 'service_charge' THEN amount_cents ELSE 0 END), 0) as total_charges,
			COALESCE(SUM(CASE WHEN type = 'discount' THEN amount_cents ELSE 0 END), 0) as total_discounts,
			COALESCE(SUM(CASE WHEN type = 'payment' THEN amount_cents ELSE 0 END), 0) as total_payments,
			COALESCE(SUM(CASE WHEN type = 'refund' THEN amount_cents ELSE 0 END), 0) as total_refunds
		FROM appointment_account_entries
		WHERE appointment_account_id = $1`

	var totalCharges, totalDiscounts, totalPayments, totalRefunds int64
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&totalCharges,
		&totalDiscounts,
		&totalPayments,
		&totalRefunds,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate balance: %w", err)
	}

	balanceDue := totalCharges + totalDiscounts - totalPayments + totalRefunds

	// Get payments by currency
	paymentsByCurrency := make(map[entities.Currency]int64)
	currencyQuery := `
		SELECT currency, COALESCE(SUM(amount_cents), 0)
		FROM appointment_account_entries
		WHERE appointment_account_id = $1 AND type = 'payment'
		GROUP BY currency`

	rows, err := r.db.QueryContext(ctx, currencyQuery, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by currency: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var currency entities.Currency
		var amount int64
		if err := rows.Scan(&currency, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan currency payment: %w", err)
		}
		paymentsByCurrency[currency] = amount
	}

	// Get payments by method
	paymentsByMethod := make(map[entities.PaymentMethod]int64)
	methodQuery := `
		SELECT payment_method, COALESCE(SUM(amount_cents), 0)
		FROM appointment_account_entries
		WHERE appointment_account_id = $1 AND type = 'payment' AND payment_method IS NOT NULL
		GROUP BY payment_method`

	rows, err = r.db.QueryContext(ctx, methodQuery, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by method: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var method entities.PaymentMethod
		var amount int64
		if err := rows.Scan(&method, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan method payment: %w", err)
		}
		paymentsByMethod[method] = amount
	}

	return &repositories.AccountBalance{
		TotalChargesCents:   totalCharges,
		TotalDiscountsCents: totalDiscounts,
		TotalPaymentsCents:  totalPayments,
		TotalRefundsCents:   totalRefunds,
		BalanceDueCents:     balanceDue,
		PaymentsByCurrency:  paymentsByCurrency,
		PaymentsByMethod:    paymentsByMethod,
	}, nil
}

// GetPaymentsByCashSession retrieves payment entries grouped by payment method and currency for a session
func (r *AppointmentAccountEntryPostgresRepository) GetPaymentsByCashSession(ctx context.Context, cashSessionID uuid.UUID) (map[entities.PaymentMethod]map[entities.Currency]int64, error) {
	query := `
		SELECT payment_method, currency, COALESCE(SUM(amount_cents), 0)
		FROM appointment_account_entries
		WHERE cash_session_id = $1 AND type = 'payment' AND payment_method IS NOT NULL
		GROUP BY payment_method, currency`

	rows, err := r.db.QueryContext(ctx, query, cashSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by cash session: %w", err)
	}
	defer rows.Close()

	result := make(map[entities.PaymentMethod]map[entities.Currency]int64)

	for rows.Next() {
		var method entities.PaymentMethod
		var currency entities.Currency
		var amount int64

		if err := rows.Scan(&method, &currency, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}

		if result[method] == nil {
			result[method] = make(map[entities.Currency]int64)
		}
		result[method][currency] = amount
	}

	return result, nil
}

// GetCorrections retrieves all correction entries for a specific entry
func (r *AppointmentAccountEntryPostgresRepository) GetCorrections(ctx context.Context, entryID uuid.UUID) ([]*entities.AppointmentAccountEntry, error) {
	query := `
		SELECT id, appointment_account_id, type, currency, amount_cents, description,
			created_by_user_id, created_at, payment_method, exchange_rate_used,
			doctor_id, corrects_entry_id, doctor_type, commission_pct,
			external_doctor_fee_cents, is_sensitive, service_id, quantity,
			unit_price_cents, notes, cash_session_id
		FROM appointment_account_entries
		WHERE corrects_entry_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get corrections: %w", err)
	}
	defer rows.Close()

	var entries []*entities.AppointmentAccountEntry
	for rows.Next() {
		var entry entities.AppointmentAccountEntry
		err := rows.Scan(
			&entry.ID,
			&entry.AppointmentAccountID,
			&entry.Type,
			&entry.Currency,
			&entry.AmountCents,
			&entry.Description,
			&entry.CreatedByUserID,
			&entry.CreatedAt,
			&entry.PaymentMethod,
			&entry.ExchangeRateUsed,
			&entry.DoctorID,
			&entry.CorrectsEntryID,
			&entry.DoctorType,
			&entry.CommissionPct,
			&entry.ExternalDoctorFeeCents,
			&entry.IsSensitive,
			&entry.ServiceID,
			&entry.Quantity,
			&entry.UnitPriceCents,
			&entry.Notes,
			&entry.CashSessionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan correction entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}
