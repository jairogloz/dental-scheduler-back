package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// PatientPostgresRepository implements the PatientRepository interface
type PatientPostgresRepository struct {
	db *sql.DB
}

// NewPatientPostgresRepository creates a new instance of PatientPostgresRepository
func NewPatientPostgresRepository(db *sql.DB) repositories.PatientRepository {
	return &PatientPostgresRepository{db: db}
}

// Create creates a new patient
func (r *PatientPostgresRepository) Create(ctx context.Context, patient *entities.Patient) error {
	query := `
		INSERT INTO patients (id, name, email, phone, date_of_birth, medical_history, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		patient.ID,
		patient.Name,
		patient.Email,
		patient.Phone,
		patient.DateOfBirth,
		patient.MedicalHistory,
		patient.CreatedAt,
		patient.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create patient: %w", err)
	}

	return nil
}

// GetByID retrieves a patient by its ID
func (r *PatientPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Patient, error) {
	query := `
		SELECT id, name, email, phone, date_of_birth, medical_history, created_at, updated_at
		FROM patients
		WHERE id = $1`

	var patient entities.Patient
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&patient.ID,
		&patient.Name,
		&patient.Email,
		&patient.Phone,
		&patient.DateOfBirth,
		&patient.MedicalHistory,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get patient: %w", err)
	}

	return &patient, nil
}

// GetAll retrieves all patients
func (r *PatientPostgresRepository) GetAll(ctx context.Context) ([]*entities.Patient, error) {
	query := `
		SELECT id, name, email, phone, date_of_birth, medical_history, created_at, updated_at
		FROM patients
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get patients: %w", err)
	}
	defer rows.Close()

	var patients []*entities.Patient
	for rows.Next() {
		var patient entities.Patient
		err := rows.Scan(
			&patient.ID,
			&patient.Name,
			&patient.Email,
			&patient.Phone,
			&patient.DateOfBirth,
			&patient.MedicalHistory,
			&patient.CreatedAt,
			&patient.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan patient: %w", err)
		}
		patients = append(patients, &patient)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over patient rows: %w", err)
	}

	return patients, nil
}

// GetByEmail retrieves a patient by email
func (r *PatientPostgresRepository) GetByEmail(ctx context.Context, email string) (*entities.Patient, error) {
	query := `
		SELECT id, name, email, phone, date_of_birth, medical_history, created_at, updated_at
		FROM patients
		WHERE email = $1`

	var patient entities.Patient
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&patient.ID,
		&patient.Name,
		&patient.Email,
		&patient.Phone,
		&patient.DateOfBirth,
		&patient.MedicalHistory,
		&patient.CreatedAt,
		&patient.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get patient by email: %w", err)
	}

	return &patient, nil
}

// Update updates an existing patient
func (r *PatientPostgresRepository) Update(ctx context.Context, patient *entities.Patient) error {
	query := `
		UPDATE patients
		SET name = $2, email = $3, phone = $4, date_of_birth = $5, medical_history = $6, updated_at = $7
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		patient.ID,
		patient.Name,
		patient.Email,
		patient.Phone,
		patient.DateOfBirth,
		patient.MedicalHistory,
		patient.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update patient: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrPatientNotFound
	}

	return nil
}

// Delete deletes a patient by its ID
func (r *PatientPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM patients WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete patient: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrPatientNotFound
	}

	return nil
}

// Exists checks if a patient exists by its ID
func (r *PatientPostgresRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM patients WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check patient existence: %w", err)
	}

	return exists, nil
}
