package entities

import (
	"time"

	"github.com/google/uuid"
)

// AppointmentStatus represents the status of an appointment
type AppointmentStatus string

const (
	AppointmentStatusScheduled           AppointmentStatus = "scheduled"
	AppointmentStatusConfirmed           AppointmentStatus = "confirmed"
	AppointmentStatusCompleted           AppointmentStatus = "completed"
	AppointmentStatusCancelled           AppointmentStatus = "cancelled"
	AppointmentStatusRescheduled         AppointmentStatus = "rescheduled"
	AppointmentStatusRescheduleRequested AppointmentStatus = "reschedule-requested"
	AppointmentStatusNoShow              AppointmentStatus = "no-show"
	AppointmentStatusWithError           AppointmentStatus = "with-error"
)

// Appointment represents an appointment entity
type Appointment struct {
	ID                uuid.UUID         `json:"id" db:"id"`
	PatientID         *uuid.UUID        `json:"patient_id,omitempty" db:"patient_id"`
	DoctorID          *uuid.UUID        `json:"doctor_id,omitempty" db:"doctor_id"`
	UnitID            *uuid.UUID        `json:"unit_id,omitempty" db:"unit_id"`
	ServiceID         *string           `json:"service_id,omitempty" db:"service_id"`
	Status            AppointmentStatus `json:"status" db:"status"`
	StartTime         time.Time         `json:"start_time" db:"start_time"`
	EndTime           time.Time         `json:"end_time" db:"end_time"`
	Notes             *string           `json:"notes,omitempty" db:"notes"`
	MigrationSourceID *string           `json:"migration_source_id,omitempty" db:"migration_source_id"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}

// Validate checks if the appointment entity is valid
func (a *Appointment) Validate() error {
	// Note: PatientID, DoctorID, and UnitID can now be nil for incomplete appointments
	// Individual validation can be performed at the business logic level when needed

	if a.StartTime.IsZero() || a.EndTime.IsZero() {
		return ErrInvalidAppointmentTime
	}

	if a.EndTime.Before(a.StartTime) || a.EndTime.Equal(a.StartTime) {
		return ErrEndTimeBeforeStartTime
	}

	if !IsValidAppointmentStatus(a.Status) {
		return ErrInvalidAppointmentStatus
	}

	// Note: Past appointment time validation removed to allow scheduling/updating past appointments

	return nil
}

// IsValid checks if the appointment has valid data
func (a *Appointment) IsValid() bool {
	return a.Validate() == nil
}

// Duration returns the duration of the appointment
func (a *Appointment) Duration() time.Duration {
	return a.EndTime.Sub(a.StartTime)
}

// IsScheduled checks if the appointment is scheduled
func (a *Appointment) IsScheduled() bool {
	return a.Status == AppointmentStatusScheduled
}

// IsCompleted checks if the appointment is completed
func (a *Appointment) IsCompleted() bool {
	return a.Status == AppointmentStatusCompleted
}

// IsCancelled checks if the appointment is cancelled
func (a *Appointment) IsCancelled() bool {
	return a.Status == AppointmentStatusCancelled
}

// IsRescheduled checks if the appointment is rescheduled
func (a *Appointment) IsRescheduled() bool {
	return a.Status == AppointmentStatusRescheduled
}

// IsValidStatus checks if the provided status is valid
func IsValidAppointmentStatus(status AppointmentStatus) bool {
	switch status {
	case AppointmentStatusScheduled,
		AppointmentStatusConfirmed,
		AppointmentStatusCompleted,
		AppointmentStatusCancelled,
		AppointmentStatusRescheduled,
		AppointmentStatusRescheduleRequested,
		AppointmentStatusNoShow,
		AppointmentStatusWithError:
		return true
	default:
		return false
	}
}

// Cancel cancels the appointment
func (a *Appointment) Cancel() {
	a.Status = AppointmentStatusCancelled
	a.UpdatedAt = time.Now()
}

// Complete marks the appointment as completed
func (a *Appointment) Complete() {
	a.Status = AppointmentStatusCompleted
	a.UpdatedAt = time.Now()
}

// Reschedule marks the appointment as rescheduled
func (a *Appointment) Reschedule() {
	a.Status = AppointmentStatusRescheduled
	a.UpdatedAt = time.Now()
}
