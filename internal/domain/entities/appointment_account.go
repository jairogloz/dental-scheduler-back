package entities

import (
	"time"

	"github.com/google/uuid"
)

// AppointmentAccount represents a financial account for an appointment.
// Created on-demand when the first charge or payment is recorded.
type AppointmentAccount struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	AppointmentID  uuid.UUID `json:"appointment_id" db:"appointment_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the appointment account is valid
func (a *AppointmentAccount) Validate() error {
	if a.OrganizationID == uuid.Nil {
		return ErrInvalidOrganizationID
	}
	if a.AppointmentID == uuid.Nil {
		return ErrInvalidAppointmentID
	}
	return nil
}

// IsValid checks if the appointment account has valid data
func (a *AppointmentAccount) IsValid() bool {
	return a.Validate() == nil
}
