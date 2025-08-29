package dto

import (
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreatePatientRequest represents the request to create a patient
type CreatePatientRequest struct {
	Name           string     `json:"name" binding:"required"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	MedicalHistory *string    `json:"medical_history,omitempty"`
}

// UpdatePatientRequest represents the request to update a patient
type UpdatePatientRequest struct {
	Name           string     `json:"name" binding:"required"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	MedicalHistory *string    `json:"medical_history,omitempty"`
}

// PatientResponse represents the response for a patient
type PatientResponse struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
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
		Name:           req.Name,
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
		Name:           p.Name,
		Email:          p.Email,
		Phone:          p.Phone,
		DateOfBirth:    p.DateOfBirth,
		MedicalHistory: p.MedicalHistory,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// ToEntityUpdate converts UpdatePatientRequest to updated entities.Patient
func (req *UpdatePatientRequest) ToEntityUpdate(existing *entities.Patient) *entities.Patient {
	existing.Name = req.Name
	existing.Email = req.Email
	existing.Phone = req.Phone
	existing.DateOfBirth = req.DateOfBirth
	existing.MedicalHistory = req.MedicalHistory
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
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Phone *string `json:"phone,omitempty"`
	Email *string `json:"email,omitempty"`
}

// PatientSearchResult represents the wrapper for search results
type PatientSearchResult struct {
	Patients []PatientSearchResponse `json:"patients"`
	Total    int                     `json:"total"`
}

// ToPatientSearchResponse converts entities.Patient to PatientSearchResponse
func ToPatientSearchResponse(p *entities.Patient) PatientSearchResponse {
	return PatientSearchResponse{
		ID:    p.ID.String(),
		Name:  p.Name,
		Phone: p.Phone,
		Email: p.Email,
	}
}
