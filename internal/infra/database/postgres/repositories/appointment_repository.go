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

// AppointmentPostgresRepository implements the AppointmentRepository interface
type AppointmentPostgresRepository struct {
	db *sql.DB
}

// NewAppointmentPostgresRepository creates a new instance of AppointmentPostgresRepository
func NewAppointmentPostgresRepository(db *sql.DB) repositories.AppointmentRepository {
	return &AppointmentPostgresRepository{db: db}
}

// Create creates a new appointment
func (r *AppointmentPostgresRepository) Create(ctx context.Context, appointment *entities.Appointment) error {
	query := `
		INSERT INTO appointments (id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		appointment.ID,
		appointment.PatientID,
		appointment.DoctorID,
		appointment.UnitID,
		appointment.TreatmentType,
		appointment.Status,
		appointment.StartTime,
		appointment.EndTime,
		appointment.Notes,
		appointment.CreatedAt,
		appointment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create appointment: %w", err)
	}

	return nil
}

// GetByID retrieves an appointment by its ID
func (r *AppointmentPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		WHERE id = $1`

	var appointment entities.Appointment
	var status string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&appointment.ID,
		&appointment.PatientID,
		&appointment.DoctorID,
		&appointment.UnitID,
		&appointment.TreatmentType,
		&status,
		&appointment.StartTime,
		&appointment.EndTime,
		&appointment.Notes,
		&appointment.CreatedAt,
		&appointment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	appointment.Status = entities.AppointmentStatus(status)
	return &appointment, nil
}

// GetAll retrieves all appointments
func (r *AppointmentPostgresRepository) GetAll(ctx context.Context) ([]*entities.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments: %w", err)
	}
	defer rows.Close()

	return r.scanAppointments(rows)
}

// GetByPatientID retrieves all appointments for a patient
func (r *AppointmentPostgresRepository) GetByPatientID(ctx context.Context, patientID uuid.UUID) ([]*entities.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		WHERE patient_id = $1
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query, patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments by patient ID: %w", err)
	}
	defer rows.Close()

	return r.scanAppointments(rows)
}

// GetByDoctorID retrieves all appointments for a doctor
func (r *AppointmentPostgresRepository) GetByDoctorID(ctx context.Context, doctorID uuid.UUID) ([]*entities.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		WHERE doctor_id = $1
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query, doctorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments by doctor ID: %w", err)
	}
	defer rows.Close()

	return r.scanAppointments(rows)
}

// GetByUnitID retrieves all appointments for a unit
func (r *AppointmentPostgresRepository) GetByUnitID(ctx context.Context, unitID uuid.UUID) ([]*entities.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		WHERE unit_id = $1
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments by unit ID: %w", err)
	}
	defer rows.Close()

	return r.scanAppointments(rows)
}

// GetByDoctorIDAndDate retrieves appointments for a doctor on a specific date
func (r *AppointmentPostgresRepository) GetByDoctorIDAndDate(ctx context.Context, doctorID uuid.UUID, date time.Time) ([]*entities.Appointment, error) {
	// Get the start and end of the day
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		WHERE doctor_id = $1 AND start_time >= $2 AND start_time < $3
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query, doctorID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments by doctor ID and date: %w", err)
	}
	defer rows.Close()

	return r.scanAppointments(rows)
}

// GetUpcoming retrieves all upcoming appointments
func (r *AppointmentPostgresRepository) GetUpcoming(ctx context.Context) ([]*entities.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		WHERE start_time > NOW() AND status = 'scheduled'
		ORDER BY start_time`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming appointments: %w", err)
	}
	defer rows.Close()

	return r.scanAppointments(rows)
}

// Update updates an existing appointment
func (r *AppointmentPostgresRepository) Update(ctx context.Context, appointment *entities.Appointment) error {
	query := `
		UPDATE appointments
		SET patient_id = $2, doctor_id = $3, unit_id = $4, treatment_type = $5, status = $6, start_time = $7, end_time = $8, notes = $9, updated_at = $10
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		appointment.ID,
		appointment.PatientID,
		appointment.DoctorID,
		appointment.UnitID,
		appointment.TreatmentType,
		appointment.Status,
		appointment.StartTime,
		appointment.EndTime,
		appointment.Notes,
		appointment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update appointment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAppointmentNotFound
	}

	return nil
}

// Delete deletes an appointment by its ID
func (r *AppointmentPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM appointments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete appointment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAppointmentNotFound
	}

	return nil
}

// CheckConflict checks if an appointment conflicts with existing appointments
func (r *AppointmentPostgresRepository) CheckConflict(ctx context.Context, doctorID, unitID uuid.UUID, startTime, endTime time.Time, excludeAppointmentID *uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM appointments
		WHERE status = 'scheduled'
		  AND (doctor_id = $1 OR unit_id = $2)
		  AND start_time < $4
		  AND end_time > $3`

	args := []interface{}{doctorID, unitID, startTime, endTime}

	if excludeAppointmentID != nil {
		query += " AND id != $5"
		args = append(args, *excludeAppointmentID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check appointment conflict: %w", err)
	}

	return count > 0, nil
}

// GetConflictingAppointments returns appointments that conflict with the given time range
func (r *AppointmentPostgresRepository) GetConflictingAppointments(ctx context.Context, doctorID, unitID uuid.UUID, startTime, endTime time.Time, excludeAppointmentID *uuid.UUID) ([]*entities.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, unit_id, treatment_type, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		WHERE status = 'scheduled'
		  AND (doctor_id = $1 OR unit_id = $2)
		  AND start_time < $4
		  AND end_time > $3`

	args := []interface{}{doctorID, unitID, startTime, endTime}

	if excludeAppointmentID != nil {
		query += " AND id != $5"
		args = append(args, *excludeAppointmentID)
	}

	query += " ORDER BY start_time"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflicting appointments: %w", err)
	}
	defer rows.Close()

	return r.scanAppointments(rows)
}

