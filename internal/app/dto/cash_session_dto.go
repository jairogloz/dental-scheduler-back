package dto

import (
	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// OpenCashSessionRequest represents a request to open a cash session
type OpenCashSessionRequest struct {
	ClinicID           uuid.UUID                       `json:"clinic_id" binding:"required"`
	OpeningType        entities.CashSessionOpeningType `json:"opening_type" binding:"required"`
	StartingFloatCents int64                           `json:"starting_float_cents" binding:"required,min=0"`
	Notes              *string                         `json:"notes,omitempty"`
}

// CashSessionResponse represents a cash session in the response
type CashSessionResponse struct {
	ID                 uuid.UUID                       `json:"id"`
	OrganizationID     uuid.UUID                       `json:"organization_id"`
	ClinicID           uuid.UUID                       `json:"clinic_id"`
	UserID             uuid.UUID                       `json:"user_id"`
	OpenedAt           string                          `json:"opened_at"`
	ClosedAt           *string                         `json:"closed_at,omitempty"`
	StartingFloatCents int64                           `json:"starting_float_cents"`
	Status             entities.CashSessionStatus      `json:"status"`
	OpeningType        entities.CashSessionOpeningType `json:"opening_type"`
	Notes              *string                         `json:"notes,omitempty"`
}

// CashSessionDetailsResponse represents full session details with entries
type CashSessionDetailsResponse struct {
	Session         CashSessionResponse                                    `json:"session"`
	Entries         []AppointmentAccountEntryResponse                      `json:"entries"`
	ExpectedAmounts map[entities.Currency]int64                            `json:"expected_amounts"`
	PaymentSummary  map[entities.PaymentMethod]map[entities.Currency]int64 `json:"payment_summary"`
}

// ListCashSessionsQuery represents query parameters for listing sessions
type ListCashSessionsQuery struct {
	ClinicID  *uuid.UUID                  `form:"clinic_id"`
	UserID    *uuid.UUID                  `form:"user_id"`
	Status    *entities.CashSessionStatus `form:"status"`
	StartDate *string                     `form:"start_date"` // YYYY-MM-DD
	EndDate   *string                     `form:"end_date"`   // YYYY-MM-DD
	Page      int                         `form:"page" binding:"min=1"`
	Limit     int                         `form:"limit" binding:"min=1,max=100"`
}
