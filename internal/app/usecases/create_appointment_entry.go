package usecases

import (
	"context"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/services"

	"github.com/google/uuid"
)

// CreateAppointmentEntryUseCase handles creating entries for appointment accounts
type CreateAppointmentEntryUseCase struct {
	accountingService  *services.AccountingService
	cashSessionService *services.CashSessionService
}

// NewCreateAppointmentEntryUseCase creates a new instance
func NewCreateAppointmentEntryUseCase(
	accountingService *services.AccountingService,
	cashSessionService *services.CashSessionService,
) *CreateAppointmentEntryUseCase {
	return &CreateAppointmentEntryUseCase{
		accountingService:  accountingService,
		cashSessionService: cashSessionService,
	}
}

// CreateServiceChargeInput contains parameters for creating a service charge
type CreateServiceChargeInput struct {
	OrganizationID         string
	AppointmentID          string
	DoctorID               string
	DoctorType             entities.DoctorType
	Currency               entities.Currency
	AmountCents            int64
	Description            string
	CreatedByUserID        string
	ServiceID              *string
	CommissionPct          *float64
	ExternalDoctorFeeCents *int64
	ClinicID               *string // For auto-creating cash session if needed
	UserID                 *string // For auto-creating cash session if needed
}

// CreatePaymentInput contains parameters for creating a payment
type CreatePaymentInput struct {
	OrganizationID  string
	AppointmentID   string
	PaymentMethod   entities.PaymentMethod
	Currency        entities.Currency
	AmountCents     int64
	Description     string
	CreatedByUserID string
	ExchangeRate    *float64
	ClinicID        string // Required for cash session lookup
	UserID          string // Required for cash session lookup
}

// Execute creates a service charge entry
func (uc *CreateAppointmentEntryUseCase) CreateServiceCharge(ctx context.Context, input CreateServiceChargeInput) (*entities.AppointmentAccountEntry, error) {
	// Parse UUIDs
	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	apptID, err := uuid.Parse(input.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid appointment ID: %w", err)
	}
	doctorID, err := uuid.Parse(input.DoctorID)
	if err != nil {
		return nil, fmt.Errorf("invalid doctor ID: %w", err)
	}
	userID, err := uuid.Parse(input.CreatedByUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Get or create appointment account
	account, err := uc.accountingService.CreateOrGetAccount(ctx, orgID, apptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create account: %w", err)
	}

	// For cash payments, get or create cash session
	var cashSessionID *uuid.UUID
	if input.ClinicID != nil && input.UserID != nil {
		clinicID, err := uuid.Parse(*input.ClinicID)
		if err != nil {
			return nil, fmt.Errorf("invalid clinic ID: %w", err)
		}
		sessionUserID, err := uuid.Parse(*input.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID for session: %w", err)
		}
		session, err := uc.cashSessionService.GetOrCreateOpenSession(ctx, orgID, clinicID, sessionUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create cash session: %w", err)
		}
		cashSessionID = &session.ID
	}

	// Create the service charge
	entry, err := uc.accountingService.CreateServiceCharge(
		ctx,
		account.ID,
		doctorID,
		input.DoctorType,
		input.AmountCents,
		input.Description,
		userID,
		cashSessionID,
		input.ServiceID,
		input.CommissionPct,
		input.ExternalDoctorFeeCents,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create service charge: %w", err)
	}

	return entry, nil
}

// CreatePayment creates a payment entry
func (uc *CreateAppointmentEntryUseCase) CreatePayment(ctx context.Context, input CreatePaymentInput) (*entities.AppointmentAccountEntry, error) {
	// Parse UUIDs
	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	apptID, err := uuid.Parse(input.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid appointment ID: %w", err)
	}
	userID, err := uuid.Parse(input.CreatedByUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	clinicID, err := uuid.Parse(input.ClinicID)
	if err != nil {
		return nil, fmt.Errorf("invalid clinic ID: %w", err)
	}
	sessionUserID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID for session: %w", err)
	}

	// Get or create appointment account
	account, err := uc.accountingService.CreateOrGetAccount(ctx, orgID, apptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create account: %w", err)
	}

	// For all payments, attach to the active cash session so session details
	// include full payment summary by method and currency.
	var cashSessionID *uuid.UUID
	session, err := uc.cashSessionService.GetOrCreateOpenSession(ctx, orgID, clinicID, sessionUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create cash session: %w", err)
	}
	cashSessionID = &session.ID

	// Create the payment
	entry, err := uc.accountingService.CreatePayment(
		ctx,
		account.ID,
		input.PaymentMethod,
		input.Currency,
		input.AmountCents,
		input.Description,
		userID,
		cashSessionID,
		input.ExchangeRate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return entry, nil
}

// CreateCorrectionInput contains parameters for creating a correction
type CreateCorrectionInput struct {
	OriginalEntryID uuid.UUID
	Description     string
	CreatedByUserID uuid.UUID
	Notes           *string
}

// CreateCorrection creates a correction entry that reverses a previous entry
func (uc *CreateAppointmentEntryUseCase) CreateCorrection(ctx context.Context, input CreateCorrectionInput) (*entities.AppointmentAccountEntry, error) {
	entry, err := uc.accountingService.CreateCorrection(
		ctx,
		input.OriginalEntryID,
		input.Description,
		input.CreatedByUserID,
		input.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create correction: %w", err)
	}

	return entry, nil
}
