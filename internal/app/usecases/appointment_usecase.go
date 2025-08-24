package usecases

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/domain/services"

	"github.com/google/uuid"
)

// AppointmentUseCase handles appointment-related business logic
type AppointmentUseCase struct {
	appointmentRepo   repositories.AppointmentRepository
	patientRepo       repositories.PatientRepository
	doctorRepo        repositories.DoctorRepository
	unitRepo          repositories.UnitRepository
	schedulingService *services.SchedulingService
}

// NewAppointmentUseCase creates a new instance of AppointmentUseCase
func NewAppointmentUseCase(
	appointmentRepo repositories.AppointmentRepository,
	patientRepo repositories.PatientRepository,
	doctorRepo repositories.DoctorRepository,
	unitRepo repositories.UnitRepository,
	schedulingService *services.SchedulingService,
) *AppointmentUseCase {
	return &AppointmentUseCase{
		appointmentRepo:   appointmentRepo,
		patientRepo:       patientRepo,
		doctorRepo:        doctorRepo,
		unitRepo:          unitRepo,
		schedulingService: schedulingService,
	}
}

// CreateAppointment creates a new appointment with conflict checking
func (uc *AppointmentUseCase) CreateAppointment(ctx context.Context, req *dto.CreateAppointmentRequest) (*dto.AppointmentResponse, error) {
	// Verify patient exists
	patientExists, err := uc.patientRepo.Exists(ctx, req.PatientID)
	if err != nil {
		return nil, err
	}
	if !patientExists {
		return nil, entities.ErrPatientNotFound
	}

	appointment := req.ToEntity()

	// Use scheduling service to create appointment with conflict checking
	if err := uc.schedulingService.ScheduleAppointment(ctx, appointment); err != nil {
		return nil, err
	}

	return dto.ToAppointmentResponse(appointment), nil
}

// GetAppointmentByID retrieves an appointment by its ID
func (uc *AppointmentUseCase) GetAppointmentByID(ctx context.Context, id uuid.UUID) (*dto.AppointmentResponse, error) {
	appointment, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if appointment == nil {
		return nil, entities.ErrAppointmentNotFound
	}

	return dto.ToAppointmentResponse(appointment), nil
}

// GetAllAppointments retrieves all appointments
func (uc *AppointmentUseCase) GetAllAppointments(ctx context.Context) ([]*dto.AppointmentResponse, error) {
	appointments, err := uc.appointmentRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		responses[i] = dto.ToAppointmentResponse(appointment)
	}

	return responses, nil
}

// GetUpcomingAppointments retrieves all upcoming appointments
func (uc *AppointmentUseCase) GetUpcomingAppointments(ctx context.Context) ([]*dto.AppointmentResponse, error) {
	appointments, err := uc.appointmentRepo.GetUpcoming(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		responses[i] = dto.ToAppointmentResponse(appointment)
	}

	return responses, nil
}

// UpdateAppointment updates an existing appointment
func (uc *AppointmentUseCase) UpdateAppointment(ctx context.Context, id uuid.UUID, req *dto.UpdateAppointmentRequest) (*dto.AppointmentResponse, error) {
	existing, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, entities.ErrAppointmentNotFound
	}

	// Verify patient exists
	patientExists, err := uc.patientRepo.Exists(ctx, req.PatientID)
	if err != nil {
		return nil, err
	}
	if !patientExists {
		return nil, entities.ErrPatientNotFound
	}

	// Verify doctor exists
	doctorExists, err := uc.doctorRepo.Exists(ctx, req.DoctorID)
	if err != nil {
		return nil, err
	}
	if !doctorExists {
		return nil, entities.ErrDoctorNotFound
	}

	// Verify unit exists
	unitExists, err := uc.unitRepo.Exists(ctx, req.UnitID)
	if err != nil {
		return nil, err
	}
	if !unitExists {
		return nil, entities.ErrUnitNotFound
	}

	updated := req.ToEntityUpdate(existing)

	if err := updated.Validate(); err != nil {
		return nil, err
	}

	// Check for conflicts if time has changed
	if !updated.StartTime.Equal(existing.StartTime) || !updated.EndTime.Equal(existing.EndTime) {
		hasConflict, err := uc.appointmentRepo.CheckConflict(
			ctx,
			updated.DoctorID,
			updated.UnitID,
			updated.StartTime,
			updated.EndTime,
			&updated.ID,
		)
		if err != nil {
			return nil, err
		}
		if hasConflict {
			return nil, entities.ErrAppointmentConflict
		}
	}

	if err := uc.appointmentRepo.Update(ctx, updated); err != nil {
		return nil, err
	}

	return dto.ToAppointmentResponse(updated), nil
}

// RescheduleAppointment reschedules an existing appointment
func (uc *AppointmentUseCase) RescheduleAppointment(ctx context.Context, id uuid.UUID, req *dto.RescheduleAppointmentRequest) (*dto.AppointmentResponse, error) {
	if err := uc.schedulingService.RescheduleAppointment(ctx, id, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	// Get the updated appointment
	appointment, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.ToAppointmentResponse(appointment), nil
}

// CancelAppointment cancels an appointment
func (uc *AppointmentUseCase) CancelAppointment(ctx context.Context, id uuid.UUID) error {
	appointment, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if appointment == nil {
		return entities.ErrAppointmentNotFound
	}

	appointment.Cancel()

	return uc.appointmentRepo.Update(ctx, appointment)
}

// CompleteAppointment marks an appointment as completed
func (uc *AppointmentUseCase) CompleteAppointment(ctx context.Context, id uuid.UUID) error {
	appointment, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if appointment == nil {
		return entities.ErrAppointmentNotFound
	}

	appointment.Complete()

	return uc.appointmentRepo.Update(ctx, appointment)
}

// DeleteAppointment deletes an appointment by its ID
func (uc *AppointmentUseCase) DeleteAppointment(ctx context.Context, id uuid.UUID) error {
	exists, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if exists == nil {
		return entities.ErrAppointmentNotFound
	}

	return uc.appointmentRepo.Delete(ctx, id)
}

// GetAvailableSlots returns available time slots for a doctor on a specific date
func (uc *AppointmentUseCase) GetAvailableSlots(ctx context.Context, doctorID uuid.UUID, date time.Time, slotDurationMinutes int) ([]*dto.AvailableSlotResponse, error) {
	slotDuration := time.Duration(slotDurationMinutes) * time.Minute

	slots, err := uc.schedulingService.GetAvailableSlots(ctx, doctorID, date, slotDuration)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AvailableSlotResponse, len(slots))
	for i, slot := range slots {
		responses[i] = &dto.AvailableSlotResponse{
			StartTime: slot,
			EndTime:   slot.Add(slotDuration),
		}
	}

	return responses, nil
}
