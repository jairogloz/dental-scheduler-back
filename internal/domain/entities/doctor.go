package entities

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Doctor represents a doctor entity
type Doctor struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	Specialty     *string    `json:"specialty,omitempty" db:"specialty"`
	Email         *string    `json:"email,omitempty" db:"email"`
	Phone         *string    `json:"phone,omitempty" db:"phone"`
	DefaultUnitID *uuid.UUID `json:"default_unit_id,omitempty" db:"default_unit_id"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Validate checks if the doctor entity is valid
func (d *Doctor) Validate() error {
	if d.Name == "" {
		return ErrInvalidDoctorName
	}

	if d.Email != nil && *d.Email != "" {
		if !isValidEmail(*d.Email) {
			return ErrInvalidEmail
		}
	}

	return nil
}

// IsValid checks if the doctor has valid data
func (d *Doctor) IsValid() bool {
	return d.Validate() == nil
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
