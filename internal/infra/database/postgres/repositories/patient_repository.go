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
		INSERT INTO patients (id, first_name, last_name, email, phone, date_of_birth, medical_history, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		patient.ID,
		patient.FirstName,
		patient.LastName,
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
		SELECT id, first_name, last_name, email, phone, date_of_birth, medical_history, first_appointment_id, created_at, updated_at
		FROM patients
		WHERE id = $1`

	var patient entities.Patient
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&patient.ID,
		&patient.FirstName,
		&patient.LastName,
		&patient.Email,
		&patient.Phone,
		&patient.DateOfBirth,
		&patient.MedicalHistory,
		&patient.FirstAppointmentID,
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
		SELECT id, first_name, last_name, email, phone, date_of_birth, medical_history, first_appointment_id, created_at, updated_at
		FROM patients
		ORDER BY first_name, last_name`

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
			&patient.FirstName,
			&patient.LastName,
			&patient.Email,
			&patient.Phone,
			&patient.DateOfBirth,
			&patient.MedicalHistory,
			&patient.FirstAppointmentID,
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
		SELECT id, first_name, last_name, email, phone, date_of_birth, medical_history, first_appointment_id, created_at, updated_at
		FROM patients
		WHERE email = $1`

	var patient entities.Patient
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&patient.ID,
		&patient.FirstName,
		&patient.LastName,
		&patient.Email,
		&patient.Phone,
		&patient.DateOfBirth,
		&patient.MedicalHistory,
		&patient.FirstAppointmentID,
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
		SET first_name = $2, last_name = $3, email = $4, phone = $5, date_of_birth = $6, medical_history = $7, updated_at = $8
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		patient.ID,
		patient.FirstName,
		patient.LastName,
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

// SearchPatients searches for patients by name, phone, or email within an organization
func (r *PatientPostgresRepository) SearchPatients(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]*entities.Patient, error) {
	searchQuery := `
		SELECT DISTINCT p.id, p.first_name, p.last_name, p.email, p.phone, p.date_of_birth, p.medical_history, p.first_appointment_id, p.created_at, p.updated_at
		FROM patients p
		INNER JOIN patient_organizations po ON p.id = po.patient_id
		WHERE po.organization_id = $1
		AND (
			LOWER(COALESCE(p.first_name, '')) LIKE LOWER($2) OR
			LOWER(COALESCE(p.last_name, '')) LIKE LOWER($2) OR
			LOWER(CONCAT(p.first_name, ' ', COALESCE(p.last_name, ''))) LIKE LOWER($2) OR
			LOWER(COALESCE(p.phone, '')) LIKE LOWER($2) OR
			LOWER(COALESCE(p.email, '')) LIKE LOWER($2)
		)
		ORDER BY p.first_name, p.last_name
		LIMIT $3`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, searchQuery, orgID, searchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search patients: %w", err)
	}
	defer rows.Close()

	var patients []*entities.Patient
	for rows.Next() {
		var patient entities.Patient
		err := rows.Scan(
			&patient.ID,
			&patient.FirstName,
			&patient.LastName,
			&patient.Email,
			&patient.Phone,
			&patient.DateOfBirth,
			&patient.MedicalHistory,
			&patient.FirstAppointmentID,
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

// AddPatientToOrganization links a patient to an organization
func (r *PatientPostgresRepository) AddPatientToOrganization(ctx context.Context, patientID, orgID uuid.UUID) error {
	query := `
		INSERT INTO patient_organizations (patient_id, organization_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (patient_id, organization_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, patientID, orgID)
	if err != nil {
		return fmt.Errorf("failed to add patient to organization: %w", err)
	}

	return nil
}

// OrganizationExists checks if an organization exists by its ID
func (r *PatientPostgresRepository) OrganizationExists(ctx context.Context, orgID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}

	return exists, nil
}

// CreatePatientWithOrganization creates a patient and links to organization in a transaction
func (r *PatientPostgresRepository) CreatePatientWithOrganization(ctx context.Context, patient *entities.Patient, orgID uuid.UUID) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create patient
	patientQuery := `
		INSERT INTO patients (id, first_name, last_name, email, phone, date_of_birth, medical_history, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = tx.ExecContext(ctx, patientQuery,
		patient.ID,
		patient.FirstName,
		patient.LastName,
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

	// Link patient to organization
	linkQuery := `
		INSERT INTO patient_organizations (patient_id, organization_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())`

	_, err = tx.ExecContext(ctx, linkQuery, patient.ID, orgID)
	if err != nil {
		return fmt.Errorf("failed to link patient to organization: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateFirstAppointmentIfNil sets the patient's first_appointment_id if it's currently NULL
func (r *PatientPostgresRepository) UpdateFirstAppointmentIfNil(ctx context.Context, patientID uuid.UUID, appointmentID uuid.UUID) error {
	query := `
		UPDATE patients
		SET first_appointment_id = $1
		WHERE id = $2 AND first_appointment_id IS NULL`

	_, err := r.db.ExecContext(ctx, query, appointmentID, patientID)
	if err != nil {
		return fmt.Errorf("failed to update first_appointment_id: %w", err)
	}

	return nil
}
