package dto

import (
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreateClinicRequest represents the request to create a clinic
type CreateClinicRequest struct {
	Name    string  `json:"name" binding:"required"`
	Address *string `json:"address,omitempty"`
	Phone   *string `json:"phone,omitempty"`
}

// UpdateClinicRequest represents the request to update a clinic
type UpdateClinicRequest struct {
	Name    string  `json:"name" binding:"required"`
	Address *string `json:"address,omitempty"`
	Phone   *string `json:"phone,omitempty"`
}

// ClinicResponse represents the response for a clinic
type ClinicResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Address   *string   `json:"address,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToEntity converts CreateClinicRequest to entities.Clinic
func (req *CreateClinicRequest) ToEntity() *entities.Clinic {
	return &entities.Clinic{
		ID:        uuid.New(),
		Name:      req.Name,
		Address:   req.Address,
		Phone:     req.Phone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ToClinicResponse converts entities.Clinic to ClinicResponse
func ToClinicResponse(c *entities.Clinic) *ClinicResponse {
	return &ClinicResponse{
		ID:        c.ID,
		Name:      c.Name,
		Address:   c.Address,
		Phone:     c.Phone,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// ToEntityUpdate converts UpdateClinicRequest to updated entities.Clinic
func (req *UpdateClinicRequest) ToEntityUpdate(existing *entities.Clinic) *entities.Clinic {
	existing.Name = req.Name
	existing.Address = req.Address
	existing.Phone = req.Phone
	existing.UpdatedAt = time.Now()
	return existing
}
