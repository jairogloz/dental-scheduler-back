package entities

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

// Doctor represents a doctor entity
type Doctor struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	OrganizationID uuid.UUID  `json:"organization_id" db:"organization_id"`
	UserID         *uuid.UUID `json:"user_id,omitempty" db:"user_id"` // Links to auth.users(id)
	Name           string     `json:"name" db:"name"`
	Specialty      *string    `json:"specialty,omitempty" db:"specialty"`
	Email          *string    `json:"email,omitempty" db:"email"`
	Phone          *string    `json:"phone,omitempty" db:"phone"`
	DefaultUnitID  *uuid.UUID `json:"default_unit_id,omitempty" db:"default_unit_id"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Validate checks if the doctor entity is valid
func (d *Doctor) Validate() error {
	if d.Name == "" {
		return ErrInvalidDoctorName
	}

	if d.OrganizationID == uuid.Nil {
		return ErrInvalidOrganizationID
	}

	if d.Email != nil && *d.Email != "" {
		if !isValidEmail(*d.Email) {
			return ErrInvalidEmail
		}
	}

	return nil
}

// ValidateWithUnit checks if the doctor entity is valid including unit organization validation
// This should be used when setting or updating the default unit
func (d *Doctor) ValidateWithUnit(unitOrganizationID *uuid.UUID) error {
	if err := d.Validate(); err != nil {
		return err
	}

	// If doctor has a default unit, ensure it belongs to the same organization
	if d.DefaultUnitID != nil && unitOrganizationID != nil {
		if *unitOrganizationID != d.OrganizationID {
			return ErrDoctorUnitOrganizationMismatch
		}
	}

	return nil
}

// NewDoctor creates a new doctor with the given name and organization
func NewDoctor(name string, organizationID uuid.UUID) *Doctor {
	now := time.Now()
	return &Doctor{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Name:           name,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// LinkToUser links the doctor to an authenticated user
func (d *Doctor) LinkToUser(userID uuid.UUID) {
	d.UserID = &userID
	d.UpdatedAt = time.Now()
}

// UnlinkFromUser removes the user link from the doctor
func (d *Doctor) UnlinkFromUser() {
	d.UserID = nil
	d.UpdatedAt = time.Now()
}

// HasUserAccount checks if the doctor has a linked user account
func (d *Doctor) HasUserAccount() bool {
	return d.UserID != nil
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
