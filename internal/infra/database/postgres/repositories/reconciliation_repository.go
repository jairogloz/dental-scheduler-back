package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// ReconciliationPostgresRepository implements the ReconciliationRepository interface
type ReconciliationPostgresRepository struct {
	db *sql.DB
}

// NewReconciliationPostgresRepository creates a new instance
func NewReconciliationPostgresRepository(db *sql.DB) repositories.ReconciliationRepository {
	return &ReconciliationPostgresRepository{db: db}
}

// Create creates a new reconciliation
func (r *ReconciliationPostgresRepository) Create(ctx context.Context, reconciliation *entities.Reconciliation) error {
	query := `
		INSERT INTO reconciliations (
			id, cash_session_id, organization_id, clinic_id, payment_method, currency,
			reconciled_at, reconciled_by_user_id, expected_amount_cents, actual_amount_cents,
			float_left_cents, deposited_cents, discrepancy_cents, status, notes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`

	_, err := r.db.ExecContext(ctx, query,
		reconciliation.ID,
		reconciliation.CashSessionID,
		reconciliation.OrganizationID,
		reconciliation.ClinicID,
		reconciliation.PaymentMethod,
		reconciliation.Currency,
		reconciliation.ReconciledAt,
		reconciliation.ReconciledByUserID,
		reconciliation.ExpectedAmountCents,
		reconciliation.ActualAmountCents,
		reconciliation.FloatLeftCents,
		reconciliation.DepositedCents,
		reconciliation.DiscrepancyCents,
		reconciliation.Status,
		reconciliation.Notes,
		reconciliation.CreatedAt,
		reconciliation.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create reconciliation: %w", err)
	}

	return nil
}

