package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// CashSessionPostgresRepository implements the CashSessionRepository interface
type CashSessionPostgresRepository struct {
	db *sql.DB
}

// NewCashSessionPostgresRepository creates a new instance
func NewCashSessionPostgresRepository(db *sql.DB) repositories.CashSessionRepository {
	return &CashSessionPostgresRepository{db: db}
}

// Create creates a new cash session
func (r *CashSessionPostgresRepository) Create(ctx context.Context, session *entities.CashSession) error {
	query := `
		INSERT INTO cash_sessions (
			id, organization_id, clinic_id, user_id, opened_at, closed_at,
			starting_float_cents, status, opening_type, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.OrganizationID,
		session.ClinicID,
		session.UserID,
		session.OpenedAt,
		session.ClosedAt,
		session.StartingFloatCents,
		session.Status,
		session.OpeningType,
		session.Notes,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create cash session: %w", err)
	}

	return nil
}

// Update updates an existing cash session
func (r *CashSessionPostgresRepository) Update(ctx context.Context, session *entities.CashSession) error {
	query := `
		UPDATE cash_sessions
		SET closed_at = $1, status = $2, notes = $3, updated_at = $4
		WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query,
		session.ClosedAt,
		session.Status,
		session.Notes,
		session.UpdatedAt,
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update cash session: %w", err)
	}

	return nil
}

// GetByID retrieves a cash session by its ID
func (r *CashSessionPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.CashSession, error) {
	query := `
		SELECT id, organization_id, clinic_id, user_id, opened_at, closed_at,
			starting_float_cents, status, opening_type, notes, created_at, updated_at
		FROM cash_sessions
		WHERE id = $1`

	var session entities.CashSession
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.OrganizationID,
		&session.ClinicID,
		&session.UserID,
		&session.OpenedAt,
		&session.ClosedAt,
		&session.StartingFloatCents,
		&session.Status,
		&session.OpeningType,
		&session.Notes,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cash session: %w", err)
	}

	return &session, nil
}

// GetCurrentOpenSession retrieves the currently open session for a user at a clinic
func (r *CashSessionPostgresRepository) GetCurrentOpenSession(ctx context.Context, userID, clinicID uuid.UUID) (*entities.CashSession, error) {
	query := `
		SELECT id, organization_id, clinic_id, user_id, opened_at, closed_at,
			starting_float_cents, status, opening_type, notes, created_at, updated_at
		FROM cash_sessions
		WHERE user_id = $1 AND clinic_id = $2 AND status = 'open'
		ORDER BY opened_at DESC
		LIMIT 1`

	var session entities.CashSession
	err := r.db.QueryRowContext(ctx, query, userID, clinicID).Scan(
		&session.ID,
		&session.OrganizationID,
		&session.ClinicID,
		&session.UserID,
		&session.OpenedAt,
		&session.ClosedAt,
		&session.StartingFloatCents,
		&session.Status,
		&session.OpeningType,
		&session.Notes,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current open session: %w", err)
	}

	return &session, nil
}

// HasOpenSession checks if a user has an open session at a clinic
func (r *CashSessionPostgresRepository) HasOpenSession(ctx context.Context, userID, clinicID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM cash_sessions 
			WHERE user_id = $1 AND clinic_id = $2 AND status = 'open'
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, clinicID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check open session: %w", err)
	}

	return exists, nil
}

// List retrieves cash sessions with optional filters
func (r *CashSessionPostgresRepository) List(ctx context.Context, filters repositories.CashSessionFilters) ([]*entities.CashSession, error) {
	query := `
		SELECT id, organization_id, clinic_id, user_id, opened_at, closed_at,
			starting_float_cents, status, opening_type, notes, created_at, updated_at
		FROM cash_sessions
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

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *filters.UserID)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND opened_at >= $%d", argPos)
		args = append(args, *filters.StartDate)
		argPos++
	}

	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND opened_at <= $%d", argPos)
		args = append(args, *filters.EndDate)
		argPos++
	}

	query += " ORDER BY opened_at DESC"

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
		return nil, fmt.Errorf("failed to list cash sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*entities.CashSession
	for rows.Next() {
		var session entities.CashSession
		err := rows.Scan(
			&session.ID,
			&session.OrganizationID,
			&session.ClinicID,
			&session.UserID,
			&session.OpenedAt,
			&session.ClosedAt,
			&session.StartingFloatCents,
			&session.Status,
			&session.OpeningType,
			&session.Notes,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cash session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// Close closes a cash session
func (r *CashSessionPostgresRepository) Close(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE cash_sessions
		SET status = 'closed', closed_at = NOW(), updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to close cash session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrCashSessionNotFound
	}

	return nil
}
