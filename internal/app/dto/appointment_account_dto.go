package dto

import (
	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreateServiceChargeRequest represents a request to create a service charge
type CreateServiceChargeRequest struct {
	DoctorID               uuid.UUID           `json:"doctor_id" binding:"required"`
	DoctorType             entities.DoctorType `json:"doctor_type" binding:"required"`
	Currency               entities.Currency   `json:"currency" binding:"required"`
	AmountCents            int64               `json:"amount_cents" binding:"required"`
	Description            string              `json:"description" binding:"required"`
	ServiceID              *string             `json:"service_id,omitempty"`
	CommissionPct          *float64            `json:"commission_pct,omitempty"`
	ExternalDoctorFeeCents *int64              `json:"external_doctor_fee_cents,omitempty"`
}

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	PaymentMethod entities.PaymentMethod `json:"payment_method" binding:"required"`
	Currency      entities.Currency      `json:"currency" binding:"required"`
	AmountCents   int64                  `json:"amount_cents" binding:"required"`
	Description   string                 `json:"description" binding:"required"`
	ExchangeRate  *float64               `json:"exchange_rate,omitempty"`
}

// CreateCorrectionRequest represents a request to create a correction entry
type CreateCorrectionRequest struct {
	Description string  `json:"description" binding:"required"`
	Notes       *string `json:"notes,omitempty"`
}

// AppointmentAccountEntryResponse represents an entry in the response
type AppointmentAccountEntryResponse struct {
	ID                     uuid.UUID               `json:"id"`
	AppointmentAccountID   uuid.UUID               `json:"appointment_account_id"`
	Type                   entities.EntryType      `json:"type"`
	Currency               entities.Currency       `json:"currency"`
	AmountCents            int64                   `json:"amount_cents"`
	Description            string                  `json:"description"`
	CreatedByUserID        uuid.UUID               `json:"created_by_user_id"`
	CreatedAt              string                  `json:"created_at"`
	PaymentMethod          *entities.PaymentMethod `json:"payment_method,omitempty"`
	ExchangeRateUsed       *float64                `json:"exchange_rate_used,omitempty"`
	DoctorID               *uuid.UUID              `json:"doctor_id,omitempty"`
	DoctorType             *entities.DoctorType    `json:"doctor_type,omitempty"`
	CommissionPct          *float64                `json:"commission_pct,omitempty"`
	ExternalDoctorFeeCents *int64                  `json:"external_doctor_fee_cents,omitempty"`
	ServiceID              *string                 `json:"service_id,omitempty"`
	CashSessionID          *uuid.UUID              `json:"cash_session_id,omitempty"`
	CorrectsEntryID        *uuid.UUID              `json:"corrects_entry_id,omitempty"`
	Notes                  *string                 `json:"notes,omitempty"`
	Quantity               int                     `json:"quantity"`
}

// AppointmentAccountResponse represents the full account with entries
type AppointmentAccountResponse struct {
	ID             uuid.UUID                         `json:"id"`
	OrganizationID uuid.UUID                         `json:"organization_id"`
	AppointmentID  uuid.UUID                         `json:"appointment_id"`
	Entries        []AppointmentAccountEntryResponse `json:"entries"`
	Balance        AccountBalanceResponse            `json:"balance"`
}

// AccountBalanceResponse represents the balance summary
type AccountBalanceResponse struct {
	TotalChargesCents   int64                            `json:"total_charges_cents"`
	TotalDiscountsCents int64                            `json:"total_discounts_cents"`
	TotalPaymentsCents  int64                            `json:"total_payments_cents"`
	TotalRefundsCents   int64                            `json:"total_refunds_cents"`
	BalanceDueCents     int64                            `json:"balance_due_cents"`
	PaymentsByCurrency  map[entities.Currency]int64      `json:"payments_by_currency"`
	PaymentsByMethod    map[entities.PaymentMethod]int64 `json:"payments_by_method"`
}
