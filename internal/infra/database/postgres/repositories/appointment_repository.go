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
		INSERT INTO appointments (id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, query,
		appointment.ID,
		appointment.PatientID,
		appointment.DoctorID,
		appointment.UnitID,
		appointment.ServiceID,
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
		SELECT id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at
		FROM appointments
		WHERE id = $1`

	var appointment entities.Appointment
	var status string
	var patientID, doctorID, unitID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&appointment.ID,
		&patientID,
		&doctorID,
		&unitID,
		&appointment.ServiceID,
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

	// Convert nullable foreign keys
	if patientID.Valid {
		if parsedID, err := uuid.Parse(patientID.String); err == nil {
			appointment.PatientID = &parsedID
		}
	}
	if doctorID.Valid {
		if parsedID, err := uuid.Parse(doctorID.String); err == nil {
			appointment.DoctorID = &parsedID
		}
	}
	if unitID.Valid {
		if parsedID, err := uuid.Parse(unitID.String); err == nil {
			appointment.UnitID = &parsedID
		}
	}

	appointment.Status = entities.AppointmentStatus(status)
	return &appointment, nil
}

// GetAll retrieves all appointments
func (r *AppointmentPostgresRepository) GetAll(ctx context.Context) ([]*entities.Appointment, error) {
	query := `
		SELECT id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at
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
		SELECT id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at
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
		SELECT id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at
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
		SELECT id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at
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
		SELECT id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at
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
		SELECT id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at
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
		SET patient_id = $2, doctor_id = $3, unit_id = $4, service_id = $5, status = $6, 
		    start_time = $7, end_time = $8, notes = $9, 
		    moved_to_needs_rescheduling_at = $10, rescheduled_to_appointment_id = $11, 
		    cancellation_reason = $12, snoozed_until = $13, updated_at = $14
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		appointment.ID,
		appointment.PatientID,
		appointment.DoctorID,
		appointment.UnitID,
		appointment.ServiceID,
		appointment.Status,
		appointment.StartTime,
		appointment.EndTime,
		appointment.Notes,
		appointment.MovedToNeedsReschedulingAt,
		appointment.RescheduledToAppointmentID,
		appointment.CancellationReason,
		appointment.SnoozedUntil,
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
		SELECT id, patient_id, doctor_id, unit_id, service_id, status, start_time, end_time, notes, created_at, updated_at
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
			&appointment.ServiceID,
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
		LEFT JOIN units u ON a.unit_id = u.id
		LEFT JOIN clinics c ON u.clinic_id = c.id
		LEFT JOIN doctors d ON a.doctor_id = d.id
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN services s ON a.service_id = s.id
		WHERE (c.organization_id = $1 OR (a.unit_id IS NULL AND d.organization_id = $1) OR (a.unit_id IS NULL AND a.doctor_id IS NULL))
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
			a.id, a.patient_id, a.doctor_id, a.unit_id, a.service_id, a.status, 
			a.start_time, a.end_time, a.notes, a.created_at, a.updated_at,
			s.name as service_name,
			p.id, p.first_name, p.last_name, p.phone, p.email, p.first_appointment_id, p.created_at, p.updated_at,
			d.id, d.organization_id, d.user_id, d.name, d.specialty, d.email, d.phone, d.is_active, d.created_at, d.updated_at,
			u.id, u.name, u.description, u.clinic_id, u.created_at, u.updated_at,
			c.id, c.name, c.address, c.phone, c.email, c.timezone, c.organization_id, c.created_at, c.updated_at`

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
		var status string
		var serviceName sql.NullString

		// Nullable appointment foreign keys
		var patientID, doctorID, unitID sql.NullString

		// Nullable patient fields
		var patientIDScan, patientFirstName, patientLastName, patientPhone, patientEmail sql.NullString
		var patientFirstAppointmentID sql.NullString
		var patientCreatedAt, patientUpdatedAt sql.NullTime

		// Nullable doctor fields
		var doctorIDScan, doctorOrgID, doctorUserID, doctorName, doctorSpecialty, doctorEmail, doctorPhone sql.NullString
		var doctorIsActive sql.NullBool
		var doctorCreatedAt, doctorUpdatedAt sql.NullTime

		// Nullable unit fields
		var unitIDScan, unitName, unitDescription, unitClinicID sql.NullString
		var unitCreatedAt, unitUpdatedAt sql.NullTime

		// Nullable clinic fields
		var clinicIDScan, clinicName, clinicAddress, clinicPhone, clinicEmail, clinicTimezone, clinicOrgID sql.NullString
		var clinicCreatedAt, clinicUpdatedAt sql.NullTime

		err := rows.Scan(
			// Appointment fields
			&appointment.ID,
			&patientID,
			&doctorID,
			&unitID,
			&appointment.ServiceID,
			&status,
			&appointment.StartTime,
			&appointment.EndTime,
			&appointment.Notes,
			&appointment.CreatedAt,
			&appointment.UpdatedAt,
			// Service name
			&serviceName,
			// Patient fields
			&patientIDScan,
			&patientFirstName,
			&patientLastName,
			&patientPhone,
			&patientEmail,
			&patientFirstAppointmentID,
			&patientCreatedAt,
			&patientUpdatedAt,
			// Doctor fields
			&doctorIDScan,
			&doctorOrgID,
			&doctorUserID,
			&doctorName,
			&doctorSpecialty,
			&doctorEmail,
			&doctorPhone,
			&doctorIsActive,
			&doctorCreatedAt,
			&doctorUpdatedAt,
			// Unit fields
			&unitIDScan,
			&unitName,
			&unitDescription,
			&unitClinicID,
			&unitCreatedAt,
			&unitUpdatedAt,
			// Clinic fields
			&clinicIDScan,
			&clinicName,
			&clinicAddress,
			&clinicPhone,
			&clinicEmail,
			&clinicTimezone,
			&clinicOrgID,
			&clinicCreatedAt,
			&clinicUpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan appointment with details: %w", err)
		}

		// Set appointment foreign keys
		if patientID.Valid {
			if parsedID, err := uuid.Parse(patientID.String); err == nil {
				appointment.PatientID = &parsedID
			}
		}
		if doctorID.Valid {
			if parsedID, err := uuid.Parse(doctorID.String); err == nil {
				appointment.DoctorID = &parsedID
			}
		}
		if unitID.Valid {
			if parsedID, err := uuid.Parse(unitID.String); err == nil {
				appointment.UnitID = &parsedID
			}
		}

		// Set appointment status
		appointment.Status = entities.AppointmentStatus(status)

		// Convert serviceName to pointer
		var serviceNamePtr *string
		if serviceName.Valid {
			serviceNamePtr = &serviceName.String
		}

		// Build entities only if they have valid data
		var patient *entities.Patient
		if patientIDScan.Valid {
			patient = &entities.Patient{}
			if parsedID, err := uuid.Parse(patientIDScan.String); err == nil {
				patient.ID = parsedID
			}
			if patientFirstName.Valid {
				patient.FirstName = patientFirstName.String
			}
			if patientLastName.Valid {
				patient.LastName = &patientLastName.String
			}
			if patientPhone.Valid {
				patient.Phone = &patientPhone.String
			}
			if patientEmail.Valid {
				patient.Email = &patientEmail.String
			}
			if patientFirstAppointmentID.Valid {
				if parsedID, err := uuid.Parse(patientFirstAppointmentID.String); err == nil {
					patient.FirstAppointmentID = &parsedID
				}
			}
			if patientCreatedAt.Valid {
				patient.CreatedAt = patientCreatedAt.Time
			}
			if patientUpdatedAt.Valid {
				patient.UpdatedAt = patientUpdatedAt.Time
			}
		}

		var doctor *entities.Doctor
		if doctorIDScan.Valid {
			doctor = &entities.Doctor{}
			if parsedID, err := uuid.Parse(doctorIDScan.String); err == nil {
				doctor.ID = parsedID
			}
			if doctorOrgID.Valid {
				if parsedID, err := uuid.Parse(doctorOrgID.String); err == nil {
					doctor.OrganizationID = parsedID
				}
			}
			if doctorUserID.Valid {
				if parsedID, err := uuid.Parse(doctorUserID.String); err == nil {
					doctor.UserID = &parsedID
				}
			}
			if doctorName.Valid {
				doctor.Name = doctorName.String
			}
			if doctorSpecialty.Valid {
				doctor.Specialty = &doctorSpecialty.String
			}
			if doctorEmail.Valid {
				doctor.Email = &doctorEmail.String
			}
			if doctorPhone.Valid {
				doctor.Phone = &doctorPhone.String
			}
			if doctorIsActive.Valid {
				doctor.IsActive = doctorIsActive.Bool
			}
			if doctorCreatedAt.Valid {
				doctor.CreatedAt = doctorCreatedAt.Time
			}
			if doctorUpdatedAt.Valid {
				doctor.UpdatedAt = doctorUpdatedAt.Time
			}
		}

		var unit *entities.Unit
		if unitIDScan.Valid {
			unit = &entities.Unit{}
			if parsedID, err := uuid.Parse(unitIDScan.String); err == nil {
				unit.ID = parsedID
			}
			if unitName.Valid {
				unit.Name = unitName.String
			}
			if unitDescription.Valid {
				unit.Description = &unitDescription.String
			}
			if unitClinicID.Valid {
				if parsedID, err := uuid.Parse(unitClinicID.String); err == nil {
					unit.ClinicID = parsedID
				}
			}
			if unitCreatedAt.Valid {
				unit.CreatedAt = unitCreatedAt.Time
			}
			if unitUpdatedAt.Valid {
				unit.UpdatedAt = unitUpdatedAt.Time
			}
		}

		var clinic *entities.Clinic
		if clinicIDScan.Valid {
			clinic = &entities.Clinic{}
			if parsedID, err := uuid.Parse(clinicIDScan.String); err == nil {
				clinic.ID = parsedID
			}
			if clinicName.Valid {
				clinic.Name = clinicName.String
			}
			if clinicAddress.Valid {
				clinic.Address = &clinicAddress.String
			}
			if clinicPhone.Valid {
				clinic.Phone = &clinicPhone.String
			}
			if clinicEmail.Valid {
				clinic.Email = &clinicEmail.String
			}
			if clinicTimezone.Valid {
				clinic.Timezone = clinicTimezone.String
			}
			if clinicOrgID.Valid {
				if parsedID, err := uuid.Parse(clinicOrgID.String); err == nil {
					clinic.OrganizationID = parsedID
				}
			}
			if clinicCreatedAt.Valid {
				clinic.CreatedAt = clinicCreatedAt.Time
			}
			if clinicUpdatedAt.Valid {
				clinic.UpdatedAt = clinicUpdatedAt.Time
			}
		}

		appointmentWithDetails := &repositories.AppointmentWithDetails{
			Appointment: &appointment,
			Patient:     patient,
			Doctor:      doctor,
			Unit:        unit,
			Clinic:      clinic,
			ServiceName: serviceNamePtr,
		}

		appointments = append(appointments, appointmentWithDetails)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating over appointment rows: %w", err)
	}

	return appointments, totalCount, nil
}

// GetReschedulingQueue retrieves appointments in rescheduling queue with pagination
func (r *AppointmentPostgresRepository) GetReschedulingQueue(ctx context.Context, filters repositories.ReschedulingQueueFilters) ([]*repositories.AppointmentWithDetails, int, error) {
	// Base query with JOINs
	baseQuery := `
		FROM appointments a
		LEFT JOIN patients p ON a.patient_id = p.id
		LEFT JOIN doctors d ON a.doctor_id = d.id
		LEFT JOIN units u ON a.unit_id = u.id
		LEFT JOIN clinics c ON u.clinic_id = c.id
		LEFT JOIN services s ON a.service_id = s.id`

	// WHERE conditions
	whereConditions := ` WHERE a.status = 'needs-rescheduling' AND c.organization_id = $1 AND (a.snoozed_until IS NULL OR a.snoozed_until < NOW())`
	params := []interface{}{filters.OrganizationID}
	paramCount := 1

	// Add optional filters
	if filters.ClinicID != nil {
		paramCount++
		whereConditions += fmt.Sprintf(" AND c.id = $%d", paramCount)
		params = append(params, *filters.ClinicID)
	}

	if filters.DoctorID != nil {
		paramCount++
		whereConditions += fmt.Sprintf(" AND a.doctor_id = $%d", paramCount)
		params = append(params, *filters.DoctorID)
	}

	// Add search filter (patient name, phone, or email)
	if filters.Search != "" {
		paramCount++
		searchPattern := "%" + filters.Search + "%"
		whereConditions += fmt.Sprintf(" AND (LOWER(p.first_name || ' ' || COALESCE(p.last_name, '')) LIKE LOWER($%d) OR LOWER(p.phone) LIKE LOWER($%d) OR LOWER(p.email) LIKE LOWER($%d))", paramCount, paramCount, paramCount)
		params = append(params, searchPattern)
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + baseQuery + whereConditions
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, params...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count rescheduling queue appointments: %w", err)
	}

	// Main query with all fields
	selectFields := `
		SELECT 
			a.id, a.patient_id, a.doctor_id, a.unit_id, a.service_id, a.status, 
			a.start_time, a.end_time, a.notes, a.moved_to_needs_rescheduling_at,
			a.rescheduled_to_appointment_id, a.cancellation_reason, a.created_at, a.updated_at,
			s.name as service_name,
			p.id, p.first_name, p.last_name, p.phone, p.email, p.first_appointment_id, p.created_at, p.updated_at,
			d.id, d.organization_id, d.user_id, d.name, d.specialty, d.email, d.phone, d.is_active, d.created_at, d.updated_at,
			u.id, u.name, u.description, u.clinic_id, u.created_at, u.updated_at,
			c.id, c.name, c.address, c.phone, c.email, c.timezone, c.organization_id, c.created_at, c.updated_at`

	// Sort order
	orderBy := " ORDER BY a.moved_to_needs_rescheduling_at"
	if filters.SortOldest {
		orderBy += " ASC" // Oldest first
	} else {
		orderBy += " DESC" // Newest first
	}

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
		return nil, 0, fmt.Errorf("failed to query rescheduling queue: %w", err)
	}
	defer rows.Close()

	var appointments []*repositories.AppointmentWithDetails

	for rows.Next() {
		var appointment entities.Appointment
		var status string
		var serviceName sql.NullString

		// Nullable appointment foreign keys
		var patientID, doctorID, unitID sql.NullString

		// New queue tracking fields
		var movedToNeedsReschedulingAt sql.NullTime
		var rescheduledToAppointmentID sql.NullString
		var cancellationReason sql.NullString

		// Nullable patient fields
		var patientIDScan, patientFirstName, patientLastName, patientPhone, patientEmail sql.NullString
		var patientFirstAppointmentID sql.NullString
		var patientCreatedAt, patientUpdatedAt sql.NullTime

		// Nullable doctor fields
		var doctorIDScan, doctorOrgID, doctorUserID, doctorName, doctorSpecialty, doctorEmail, doctorPhone sql.NullString
		var doctorIsActive sql.NullBool
		var doctorCreatedAt, doctorUpdatedAt sql.NullTime

		// Nullable unit fields
		var unitIDScan, unitName, unitDescription, unitClinicID sql.NullString
		var unitCreatedAt, unitUpdatedAt sql.NullTime

		// Nullable clinic fields
		var clinicIDScan, clinicName, clinicAddress, clinicPhone, clinicEmail, clinicTimezone, clinicOrgID sql.NullString
		var clinicCreatedAt, clinicUpdatedAt sql.NullTime

		err := rows.Scan(
			// Appointment fields
			&appointment.ID,
			&patientID,
			&doctorID,
			&unitID,
			&appointment.ServiceID,
			&status,
			&appointment.StartTime,
			&appointment.EndTime,
			&appointment.Notes,
			&movedToNeedsReschedulingAt,
			&rescheduledToAppointmentID,
			&cancellationReason,
			&appointment.CreatedAt,
			&appointment.UpdatedAt,
			// Service name
			&serviceName,
			// Patient fields
			&patientIDScan,
			&patientFirstName,
			&patientLastName,
			&patientPhone,
			&patientEmail,
			&patientFirstAppointmentID,
			&patientCreatedAt,
			&patientUpdatedAt,
			// Doctor fields
			&doctorIDScan,
			&doctorOrgID,
			&doctorUserID,
			&doctorName,
			&doctorSpecialty,
			&doctorEmail,
			&doctorPhone,
			&doctorIsActive,
			&doctorCreatedAt,
			&doctorUpdatedAt,
			// Unit fields
			&unitIDScan,
			&unitName,
			&unitDescription,
			&unitClinicID,
			&unitCreatedAt,
			&unitUpdatedAt,
			// Clinic fields
			&clinicIDScan,
			&clinicName,
			&clinicAddress,
			&clinicPhone,
			&clinicEmail,
			&clinicTimezone,
			&clinicOrgID,
			&clinicCreatedAt,
			&clinicUpdatedAt,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan appointment row: %w", err)
		}

		// Convert nullable foreign keys
		if patientID.Valid {
			if parsedID, err := uuid.Parse(patientID.String); err == nil {
				appointment.PatientID = &parsedID
			}
		}
		if doctorID.Valid {
			if parsedID, err := uuid.Parse(doctorID.String); err == nil {
				appointment.DoctorID = &parsedID
			}
		}
		if unitID.Valid {
			if parsedID, err := uuid.Parse(unitID.String); err == nil {
				appointment.UnitID = &parsedID
			}
		}

		// Convert queue tracking fields
		if movedToNeedsReschedulingAt.Valid {
			appointment.MovedToNeedsReschedulingAt = &movedToNeedsReschedulingAt.Time
		}
		if rescheduledToAppointmentID.Valid {
			if parsedID, err := uuid.Parse(rescheduledToAppointmentID.String); err == nil {
				appointment.RescheduledToAppointmentID = &parsedID
			}
		}
		if cancellationReason.Valid {
			appointment.CancellationReason = &cancellationReason.String
		}

		appointment.Status = entities.AppointmentStatus(status)

		var serviceNamePtr *string
		if serviceName.Valid {
			serviceNamePtr = &serviceName.String
		}

		// Build patient object
		var patient *entities.Patient
		if patientIDScan.Valid {
			patient = &entities.Patient{}
			if parsedID, err := uuid.Parse(patientIDScan.String); err == nil {
				patient.ID = parsedID
			}
			if patientFirstName.Valid {
				patient.FirstName = patientFirstName.String
			}
			if patientLastName.Valid {
				patient.LastName = &patientLastName.String
			}
			if patientPhone.Valid {
				patient.Phone = &patientPhone.String
			}
			if patientEmail.Valid {
				patient.Email = &patientEmail.String
			}
			if patientFirstAppointmentID.Valid {
				if parsedID, err := uuid.Parse(patientFirstAppointmentID.String); err == nil {
					patient.FirstAppointmentID = &parsedID
				}
			}
			if patientCreatedAt.Valid {
				patient.CreatedAt = patientCreatedAt.Time
			}
			if patientUpdatedAt.Valid {
				patient.UpdatedAt = patientUpdatedAt.Time
			}
		}

		// Build doctor object
		var doctor *entities.Doctor
		if doctorIDScan.Valid {
			doctor = &entities.Doctor{}
			if parsedID, err := uuid.Parse(doctorIDScan.String); err == nil {
				doctor.ID = parsedID
			}
			if doctorOrgID.Valid {
				if parsedID, err := uuid.Parse(doctorOrgID.String); err == nil {
					doctor.OrganizationID = parsedID
				}
			}
			if doctorUserID.Valid {
				if parsedID, err := uuid.Parse(doctorUserID.String); err == nil {
					doctor.UserID = &parsedID
				}
			}
			if doctorName.Valid {
				doctor.Name = doctorName.String
			}
			if doctorSpecialty.Valid {
				doctor.Specialty = &doctorSpecialty.String
			}
			if doctorEmail.Valid {
				doctor.Email = &doctorEmail.String
			}
			if doctorPhone.Valid {
				doctor.Phone = &doctorPhone.String
			}
			if doctorIsActive.Valid {
				doctor.IsActive = doctorIsActive.Bool
			}
			if doctorCreatedAt.Valid {
				doctor.CreatedAt = doctorCreatedAt.Time
			}
			if doctorUpdatedAt.Valid {
				doctor.UpdatedAt = doctorUpdatedAt.Time
			}
		}

		// Build unit object
		var unit *entities.Unit
		if unitIDScan.Valid {
			unit = &entities.Unit{}
			if parsedID, err := uuid.Parse(unitIDScan.String); err == nil {
				unit.ID = parsedID
			}
			if unitName.Valid {
				unit.Name = unitName.String
			}
			if unitDescription.Valid {
				unit.Description = &unitDescription.String
			}
			if unitClinicID.Valid {
				if parsedID, err := uuid.Parse(unitClinicID.String); err == nil {
					unit.ClinicID = parsedID
				}
			}
			if unitCreatedAt.Valid {
				unit.CreatedAt = unitCreatedAt.Time
			}
			if unitUpdatedAt.Valid {
				unit.UpdatedAt = unitUpdatedAt.Time
			}
		}

		// Build clinic object
		var clinic *entities.Clinic
		if clinicIDScan.Valid {
			clinic = &entities.Clinic{}
			if parsedID, err := uuid.Parse(clinicIDScan.String); err == nil {
				clinic.ID = parsedID
			}
			if clinicName.Valid {
				clinic.Name = clinicName.String
			}
			if clinicAddress.Valid {
				clinic.Address = &clinicAddress.String
			}
			if clinicPhone.Valid {
				clinic.Phone = &clinicPhone.String
			}
			if clinicEmail.Valid {
				clinic.Email = &clinicEmail.String
			}
			if clinicTimezone.Valid {
				clinic.Timezone = clinicTimezone.String
			}
			if clinicOrgID.Valid {
				if parsedID, err := uuid.Parse(clinicOrgID.String); err == nil {
					clinic.OrganizationID = parsedID
				}
			}
			if clinicCreatedAt.Valid {
				clinic.CreatedAt = clinicCreatedAt.Time
			}
			if clinicUpdatedAt.Valid {
				clinic.UpdatedAt = clinicUpdatedAt.Time
			}
		}

		appointmentWithDetails := &repositories.AppointmentWithDetails{
			Appointment: &appointment,
			Patient:     patient,
			Doctor:      doctor,
			Unit:        unit,
			Clinic:      clinic,
			ServiceName: serviceNamePtr,
		}

		appointments = append(appointments, appointmentWithDetails)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating over rescheduling queue rows: %w", err)
	}

	return appointments, totalCount, nil
}

// CancelWithReason cancels an appointment and stores the cancellation reason
func (r *AppointmentPostgresRepository) CancelWithReason(ctx context.Context, appointmentID uuid.UUID, reason string) error {
	query := `
		UPDATE appointments
		SET status = 'cancelled',
		    cancellation_reason = $1,
		    updated_at = NOW()
		WHERE id = $2 AND status = 'needs-rescheduling'`

	result, err := r.db.ExecContext(ctx, query, reason, appointmentID)
	if err != nil {
		return fmt.Errorf("failed to cancel appointment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAppointmentNotInQueue
	}

	return nil
}

// SnoozeAppointment temporarily hides an appointment from the rescheduling queue until specified time
func (r *AppointmentPostgresRepository) SnoozeAppointment(ctx context.Context, appointmentID uuid.UUID, until time.Time) error {
	query := `
		UPDATE appointments
		SET snoozed_until = $1,
		    updated_at = NOW()
		WHERE id = $2 AND status = 'needs-rescheduling'`

	result, err := r.db.ExecContext(ctx, query, until, appointmentID)
	if err != nil {
		return fmt.Errorf("failed to snooze appointment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return entities.ErrAppointmentNotInQueue
	}

	return nil
}
