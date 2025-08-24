package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// DoctorPostgresRepository implements the DoctorRepository interface
type DoctorPostgresRepository struct {
	db *sql.DB
}

// NewDoctorPostgresRepository creates a new instance of DoctorPostgresRepository
func NewDoctorPostgresRepository(db *sql.DB) repositories.DoctorRepository {
	return &DoctorPostgresRepository{db: db}
}

// Create creates a new doctor
func (r *DoctorPostgresRepository) Create(ctx context.Context, doctor *entities.Doctor) error {
	query := `
		INSERT INTO doctors (id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		doctor.ID,
		doctor.Name,
		doctor.Specialty,
		doctor.Email,
		doctor.Phone,
		doctor.DefaultUnitID,
		doctor.IsActive,
		doctor.CreatedAt,
		doctor.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create doctor: %w", err)
	}

	return nil
}

// GetByID retrieves a doctor by its ID
func (r *DoctorPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Doctor, error) {
	query := `
		SELECT id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at
		FROM doctors
		WHERE id = $1`

	var doctor entities.Doctor
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doctor.ID,
		&doctor.Name,
		&doctor.Specialty,
		&doctor.Email,
		&doctor.Phone,
		&doctor.DefaultUnitID,
		&doctor.IsActive,
		&doctor.CreatedAt,
		&doctor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get doctor: %w", err)
	}

	return &doctor, nil
}

// GetAll retrieves all doctors
func (r *DoctorPostgresRepository) GetAll(ctx context.Context) ([]*entities.Doctor, error) {
	query := `
		SELECT id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at
		FROM doctors
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctors: %w", err)
	}
	defer rows.Close()

	var doctors []*entities.Doctor
	for rows.Next() {
		var doctor entities.Doctor
		err := rows.Scan(
			&doctor.ID,
			&doctor.Name,
			&doctor.Specialty,
			&doctor.Email,
			&doctor.Phone,
			&doctor.DefaultUnitID,
			&doctor.IsActive,
			&doctor.CreatedAt,
			&doctor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan doctor: %w", err)
		}
		doctors = append(doctors, &doctor)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over doctor rows: %w", err)
	}

	return doctors, nil
}

// GetByEmail retrieves a doctor by email
func (r *DoctorPostgresRepository) GetByEmail(ctx context.Context, email string) (*entities.Doctor, error) {
	query := `
		SELECT id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at
		FROM doctors
		WHERE email = $1`

	var doctor entities.Doctor
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&doctor.ID,
		&doctor.Name,
		&doctor.Specialty,
		&doctor.Email,
		&doctor.Phone,
		&doctor.DefaultUnitID,
		&doctor.IsActive,
		&doctor.CreatedAt,
		&doctor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get doctor by email: %w", err)
	}

	return &doctor, nil
}

// Update updates an existing doctor
func (r *DoctorPostgresRepository) Update(ctx context.Context, doctor *entities.Doctor) error {
	query := `
		UPDATE doctors
		SET name = $2, specialty = $3, email = $4, phone = $5, default_unit_id = $6, is_active = $7, updated_at = $8
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		doctor.ID,
		doctor.Name,
		doctor.Specialty,
		doctor.Email,
		doctor.Phone,
		doctor.DefaultUnitID,
		doctor.IsActive,
		doctor.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update doctor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrDoctorNotFound
	}

	return nil
}

// Delete deletes a doctor by its ID
func (r *DoctorPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM doctors WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete doctor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrDoctorNotFound
	}

	return nil
}

// Exists checks if a doctor exists by its ID
func (r *DoctorPostgresRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM doctors WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check doctor existence: %w", err)
	}

	return exists, nil
}
