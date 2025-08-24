package usecases

import (
	"context"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// PatientUseCase handles patient-related business logic
type PatientUseCase struct {
	patientRepo repositories.PatientRepository
}

// NewPatientUseCase creates a new instance of PatientUseCase
func NewPatientUseCase(patientRepo repositories.PatientRepository) *PatientUseCase {
	return &PatientUseCase{
		patientRepo: patientRepo,
	}
}

// CreatePatient creates a new patient
func (uc *PatientUseCase) CreatePatient(ctx context.Context, req *dto.CreatePatientRequest) (*dto.PatientResponse, error) {
	patient := req.ToEntity()

	if err := patient.Validate(); err != nil {
		return nil, err
	}

	if err := uc.patientRepo.Create(ctx, patient); err != nil {
		return nil, err
	}

	return dto.ToPatientResponse(patient), nil
}

// GetPatientByID retrieves a patient by its ID
func (uc *PatientUseCase) GetPatientByID(ctx context.Context, id uuid.UUID) (*dto.PatientResponse, error) {
	patient, err := uc.patientRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if patient == nil {
		return nil, entities.ErrPatientNotFound
	}

	return dto.ToPatientResponse(patient), nil
}

// GetAllPatients retrieves all patients
func (uc *PatientUseCase) GetAllPatients(ctx context.Context) ([]*dto.PatientResponse, error) {
	patients, err := uc.patientRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.PatientResponse, len(patients))
	for i, patient := range patients {
		responses[i] = dto.ToPatientResponse(patient)
	}

	return responses, nil
}

// UpdatePatient updates an existing patient
func (uc *PatientUseCase) UpdatePatient(ctx context.Context, id uuid.UUID, req *dto.UpdatePatientRequest) (*dto.PatientResponse, error) {
	existing, err := uc.patientRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, entities.ErrPatientNotFound
	}

	updated := req.ToEntityUpdate(existing)

	if err := updated.Validate(); err != nil {
		return nil, err
	}

	if err := uc.patientRepo.Update(ctx, updated); err != nil {
		return nil, err
	}

	return dto.ToPatientResponse(updated), nil
}

// DeletePatient deletes a patient by its ID
func (uc *PatientUseCase) DeletePatient(ctx context.Context, id uuid.UUID) error {
	exists, err := uc.patientRepo.Exists(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return entities.ErrPatientNotFound
	}

	return uc.patientRepo.Delete(ctx, id)
}
