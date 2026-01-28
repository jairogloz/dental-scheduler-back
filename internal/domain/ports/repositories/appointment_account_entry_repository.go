package repositories

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// AppointmentAccountEntryFilters represents filters for entry queries
type AppointmentAccountEntryFilters struct {
	AppointmentAccountID *uuid.UUID
	CashSessionID        *uuid.UUID
	DoctorID             *uuid.UUID
	Type                 *entities.EntryType
	PaymentMethod        *entities.PaymentMethod
	Currency             *entities.Currency
	CreatedByUserID      *uuid.UUID
	StartDate            *time.Time
	EndDate              *time.Time
	Page                 int
	Limit                int
}

// AccountBalance represents the balance summary of an appointment account
type AccountBalance struct {
	TotalChargesCents   int64                            `json:"total_charges_cents"`
	TotalDiscountsCents int64                            `json:"total_discounts_cents"` // negative value
	TotalPaymentsCents  int64                            `json:"total_payments_cents"`
	TotalRefundsCents   int64                            `json:"total_refunds_cents"` // negative value
	BalanceDueCents     int64                            `json:"balance_due_cents"`
	PaymentsByCurrency  map[entities.Currency]int64      `json:"payments_by_currency"`
	PaymentsByMethod    map[entities.PaymentMethod]int64 `json:"payments_by_method"`
}

// AppointmentAccountEntryRepository defines the interface for entry data operations
type AppointmentAccountEntryRepository interface {
	// Create creates a new entry (immutable - no update or delete)
	Create(ctx context.Context, entry *entities.AppointmentAccountEntry) error

	// GetByID retrieves an entry by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.AppointmentAccountEntry, error)

	// List retrieves entries with optional filters
	List(ctx context.Context, filters AppointmentAccountEntryFilters) ([]*entities.AppointmentAccountEntry, error)

	// GetByAccountID retrieves all entries for an appointment account
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entities.AppointmentAccountEntry, error)

	// GetByCashSessionID retrieves all entries for a cash session
	GetByCashSessionID(ctx context.Context, cashSessionID uuid.UUID) ([]*entities.AppointmentAccountEntry, error)

	// GetBalance calculates the balance for an appointment account
	GetBalance(ctx context.Context, accountID uuid.UUID) (*AccountBalance, error)

	// GetPaymentsByCashSession retrieves payment entries grouped by payment method and currency for a session
	GetPaymentsByCashSession(ctx context.Context, cashSessionID uuid.UUID) (map[entities.PaymentMethod]map[entities.Currency]int64, error)

	// GetCorrections retrieves all correction entries for a specific entry
	GetCorrections(ctx context.Context, entryID uuid.UUID) ([]*entities.AppointmentAccountEntry, error)
}
