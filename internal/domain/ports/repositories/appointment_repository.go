package repositories

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// AppointmentRepository defines the interface for appointment data operations
type AppointmentRepository interface {
	// Create creates a new appointment
	Create(ctx context.Context, appointment *entities.Appointment) error

	// GetByID retrieves an appointment by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Appointment, error)

	// GetAll retrieves all appointments
	GetAll(ctx context.Context) ([]*entities.Appointment, error)

	// GetByPatientID retrieves all appointments for a patient
	GetByPatientID(ctx context.Context, patientID uuid.UUID) ([]*entities.Appointment, error)

	// GetByDoctorID retrieves all appointments for a doctor
	GetByDoctorID(ctx context.Context, doctorID uuid.UUID) ([]*entities.Appointment, error)

	// GetByUnitID retrieves all appointments for a unit
	GetByUnitID(ctx context.Context, unitID uuid.UUID) ([]*entities.Appointment, error)

	// GetByDoctorIDAndDate retrieves appointments for a doctor on a specific date
	GetByDoctorIDAndDate(ctx context.Context, doctorID uuid.UUID, date time.Time) ([]*entities.Appointment, error)

	// GetUpcoming retrieves all upcoming appointments
	GetUpcoming(ctx context.Context) ([]*entities.Appointment, error)

	// Update updates an existing appointment
	Update(ctx context.Context, appointment *entities.Appointment) error

	// Delete deletes an appointment by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// CheckConflict checks if an appointment conflicts with existing appointments
	CheckConflict(ctx context.Context, doctorID, unitID uuid.UUID, startTime, endTime time.Time, excludeAppointmentID *uuid.UUID) (bool, error)

	// GetConflictingAppointments returns appointments that conflict with the given time range
	GetConflictingAppointments(ctx context.Context, doctorID, unitID uuid.UUID, startTime, endTime time.Time, excludeAppointmentID *uuid.UUID) ([]*entities.Appointment, error)
}
