package entities

import (
	"time"

	"github.com/google/uuid"
)

// ReconciliationStatus represents the status of a reconciliation
type ReconciliationStatus string

const (
	ReconciliationStatusPending  ReconciliationStatus = "pending"
	ReconciliationStatusClosed   ReconciliationStatus = "closed"
	ReconciliationStatusDisputed ReconciliationStatus = "disputed"
)

// Reconciliation represents a cash reconciliation when closing a cash session
type Reconciliation struct {
	ID                  uuid.UUID            `json:"id" db:"id"`
	CashSessionID       uuid.UUID            `json:"cash_session_id" db:"cash_session_id"`
	OrganizationID      uuid.UUID            `json:"organization_id" db:"organization_id"`
	ClinicID            uuid.UUID            `json:"clinic_id" db:"clinic_id"`
	PaymentMethod       PaymentMethod        `json:"payment_method" db:"payment_method"`
	Currency            Currency             `json:"currency" db:"currency"`
	ReconciledAt        time.Time            `json:"reconciled_at" db:"reconciled_at"`
	ReconciledByUserID  uuid.UUID            `json:"reconciled_by_user_id" db:"reconciled_by_user_id"`
	ExpectedAmountCents int64                `json:"expected_amount_cents" db:"expected_amount_cents"`
	ActualAmountCents   int64                `json:"actual_amount_cents" db:"actual_amount_cents"`
	FloatLeftCents      int64                `json:"float_left_cents" db:"float_left_cents"`
	DepositedCents      int64                `json:"deposited_cents" db:"deposited_cents"`
	DiscrepancyCents    int64                `json:"discrepancy_cents" db:"discrepancy_cents"`
	Status              ReconciliationStatus `json:"status" db:"status"`
	Notes               *string              `json:"notes,omitempty" db:"notes"`
	CreatedAt           time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at" db:"updated_at"`
}

// Validate checks if the reconciliation is valid
func (r *Reconciliation) Validate() error {
	if r.CashSessionID == uuid.Nil {
		return ErrInvalidCashSessionID
	}
	if r.OrganizationID == uuid.Nil {
		return ErrInvalidOrganizationID
	}
	if r.ClinicID == uuid.Nil {
		return ErrInvalidClinicID
	}
	if !IsValidPaymentMethod(r.PaymentMethod) {
		return ErrInvalidPaymentMethod
	}
	if !IsValidCurrency(r.Currency) {
		return ErrInvalidCurrency
	}
	if r.ReconciledByUserID == uuid.Nil {
		return ErrReconciledByUserRequired
	}
	if !IsValidReconciliationStatus(r.Status) {
		return ErrInvalidReconciliationStatus
	}
	if r.FloatLeftCents < 0 {
		return ErrInvalidFloatLeft
	}

	// Validate computed fields
	expectedDeposited := r.ActualAmountCents - r.FloatLeftCents
	if r.DepositedCents != expectedDeposited {
		return ErrInvalidDepositedAmount
	}

	expectedDiscrepancy := r.ActualAmountCents - r.ExpectedAmountCents
	if r.DiscrepancyCents != expectedDiscrepancy {
		return ErrInvalidDiscrepancyAmount
	}

	return nil
}

// IsValid checks if the reconciliation has valid data
func (r *Reconciliation) IsValid() bool {
	return r.Validate() == nil
}

// HasDiscrepancy returns true if there's a difference between expected and actual
func (r *Reconciliation) HasDiscrepancy() bool {
	return r.DiscrepancyCents != 0
}

// IsOverage returns true if actual is greater than expected
func (r *Reconciliation) IsOverage() bool {
	return r.DiscrepancyCents > 0
}

// IsShortage returns true if actual is less than expected
func (r *Reconciliation) IsShortage() bool {
	return r.DiscrepancyCents < 0
}

// CalculateDeposited calculates the deposited amount (actual - float_left)
func CalculateDeposited(actualAmountCents, floatLeftCents int64) int64 {
	return actualAmountCents - floatLeftCents
}

// CalculateDiscrepancy calculates the discrepancy (actual - expected)
func CalculateDiscrepancy(actualAmountCents, expectedAmountCents int64) int64 {
	return actualAmountCents - expectedAmountCents
}

// IsValidReconciliationStatus checks if the reconciliation status is valid
func IsValidReconciliationStatus(status ReconciliationStatus) bool {
	switch status {
	case ReconciliationStatusPending, ReconciliationStatusClosed, ReconciliationStatusDisputed:
		return true
	}
	return false
}
