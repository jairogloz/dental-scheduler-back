package dto

import (
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreateDoctorRequest represents the request to create a doctor
type CreateDoctorRequest struct {
	Name          string     `json:"name" binding:"required"`
	Specialty     *string    `json:"specialty,omitempty"`
	Email         *string    `json:"email,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	DefaultUnitID *uuid.UUID `json:"default_unit_id,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
}

// UpdateDoctorRequest represents the request to update a doctor
type UpdateDoctorRequest struct {
	Name          string     `json:"name" binding:"required"`
	Specialty     *string    `json:"specialty,omitempty"`
	Email         *string    `json:"email,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	DefaultUnitID *uuid.UUID `json:"default_unit_id,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
}

// DoctorResponse represents the response for a doctor
type DoctorResponse struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Specialty     *string    `json:"specialty,omitempty"`
	Email         *string    `json:"email,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	DefaultUnitID *uuid.UUID `json:"default_unit_id,omitempty"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ToEntity converts CreateDoctorRequest to entities.Doctor
func (req *CreateDoctorRequest) ToEntity() *entities.Doctor {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return &entities.Doctor{
		ID:            uuid.New(),
		Name:          req.Name,
		Specialty:     req.Specialty,
		Email:         req.Email,
		Phone:         req.Phone,
		DefaultUnitID: req.DefaultUnitID,
		IsActive:      isActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// ToDoctorResponse converts entities.Doctor to DoctorResponse
func ToDoctorResponse(d *entities.Doctor) *DoctorResponse {
	return &DoctorResponse{
		ID:            d.ID,
		Name:          d.Name,
		Specialty:     d.Specialty,
		Email:         d.Email,
		Phone:         d.Phone,
		DefaultUnitID: d.DefaultUnitID,
		IsActive:      d.IsActive,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}

// ToEntityUpdate converts UpdateDoctorRequest to updated entities.Doctor
func (req *UpdateDoctorRequest) ToEntityUpdate(existing *entities.Doctor) *entities.Doctor {
	existing.Name = req.Name
	existing.Specialty = req.Specialty
	existing.Email = req.Email
	existing.Phone = req.Phone
	existing.DefaultUnitID = req.DefaultUnitID
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.UpdatedAt = time.Now()
	return existing
}

// DoctorWithOrgInfoResponse represents a doctor with organization and clinic info
type DoctorWithOrgInfoResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Specialty         *string `json:"specialty,omitempty"`
	DefaultUnitID     *string `json:"default_unit_id,omitempty"`
	DefaultClinicID   *string `json:"default_clinic_id,omitempty"`
	DefaultClinicName *string `json:"default_clinic_name,omitempty"`
	OrgID             string  `json:"org_id"`
	OrgName           string  `json:"org_name"`
}

// GetDoctorsByOrgRequest represents the query parameters for getting doctors by organization
type GetDoctorsByOrgRequest struct {
	OrgID    string  `form:"orgId" binding:"required"`
	ClinicID *string `form:"clinicId"`
}

// ParsedOrgID returns the parsed UUID for OrgID
func (req *GetDoctorsByOrgRequest) ParsedOrgID() (uuid.UUID, error) {
	return uuid.Parse(req.OrgID)
}

// ParsedClinicID returns the parsed UUID for ClinicID if provided
func (req *GetDoctorsByOrgRequest) ParsedClinicID() (*uuid.UUID, error) {
	if req.ClinicID == nil || *req.ClinicID == "" {
		return nil, nil
	}

	clinicUUID, err := uuid.Parse(*req.ClinicID)
	if err != nil {
		return nil, err
	}
	return &clinicUUID, nil
}
