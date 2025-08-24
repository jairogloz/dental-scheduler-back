package repositories

import (
	"context"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// DoctorRepository defines the interface for doctor data operations
type DoctorRepository interface {
	// Create creates a new doctor
	Create(ctx context.Context, doctor *entities.Doctor) error

	// GetByID retrieves a doctor by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Doctor, error)

	// GetAll retrieves all doctors
	GetAll(ctx context.Context) ([]*entities.Doctor, error)

	// GetByEmail retrieves a doctor by email
	GetByEmail(ctx context.Context, email string) (*entities.Doctor, error)

	// Update updates an existing doctor
	Update(ctx context.Context, doctor *entities.Doctor) error

	// Delete deletes a doctor by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a doctor exists by its ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
