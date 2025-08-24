package repositories

import (
	"context"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// PatientRepository defines the interface for patient data operations
type PatientRepository interface {
	// Create creates a new patient
	Create(ctx context.Context, patient *entities.Patient) error

	// GetByID retrieves a patient by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Patient, error)

	// GetAll retrieves all patients
	GetAll(ctx context.Context) ([]*entities.Patient, error)

	// GetByEmail retrieves a patient by email
	GetByEmail(ctx context.Context, email string) (*entities.Patient, error)

	// Update updates an existing patient
	Update(ctx context.Context, patient *entities.Patient) error

	// Delete deletes a patient by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// Exists checks if a patient exists by its ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
