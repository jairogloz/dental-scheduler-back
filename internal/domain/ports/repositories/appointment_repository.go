package repositories

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// AppointmentFilters represents filters for appointment queries
type AppointmentFilters struct {
	ClinicID *uuid.UUID
	DoctorID *uuid.UUID
	Status   *entities.AppointmentStatus
	Page     int
	Limit    int
}

// ReschedulingQueueFilters represents filters for rescheduling queue queries
type ReschedulingQueueFilters struct {
	OrganizationID uuid.UUID
	ClinicID       *uuid.UUID
	DoctorID       *uuid.UUID
	Search         string // Search in patient name, phone, email
	Page           int
	Limit          int
	SortOldest     bool // true = ASC (oldest first), false = DESC (newest first)
}

// AppointmentWithDetails represents an appointment with all related entity details
type AppointmentWithDetails struct {
	Appointment *entities.Appointment
	Patient     *entities.Patient
	Doctor      *entities.Doctor
	Unit        *entities.Unit
	Clinic      *entities.Clinic
	ServiceName *string
}

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

	// GetByOrganizationAndDateRange retrieves appointments for an organization within a date range with filters
	GetByOrganizationAndDateRange(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, filters AppointmentFilters) ([]*AppointmentWithDetails, int, error)

	// GetReschedulingQueue retrieves appointments in rescheduling queue with pagination
	GetReschedulingQueue(ctx context.Context, filters ReschedulingQueueFilters) ([]*AppointmentWithDetails, int, error)

	// CancelWithReason cancels an appointment and stores the cancellation reason
	CancelWithReason(ctx context.Context, appointmentID uuid.UUID, reason string) error
}
