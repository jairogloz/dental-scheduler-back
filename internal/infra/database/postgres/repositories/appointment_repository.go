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
