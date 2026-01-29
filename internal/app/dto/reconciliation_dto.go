package dto

import (
	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreateReconciliationRequest represents a request to create a reconciliation
type CreateReconciliationRequest struct {
	PaymentMethod  entities.PaymentMethod `json:"payment_method" binding:"required"`
	Currency       entities.Currency      `json:"currency" binding:"required"`
	ExpectedCents  int64                  `json:"expected_cents" binding:"required"`
	ActualCents    int64                  `json:"actual_cents" binding:"required"`
	FloatLeftCents int64                  `json:"float_left_cents" binding:"required,min=0"`
	DepositedCents int64                  `json:"deposited_cents" binding:"required"`
	Notes          *string                `json:"notes,omitempty"`
}

// ReconciliationResponse represents a reconciliation in the response
type ReconciliationResponse struct {
	ID                  uuid.UUID                      `json:"id"`
	CashSessionID       uuid.UUID                      `json:"cash_session_id"`
	OrganizationID      uuid.UUID                      `json:"organization_id"`
	ClinicID            uuid.UUID                      `json:"clinic_id"`
	PaymentMethod       entities.PaymentMethod         `json:"payment_method"`
	Currency            entities.Currency              `json:"currency"`
	ReconciledAt        string                         `json:"reconciled_at"`
	ReconciledByUserID  uuid.UUID                      `json:"reconciled_by_user_id"`
	ExpectedAmountCents int64                          `json:"expected_amount_cents"`
	ActualAmountCents   int64                          `json:"actual_amount_cents"`
	FloatLeftCents      int64                          `json:"float_left_cents"`
	DepositedCents      int64                          `json:"deposited_cents"`
	DiscrepancyCents    int64                          `json:"discrepancy_cents"`
	Status              entities.ReconciliationStatus  `json:"status"`
	Notes               *string                        `json:"notes,omitempty"`
}

// ReconciliationPreviewResponse represents expected amounts for reconciliation
type ReconciliationPreviewResponse struct {
	Session                   CashSessionResponse         `json:"session"`
	ExpectedAmountsByCurrency map[entities.Currency]int64 `json:"expected_amounts_by_currency"`
	ExistingReconciliations   []ReconciliationResponse    `json:"existing_reconciliations"`
}

// ListReconciliationsQuery represents query parameters for listing reconciliations
type ListReconciliationsQuery struct {
	ClinicID       *uuid.UUID                        `form:"clinic_id"`
	CashSessionID  *uuid.UUID                        `form:"cash_session_id"`
	PaymentMethod  *entities.PaymentMethod           `form:"payment_method"`
	Currency       *entities.Currency                `form:"currency"`
	Status         *entities.ReconciliationStatus    `form:"status"`
	StartDate      *string                           `form:"start_date"` // YYYY-MM-DD
	EndDate        *string                           `form:"end_date"`   // YYYY-MM-DD
	HasDiscrepancy *bool                             `form:"has_discrepancy"`
	Page           int                               `form:"page" binding:"min=1"`
	Limit          int                               `form:"limit" binding:"min=1,max=100"`
}
