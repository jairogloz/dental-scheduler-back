package entities

import (
	"time"

	"github.com/google/uuid"
)

// CashSessionStatus represents the status of a cash session
type CashSessionStatus string

const (
	CashSessionStatusOpen   CashSessionStatus = "open"
	CashSessionStatusClosed CashSessionStatus = "closed"
)

// CashSessionOpeningType represents how the session was opened
type CashSessionOpeningType string

const (
	CashSessionOpeningTypeManual CashSessionOpeningType = "manual"
	CashSessionOpeningTypeAuto   CashSessionOpeningType = "auto"
)

// CashSession represents a cash handling period (apertura de caja)
type CashSession struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	OrganizationID     uuid.UUID              `json:"organization_id" db:"organization_id"`
	ClinicID           uuid.UUID              `json:"clinic_id" db:"clinic_id"`
	UserID             uuid.UUID              `json:"user_id" db:"user_id"`
	OpenedAt           time.Time              `json:"opened_at" db:"opened_at"`
	ClosedAt           *time.Time             `json:"closed_at,omitempty" db:"closed_at"`
	StartingFloatCents int64                  `json:"starting_float_cents" db:"starting_float_cents"`
	Status             CashSessionStatus      `json:"status" db:"status"`
	OpeningType        CashSessionOpeningType `json:"opening_type" db:"opening_type"`
	Notes              *string                `json:"notes,omitempty" db:"notes"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// Validate checks if the cash session is valid
func (cs *CashSession) Validate() error {
	if cs.OrganizationID == uuid.Nil {
		return ErrInvalidOrganizationID
	}
	if cs.ClinicID == uuid.Nil {
		return ErrInvalidClinicID
	}
	if cs.UserID == uuid.Nil {
		return ErrUserIDRequired
	}
	if !IsValidCashSessionStatus(cs.Status) {
		return ErrInvalidCashSessionStatus
	}
	if !IsValidCashSessionOpeningType(cs.OpeningType) {
		return ErrInvalidCashSessionOpeningType
	}
	if cs.StartingFloatCents < 0 {
		return ErrInvalidStartingFloat
	}
	if cs.Status == CashSessionStatusClosed && cs.ClosedAt == nil {
		return ErrClosedAtRequired
	}
	if cs.Status == CashSessionStatusOpen && cs.ClosedAt != nil {
		return ErrClosedAtMustBeNull
	}
	return nil
}

// IsValid checks if the cash session has valid data
func (cs *CashSession) IsValid() bool {
	return cs.Validate() == nil
}

// IsOpen returns true if the session is currently open
func (cs *CashSession) IsOpen() bool {
	return cs.Status == CashSessionStatusOpen
}

// IsClosed returns true if the session is closed
func (cs *CashSession) IsClosed() bool {
	return cs.Status == CashSessionStatusClosed
}

// Close closes the cash session
func (cs *CashSession) Close() error {
	if cs.IsClosed() {
		return ErrCashSessionAlreadyClosed
	}
	now := time.Now()
	cs.ClosedAt = &now
	cs.Status = CashSessionStatusClosed
	return nil
}

// IsValidCashSessionStatus checks if the status is valid
func IsValidCashSessionStatus(status CashSessionStatus) bool {
	return status == CashSessionStatusOpen || status == CashSessionStatusClosed
}

// IsValidCashSessionOpeningType checks if the opening type is valid
func IsValidCashSessionOpeningType(openingType CashSessionOpeningType) bool {
	return openingType == CashSessionOpeningTypeManual || openingType == CashSessionOpeningTypeAuto
}
