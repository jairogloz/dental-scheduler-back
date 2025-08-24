package entities

import (
	"time"

	"github.com/google/uuid"
)

// Unit represents a dental unit entity
type Unit struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ClinicID    uuid.UUID `json:"clinic_id" db:"clinic_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the unit entity is valid
func (u *Unit) Validate() error {
	if u.Name == "" {
		return ErrInvalidUnitName
	}
	if u.ClinicID == uuid.Nil {
		return ErrInvalidClinicID
	}
	return nil
}

// IsValid checks if the unit has valid data
func (u *Unit) IsValid() bool {
	return u.Validate() == nil
}
