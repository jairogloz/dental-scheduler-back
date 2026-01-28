package services

import (
	"context"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// ReconciliationService provides reconciliation business logic
type ReconciliationService struct {
	reconciliationRepo repositories.ReconciliationRepository
	cashSessionRepo    repositories.CashSessionRepository
	entryRepo          repositories.AppointmentAccountEntryRepository
}

// NewReconciliationService creates a new instance of ReconciliationService
func NewReconciliationService(
	reconciliationRepo repositories.ReconciliationRepository,
	cashSessionRepo repositories.CashSessionRepository,
	entryRepo repositories.AppointmentAccountEntryRepository,
) *ReconciliationService {
	return &ReconciliationService{
		reconciliationRepo: reconciliationRepo,
		cashSessionRepo:    cashSessionRepo,
		entryRepo:          entryRepo,
	}
}

// CalculateExpectedAmounts calculates expected cash amounts for a session
// Returns expected amounts by currency based on all cash payments in the session
func (s *ReconciliationService) CalculateExpectedAmounts(
	ctx context.Context,
	sessionID uuid.UUID,
) (map[entities.Currency]int64, error) {
	// Get all payments for this cash session
	payments, err := s.entryRepo.GetPaymentsByCashSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	// Extract cash payments only
	cashPayments, ok := payments[entities.PaymentMethodCash]
	if !ok {
		// No cash payments, return empty map
		return make(map[entities.Currency]int64), nil
	}

	return cashPayments, nil
}

// CreateReconciliation creates a reconciliation for a cash session
func (s *ReconciliationService) CreateReconciliation(
	ctx context.Context,
	cashSessionID uuid.UUID,
	organizationID uuid.UUID,
	clinicID uuid.UUID,
	paymentMethod entities.PaymentMethod,
	currency entities.Currency,
	expectedAmountCents int64,
	actualAmountCents int64,
	floatLeftCents int64,
	depositedCents int64,
	reconciledByUserID uuid.UUID,
	notes *string,
) (*entities.Reconciliation, error) {
	// Verify cash session exists and is open
	session, err := s.cashSessionRepo.GetByID(ctx, cashSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash session: %w", err)
	}
	if session == nil {
		return nil, entities.ErrCashSessionNotFound
	}
	if session.Status != entities.CashSessionStatusOpen {
		return nil, entities.ErrCashSessionAlreadyClosed
	}

	// Check if reconciliation already exists for this session, payment method, and currency
	exists, err := s.reconciliationRepo.Exists(ctx, cashSessionID, paymentMethod, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing reconciliation: %w", err)
	}
	if exists {
		return nil, entities.ErrReconciliationAlreadyExists
	}

	// Create reconciliation
	reconciliation := &entities.Reconciliation{
		ID:                  uuid.New(),
		CashSessionID:       cashSessionID,
		OrganizationID:      organizationID,
		ClinicID:            clinicID,
		PaymentMethod:       paymentMethod,
		Currency:            currency,
		ExpectedAmountCents: expectedAmountCents,
		ActualAmountCents:   actualAmountCents,
		FloatLeftCents:      floatLeftCents,
		DepositedCents:      depositedCents,
		DiscrepancyCents:    actualAmountCents - expectedAmountCents,
		ReconciledByUserID:  reconciledByUserID,
		Status:              entities.ReconciliationStatusPending,
		Notes:               notes,
	}

	// Validate reconciliation
	if err := reconciliation.Validate(); err != nil {
		return nil, err
	}

	// Create in database
	if err := s.reconciliationRepo.Create(ctx, reconciliation); err != nil {
		return nil, fmt.Errorf("failed to create reconciliation: %w", err)
	}

	return reconciliation, nil
}

// GetReconciliationByID retrieves a reconciliation by ID
func (s *ReconciliationService) GetReconciliationByID(ctx context.Context, reconciliationID uuid.UUID) (*entities.Reconciliation, error) {
	return s.reconciliationRepo.GetByID(ctx, reconciliationID)
}

// GetReconciliationsByCashSession retrieves all reconciliations for a cash session
func (s *ReconciliationService) GetReconciliationsByCashSession(
	ctx context.Context,
	cashSessionID uuid.UUID,
) ([]*entities.Reconciliation, error) {
	return s.reconciliationRepo.GetByCashSessionID(ctx, cashSessionID)
}

// GetDiscrepancies retrieves reconciliations with discrepancies for a clinic in a date range
func (s *ReconciliationService) GetDiscrepancies(
	ctx context.Context,
	clinicID uuid.UUID,
	startDate, endDate time.Time,
) ([]*entities.Reconciliation, error) {
	return s.reconciliationRepo.GetDiscrepancies(ctx, clinicID, startDate, endDate)
}

// ValidateReconciliationAmounts validates that reconciliation amounts make sense
func (s *ReconciliationService) ValidateReconciliationAmounts(
	expectedAmountCents int64,
	actualAmountCents int64,
	floatLeftCents int64,
	depositedCents int64,
) error {
	// Deposited must equal actual minus float
	if depositedCents != (actualAmountCents - floatLeftCents) {
		return entities.ErrInvalidDepositedAmount
	}

	// Float left cannot be negative
	if floatLeftCents < 0 {
		return entities.ErrInvalidFloatLeft
	}

	return nil
}

// PrepareReconciliationData prepares reconciliation data for a cash session
// Returns expected amounts by currency and session details
func (s *ReconciliationService) PrepareReconciliationData(
	ctx context.Context,
	sessionID uuid.UUID,
) (*ReconciliationData, error) {
	// Get cash session
	session, err := s.cashSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash session: %w", err)
	}
	if session == nil {
		return nil, entities.ErrCashSessionNotFound
	}

	// Calculate expected amounts
	expectedAmounts, err := s.CalculateExpectedAmounts(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get existing reconciliations
	existingReconciliations, err := s.reconciliationRepo.GetByCashSessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing reconciliations: %w", err)
	}

	return &ReconciliationData{
		Session:                   session,
		ExpectedAmountsByCurrency: expectedAmounts,
		ExistingReconciliations:   existingReconciliations,
	}, nil
}

// ReconciliationData contains prepared data for reconciliation
type ReconciliationData struct {
	Session                   *entities.CashSession
	ExpectedAmountsByCurrency map[entities.Currency]int64
	ExistingReconciliations   []*entities.Reconciliation
}
