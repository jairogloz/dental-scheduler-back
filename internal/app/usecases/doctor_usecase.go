package usecases

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// DoctorUseCase handles doctor-related business logic
type DoctorUseCase struct {
	doctorRepo      repositories.DoctorRepository
	unitRepo        repositories.UnitRepository
	appointmentRepo repositories.AppointmentRepository
}

// NewDoctorUseCase creates a new instance of DoctorUseCase
func NewDoctorUseCase(
	doctorRepo repositories.DoctorRepository,
	unitRepo repositories.UnitRepository,
	appointmentRepo repositories.AppointmentRepository,
) *DoctorUseCase {
	return &DoctorUseCase{
		doctorRepo:      doctorRepo,
		unitRepo:        unitRepo,
		appointmentRepo: appointmentRepo,
	}
}

// CreateDoctor creates a new doctor
func (uc *DoctorUseCase) CreateDoctor(ctx context.Context, req *dto.CreateDoctorRequest) (*dto.DoctorResponse, error) {
	// Verify default unit exists if provided
	if req.DefaultUnitID != nil {
		unitExists, err := uc.unitRepo.Exists(ctx, *req.DefaultUnitID)
		if err != nil {
			return nil, err
		}
		if !unitExists {
			return nil, entities.ErrUnitNotFound
		}
	}

	doctor := req.ToEntity()

	if err := doctor.Validate(); err != nil {
		return nil, err
	}

	if err := uc.doctorRepo.Create(ctx, doctor); err != nil {
		return nil, err
	}

	return dto.ToDoctorResponse(doctor), nil
}

// GetDoctorByID retrieves a doctor by its ID
func (uc *DoctorUseCase) GetDoctorByID(ctx context.Context, id uuid.UUID) (*dto.DoctorResponse, error) {
	doctor, err := uc.doctorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if doctor == nil {
		return nil, entities.ErrDoctorNotFound
	}

	return dto.ToDoctorResponse(doctor), nil
}

// GetAllDoctors retrieves all doctors
func (uc *DoctorUseCase) GetAllDoctors(ctx context.Context) ([]*dto.DoctorResponse, error) {
	doctors, err := uc.doctorRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.DoctorResponse, len(doctors))
	for i, doctor := range doctors {
		responses[i] = dto.ToDoctorResponse(doctor)
	}

	return responses, nil
}

// GetDoctorAvailability retrieves a doctor's appointments for a specific date
func (uc *DoctorUseCase) GetDoctorAvailability(ctx context.Context, doctorID uuid.UUID, date time.Time) ([]*dto.AppointmentResponse, error) {
	// Verify doctor exists
	doctorExists, err := uc.doctorRepo.Exists(ctx, doctorID)
	if err != nil {
		return nil, err
	}
	if !doctorExists {
		return nil, entities.ErrDoctorNotFound
	}

	appointments, err := uc.appointmentRepo.GetByDoctorIDAndDate(ctx, doctorID, date)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		responses[i] = dto.ToAppointmentResponse(appointment)
	}

	return responses, nil
}

// UpdateDoctor updates an existing doctor
func (uc *DoctorUseCase) UpdateDoctor(ctx context.Context, id uuid.UUID, req *dto.UpdateDoctorRequest) (*dto.DoctorResponse, error) {
	existing, err := uc.doctorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, entities.ErrDoctorNotFound
	}

	// Verify default unit exists if provided
	if req.DefaultUnitID != nil {
		unitExists, err := uc.unitRepo.Exists(ctx, *req.DefaultUnitID)
		if err != nil {
			return nil, err
		}
		if !unitExists {
			return nil, entities.ErrUnitNotFound
		}
	}

	updated := req.ToEntityUpdate(existing)

	if err := updated.Validate(); err != nil {
		return nil, err
	}

	if err := uc.doctorRepo.Update(ctx, updated); err != nil {
		return nil, err
	}

	return dto.ToDoctorResponse(updated), nil
}

// DeleteDoctor deletes a doctor by its ID
func (uc *DoctorUseCase) DeleteDoctor(ctx context.Context, id uuid.UUID) error {
	exists, err := uc.doctorRepo.Exists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return entities.ErrDoctorNotFound
	}

	return uc.doctorRepo.Delete(ctx, id)
}
