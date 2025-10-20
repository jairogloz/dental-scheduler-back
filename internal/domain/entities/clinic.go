package entities

import (
	"time"

	"github.com/google/uuid"
)

// Clinic represents a dental clinic entity
type Clinic struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	Name           string    `json:"name" db:"name"`
	Address        *string   `json:"address,omitempty" db:"address"`
	Phone          *string   `json:"phone,omitempty" db:"phone"`
	Email          *string   `json:"email,omitempty" db:"email"`
	Timezone       string    `json:"timezone" db:"timezone"` // IANA timezone (e.g., "America/Mexico_City")
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the clinic entity is valid
func (c *Clinic) Validate() error {
	if c.Name == "" {
		return ErrInvalidClinicName
	}
	if c.OrganizationID == uuid.Nil {
		return ErrInvalidOrganizationID
	}
	return nil
}

// NewClinic creates a new clinic with the given name and organization
func NewClinic(name string, organizationID uuid.UUID) *Clinic {
	now := time.Now()
	return &Clinic{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           name,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// IsValid checks if the clinic has valid data
func (c *Clinic) IsValid() bool {
	return c.Validate() == nil
}