// scanAppointments is a helper method to scan multiple appointment rows
func (r *AppointmentPostgresRepository) scanAppointments(rows *sql.Rows) ([]*entities.Appointment, error) {
	var appointments []*entities.Appointment
	for rows.Next() {
		var appointment entities.Appointment
		var status string
		err := rows.Scan(
			&appointment.ID,
			&appointment.PatientID,
			&appointment.DoctorID,
			&appointment.UnitID,
			&appointment.TreatmentType,
			&status,
			&appointment.StartTime,
			&appointment.EndTime,
			&appointment.Notes,
			&appointment.CreatedAt,
			&appointment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan appointment: %w", err)
		}
		appointment.Status = entities.AppointmentStatus(status)
		appointments = append(appointments, &appointment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over appointment rows: %w", err)
	}

	return appointments, nil
}

// GetByOrganizationAndDateRange retrieves appointments for an organization within a date range with filters
func (r *AppointmentPostgresRepository) GetByOrganizationAndDateRange(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, filters repositories.AppointmentFilters) ([]*repositories.AppointmentWithDetails, int, error) {
	// Build the base query with joins
	baseQuery := `
		FROM appointments a
		INNER JOIN units u ON a.unit_id = u.id
		INNER JOIN clinics c ON u.clinic_id = c.id
		INNER JOIN doctors d ON a.doctor_id = d.id
		INNER JOIN patients p ON a.patient_id = p.id
		WHERE c.organization_id = $1 
		AND a.start_time >= $2 
		AND a.start_time < $3`

	// Build WHERE conditions and parameters
	params := []interface{}{orgID, startDate, endDate.AddDate(0, 0, 1)} // Add 1 day to include the end date
	paramIndex := 4

	whereConditions := ""

	if filters.ClinicID != nil {
		whereConditions += fmt.Sprintf(" AND c.id = $%d", paramIndex)
		params = append(params, *filters.ClinicID)
		paramIndex++
	}

	if filters.DoctorID != nil {
		whereConditions += fmt.Sprintf(" AND d.id = $%d", paramIndex)
		params = append(params, *filters.DoctorID)
		paramIndex++
	}

	if filters.Status != nil {
		whereConditions += fmt.Sprintf(" AND a.status = $%d", paramIndex)
		params = append(params, string(*filters.Status))
		paramIndex++
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + baseQuery + whereConditions

	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, params...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count appointments: %w", err)
	}

	// Main query with all fields
	selectFields := `
		SELECT 
			a.id, a.patient_id, a.doctor_id, a.unit_id, a.treatment_type, a.status, 
			a.start_time, a.end_time, a.notes, a.created_at, a.updated_at,
			p.id, p.first_name, p.last_name, p.phone, p.email, p.created_at, p.updated_at,
			d.id, d.organization_id, d.user_id, d.name, d.specialty, d.email, d.phone, d.is_active, d.created_at, d.updated_at,
			u.id, u.name, u.description, u.clinic_id, u.created_at, u.updated_at,
			c.id, c.name, c.address, c.phone, c.email, c.organization_id, c.created_at, c.updated_at`

	orderBy := " ORDER BY a.start_time ASC"

	// Add pagination
	if filters.Limit > 0 {
		offset := 0
		if filters.Page > 1 {
			offset = (filters.Page - 1) * filters.Limit
		}
		orderBy += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.Limit, offset)
	}

	fullQuery := selectFields + " " + baseQuery + whereConditions + orderBy

	rows, err := r.db.QueryContext(ctx, fullQuery, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query appointments: %w", err)
	}
	defer rows.Close()

	var appointments []*repositories.AppointmentWithDetails

	for rows.Next() {
		var appointment entities.Appointment
		var patient entities.Patient
		var doctor entities.Doctor
		var unit entities.Unit
		var clinic entities.Clinic
		var status string

		err := rows.Scan(
			// Appointment fields
			&appointment.ID,
			&appointment.PatientID,
			&appointment.DoctorID,
			&appointment.UnitID,
			&appointment.TreatmentType,
			&status,
			&appointment.StartTime,
			&appointment.EndTime,
			&appointment.Notes,
			&appointment.CreatedAt,
			&appointment.UpdatedAt,
			// Patient fields
			&patient.ID,
			&patient.FirstName,
			&patient.LastName,
			&patient.Phone,
			&patient.Email,
			&patient.CreatedAt,
			&patient.UpdatedAt,
			// Doctor fields
			&doctor.ID,
			&doctor.OrganizationID,
			&doctor.UserID,
			&doctor.Name,
			&doctor.Specialty,
			&doctor.Email,
			&doctor.Phone,
			&doctor.IsActive,
			&doctor.CreatedAt,
			&doctor.UpdatedAt,
			// Unit fields
			&unit.ID,
			&unit.Name,
			&unit.Description,
			&unit.ClinicID,
			&unit.CreatedAt,
			&unit.UpdatedAt,
			// Clinic fields
			&clinic.ID,
			&clinic.Name,
			&clinic.Address,
			&clinic.Phone,
			&clinic.Email,
			&clinic.OrganizationID,
			&clinic.CreatedAt,
			&clinic.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan appointment with details: %w", err)
		}

		// Set appointment status
		appointment.Status = entities.AppointmentStatus(status)

		appointmentWithDetails := &repositories.AppointmentWithDetails{
			Appointment: &appointment,
			Patient:     &patient,
			Doctor:      &doctor,
			Unit:        &unit,
			Clinic:      &clinic,
		}

		appointments = append(appointments, appointmentWithDetails)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating over appointment rows: %w", err)
	}

	return appointments, totalCount, nil
}
