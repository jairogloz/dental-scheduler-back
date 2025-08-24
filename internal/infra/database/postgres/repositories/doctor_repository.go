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
		INSERT INTO doctors (id, organization_id, user_id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		doctor.ID,
		doctor.OrganizationID,
		doctor.UserID,
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
		SELECT id, organization_id, user_id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at
		FROM doctors
		WHERE id = $1`

	var doctor entities.Doctor
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doctor.ID,
		&doctor.OrganizationID,
		&doctor.UserID,
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
		SELECT id, organization_id, user_id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at
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
			&doctor.OrganizationID,
			&doctor.UserID,
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
		SELECT id, organization_id, user_id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at
		FROM doctors
		WHERE email = $1`

	var doctor entities.Doctor
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&doctor.ID,
		&doctor.OrganizationID,
		&doctor.UserID,
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
		SET organization_id = $2, user_id = $3, name = $4, specialty = $5, email = $6, phone = $7, default_unit_id = $8, is_active = $9, updated_at = $10
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		doctor.ID,
		doctor.OrganizationID,
		doctor.UserID,
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

// GetByOrganizationID retrieves doctors by organization ID with clinic info
func (r *DoctorPostgresRepository) GetByOrganizationID(ctx context.Context, orgID uuid.UUID, clinicID *uuid.UUID) ([]*repositories.DoctorWithOrgInfo, error) {
	query := `
		SELECT 
			d.id, d.organization_id, d.user_id, d.name, d.specialty, d.email, d.phone, 
			d.default_unit_id, d.is_active, d.created_at, d.updated_at,
			c.id as clinic_id, c.name as clinic_name, o.name as org_name,
			CASE WHEN $2::UUID IS NOT NULL AND c.id = $2 THEN 0 ELSE 1 END as sort_priority
		FROM doctors d
		JOIN organizations o ON d.organization_id = o.id
		LEFT JOIN units u ON d.default_unit_id = u.id
		LEFT JOIN clinics c ON u.clinic_id = c.id
		WHERE d.organization_id = $1 AND d.is_active = true
		ORDER BY sort_priority, c.name NULLS LAST, d.name ASC`

	rows, err := r.db.QueryContext(ctx, query, orgID, clinicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctors by organization: %w", err)
	}
	defer rows.Close()

	var doctors []*repositories.DoctorWithOrgInfo
	for rows.Next() {
		var doctor entities.Doctor
		var clinicIDPtr *uuid.UUID
		var clinicNamePtr *string
		var orgName string
		var sortPriority int

		err := rows.Scan(
			&doctor.ID,
			&doctor.OrganizationID,
			&doctor.UserID,
			&doctor.Name,
			&doctor.Specialty,
			&doctor.Email,
			&doctor.Phone,
			&doctor.DefaultUnitID,
			&doctor.IsActive,
			&doctor.CreatedAt,
			&doctor.UpdatedAt,
			&clinicIDPtr,
			&clinicNamePtr,
			&orgName,
			&sortPriority,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan doctor: %w", err)
		}

		doctorWithInfo := &repositories.DoctorWithOrgInfo{
			Doctor:            &doctor,
			DefaultClinicID:   clinicIDPtr,
			DefaultClinicName: clinicNamePtr,
			OrgName:           orgName,
		}

		doctors = append(doctors, doctorWithInfo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating doctors: %w", err)
	}

	return doctors, nil
}
