package dto

import (
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreatePatientRequest represents the request to create a patient
type CreatePatientRequest struct {
	FirstName      string     `json:"first_name" binding:"required"`
	LastName       *string    `json:"last_name,omitempty"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	MedicalHistory *string    `json:"medical_history,omitempty"`
}

// CreatePatientWithOrgRequest represents the request to create a patient with organization ID
type CreatePatientWithOrgRequest struct {
	CreatePatientRequest
	OrganizationIDStr *string `form:"organization_id,omitempty"`
}

// GetOrganizationID parses and returns the organization ID as UUID
func (req *CreatePatientWithOrgRequest) GetOrganizationID() (*uuid.UUID, error) {
	if req.OrganizationIDStr == nil || *req.OrganizationIDStr == "" {
		return nil, nil
	}

	orgID, err := uuid.Parse(*req.OrganizationIDStr)
	if err != nil {
		return nil, err
	}

	return &orgID, nil
}

// UpdatePatientRequest represents the request to update a patient
// All fields are optional for partial updates
type UpdatePatientRequest struct {
	FirstName      *string    `json:"first_name,omitempty"`
	LastName       *string    `json:"last_name,omitempty"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	MedicalHistory *string    `json:"medical_history,omitempty"`
}

// PatientResponse represents the response for a patient
type PatientResponse struct {
	ID             uuid.UUID  `json:"id"`
	FirstName      string     `json:"first_name"`
	LastName       *string    `json:"last_name,omitempty"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	MedicalHistory *string    `json:"medical_history,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ToEntity converts CreatePatientRequest to entities.Patient
func (req *CreatePatientRequest) ToEntity() *entities.Patient {
	return &entities.Patient{
		ID:             uuid.New(),
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		Phone:          req.Phone,
		DateOfBirth:    req.DateOfBirth,
		MedicalHistory: req.MedicalHistory,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// ToPatientResponse converts entities.Patient to PatientResponse
func ToPatientResponse(p *entities.Patient) *PatientResponse {
	return &PatientResponse{
		ID:             p.ID,
		FirstName:      p.FirstName,
		LastName:       p.LastName,
		Email:          p.Email,
		Phone:          p.Phone,
		DateOfBirth:    p.DateOfBirth,
		MedicalHistory: p.MedicalHistory,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// ToEntityUpdate converts UpdatePatientRequest to updated entities.Patient
// Only updates fields that are present in the request
// If a field is an empty string, sets it to nil
func (req *UpdatePatientRequest) ToEntityUpdate(existing *entities.Patient) *entities.Patient {
	// Update FirstName if provided
	if req.FirstName != nil {
		if *req.FirstName == "" {
			// Empty string not allowed for FirstName, keep existing
			// FirstName is required field
		} else {
			existing.FirstName = *req.FirstName
		}
	}

	// Update LastName if provided
	if req.LastName != nil {
		if *req.LastName == "" {
			existing.LastName = nil // Empty string -> set to null
		} else {
			existing.LastName = req.LastName
		}
	}

	// Update Email if provided
	if req.Email != nil {
		if *req.Email == "" {
			existing.Email = nil // Empty string -> set to null
		} else {
			existing.Email = req.Email
		}
	}

	// Update Phone if provided
	if req.Phone != nil {
		if *req.Phone == "" {
			existing.Phone = nil // Empty string -> set to null
		} else {
			existing.Phone = req.Phone
		}
	}

	// Update DateOfBirth if provided
	if req.DateOfBirth != nil {
		existing.DateOfBirth = req.DateOfBirth
	}

	// Update MedicalHistory if provided
	if req.MedicalHistory != nil {
		if *req.MedicalHistory == "" {
			existing.MedicalHistory = nil // Empty string -> set to null
		} else {
			existing.MedicalHistory = req.MedicalHistory
		}
	}

	existing.UpdatedAt = time.Now()
	return existing
}

// PatientSearchRequest represents the request to search patients
type PatientSearchRequest struct {
	Query string `form:"q,omitempty"`     // Search by name, phone, or email
	Limit int    `form:"limit,omitempty"` // Max results (default: 50, max: 100)
}

// PatientSearchResponse represents minimal patient data for autocomplete
type PatientSearchResponse struct {
	ID        string  `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty"`
}

// PatientSearchResult represents the wrapper for search results
type PatientSearchResult struct {
	Patients []PatientSearchResponse `json:"patients"`
	Total    int                     `json:"total"`
}

// ToPatientSearchResponse converts entities.Patient to PatientSearchResponse
func ToPatientSearchResponse(p *entities.Patient) PatientSearchResponse {
	return PatientSearchResponse{
		ID:        p.ID.String(),
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Phone:     p.Phone,
		Email:     p.Email,
	}
}
