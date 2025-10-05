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

	// SearchPatients searches for patients by name, phone, or email within an organization
	SearchPatients(ctx context.Context, orgID uuid.UUID, query string, limit int) ([]*entities.Patient, error)

	// AddPatientToOrganization links a patient to an organization
	AddPatientToOrganization(ctx context.Context, patientID, orgID uuid.UUID) error

	// OrganizationExists checks if an organization exists by its ID
	OrganizationExists(ctx context.Context, orgID uuid.UUID) (bool, error)

	// CreatePatientWithOrganization creates a patient and links to organization in a transaction
	CreatePatientWithOrganization(ctx context.Context, patient *entities.Patient, orgID uuid.UUID) error

	// UpdateFirstAppointmentIfNil sets the patient's first_appointment_id if it's currently NULL
	UpdateFirstAppointmentIfNil(ctx context.Context, patientID uuid.UUID, appointmentID uuid.UUID) error

	// PatientBelongsToOrganization checks if a patient belongs to an organization
	PatientBelongsToOrganization(ctx context.Context, patientID, orgID uuid.UUID) (bool, error)
}
