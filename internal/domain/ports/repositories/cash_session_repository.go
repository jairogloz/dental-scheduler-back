package repositories

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CashSessionFilters represents filters for cash session queries
type CashSessionFilters struct {
	OrganizationID *uuid.UUID
	ClinicID       *uuid.UUID
	UserID         *uuid.UUID
	Status         *entities.CashSessionStatus
	StartDate      *time.Time
	EndDate        *time.Time
	Page           int
	Limit          int
}

// CashSessionRepository defines the interface for cash session data operations
type CashSessionRepository interface {
	// Create creates a new cash session
	Create(ctx context.Context, session *entities.CashSession) error

	// Update updates an existing cash session
	Update(ctx context.Context, session *entities.CashSession) error

	// GetByID retrieves a cash session by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.CashSession, error)

	// GetCurrentOpenSession retrieves the currently open session for a user at a clinic
	GetCurrentOpenSession(ctx context.Context, userID, clinicID uuid.UUID) (*entities.CashSession, error)

	// HasOpenSession checks if a user has an open session at a clinic
	HasOpenSession(ctx context.Context, userID, clinicID uuid.UUID) (bool, error)

	// List retrieves cash sessions with optional filters
	List(ctx context.Context, filters CashSessionFilters) ([]*entities.CashSession, error)

	// Close closes a cash session
	Close(ctx context.Context, sessionID uuid.UUID) error
}
