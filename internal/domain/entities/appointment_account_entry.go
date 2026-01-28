package entities

import (
	"time"

	"github.com/google/uuid"
)

// EntryType represents the type of an accounting entry
type EntryType string

const (
	EntryTypeServiceCharge EntryType = "service_charge"
	EntryTypeDiscount      EntryType = "discount"
	EntryTypePayment       EntryType = "payment"
	EntryTypeRefund        EntryType = "refund"
	EntryTypeCorrection    EntryType = "correction"
)

// Currency represents supported currencies
type Currency string

const (
	CurrencyMXN Currency = "MXN"
	CurrencyUSD Currency = "USD"
)

// PaymentMethod represents payment methods
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodTransfer PaymentMethod = "transfer"
)

// DoctorType represents the type of doctor for commission calculation
type DoctorType string

const (
	DoctorTypeInternal DoctorType = "internal"
	DoctorTypeExternal DoctorType = "external"
)

// AppointmentAccountEntry represents an immutable financial transaction entry
type AppointmentAccountEntry struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	AppointmentAccountID uuid.UUID `json:"appointment_account_id" db:"appointment_account_id"`
	Type                 EntryType `json:"type" db:"type"`
	Currency             Currency  `json:"currency" db:"currency"`
	AmountCents          int64     `json:"amount_cents" db:"amount_cents"`
	Description          string    `json:"description" db:"description"`
	CreatedByUserID      uuid.UUID `json:"created_by_user_id" db:"created_by_user_id"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`

	// Conditional fields
	PaymentMethod    *PaymentMethod `json:"payment_method,omitempty" db:"payment_method"`
	ExchangeRateUsed *float64       `json:"exchange_rate_used,omitempty" db:"exchange_rate_used"`
	DoctorID         *uuid.UUID     `json:"doctor_id,omitempty" db:"doctor_id"`
	CorrectsEntryID  *uuid.UUID     `json:"corrects_entry_id,omitempty" db:"corrects_entry_id"`

	// Doctor commission fields
	DoctorType             *DoctorType `json:"doctor_type,omitempty" db:"doctor_type"`
	CommissionPct          *float64    `json:"commission_pct,omitempty" db:"commission_pct"`
	ExternalDoctorFeeCents *int64      `json:"external_doctor_fee_cents,omitempty" db:"external_doctor_fee_cents"`
	IsSensitive            bool        `json:"is_sensitive" db:"is_sensitive"`

	// Optional fields
	ServiceID      *string    `json:"service_id,omitempty" db:"service_id"`
	Quantity       int        `json:"quantity" db:"quantity"`
	UnitPriceCents *int64     `json:"unit_price_cents,omitempty" db:"unit_price_cents"`
	Notes          *string    `json:"notes,omitempty" db:"notes"`
	CashSessionID  *uuid.UUID `json:"cash_session_id,omitempty" db:"cash_session_id"`
}

