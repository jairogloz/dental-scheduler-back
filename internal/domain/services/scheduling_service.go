package services

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// SchedulingService provides scheduling-related business logic
type SchedulingService struct {
	appointmentRepo  repositories.AppointmentRepository
	availabilityRepo repositories.DoctorAvailabilityRepository
	doctorRepo       repositories.DoctorRepository
	unitRepo         repositories.UnitRepository
	conflictChecker  *AppointmentConflictChecker
}

// NewSchedulingService creates a new instance of SchedulingService
func NewSchedulingService(
	appointmentRepo repositories.AppointmentRepository,
	availabilityRepo repositories.DoctorAvailabilityRepository,
	doctorRepo repositories.DoctorRepository,
	unitRepo repositories.UnitRepository,
	conflictChecker *AppointmentConflictChecker,
) *SchedulingService {
	return &SchedulingService{
		appointmentRepo:  appointmentRepo,
		availabilityRepo: availabilityRepo,
		doctorRepo:       doctorRepo,
		unitRepo:         unitRepo,
		conflictChecker:  conflictChecker,
	}
}

// ScheduleAppointment schedules a new appointment with conflict checking
func (ss *SchedulingService) ScheduleAppointment(
	ctx context.Context,
	appointment *entities.Appointment,
) error {
	// Validate the appointment
	if err := appointment.Validate(); err != nil {
		return err
	}

	// Check for conflicts
	if err := ss.conflictChecker.CheckForConflicts(ctx, appointment); err != nil {
		return err
	}

	// Verify that the doctor exists (if specified)
	if appointment.DoctorID != nil {
		doctorExists, err := ss.doctorRepo.Exists(ctx, *appointment.DoctorID)
		if err != nil {
			return err
		}
		if !doctorExists {
			return entities.ErrDoctorNotFound
		}
	}

	// Verify that the unit exists (if specified)
	if appointment.UnitID != nil {
		unitExists, err := ss.unitRepo.Exists(ctx, *appointment.UnitID)
		if err != nil {
			return err
		}
		if !unitExists {
			return entities.ErrUnitNotFound
		}
	}

	// Create the appointment
	return ss.appointmentRepo.Create(ctx, appointment)
}

// RescheduleAppointment reschedules an existing appointment
func (ss *SchedulingService) RescheduleAppointment(
	ctx context.Context,
	appointmentID uuid.UUID,
	newStartTime, newEndTime time.Time,
) error {
	// Get the existing appointment
	appointment, err := ss.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return err
	}
	if appointment == nil {
		return entities.ErrAppointmentNotFound
	}

	// Update the times
	appointment.StartTime = newStartTime
	appointment.EndTime = newEndTime
	appointment.UpdatedAt = time.Now()

	// Validate the updated appointment
	if err := appointment.Validate(); err != nil {
		return err
	}

	// Check for conflicts (excluding the current appointment)
	if err := ss.conflictChecker.CheckForConflicts(ctx, appointment); err != nil {
		return err
	}

	// Update the appointment
	return ss.appointmentRepo.Update(ctx, appointment)
}

// GetAvailableSlots returns available time slots for a doctor on a specific date
func (ss *SchedulingService) GetAvailableSlots(
	ctx context.Context,
	doctorID uuid.UUID,
	date time.Time,
	slotDuration time.Duration,
) ([]time.Time, error) {
	// Get doctor's availability for the date
	availability, err := ss.availabilityRepo.GetByDoctorIDAndDate(ctx, doctorID, date)
	if err != nil {
		return nil, err
	}

	// Get existing appointments for the date
	appointments, err := ss.appointmentRepo.GetByDoctorIDAndDate(ctx, doctorID, date)
	if err != nil {
		return nil, err
	}

	var availableSlots []time.Time

	// For each availability period, calculate available slots
	for _, avail := range availability {
		if !avail.IsAvailable {
			continue
		}

		// Generate slots within this availability period
		slots := ss.generateSlotsInPeriod(avail.StartTime, avail.EndTime, slotDuration, appointments)
		availableSlots = append(availableSlots, slots...)
	}

	return availableSlots, nil
}

// generateSlotsInPeriod generates available time slots within a period, excluding conflicts
func (ss *SchedulingService) generateSlotsInPeriod(
	startTime, endTime time.Time,
	slotDuration time.Duration,
	existingAppointments []*entities.Appointment,
) []time.Time {
	var slots []time.Time

	current := startTime
	for current.Add(slotDuration).Before(endTime) || current.Add(slotDuration).Equal(endTime) {
		slotEnd := current.Add(slotDuration)

		// Check if this slot conflicts with any existing appointment
		hasConflict := false
		for _, apt := range existingAppointments {
			if apt.IsScheduled() && ss.timesOverlap(current, slotEnd, apt.StartTime, apt.EndTime) {
				hasConflict = true
				break
			}
		}

		if !hasConflict {
			slots = append(slots, current)
		}

		current = current.Add(slotDuration)
	}

	return slots
}

// timesOverlap checks if two time ranges overlap
func (ss *SchedulingService) timesOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && end1.After(start2)
}
