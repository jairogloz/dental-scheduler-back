package repositories

import (
	"context"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// ClinicRepository defines the interface for clinic data operations
type ClinicRepository interface {
	// Create creates a new clinic
	Create(ctx context.Context, clinic *entities.Clinic) error

	// GetByID retrieves a clinic by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Clinic, error)

	// GetAll retrieves all clinics
	GetAll(ctx context.Context) ([]*entities.Clinic, error)

	// Update updates an existing clinic
	Update(ctx context.Context, clinic *entities.Clinic) error

	// Delete deletes a clinic by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a clinic exists by its ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
