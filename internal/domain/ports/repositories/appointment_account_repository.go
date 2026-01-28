package repositories

import (
	"context"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// AppointmentAccountRepository defines the interface for appointment account data operations
type AppointmentAccountRepository interface {
	// Create creates a new appointment account
	Create(ctx context.Context, account *entities.AppointmentAccount) error

	// GetByID retrieves an appointment account by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.AppointmentAccount, error)

	// GetByAppointmentID retrieves an appointment account by appointment ID
	GetByAppointmentID(ctx context.Context, appointmentID uuid.UUID) (*entities.AppointmentAccount, error)

	// Exists checks if an appointment account exists for an appointment
	Exists(ctx context.Context, appointmentID uuid.UUID) (bool, error)

	// GetOrCreate gets an existing account or creates a new one if it doesn't exist
	GetOrCreate(ctx context.Context, organizationID, appointmentID uuid.UUID) (*entities.AppointmentAccount, error)
}
