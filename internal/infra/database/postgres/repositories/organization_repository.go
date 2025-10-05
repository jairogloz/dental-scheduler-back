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

// OrganizationPostgresRepository implements the OrganizationRepository interface
type OrganizationPostgresRepository struct {
	db *sql.DB
}

// NewOrganizationPostgresRepository creates a new instance of OrganizationPostgresRepository
func NewOrganizationPostgresRepository(db *sql.DB) repositories.OrganizationRepository {
	return &OrganizationPostgresRepository{db: db}
}

// GetByID retrieves an organization by its ID
func (r *OrganizationPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	query := `
		SELECT id, name, description, address, phone, email, website, is_active, created_at, updated_at
		FROM organizations
		WHERE id = $1`

	var org entities.Organization
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Description,
		&org.Address,
		&org.Phone,
		&org.Email,
		&org.Website,
		&org.IsActive,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return &org, nil
}

// Exists checks if an organization exists by its ID
func (r *OrganizationPostgresRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}

	return exists, nil
}

// GetOrganizationData retrieves complete organization data for calendar loading
func (r *OrganizationPostgresRepository) GetOrganizationData(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, limit int) (*repositories.OrganizationData, error) {
	// Get organization
	org, err := r.GetByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, entities.ErrOrganizationNotFound
	}

	// Get clinics for this organization
	clinics, err := r.getClinicsByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get clinics: %w", err)
	}

	// Get units for this organization
	units, err := r.getUnitsByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get units: %w", err)
	}

	// Get doctors for this organization
	doctors, err := r.getDoctorsByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctors: %w", err)
	}

	// Get appointments for this organization (excluding cancelled)
	appointments, err := r.getAppointmentsByOrganization(ctx, orgID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments: %w", err)
	}

	// Get services for this organization
	services, err := r.getServicesByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	return &repositories.OrganizationData{
		Organization: org,
		Clinics:      clinics,
		Units:        units,
		Doctors:      doctors,
		Appointments: appointments,
		Services:     services,
	}, nil
}

// getClinicsByOrganization retrieves all clinics for an organization
func (r *OrganizationPostgresRepository) getClinicsByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entities.Clinic, error) {
	query := `
		SELECT id, organization_id, name, address, phone, created_at, updated_at
		FROM clinics
		WHERE organization_id = $1
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clinics []*entities.Clinic
	for rows.Next() {
		var clinic entities.Clinic
		err := rows.Scan(
			&clinic.ID,
			&clinic.OrganizationID,
			&clinic.Name,
			&clinic.Address,
			&clinic.Phone,
			&clinic.CreatedAt,
			&clinic.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		clinics = append(clinics, &clinic)
	}

	return clinics, rows.Err()
}

// getUnitsByOrganization retrieves all units for an organization
func (r *OrganizationPostgresRepository) getUnitsByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entities.Unit, error) {
	query := `
		SELECT u.id, u.clinic_id, u.name, u.description, u.is_active, u.created_at, u.updated_at
		FROM units u
		INNER JOIN clinics c ON u.clinic_id = c.id
		WHERE c.organization_id = $1
		ORDER BY c.name, u.name`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		units = append(units, &unit)
	}

	return units, rows.Err()
}

// getDoctorsByOrganization retrieves all doctors for an organization
func (r *OrganizationPostgresRepository) getDoctorsByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entities.Doctor, error) {
	query := `
		SELECT id, organization_id, name, specialty, email, phone, default_unit_id, is_active, created_at, updated_at
		FROM doctors
		WHERE organization_id = $1
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []*entities.Doctor
	for rows.Next() {
		var doctor entities.Doctor
		err := rows.Scan(
			&doctor.ID,
			&doctor.OrganizationID,
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
			return nil, err
		}
		doctors = append(doctors, &doctor)
	}

	return doctors, rows.Err()
}

// getAppointmentsByOrganization retrieves appointments for calendar view (excluding cancelled)
func (r *OrganizationPostgresRepository) getAppointmentsByOrganization(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, limit int) ([]*repositories.AppointmentCalendarData, error) {
	query := `
		SELECT DISTINCT 
			a.id, 
			a.patient_id,
			CONCAT(p.first_name, ' ', COALESCE(p.last_name, '')) as patient_name,
			p.phone as patient_phone,
			a.doctor_id,
			c.id as clinic_id,
			a.unit_id,
			a.start_time,
			a.end_time,
			a.status,
			a.service_id,
			s.name as service_name,
			CASE 
				WHEN p.first_appointment_id = a.id THEN true
				ELSE false
			END as is_first_visit
		FROM appointments a
		INNER JOIN units u ON a.unit_id = u.id
		INNER JOIN clinics c ON u.clinic_id = c.id
		INNER JOIN patients p ON a.patient_id = p.id
		LEFT JOIN services s ON a.service_id = s.id
		WHERE c.organization_id = $1
		AND a.start_time >= $2
		AND a.start_time <= $3
		AND a.status != 'cancelled'
		ORDER BY a.start_time
		LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, orgID, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []*repositories.AppointmentCalendarData
	for rows.Next() {
		var appt repositories.AppointmentCalendarData
		err := rows.Scan(
			&appt.ID,
			&appt.PatientID,
			&appt.PatientName,
			&appt.PatientPhone,
			&appt.DoctorID,
			&appt.ClinicID,
			&appt.UnitID,
			&appt.StartTime,
			&appt.EndTime,
			&appt.Status,
			&appt.ServiceID,
			&appt.ServiceName,
			&appt.IsFirstVisit,
		)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, &appt)
	}

	return appointments, rows.Err()
}

// getServicesByOrganization retrieves all services for an organization
func (r *OrganizationPostgresRepository) getServicesByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entities.Service, error) {
	query := `
		SELECT id, name, base_price, organization_id, created_at, updated_at
		FROM services
		WHERE organization_id = $1
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*entities.Service
	for rows.Next() {
		var service entities.Service
		err := rows.Scan(
			&service.ID,
			&service.Name,
			&service.BasePrice,
			&service.OrganizationID,
			&service.CreatedAt,
			&service.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		services = append(services, &service)
	}

	return services, rows.Err()
}
