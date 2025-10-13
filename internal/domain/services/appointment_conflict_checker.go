package services

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// AppointmentConflictChecker provides methods to check for appointment conflicts
type AppointmentConflictChecker struct {
	appointmentRepo  repositories.AppointmentRepository
	availabilityRepo repositories.DoctorAvailabilityRepository
}

// NewAppointmentConflictChecker creates a new instance of AppointmentConflictChecker
func NewAppointmentConflictChecker(
	appointmentRepo repositories.AppointmentRepository,
	availabilityRepo repositories.DoctorAvailabilityRepository,
) *AppointmentConflictChecker {
	return &AppointmentConflictChecker{
		appointmentRepo:  appointmentRepo,
		availabilityRepo: availabilityRepo,
	}
}

// CheckForConflicts checks if an appointment has any conflicts
func (acc *AppointmentConflictChecker) CheckForConflicts(
	ctx context.Context,
	appointment *entities.Appointment,
) error {
	// Note: Past appointment time validation removed to allow scheduling/updating past appointments

	// Check if end time is after start time
	if appointment.EndTime.Before(appointment.StartTime) || appointment.EndTime.Equal(appointment.StartTime) {
		return entities.ErrEndTimeBeforeStartTime
	}

	// Check for conflicts only if both doctor and unit are specified
	if appointment.DoctorID != nil && appointment.UnitID != nil {
		hasConflict, err := acc.appointmentRepo.CheckConflict(
			ctx,
			*appointment.DoctorID,
			*appointment.UnitID,
			appointment.StartTime,
			appointment.EndTime,
			&appointment.ID,
		)
		if err != nil {
			return err
		}

		if hasConflict {
			return entities.ErrAppointmentConflict
		}
	}

	// Check doctor availability (only if doctor is specified)
	if appointment.DoctorID != nil {
		isAvailable, err := acc.availabilityRepo.IsAvailable(
			ctx,
			*appointment.DoctorID,
			appointment.StartTime,
			appointment.EndTime,
		)
		if err != nil {
			return err
		}

		if !isAvailable {
			return entities.ErrDoctorNotAvailable
		}
	}

	return nil
}

// GetConflictingAppointments returns appointments that conflict with the given time range
func (acc *AppointmentConflictChecker) GetConflictingAppointments(
	ctx context.Context,
	doctorID, unitID uuid.UUID,
	startTime, endTime time.Time,
	excludeAppointmentID *uuid.UUID,
) ([]*entities.Appointment, error) {
	return acc.appointmentRepo.GetConflictingAppointments(
		ctx,
		doctorID,
		unitID,
		startTime,
		endTime,
		excludeAppointmentID,
	)
}
