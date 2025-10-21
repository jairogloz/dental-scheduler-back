package usecases

import (
	"context"
	"fmt"

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

// CreatePatientWithOrganization creates a new patient and optionally links to organization
func (uc *PatientUseCase) CreatePatientWithOrganization(ctx context.Context, req *dto.CreatePatientWithOrgRequest) (*dto.PatientResponse, error) {
	// Parse organization ID if provided
	orgID, err := req.GetOrganizationID()
	if err != nil {
		return nil, fmt.Errorf("invalid organization_id format: %w", err)
	}

	// Validate organization exists if provided
	if orgID != nil {
		exists, err := uc.patientRepo.OrganizationExists(ctx, *orgID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, entities.ErrOrganizationNotFound
		}
	}

	// Create patient entity
	patient := req.CreatePatientRequest.ToEntity()

	if err := patient.Validate(); err != nil {
		return nil, err
	}

	// If organization ID is provided, use transactional creation
	if orgID != nil {
		if err := uc.patientRepo.CreatePatientWithOrganization(ctx, patient, *orgID); err != nil {
			return nil, err
		}
	} else {
		// Create patient only
		if err := uc.patientRepo.Create(ctx, patient); err != nil {
			return nil, err
		}
	}

	return dto.ToPatientResponse(patient), nil
}

// CreatePatientInOrganization creates a new patient and links to organization
func (uc *PatientUseCase) CreatePatientInOrganization(ctx context.Context, req *dto.CreatePatientRequest, orgID uuid.UUID) (*dto.PatientResponse, error) {
	// Validate organization exists
	exists, err := uc.patientRepo.OrganizationExists(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, entities.ErrOrganizationNotFound
	}

	// Create patient entity
	patient := req.ToEntity()

	if err := patient.Validate(); err != nil {
		return nil, err
	}

	// Create patient with organization link in transaction
	if err := uc.patientRepo.CreatePatientWithOrganization(ctx, patient, orgID); err != nil {
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
func (uc *PatientUseCase) UpdatePatient(ctx context.Context, id uuid.UUID, orgID uuid.UUID, req *dto.UpdatePatientRequest) (*dto.PatientResponse, error) {
	existing, err := uc.patientRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, entities.ErrPatientNotFound
	}

	// Verify patient belongs to the organization
	belongs, err := uc.patientRepo.PatientBelongsToOrganization(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, entities.ErrPatientNotFound // Return not found to avoid leaking patient existence
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

// SearchPatients searches for patients within an organization for autocomplete
func (uc *PatientUseCase) SearchPatients(ctx context.Context, orgID uuid.UUID, req *dto.PatientSearchRequest) (*dto.PatientSearchResult, error) {
	// Set default limit if not provided
	limit := req.Limit
	if limit == 0 {
		limit = 50 // Default limit
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	// Search for patients
	patients, err := uc.patientRepo.SearchPatients(ctx, orgID, req.Query, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	patientResponses := make([]dto.PatientSearchResponse, len(patients))
	for i, patient := range patients {
		patientResponses[i] = dto.ToPatientSearchResponse(patient)
	}

	return &dto.PatientSearchResult{
		Patients: patientResponses,
		Total:    len(patientResponses),
	}, nil
}

// AddPatientToOrganization links a patient to an organization
func (uc *PatientUseCase) AddPatientToOrganization(ctx context.Context, patientID, orgID uuid.UUID) error {
	// Verify patient exists
	exists, err := uc.patientRepo.Exists(ctx, patientID)
	if err != nil {
		return err
	}
	if !exists {
		return entities.ErrPatientNotFound
	}

	return uc.patientRepo.AddPatientToOrganization(ctx, patientID, orgID)
}