// Validate performs comprehensive validation of the entry
func (e *AppointmentAccountEntry) Validate() error {
	// Basic validation
	if e.AppointmentAccountID == uuid.Nil {
		return ErrInvalidAppointmentAccountID
	}
	if !IsValidEntryType(e.Type) {
		return ErrInvalidEntryType
	}
	if !IsValidCurrency(e.Currency) {
		return ErrInvalidCurrency
	}
	if e.AmountCents == 0 {
		return ErrAmountCannotBeZero
	}
	if e.Description == "" {
		return ErrDescriptionRequired
	}
	if e.CreatedByUserID == uuid.Nil {
		return ErrCreatedByUserRequired
	}

	// Currency-specific validation
	if e.Currency == CurrencyUSD && e.ExchangeRateUsed == nil {
		return ErrExchangeRateRequired
	}

	// Type-specific validation
	switch e.Type {
	case EntryTypePayment:
		if e.PaymentMethod == nil {
			return ErrPaymentMethodRequired
		}
		if !IsValidPaymentMethod(*e.PaymentMethod) {
			return ErrInvalidPaymentMethod
		}
		if e.AmountCents < 0 {
			return ErrPaymentAmountMustBePositive
		}

	case EntryTypeServiceCharge:
		if e.DoctorID == nil {
			return ErrDoctorIDRequired
		}
		if e.DoctorType == nil {
			return ErrDoctorTypeRequired
		}
		if !IsValidDoctorType(*e.DoctorType) {
			return ErrInvalidDoctorType
		}
		if e.AmountCents < 0 {
			return ErrServiceChargeAmountMustBePositive
		}

		// Commission validation based on doctor type
		if *e.DoctorType == DoctorTypeInternal {
			if e.CommissionPct == nil {
				return ErrCommissionPctRequired
			}
			if *e.CommissionPct < 0 || *e.CommissionPct > 100 {
				return ErrInvalidCommissionPct
			}
		} else if *e.DoctorType == DoctorTypeExternal {
			if e.ExternalDoctorFeeCents == nil {
				return ErrExternalDoctorFeeRequired
			}
			if *e.ExternalDoctorFeeCents < 0 {
				return ErrInvalidExternalDoctorFee
			}
		}

	case EntryTypeDiscount, EntryTypeRefund:
		if e.AmountCents > 0 {
			return ErrDiscountRefundAmountMustBeNegative
		}

	case EntryTypeCorrection:
		if e.CorrectsEntryID == nil {
			return ErrCorrectsEntryIDRequired
		}
	}

	// Quantity validation
	if e.Quantity < 0 {
		return ErrInvalidQuantity
	}

	return nil
}

// IsValid checks if the entry has valid data
func (e *AppointmentAccountEntry) IsValid() bool {
	return e.Validate() == nil
}

// IsIncome returns true if this entry represents income (charge or payment)
func (e *AppointmentAccountEntry) IsIncome() bool {
	return e.Type == EntryTypeServiceCharge || e.Type == EntryTypePayment
}

// IsExpense returns true if this entry represents an expense (discount or refund)
func (e *AppointmentAccountEntry) IsExpense() bool {
	return e.Type == EntryTypeDiscount || e.Type == EntryTypeRefund
}

// IsCorrection returns true if this entry corrects another entry
func (e *AppointmentAccountEntry) IsCorrection() bool {
	return e.Type == EntryTypeCorrection
}

// CalculateCommission calculates the doctor's commission amount in cents
func (e *AppointmentAccountEntry) CalculateCommission() int64 {
	if e.Type != EntryTypeServiceCharge || e.DoctorType == nil {
		return 0
	}

	switch *e.DoctorType {
	case DoctorTypeInternal:
		if e.CommissionPct == nil {
			return 0
		}
		return int64(float64(e.AmountCents) * (*e.CommissionPct / 100.0))
	case DoctorTypeExternal:
		if e.ExternalDoctorFeeCents == nil {
			return 0
		}
		return *e.ExternalDoctorFeeCents
	}

	return 0
}

// Helper functions for validation

// IsValidEntryType checks if the entry type is valid
func IsValidEntryType(t EntryType) bool {
	switch t {
	case EntryTypeServiceCharge, EntryTypeDiscount, EntryTypePayment, EntryTypeRefund, EntryTypeCorrection:
		return true
	}
	return false
}

// IsValidCurrency checks if the currency is valid
func IsValidCurrency(c Currency) bool {
	return c == CurrencyMXN || c == CurrencyUSD
}

// IsValidPaymentMethod checks if the payment method is valid
func IsValidPaymentMethod(pm PaymentMethod) bool {
	switch pm {
	case PaymentMethodCash, PaymentMethodCard, PaymentMethodTransfer:
		return true
	}
	return false
}

// IsValidDoctorType checks if the doctor type is valid
func IsValidDoctorType(dt DoctorType) bool {
	return dt == DoctorTypeInternal || dt == DoctorTypeExternal
}
