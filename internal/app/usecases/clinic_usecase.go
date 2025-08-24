package usecases

import (
	"context"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// ClinicUseCase handles clinic-related business logic
type ClinicUseCase struct {
	clinicRepo repositories.ClinicRepository
}

// NewClinicUseCase creates a new instance of ClinicUseCase
func NewClinicUseCase(clinicRepo repositories.ClinicRepository) *ClinicUseCase {
	return &ClinicUseCase{
		clinicRepo: clinicRepo,
	}
}

// CreateClinic creates a new clinic
func (uc *ClinicUseCase) CreateClinic(ctx context.Context, req *dto.CreateClinicRequest) (*dto.ClinicResponse, error) {
	clinic := req.ToEntity()

	if err := clinic.Validate(); err != nil {
		return nil, err
	}

	if err := uc.clinicRepo.Create(ctx, clinic); err != nil {
		return nil, err
	}

	return dto.ToClinicResponse(clinic), nil
}

// GetClinicByID retrieves a clinic by its ID
func (uc *ClinicUseCase) GetClinicByID(ctx context.Context, id uuid.UUID) (*dto.ClinicResponse, error) {
	clinic, err := uc.clinicRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if clinic == nil {
		return nil, entities.ErrClinicNotFound
	}

	return dto.ToClinicResponse(clinic), nil
}

// GetAllClinics retrieves all clinics
func (uc *ClinicUseCase) GetAllClinics(ctx context.Context) ([]*dto.ClinicResponse, error) {
	clinics, err := uc.clinicRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ClinicResponse, len(clinics))
	for i, clinic := range clinics {
		responses[i] = dto.ToClinicResponse(clinic)
	}

	return responses, nil
}

// UpdateClinic updates an existing clinic
func (uc *ClinicUseCase) UpdateClinic(ctx context.Context, id uuid.UUID, req *dto.UpdateClinicRequest) (*dto.ClinicResponse, error) {
	existing, err := uc.clinicRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, entities.ErrClinicNotFound
	}

	updated := req.ToEntityUpdate(existing)

	if err := updated.Validate(); err != nil {
		return nil, err
	}

	if err := uc.clinicRepo.Update(ctx, updated); err != nil {
		return nil, err
	}

	return dto.ToClinicResponse(updated), nil
}

// DeleteClinic deletes a clinic by its ID
func (uc *ClinicUseCase) DeleteClinic(ctx context.Context, id uuid.UUID) error {
	exists, err := uc.clinicRepo.Exists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return entities.ErrClinicNotFound
	}

	return uc.clinicRepo.Delete(ctx, id)
}
