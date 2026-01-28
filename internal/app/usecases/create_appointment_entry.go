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
	OrganizationID         uuid.UUID
	AppointmentID          uuid.UUID
	DoctorID               uuid.UUID
	DoctorType             entities.DoctorType
	Currency               entities.Currency
	AmountCents            int64
	Description            string
	CreatedByUserID        uuid.UUID
	ServiceID              *string
	CommissionPct          *float64
	ExternalDoctorFeeCents *int64
	ClinicID               *uuid.UUID // For auto-creating cash session if needed
	UserID                 *uuid.UUID // For auto-creating cash session if needed
}

// CreatePaymentInput contains parameters for creating a payment
type CreatePaymentInput struct {
	OrganizationID  uuid.UUID
	AppointmentID   uuid.UUID
	PaymentMethod   entities.PaymentMethod
	Currency        entities.Currency
	AmountCents     int64
	Description     string
	CreatedByUserID uuid.UUID
	ExchangeRate    *float64
	ClinicID        uuid.UUID // Required for cash session lookup
	UserID          uuid.UUID // Required for cash session lookup
}

// Execute creates a service charge entry
func (uc *CreateAppointmentEntryUseCase) CreateServiceCharge(ctx context.Context, input CreateServiceChargeInput) (*entities.AppointmentAccountEntry, error) {
	// Get or create appointment account
	account, err := uc.accountingService.CreateOrGetAccount(ctx, input.OrganizationID, input.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create account: %w", err)
	}

	// For cash payments, get or create cash session
	var cashSessionID *uuid.UUID
	if input.ClinicID != nil && input.UserID != nil {
		session, err := uc.cashSessionService.GetOrCreateOpenSession(ctx, input.OrganizationID, *input.ClinicID, *input.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create cash session: %w", err)
		}
		cashSessionID = &session.ID
	}

	// Create the service charge
	entry, err := uc.accountingService.CreateServiceCharge(
		ctx,
		account.ID,
		input.DoctorID,
		input.DoctorType,
		input.AmountCents,
		input.Description,
		input.CreatedByUserID,
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
	// Get or create appointment account
	account, err := uc.accountingService.CreateOrGetAccount(ctx, input.OrganizationID, input.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create account: %w", err)
	}

	// For cash payments, get or create cash session
	var cashSessionID *uuid.UUID
	if input.PaymentMethod == entities.PaymentMethodCash {
		session, err := uc.cashSessionService.GetOrCreateOpenSession(ctx, input.OrganizationID, input.ClinicID, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create cash session: %w", err)
		}
		cashSessionID = &session.ID
	}

	// Create the payment
	entry, err := uc.accountingService.CreatePayment(
		ctx,
		account.ID,
		input.PaymentMethod,
		input.Currency,
		input.AmountCents,
		input.Description,
		input.CreatedByUserID,
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
