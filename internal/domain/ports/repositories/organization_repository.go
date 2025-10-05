package repositories

import (
	"context"
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// OrganizationData represents the complete organization data for calendar loading
type OrganizationData struct {
	Organization *entities.Organization
	Clinics      []*entities.Clinic
	Units        []*entities.Unit
	Doctors      []*entities.Doctor
	Appointments []*AppointmentCalendarData
	Services     []*entities.Service
}

// AppointmentCalendarData represents minimal appointment data for calendar view
type AppointmentCalendarData struct {
	ID            uuid.UUID `json:"id"`
	PatientID     uuid.UUID `json:"patient_id"`
	PatientName   string    `json:"patient_name"`
	PatientPhone  *string   `json:"patient_phone"`
	DoctorID      uuid.UUID `json:"doctor_id"`
	ClinicID      uuid.UUID `json:"clinic_id"`
	UnitID        uuid.UUID `json:"unit_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	TreatmentType *string   `json:"treatment_type"`
	IsFirstVisit  bool      `json:"is_first_visit"`
}

// OrganizationRepository defines the interface for organization data operations
type OrganizationRepository interface {
	// GetByID retrieves an organization by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error)

	// GetOrganizationData retrieves complete organization data for calendar loading
	GetOrganizationData(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, limit int) (*OrganizationData, error)

	// Exists checks if an organization exists by its ID
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}
