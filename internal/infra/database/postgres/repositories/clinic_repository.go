package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// ClinicPostgresRepository implements the ClinicRepository interface
type ClinicPostgresRepository struct {
	db *sql.DB
}

// NewClinicPostgresRepository creates a new instance of ClinicPostgresRepository
func NewClinicPostgresRepository(db *sql.DB) repositories.ClinicRepository {
	return &ClinicPostgresRepository{db: db}
}

// Create creates a new clinic
func (r *ClinicPostgresRepository) Create(ctx context.Context, clinic *entities.Clinic) error {
	query := `
		INSERT INTO clinics (id, name, address, phone, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		clinic.ID,
		clinic.Name,
		clinic.Address,
		clinic.Phone,
		clinic.Timezone,
		clinic.CreatedAt,
		clinic.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create clinic: %w", err)
	}

	return nil
}

// GetByID retrieves a clinic by its ID
func (r *ClinicPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Clinic, error) {
	query := `
		SELECT id, name, address, phone, timezone, created_at, updated_at
		FROM clinics
		WHERE id = $1`

	var clinic entities.Clinic
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&clinic.ID,
		&clinic.Name,
		&clinic.Address,
		&clinic.Phone,
		&clinic.Timezone,
		&clinic.CreatedAt,
		&clinic.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get clinic: %w", err)
	}

	return &clinic, nil
}

// GetAll retrieves all clinics
func (r *ClinicPostgresRepository) GetAll(ctx context.Context) ([]*entities.Clinic, error) {
	query := `
		SELECT id, name, address, phone, timezone, created_at, updated_at
		FROM clinics
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get clinics: %w", err)
	}
	defer rows.Close()

	var clinics []*entities.Clinic
	for rows.Next() {
		var clinic entities.Clinic
		err := rows.Scan(
			&clinic.ID,
			&clinic.Name,
			&clinic.Address,
			&clinic.Phone,
			&clinic.Timezone,
			&clinic.CreatedAt,
			&clinic.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan clinic: %w", err)
		}
		clinics = append(clinics, &clinic)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over clinic rows: %w", err)
	}

	return clinics, nil
}

// Update updates an existing clinic
func (r *ClinicPostgresRepository) Update(ctx context.Context, clinic *entities.Clinic) error {
	query := `
		UPDATE clinics
		SET name = $2, address = $3, phone = $4, timezone = $5, updated_at = $6
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		clinic.ID,
		clinic.Name,
		clinic.Address,
		clinic.Phone,
		clinic.Timezone,
		clinic.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update clinic: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrClinicNotFound
	}

	return nil
}

// Delete deletes a clinic by its ID
func (r *ClinicPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM clinics WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete clinic: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrClinicNotFound
	}

	return nil
}

// Exists checks if a clinic exists by its ID
func (r *ClinicPostgresRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM clinics WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check clinic existence: %w", err)
	}

	return exists, nil
}
