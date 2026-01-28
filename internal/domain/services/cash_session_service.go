package services

import (
	"context"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// CashSessionService provides cash session management business logic
type CashSessionService struct {
	cashSessionRepo repositories.CashSessionRepository
	entryRepo       repositories.AppointmentAccountEntryRepository
}

// NewCashSessionService creates a new instance of CashSessionService
func NewCashSessionService(
	cashSessionRepo repositories.CashSessionRepository,
	entryRepo repositories.AppointmentAccountEntryRepository,
) *CashSessionService {
	return &CashSessionService{
		cashSessionRepo: cashSessionRepo,
		entryRepo:       entryRepo,
	}
}

// OpenSession opens a new cash session with validation
func (s *CashSessionService) OpenSession(
	ctx context.Context,
	organizationID uuid.UUID,
	clinicID uuid.UUID,
	userID uuid.UUID,
	openingType entities.CashSessionOpeningType,
	startingFloatCents int64,
	notes *string,
) (*entities.CashSession, error) {
	// Check if there's already an open session for this user at this clinic
	hasOpen, err := s.cashSessionRepo.HasOpenSession(ctx, userID, clinicID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for open sessions: %w", err)
	}
	if hasOpen {
		return nil, entities.ErrCashSessionAlreadyOpen
	}

	// Validate opening type
	if !entities.IsValidCashSessionOpeningType(openingType) {
		return nil, entities.ErrInvalidCashSessionOpeningType
	}

	// Validate starting float
	if startingFloatCents < 0 {
		return nil, entities.ErrInvalidStartingFloat
	}

	session := &entities.CashSession{
		ID:                 uuid.New(),
		OrganizationID:     organizationID,
		ClinicID:           clinicID,
		UserID:             userID,
		OpenedAt:           time.Now(),
		OpeningType:        openingType,
		StartingFloatCents: startingFloatCents,
		Status:             entities.CashSessionStatusOpen,
		Notes:              notes,
	}

	if err := s.cashSessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create cash session: %w", err)
	}

	return session, nil
}

// GetOrCreateOpenSession gets the current open session or creates one automatically
// This is used when a cash payment is made but no session exists
func (s *CashSessionService) GetOrCreateOpenSession(
	ctx context.Context,
	organizationID uuid.UUID,
	clinicID uuid.UUID,
	userID uuid.UUID,
) (*entities.CashSession, error) {
	// Try to get existing open session
	session, err := s.cashSessionRepo.GetCurrentOpenSession(ctx, userID, clinicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current session: %w", err)
	}

	if session != nil {
		return session, nil
	}

	// Auto-create a new session with zero float
	session = &entities.CashSession{
		ID:                 uuid.New(),
		OrganizationID:     organizationID,
		ClinicID:           clinicID,
		UserID:             userID,
		OpenedAt:           time.Now(),
		OpeningType:        entities.CashSessionOpeningTypeAuto,
		StartingFloatCents: 0,
		Status:             entities.CashSessionStatusOpen,
	}

	if err := s.cashSessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to auto-create cash session: %w", err)
	}

	return session, nil
}

// CloseSession closes a cash session
// Note: Reconciliation is created separately and added via AddReconciliation
func (s *CashSessionService) CloseSession(
	ctx context.Context,
	sessionID uuid.UUID,
) error {
	// Get session
	session, err := s.cashSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return entities.ErrCashSessionNotFound
	}

	// Validate session is open
	if session.Status != entities.CashSessionStatusOpen {
		return entities.ErrCashSessionAlreadyClosed
	}

	// Close in database
	if err := s.cashSessionRepo.Close(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}

	return nil
}

// GetCurrentOpenSession gets the currently open session for a user at a clinic
func (s *CashSessionService) GetCurrentOpenSession(ctx context.Context, userID, clinicID uuid.UUID) (*entities.CashSession, error) {
	return s.cashSessionRepo.GetCurrentOpenSession(ctx, userID, clinicID)
}

// GetSessionByID retrieves a cash session by ID
func (s *CashSessionService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*entities.CashSession, error) {
	return s.cashSessionRepo.GetByID(ctx, sessionID)
}

// ListSessions lists cash sessions with optional filters
func (s *CashSessionService) ListSessions(
	ctx context.Context,
	filters repositories.CashSessionFilters,
) ([]*entities.CashSession, error) {
	return s.cashSessionRepo.List(ctx, filters)
}

// CalculateExpectedCash calculates the expected cash amounts for a session
// Returns: map[currency]amountCents
func (s *CashSessionService) CalculateExpectedCash(
	ctx context.Context,
	sessionID uuid.UUID,
) (map[entities.Currency]int64, error) {
	// Get all cash payments for this session
	payments, err := s.entryRepo.GetPaymentsByCashSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	// Extract cash payments only
	cashPayments, ok := payments[entities.PaymentMethodCash]
	if !ok {
		// No cash payments
		return make(map[entities.Currency]int64), nil
	}

	return cashPayments, nil
}
