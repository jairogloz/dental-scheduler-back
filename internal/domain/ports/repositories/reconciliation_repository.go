package repositories

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// ReconciliationFilters represents filters for reconciliation queries
type ReconciliationFilters struct {
	OrganizationID *uuid.UUID
	ClinicID       *uuid.UUID
	CashSessionID  *uuid.UUID
	UserID         *uuid.UUID
	PaymentMethod  *entities.PaymentMethod
	Currency       *entities.Currency
	Status         *entities.ReconciliationStatus
	StartDate      *time.Time
	EndDate        *time.Time
	HasDiscrepancy *bool
	Page           int
	Limit          int
}

// ReconciliationRepository defines the interface for reconciliation data operations
type ReconciliationRepository interface {
	// Create creates a new reconciliation
	Create(ctx context.Context, reconciliation *entities.Reconciliation) error

	// Update updates an existing reconciliation
	Update(ctx context.Context, reconciliation *entities.Reconciliation) error

	// GetByID retrieves a reconciliation by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Reconciliation, error)

	// GetByCashSessionID retrieves all reconciliations for a cash session
	GetByCashSessionID(ctx context.Context, cashSessionID uuid.UUID) ([]*entities.Reconciliation, error)

	// Exists checks if a reconciliation exists for specific session, payment method, and currency
	Exists(ctx context.Context, cashSessionID uuid.UUID, paymentMethod entities.PaymentMethod, currency entities.Currency) (bool, error)

	// List retrieves reconciliations with optional filters
	List(ctx context.Context, filters ReconciliationFilters) ([]*entities.Reconciliation, error)

	// GetDiscrepancies retrieves reconciliations with discrepancies
	GetDiscrepancies(ctx context.Context, clinicID uuid.UUID, startDate, endDate time.Time) ([]*entities.Reconciliation, error)
}
