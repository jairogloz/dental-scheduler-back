package usecases

import (
	"context"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/services"

	"github.com/google/uuid"
)

// ReconciliationUseCase handles reconciliation operations
type ReconciliationUseCase struct {
	reconciliationService *services.ReconciliationService
	cashSessionService    *services.CashSessionService
}

// NewReconciliationUseCase creates a new instance
func NewReconciliationUseCase(
	reconciliationService *services.ReconciliationService,
	cashSessionService *services.CashSessionService,
) *ReconciliationUseCase {
	return &ReconciliationUseCase{
		reconciliationService: reconciliationService,
		cashSessionService:    cashSessionService,
	}
}

// ReconciliationPreviewOutput contains expected amounts for reconciliation
type ReconciliationPreviewOutput struct {
	Session                   *entities.CashSession
	ExpectedAmountsByCurrency map[entities.Currency]int64
	ExistingReconciliations   []*entities.Reconciliation
}

// GetReconciliationPreview calculates expected amounts for a cash session
func (uc *ReconciliationUseCase) GetReconciliationPreview(ctx context.Context, sessionID uuid.UUID) (*ReconciliationPreviewOutput, error) {
	// Get session
	session, err := uc.cashSessionService.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session == nil {
		return nil, entities.ErrCashSessionNotFound
	}

	// Get reconciliation data
	data, err := uc.reconciliationService.PrepareReconciliationData(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare reconciliation data: %w", err)
	}

	return &ReconciliationPreviewOutput{
		Session:                   data.Session,
		ExpectedAmountsByCurrency: data.ExpectedAmountsByCurrency,
		ExistingReconciliations:   data.ExistingReconciliations,
	}, nil
}

// CreateReconciliationInput contains parameters for creating a reconciliation
type CreateReconciliationInput struct {
	CashSessionID      uuid.UUID
	OrganizationID     uuid.UUID
	ClinicID           uuid.UUID
	PaymentMethod      entities.PaymentMethod
	Currency           entities.Currency
	ExpectedCents      int64
	ActualCents        int64
	FloatLeftCents     int64
	DepositedCents     int64
	ReconciledByUserID uuid.UUID
	Notes              *string
}

// CreateReconciliation creates a reconciliation record
func (uc *ReconciliationUseCase) CreateReconciliation(ctx context.Context, input CreateReconciliationInput) (*entities.Reconciliation, error) {
	reconciliation, err := uc.reconciliationService.CreateReconciliation(
		ctx,
		input.CashSessionID,
		input.OrganizationID,
		input.ClinicID,
		input.PaymentMethod,
		input.Currency,
		input.ExpectedCents,
		input.ActualCents,
		input.FloatLeftCents,
		input.DepositedCents,
		input.ReconciledByUserID,
		input.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reconciliation: %w", err)
	}

	return reconciliation, nil
}

// GetReconciliation retrieves a reconciliation by ID
func (uc *ReconciliationUseCase) GetReconciliation(ctx context.Context, reconciliationID uuid.UUID) (*entities.Reconciliation, error) {
	reconciliation, err := uc.reconciliationService.GetReconciliationByID(ctx, reconciliationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reconciliation: %w", err)
	}

	if reconciliation == nil {
		return nil, entities.ErrReconciliationNotFound
	}

	return reconciliation, nil
}

// GetReconciliationsByCashSession retrieves all reconciliations for a session
func (uc *ReconciliationUseCase) GetReconciliationsByCashSession(ctx context.Context, cashSessionID uuid.UUID) ([]*entities.Reconciliation, error) {
	reconciliations, err := uc.reconciliationService.GetReconciliationsByCashSession(ctx, cashSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reconciliations: %w", err)
	}

	return reconciliations, nil
}

// GetDiscrepancies retrieves reconciliations with discrepancies for a clinic
func (uc *ReconciliationUseCase) GetDiscrepancies(ctx context.Context, clinicID uuid.UUID, startDate, endDate time.Time) ([]*entities.Reconciliation, error) {
	reconciliations, err := uc.reconciliationService.GetDiscrepancies(ctx, clinicID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get discrepancies: %w", err)
	}

	return reconciliations, nil
}

// ValidateReconciliationAmounts validates reconciliation amounts before submission
func (uc *ReconciliationUseCase) ValidateReconciliationAmounts(
	expectedAmountCents int64,
	actualAmountCents int64,
	floatLeftCents int64,
	depositedCents int64,
) error {
	return uc.reconciliationService.ValidateReconciliationAmounts(
		expectedAmountCents,
		actualAmountCents,
		floatLeftCents,
		depositedCents,
	)
}
