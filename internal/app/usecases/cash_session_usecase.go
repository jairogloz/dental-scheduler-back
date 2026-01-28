package usecases

import (
	"context"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/domain/services"

	"github.com/google/uuid"
)

// CashSessionUseCase handles cash session operations
type CashSessionUseCase struct {
	cashSessionService *services.CashSessionService
	entryRepo          repositories.AppointmentAccountEntryRepository
}

// NewCashSessionUseCase creates a new instance
func NewCashSessionUseCase(
	cashSessionService *services.CashSessionService,
	entryRepo repositories.AppointmentAccountEntryRepository,
) *CashSessionUseCase {
	return &CashSessionUseCase{
		cashSessionService: cashSessionService,
		entryRepo:          entryRepo,
	}
}

// OpenSessionInput contains parameters for opening a cash session
type OpenSessionInput struct {
	OrganizationID     uuid.UUID
	ClinicID           uuid.UUID
	UserID             uuid.UUID
	OpeningType        entities.CashSessionOpeningType
	StartingFloatCents int64
	Notes              *string
}

// OpenSession opens a new cash session
func (uc *CashSessionUseCase) OpenSession(ctx context.Context, input OpenSessionInput) (*entities.CashSession, error) {
	session, err := uc.cashSessionService.OpenSession(
		ctx,
		input.OrganizationID,
		input.ClinicID,
		input.UserID,
		input.OpeningType,
		input.StartingFloatCents,
		input.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open session: %w", err)
	}

	return session, nil
}

// GetCurrentSession retrieves the current open session for a user at a clinic
func (uc *CashSessionUseCase) GetCurrentSession(ctx context.Context, userID, clinicID uuid.UUID) (*entities.CashSession, error) {
	session, err := uc.cashSessionService.GetCurrentOpenSession(ctx, userID, clinicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current session: %w", err)
	}

	if session == nil {
		return nil, entities.ErrNoCashSessionOpen
	}

	return session, nil
}

// CashSessionDetailsOutput contains session with entries and expected amounts
type CashSessionDetailsOutput struct {
	Session         *entities.CashSession
	Entries         []*entities.AppointmentAccountEntry
	ExpectedAmounts map[entities.Currency]int64
	PaymentSummary  map[entities.PaymentMethod]map[entities.Currency]int64
}

// GetSessionDetails retrieves full details of a cash session
func (uc *CashSessionUseCase) GetSessionDetails(ctx context.Context, sessionID uuid.UUID) (*CashSessionDetailsOutput, error) {
	// Get session
	session, err := uc.cashSessionService.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		return nil, entities.ErrCashSessionNotFound
	}

	// Get all entries for this session
	entries, err := uc.entryRepo.GetByCashSessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// Calculate expected cash amounts
	expectedAmounts, err := uc.cashSessionService.CalculateExpectedCash(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate expected amounts: %w", err)
	}

	// Get payment summary by method and currency
	paymentSummary, err := uc.entryRepo.GetPaymentsByCashSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment summary: %w", err)
	}

	return &CashSessionDetailsOutput{
		Session:         session,
		Entries:         entries,
		ExpectedAmounts: expectedAmounts,
		PaymentSummary:  paymentSummary,
	}, nil
}

// CloseSession closes a cash session
func (uc *CashSessionUseCase) CloseSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := uc.cashSessionService.CloseSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}

	return nil
}

// ListSessions lists cash sessions with filters
func (uc *CashSessionUseCase) ListSessions(ctx context.Context, filters repositories.CashSessionFilters) ([]*entities.CashSession, error) {
	sessions, err := uc.cashSessionService.ListSessions(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}
