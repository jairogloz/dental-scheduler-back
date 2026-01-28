package entities

import "errors"

// Domain errors
var (
	// Organization errors
	ErrInvalidOrganizationName = errors.New("organization name is required")
	ErrOrganizationNotFound    = errors.New("organization not found")
	ErrInvalidOrganizationID   = errors.New("organization ID is required")

	// Profile errors
	ErrInvalidProfileID = errors.New("profile ID is required")
	ErrInvalidRoles     = errors.New("at least one role is required")
	ErrProfileNotFound  = errors.New("profile not found")

	// Clinic errors
	ErrInvalidClinicName = errors.New("clinic name is required")
	ErrClinicNotFound    = errors.New("clinic not found")

	// Unit errors
	ErrInvalidUnitName = errors.New("unit name is required")
	ErrInvalidClinicID = errors.New("clinic ID is required")
	ErrUnitNotFound    = errors.New("unit not found")

	// Doctor errors
	ErrInvalidDoctorName              = errors.New("doctor name is required")
	ErrDoctorNotFound                 = errors.New("doctor not found")
	ErrInvalidEmail                   = errors.New("invalid email format")
	ErrInvalidColor                   = errors.New("invalid color format (must be hex color like #3B82F6)")
	ErrDoctorUnitOrganizationMismatch = errors.New("doctor's default unit must belong to the same organization")

	// Patient errors
	ErrInvalidPatientName = errors.New("patient name is required")
	ErrPatientNotFound    = errors.New("patient not found")

	// Appointment errors
	ErrInvalidPatientID           = errors.New("patient ID is required")
	ErrInvalidDoctorID            = errors.New("doctor ID is required")
	ErrInvalidUnitID              = errors.New("unit ID is required")
	ErrInvalidAppointmentTime     = errors.New("invalid appointment time")
	ErrEndTimeBeforeStartTime     = errors.New("end time must be after start time")
	ErrAppointmentNotFound        = errors.New("appointment not found")
	ErrAppointmentConflict        = errors.New("appointment conflicts with existing appointment")
	ErrPastAppointmentTime        = errors.New("appointment time cannot be in the past")
	ErrInvalidAppointmentStatus   = errors.New("invalid appointment status")
	ErrAppointmentNotInQueue      = errors.New("appointment is not in rescheduling queue")
	ErrInvalidStatusTransition    = errors.New("invalid appointment status transition")
	ErrCancellationReasonRequired = errors.New("cancellation reason is required")
	ErrUnauthorizedAccess         = errors.New("unauthorized access to resource")

	// Doctor Availability errors
	ErrInvalidAvailabilityTime = errors.New("invalid availability time")
	ErrAvailabilityNotFound    = errors.New("availability not found")
	ErrDoctorNotAvailable      = errors.New("doctor is not available at the requested time")

	// General errors
	ErrInvalidID = errors.New("invalid ID format")

	// Accounting errors - Appointment Account
	ErrInvalidAppointmentAccountID = errors.New("appointment account ID is required")
	ErrInvalidAppointmentID        = errors.New("appointment ID is required")
	ErrAppointmentAccountNotFound  = errors.New("appointment account not found")

	// Accounting errors - Appointment Account Entry
	ErrInvalidEntryType                   = errors.New("invalid entry type")
	ErrInvalidCurrency                    = errors.New("invalid currency")
	ErrAmountCannotBeZero                 = errors.New("amount cannot be zero")
	ErrDescriptionRequired                = errors.New("description is required")
	ErrCreatedByUserRequired              = errors.New("created by user ID is required")
	ErrExchangeRateRequired               = errors.New("exchange rate is required for USD transactions")
	ErrPaymentMethodRequired              = errors.New("payment method is required for payments")
	ErrInvalidPaymentMethod               = errors.New("invalid payment method")
	ErrPaymentAmountMustBePositive        = errors.New("payment amount must be positive")
	ErrDoctorIDRequired                   = errors.New("doctor ID is required for service charges")
	ErrDoctorTypeRequired                 = errors.New("doctor type is required for service charges")
	ErrInvalidDoctorType                  = errors.New("invalid doctor type")
	ErrServiceChargeAmountMustBePositive  = errors.New("service charge amount must be positive")
	ErrCommissionPctRequired              = errors.New("commission percentage is required for internal doctors")
	ErrInvalidCommissionPct               = errors.New("commission percentage must be between 0 and 100")
	ErrExternalDoctorFeeRequired          = errors.New("external doctor fee is required for external doctors")
	ErrInvalidExternalDoctorFee           = errors.New("external doctor fee must be non-negative")
	ErrDiscountRefundAmountMustBeNegative = errors.New("discount and refund amounts must be negative")
	ErrCorrectsEntryIDRequired            = errors.New("corrects entry ID is required for corrections")
	ErrInvalidQuantity                    = errors.New("quantity must be non-negative")

	// Accounting errors - Cash Session
	ErrInvalidCashSessionID          = errors.New("cash session ID is required")
	ErrInvalidCashSessionStatus      = errors.New("invalid cash session status")
	ErrInvalidCashSessionOpeningType = errors.New("invalid cash session opening type")
	ErrUserIDRequired                = errors.New("user ID is required")
	ErrInvalidStartingFloat          = errors.New("starting float cannot be negative")
	ErrClosedAtRequired              = errors.New("closed_at is required for closed sessions")
	ErrClosedAtMustBeNull            = errors.New("closed_at must be null for open sessions")
	ErrCashSessionNotFound           = errors.New("cash session not found")
	ErrCashSessionAlreadyClosed      = errors.New("cash session is already closed")
	ErrCashSessionAlreadyOpen        = errors.New("user already has an open cash session for this clinic")
	ErrNoCashSessionOpen             = errors.New("no cash session is currently open")

	// Accounting errors - Reconciliation
	ErrReconciledByUserRequired    = errors.New("reconciled by user ID is required")
	ErrInvalidReconciliationStatus = errors.New("invalid reconciliation status")
	ErrInvalidFloatLeft            = errors.New("float left cannot be negative")
	ErrInvalidDepositedAmount      = errors.New("deposited amount must equal actual amount minus float left")
	ErrInvalidDiscrepancyAmount    = errors.New("discrepancy amount must equal actual amount minus expected amount")
	ErrReconciliationNotFound      = errors.New("reconciliation not found")
	ErrReconciliationAlreadyExists = errors.New("reconciliation already exists for this session, payment method, and currency")
)
