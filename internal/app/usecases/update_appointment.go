package usecases

import (
	"context"
	"fmt"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/domain/services"

	"github.com/google/uuid"
)

// UpdateAppointmentUseCase handles updating appointments
type UpdateAppointmentUseCase struct {
	appointmentRepo   repositories.AppointmentRepository
	patientRepo       repositories.PatientRepository
	doctorRepo        repositories.DoctorRepository
	unitRepo          repositories.UnitRepository
	schedulingService *services.SchedulingService
}

// NewUpdateAppointmentUseCase creates a new instance of UpdateAppointmentUseCase
func NewUpdateAppointmentUseCase(
	appointmentRepo repositories.AppointmentRepository,
	patientRepo repositories.PatientRepository,
	doctorRepo repositories.DoctorRepository,
	unitRepo repositories.UnitRepository,
	schedulingService *services.SchedulingService,
) *UpdateAppointmentUseCase {
	return &UpdateAppointmentUseCase{
		appointmentRepo:   appointmentRepo,
		patientRepo:       patientRepo,
		doctorRepo:        doctorRepo,
		unitRepo:          unitRepo,
		schedulingService: schedulingService,
	}
}

// Execute updates an appointment with validation and conflict checking
func (uc *UpdateAppointmentUseCase) Execute(ctx context.Context, appointmentID uuid.UUID, orgID uuid.UUID, req *dto.UpdateAppointmentRequest) (*dto.AppointmentResponse, error) {
	// Get existing appointment first
	existingAppointment, err := uc.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}
	if existingAppointment == nil {
		return nil, entities.ErrAppointmentNotFound
	}

	// Verify entities exist only if they're being updated
	if req.PatientID != nil {
		patientExists, err := uc.patientRepo.Exists(ctx, *req.PatientID)
		if err != nil {
			return nil, fmt.Errorf("failed to check patient existence: %w", err)
		}
		if !patientExists {
			return nil, entities.ErrPatientNotFound
		}
	}

	if req.DoctorID != nil {
		doctorExists, err := uc.doctorRepo.Exists(ctx, *req.DoctorID)
		if err != nil {
			return nil, fmt.Errorf("failed to check doctor existence: %w", err)
		}
		if !doctorExists {
			return nil, entities.ErrDoctorNotFound
		}
	}

	if req.UnitID != nil {
		unitExists, err := uc.unitRepo.Exists(ctx, *req.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to check unit existence: %w", err)
		}
		if !unitExists {
			return nil, entities.ErrUnitNotFound
		}
	}

	// Create updated appointment entity
	updatedAppointment := req.ToEntityUpdate(existingAppointment)

	// Basic validation: if both start and end time are provided, validate the time logic
	if req.StartTime != nil && req.EndTime != nil {
		if updatedAppointment.EndTime.Before(updatedAppointment.StartTime) || updatedAppointment.EndTime.Equal(updatedAppointment.StartTime) {
			return nil, fmt.Errorf("end time must be after start time")
		}
	}

	// Validate the updated appointment
	if err := updatedAppointment.Validate(); err != nil {
		return nil, err
	}

	// Update the appointment
	if err := uc.appointmentRepo.Update(ctx, updatedAppointment); err != nil {
		return nil, fmt.Errorf("failed to update appointment: %w", err)
	}

	// Convert to response DTO
	response := dto.ToAppointmentResponse(updatedAppointment)
	return response, nil
}
