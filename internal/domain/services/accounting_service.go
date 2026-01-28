package services

import (
	"context"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// AccountingService provides accounting-related business logic
type AccountingService struct {
	accountRepo repositories.AppointmentAccountRepository
	entryRepo   repositories.AppointmentAccountEntryRepository
}

// NewAccountingService creates a new instance of AccountingService
func NewAccountingService(
	accountRepo repositories.AppointmentAccountRepository,
	entryRepo repositories.AppointmentAccountEntryRepository,
) *AccountingService {
	return &AccountingService{
		accountRepo: accountRepo,
		entryRepo:   entryRepo,
	}
}

// CreateOrGetAccount creates an appointment account if it doesn't exist, or returns existing one
func (s *AccountingService) CreateOrGetAccount(ctx context.Context, organizationID, appointmentID uuid.UUID) (*entities.AppointmentAccount, error) {
	return s.accountRepo.GetOrCreate(ctx, organizationID, appointmentID)
}

// ValidateEntry performs comprehensive validation on an entry before creation
func (s *AccountingService) ValidateEntry(ctx context.Context, entry *entities.AppointmentAccountEntry) error {
	// Basic entity validation
	if err := entry.Validate(); err != nil {
		return err
	}

	// Verify appointment account exists
	account, err := s.accountRepo.GetByID(ctx, entry.AppointmentAccountID)
	if err != nil {
		return fmt.Errorf("failed to verify appointment account: %w", err)
	}
	if account == nil {
		return entities.ErrAppointmentAccountNotFound
	}

	// If correcting an entry, verify the original entry exists
	if entry.Type == entities.EntryTypeCorrection && entry.CorrectsEntryID != nil {
		originalEntry, err := s.entryRepo.GetByID(ctx, *entry.CorrectsEntryID)
		if err != nil {
			return fmt.Errorf("failed to verify original entry: %w", err)
		}
		if originalEntry == nil {
			return fmt.Errorf("original entry not found for correction")
		}

		// Correction amount should be opposite sign of original
		if (originalEntry.AmountCents > 0 && entry.AmountCents > 0) ||
			(originalEntry.AmountCents < 0 && entry.AmountCents < 0) {
			return fmt.Errorf("correction amount must have opposite sign of original entry")
		}
	}

	return nil
}

// CalculateBalance calculates the balance for an appointment account
func (s *AccountingService) CalculateBalance(ctx context.Context, accountID uuid.UUID) (*repositories.AccountBalance, error) {
	return s.entryRepo.GetBalance(ctx, accountID)
}

// GetEntriesForAccount retrieves all entries for an appointment account
func (s *AccountingService) GetEntriesForAccount(ctx context.Context, accountID uuid.UUID) ([]*entities.AppointmentAccountEntry, error) {
	return s.entryRepo.GetByAccountID(ctx, accountID)
}

// CreateServiceCharge creates a service charge entry with doctor commission information
func (s *AccountingService) CreateServiceCharge(
	ctx context.Context,
	accountID uuid.UUID,
	doctorID uuid.UUID,
	doctorType entities.DoctorType,
	amountCents int64,
	description string,
	createdByUserID uuid.UUID,
	cashSessionID *uuid.UUID,
	serviceID *string,
	commissionPct *float64,
	externalDoctorFeeCents *int64,
) (*entities.AppointmentAccountEntry, error) {
	entry := &entities.AppointmentAccountEntry{
		ID:                     uuid.New(),
		AppointmentAccountID:   accountID,
		Type:                   entities.EntryTypeServiceCharge,
		Currency:               entities.CurrencyMXN, // Default, can be parameterized
		AmountCents:            amountCents,
		Description:            description,
		CreatedByUserID:        createdByUserID,
		DoctorID:               &doctorID,
		DoctorType:             &doctorType,
		ServiceID:              serviceID,
		CashSessionID:          cashSessionID,
		CommissionPct:          commissionPct,
		ExternalDoctorFeeCents: externalDoctorFeeCents,
		IsSensitive:            doctorType == entities.DoctorTypeExternal,
		Quantity:               1,
	}

	if err := s.ValidateEntry(ctx, entry); err != nil {
		return nil, err
	}

	if err := s.entryRepo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create service charge: %w", err)
	}

	return entry, nil
}

// CreatePayment creates a payment entry
func (s *AccountingService) CreatePayment(
	ctx context.Context,
	accountID uuid.UUID,
	paymentMethod entities.PaymentMethod,
	currency entities.Currency,
	amountCents int64,
	description string,
	createdByUserID uuid.UUID,
	cashSessionID *uuid.UUID,
	exchangeRate *float64,
) (*entities.AppointmentAccountEntry, error) {
	entry := &entities.AppointmentAccountEntry{
		ID:                   uuid.New(),
		AppointmentAccountID: accountID,
		Type:                 entities.EntryTypePayment,
		Currency:             currency,
		AmountCents:          amountCents,
		Description:          description,
		CreatedByUserID:      createdByUserID,
		PaymentMethod:        &paymentMethod,
		CashSessionID:        cashSessionID,
		ExchangeRateUsed:     exchangeRate,
		Quantity:             1,
	}

	if err := s.ValidateEntry(ctx, entry); err != nil {
		return nil, err
	}

	if err := s.entryRepo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return entry, nil
}

// CreateCorrection creates a correction entry that reverses a previous entry
func (s *AccountingService) CreateCorrection(
	ctx context.Context,
	originalEntryID uuid.UUID,
	description string,
	createdByUserID uuid.UUID,
	notes *string,
) (*entities.AppointmentAccountEntry, error) {
	// Get original entry
	originalEntry, err := s.entryRepo.GetByID(ctx, originalEntryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original entry: %w", err)
	}
	if originalEntry == nil {
		return nil, fmt.Errorf("original entry not found")
	}

	// Create correction with opposite amount
	entry := &entities.AppointmentAccountEntry{
		ID:                   uuid.New(),
		AppointmentAccountID: originalEntry.AppointmentAccountID,
		Type:                 entities.EntryTypeCorrection,
		Currency:             originalEntry.Currency,
		AmountCents:          -originalEntry.AmountCents, // Opposite sign
		Description:          description,
		CreatedByUserID:      createdByUserID,
		CorrectsEntryID:      &originalEntryID,
		Notes:                notes,
		Quantity:             1,
	}

	if err := s.ValidateEntry(ctx, entry); err != nil {
		return nil, err
	}

	if err := s.entryRepo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create correction: %w", err)
	}

	return entry, nil
}
