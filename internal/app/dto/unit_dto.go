package dto

import (
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreateUnitRequest represents the request to create a unit
type CreateUnitRequest struct {
	ClinicID    uuid.UUID `json:"clinic_id" binding:"required"`
	Name        string    `json:"name" binding:"required"`
	Description *string   `json:"description,omitempty"`
	IsActive    *bool     `json:"is_active,omitempty"`
}

// UpdateUnitRequest represents the request to update a unit
type UpdateUnitRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// UnitResponse represents the response for a unit
type UnitResponse struct {
	ID          uuid.UUID `json:"id"`
	ClinicID    uuid.UUID `json:"clinic_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToEntity converts CreateUnitRequest to entities.Unit
func (req *CreateUnitRequest) ToEntity() *entities.Unit {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	return &entities.Unit{
		ID:          uuid.New(),
		ClinicID:    req.ClinicID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    isActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ToUnitResponse converts entities.Unit to UnitResponse
func ToUnitResponse(u *entities.Unit) *UnitResponse {
	return &UnitResponse{
		ID:          u.ID,
		ClinicID:    u.ClinicID,
		Name:        u.Name,
		Description: u.Description,
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// ToEntityUpdate converts UpdateUnitRequest to updated entities.Unit
func (req *UpdateUnitRequest) ToEntityUpdate(existing *entities.Unit) *entities.Unit {
	existing.Name = req.Name
	existing.Description = req.Description
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.UpdatedAt = time.Now()
	return existing
}
