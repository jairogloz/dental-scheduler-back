package repositories

import (
	"context"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// UnitRepository defines the interface for unit data operations
type UnitRepository interface {
	// Create creates a new unit
	Create(ctx context.Context, unit *entities.Unit) error

	// GetByID retrieves a unit by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Unit, error)

	// GetAll retrieves all units
	GetAll(ctx context.Context) ([]*entities.Unit, error)

	// GetByClinicID retrieves all units for a specific clinic
	GetByClinicID(ctx context.Context, clinicID uuid.UUID) ([]*entities.Unit, error)

	// Update updates an existing unit
	Update(ctx context.Context, unit *entities.Unit) error

	// Delete deletes a unit by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a unit exists by its ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
