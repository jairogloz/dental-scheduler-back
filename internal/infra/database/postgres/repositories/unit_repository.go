package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// UnitPostgresRepository implements the UnitRepository interface
type UnitPostgresRepository struct {
	db *sql.DB
}

// NewUnitPostgresRepository creates a new instance of UnitPostgresRepository
func NewUnitPostgresRepository(db *sql.DB) repositories.UnitRepository {
	return &UnitPostgresRepository{db: db}
}

// Create creates a new unit
func (r *UnitPostgresRepository) Create(ctx context.Context, unit *entities.Unit) error {
	query := `
		INSERT INTO units (id, clinic_id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		unit.ID,
		unit.ClinicID,
		unit.Name,
		unit.Description,
		unit.IsActive,
		unit.CreatedAt,
		unit.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create unit: %w", err)
	}

	return nil
}

// GetByID retrieves a unit by its ID
func (r *UnitPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Unit, error) {
	query := `
		SELECT id, clinic_id, name, description, is_active, created_at, updated_at
		FROM units
		WHERE id = $1`

	var unit entities.Unit
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&unit.ID,
		&unit.ClinicID,
		&unit.Name,
		&unit.Description,
		&unit.IsActive,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	return &unit, nil
}

// GetAll retrieves all units
func (r *UnitPostgresRepository) GetAll(ctx context.Context) ([]*entities.Unit, error) {
	query := `
		SELECT id, clinic_id, name, description, is_active, created_at, updated_at
		FROM units
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get units: %w", err)
	}
	defer rows.Close()

	var units []*entities.Unit
	for rows.Next() {
		var unit entities.Unit
		err := rows.Scan(
			&unit.ID,
			&unit.ClinicID,
			&unit.Name,
			&unit.Description,
			&unit.IsActive,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, &unit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over unit rows: %w", err)
	}

	return units, nil
}

// GetByClinicID retrieves all units for a specific clinic
func (r *UnitPostgresRepository) GetByClinicID(ctx context.Context, clinicID uuid.UUID) ([]*entities.Unit, error) {
	query := `
		SELECT id, clinic_id, name, description, is_active, created_at, updated_at
		FROM units
		WHERE clinic_id = $1
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, clinicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get units by clinic ID: %w", err)
	}
	defer rows.Close()

	var units []*entities.Unit
	for rows.Next() {
		var unit entities.Unit
		err := rows.Scan(
			&unit.ID,
			&unit.ClinicID,
			&unit.Name,
			&unit.Description,
			&unit.IsActive,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, &unit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over unit rows: %w", err)
	}

	return units, nil
}

// Update updates an existing unit
func (r *UnitPostgresRepository) Update(ctx context.Context, unit *entities.Unit) error {
	query := `
		UPDATE units
		SET name = $2, description = $3, is_active = $4, updated_at = $5
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		unit.ID,
		unit.Name,
		unit.Description,
		unit.IsActive,
		unit.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update unit: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrUnitNotFound
	}

	return nil
}

// Delete deletes a unit by its ID
func (r *UnitPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM units WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete unit: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrUnitNotFound
	}

	return nil
}

// Exists checks if a unit exists by its ID
func (r *UnitPostgresRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM units WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check unit existence: %w", err)
	}

	return exists, nil
}
