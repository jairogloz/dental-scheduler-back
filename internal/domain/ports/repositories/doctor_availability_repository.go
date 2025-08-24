package repositories

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// DoctorAvailabilityRepository defines the interface for doctor availability data operations
type DoctorAvailabilityRepository interface {
	// Create creates a new doctor availability entry
	Create(ctx context.Context, availability *entities.DoctorAvailability) error

	// GetByID retrieves a doctor availability by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.DoctorAvailability, error)

	// GetByDoctorID retrieves all availability entries for a doctor
	GetByDoctorID(ctx context.Context, doctorID uuid.UUID) ([]*entities.DoctorAvailability, error)

	// GetByDoctorIDAndDate retrieves availability for a doctor on a specific date
	GetByDoctorIDAndDate(ctx context.Context, doctorID uuid.UUID, date time.Time) ([]*entities.DoctorAvailability, error)

	// Update updates an existing doctor availability
	Update(ctx context.Context, availability *entities.DoctorAvailability) error

	// Delete deletes a doctor availability by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// IsAvailable checks if a doctor is available during a specific time range
	IsAvailable(ctx context.Context, doctorID uuid.UUID, startTime, endTime time.Time) (bool, error)
}
