package entities

import (
	"time"

	"github.com/google/uuid"
)

// Patient represents a patient entity
type Patient struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Email          *string    `json:"email,omitempty" db:"email"`
	Phone          *string    `json:"phone,omitempty" db:"phone"`
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	MedicalHistory *string    `json:"medical_history,omitempty" db:"medical_history"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Validate checks if the patient entity is valid
func (p *Patient) Validate() error {
	if p.Name == "" {
		return ErrInvalidPatientName
	}

	if p.Email != nil && *p.Email != "" {
		if !isValidEmail(*p.Email) {
			return ErrInvalidEmail
		}
	}

	return nil
}

// IsValid checks if the patient has valid data
func (p *Patient) IsValid() bool {
	return p.Validate() == nil
}