// Update updates an existing reconciliation
func (r *ReconciliationPostgresRepository) Update(ctx context.Context, reconciliation *entities.Reconciliation) error {
	query := `
		UPDATE reconciliations
		SET status = $1, notes = $2, updated_at = $3
		WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query,
		reconciliation.Status,
		reconciliation.Notes,
		reconciliation.UpdatedAt,
		reconciliation.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update reconciliation: %w", err)
	}

	return nil
}

// GetByID retrieves a reconciliation by its ID
func (r *ReconciliationPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Reconciliation, error) {
	query := `
		SELECT id, cash_session_id, organization_id, clinic_id, payment_method, currency,
			reconciled_at, reconciled_by_user_id, expected_amount_cents, actual_amount_cents,
			float_left_cents, deposited_cents, discrepancy_cents, status, notes,
			created_at, updated_at
		FROM reconciliations
		WHERE id = $1`

	var recon entities.Reconciliation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&recon.ID,
		&recon.CashSessionID,
		&recon.OrganizationID,
		&recon.ClinicID,
		&recon.PaymentMethod,
		&recon.Currency,
		&recon.ReconciledAt,
		&recon.ReconciledByUserID,
		&recon.ExpectedAmountCents,
		&recon.ActualAmountCents,
		&recon.FloatLeftCents,
		&recon.DepositedCents,
		&recon.DiscrepancyCents,
		&recon.Status,
		&recon.Notes,
		&recon.CreatedAt,
		&recon.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get reconciliation: %w", err)
	}

	return &recon, nil
}

// GetByCashSessionID retrieves all reconciliations for a cash session
func (r *ReconciliationPostgresRepository) GetByCashSessionID(ctx context.Context, cashSessionID uuid.UUID) ([]*entities.Reconciliation, error) {
	query := `
		SELECT id, cash_session_id, organization_id, clinic_id, payment_method, currency,
			reconciled_at, reconciled_by_user_id, expected_amount_cents, actual_amount_cents,
			float_left_cents, deposited_cents, discrepancy_cents, status, notes,
			created_at, updated_at
		FROM reconciliations
		WHERE cash_session_id = $1
		ORDER BY payment_method, currency`

	rows, err := r.db.QueryContext(ctx, query, cashSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reconciliations by cash session: %w", err)
	}
	defer rows.Close()

	var reconciliations []*entities.Reconciliation
	for rows.Next() {
		var recon entities.Reconciliation
		err := rows.Scan(
			&recon.ID,
			&recon.CashSessionID,
			&recon.OrganizationID,
			&recon.ClinicID,
			&recon.PaymentMethod,
			&recon.Currency,
			&recon.ReconciledAt,
			&recon.ReconciledByUserID,
			&recon.ExpectedAmountCents,
			&recon.ActualAmountCents,
			&recon.FloatLeftCents,
			&recon.DepositedCents,
			&recon.DiscrepancyCents,
			&recon.Status,
			&recon.Notes,
			&recon.CreatedAt,
			&recon.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reconciliation: %w", err)
		}
		reconciliations = append(reconciliations, &recon)
	}

	return reconciliations, nil
}

// Exists checks if a reconciliation exists for specific session, payment method, and currency
func (r *ReconciliationPostgresRepository) Exists(ctx context.Context, cashSessionID uuid.UUID, paymentMethod entities.PaymentMethod, currency entities.Currency) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM reconciliations 
			WHERE cash_session_id = $1 AND payment_method = $2 AND currency = $3
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, cashSessionID, paymentMethod, currency).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check reconciliation existence: %w", err)
	}

	return exists, nil
}

// List retrieves reconciliations with optional filters
func (r *ReconciliationPostgresRepository) List(ctx context.Context, filters repositories.ReconciliationFilters) ([]*entities.Reconciliation, error) {
	query := `
		SELECT id, cash_session_id, organization_id, clinic_id, payment_method, currency,
			reconciled_at, reconciled_by_user_id, expected_amount_cents, actual_amount_cents,
			float_left_cents, deposited_cents, discrepancy_cents, status, notes,
			created_at, updated_at
		FROM reconciliations
		WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	if filters.OrganizationID != nil {
		query += fmt.Sprintf(" AND organization_id = $%d", argPos)
		args = append(args, *filters.OrganizationID)
		argPos++
	}

	if filters.ClinicID != nil {
		query += fmt.Sprintf(" AND clinic_id = $%d", argPos)
		args = append(args, *filters.ClinicID)
		argPos++
	}

	if filters.CashSessionID != nil {
		query += fmt.Sprintf(" AND cash_session_id = $%d", argPos)
		args = append(args, *filters.CashSessionID)
		argPos++
	}

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND reconciled_by_user_id = $%d", argPos)
		args = append(args, *filters.UserID)
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

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND reconciled_at >= $%d", argPos)
		args = append(args, *filters.StartDate)
		argPos++
	}

	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND reconciled_at <= $%d", argPos)
		args = append(args, *filters.EndDate)
		argPos++
	}

	if filters.HasDiscrepancy != nil && *filters.HasDiscrepancy {
		query += " AND discrepancy_cents != 0"
	}

	query += " ORDER BY reconciled_at DESC"

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
		return nil, fmt.Errorf("failed to list reconciliations: %w", err)
	}
	defer rows.Close()

	var reconciliations []*entities.Reconciliation
	for rows.Next() {
		var recon entities.Reconciliation
		err := rows.Scan(
			&recon.ID,
			&recon.CashSessionID,
			&recon.OrganizationID,
			&recon.ClinicID,
			&recon.PaymentMethod,
			&recon.Currency,
			&recon.ReconciledAt,
			&recon.ReconciledByUserID,
			&recon.ExpectedAmountCents,
			&recon.ActualAmountCents,
			&recon.FloatLeftCents,
			&recon.DepositedCents,
			&recon.DiscrepancyCents,
			&recon.Status,
			&recon.Notes,
			&recon.CreatedAt,
			&recon.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reconciliation: %w", err)
		}
		reconciliations = append(reconciliations, &recon)
	}

	return reconciliations, nil
}

// GetDiscrepancies retrieves reconciliations with discrepancies
func (r *ReconciliationPostgresRepository) GetDiscrepancies(ctx context.Context, clinicID uuid.UUID, startDate, endDate time.Time) ([]*entities.Reconciliation, error) {
	hasDiscrepancy := true
	filters := repositories.ReconciliationFilters{
		ClinicID:       &clinicID,
		StartDate:      &startDate,
		EndDate:        &endDate,
		HasDiscrepancy: &hasDiscrepancy,
	}
	return r.List(ctx, filters)
}
