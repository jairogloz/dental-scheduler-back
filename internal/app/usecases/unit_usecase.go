package usecases

import (
	"context"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// UnitUseCase handles unit-related business logic
type UnitUseCase struct {
	unitRepo   repositories.UnitRepository
	clinicRepo repositories.ClinicRepository
}

// NewUnitUseCase creates a new instance of UnitUseCase
func NewUnitUseCase(unitRepo repositories.UnitRepository, clinicRepo repositories.ClinicRepository) *UnitUseCase {
	return &UnitUseCase{
		unitRepo:   unitRepo,
		clinicRepo: clinicRepo,
	}
}

// CreateUnit creates a new unit
func (uc *UnitUseCase) CreateUnit(ctx context.Context, req *dto.CreateUnitRequest) (*dto.UnitResponse, error) {
	// Verify clinic exists
	clinicExists, err := uc.clinicRepo.Exists(ctx, req.ClinicID)
	if err != nil {
		return nil, err
	}
	if !clinicExists {
		return nil, entities.ErrClinicNotFound
	}

	unit := req.ToEntity()

	if err := unit.Validate(); err != nil {
		return nil, err
	}

	if err := uc.unitRepo.Create(ctx, unit); err != nil {
		return nil, err
	}

	return dto.ToUnitResponse(unit), nil
}

// GetUnitByID retrieves a unit by its ID
func (uc *UnitUseCase) GetUnitByID(ctx context.Context, id uuid.UUID) (*dto.UnitResponse, error) {
	unit, err := uc.unitRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if unit == nil {
		return nil, entities.ErrUnitNotFound
	}

	return dto.ToUnitResponse(unit), nil
}

// GetAllUnits retrieves all units
func (uc *UnitUseCase) GetAllUnits(ctx context.Context) ([]*dto.UnitResponse, error) {
	units, err := uc.unitRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.UnitResponse, len(units))
	for i, unit := range units {
		responses[i] = dto.ToUnitResponse(unit)
	}

	return responses, nil
}

// GetUnitsByClinicID retrieves all units for a specific clinic
func (uc *UnitUseCase) GetUnitsByClinicID(ctx context.Context, clinicID uuid.UUID) ([]*dto.UnitResponse, error) {
	units, err := uc.unitRepo.GetByClinicID(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.UnitResponse, len(units))
	for i, unit := range units {
		responses[i] = dto.ToUnitResponse(unit)
	}

	return responses, nil
}

// UpdateUnit updates an existing unit
func (uc *UnitUseCase) UpdateUnit(ctx context.Context, id uuid.UUID, req *dto.UpdateUnitRequest) (*dto.UnitResponse, error) {
	existing, err := uc.unitRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, entities.ErrUnitNotFound
	}

	updated := req.ToEntityUpdate(existing)

	if err := updated.Validate(); err != nil {
		return nil, err
	}

	if err := uc.unitRepo.Update(ctx, updated); err != nil {
		return nil, err
	}

	return dto.ToUnitResponse(updated), nil
}

// DeleteUnit deletes a unit by its ID
func (uc *UnitUseCase) DeleteUnit(ctx context.Context, id uuid.UUID) error {
	exists, err := uc.unitRepo.Exists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return entities.ErrUnitNotFound
	}

	return uc.unitRepo.Delete(ctx, id)
}
