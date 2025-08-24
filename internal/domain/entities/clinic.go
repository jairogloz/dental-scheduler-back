package entities

import (
	"time"

	"github.com/google/uuid"
)

// Clinic represents a dental clinic entity
type Clinic struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Address   *string   `json:"address,omitempty" db:"address"`
	Phone     *string   `json:"phone,omitempty" db:"phone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the clinic entity is valid
func (c *Clinic) Validate() error {
	if c.Name == "" {
		return ErrInvalidClinicName
	}
	return nil
}

// IsValid checks if the clinic has valid data
func (c *Clinic) IsValid() bool {
	return c.Validate() == nil
}
