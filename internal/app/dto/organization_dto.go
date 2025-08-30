package dto

import (
	"time"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// OrganizationDataRequest represents the request for organization data
type OrganizationDataRequest struct {
	StartDate string `form:"start_date" binding:"required" example:"2024-01-01"`
	EndDate   string `form:"end_date" binding:"required" example:"2024-12-31"`
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=1000" example:"500"`
}

// OrganizationDataResponse represents the complete organization data response
type OrganizationDataResponse struct {
	Organization *OrganizationDTO              `json:"organization"`
	Clinics      []*ClinicDTO                  `json:"clinics"`
	Units        []*UnitDTO                    `json:"units"`
	Doctors      []*DoctorDTO                  `json:"doctors"`
	Appointments []*AppointmentCalendarDataDTO `json:"appointments"`
}

// OrganizationDTO represents organization data in API responses
type OrganizationDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Address     *string   `json:"address,omitempty"`
	Phone       *string   `json:"phone,omitempty"`
	Email       *string   `json:"email,omitempty"`
	Website     *string   `json:"website,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ClinicDTO represents clinic data in API responses
type ClinicDTO struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	Address        *string   `json:"address,omitempty"`
	Phone          *string   `json:"phone,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UnitDTO represents unit data in API responses
type UnitDTO struct {
	ID          uuid.UUID `json:"id"`
	ClinicID    uuid.UUID `json:"clinic_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DoctorDTO represents doctor data in API responses
type DoctorDTO struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Name           string     `json:"name"`
	Specialty      *string    `json:"specialty,omitempty"`
	Email          *string    `json:"email,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	DefaultUnitID  *uuid.UUID `json:"default_unit_id,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AppointmentCalendarDataDTO represents minimal appointment data for calendar view
type AppointmentCalendarDataDTO struct {
	ID            uuid.UUID `json:"id"`
	PatientName   string    `json:"patient_name"`
	PatientPhone  *string   `json:"patient_phone,omitempty"`
	DoctorID      uuid.UUID `json:"doctor_id"`
	ClinicID      uuid.UUID `json:"clinic_id"`
	UnitID        uuid.UUID `json:"unit_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	TreatmentType *string   `json:"treatment_type,omitempty"`
}

// ToOrganizationDTO converts an Organization entity to DTO
func ToOrganizationDTO(org *entities.Organization) *OrganizationDTO {
	if org == nil {
		return nil
	}
	return &OrganizationDTO{
		ID:          org.ID,
		Name:        org.Name,
		Description: org.Description,
		Address:     org.Address,
		Phone:       org.Phone,
		Email:       org.Email,
		Website:     org.Website,
		IsActive:    org.IsActive,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}
}

// ToClinicDTO converts a Clinic entity to DTO
func ToClinicDTO(clinic *entities.Clinic) *ClinicDTO {
	if clinic == nil {
		return nil
	}
	return &ClinicDTO{
		ID:             clinic.ID,
		OrganizationID: clinic.OrganizationID,
		Name:           clinic.Name,
		Address:        clinic.Address,
		Phone:          clinic.Phone,
		CreatedAt:      clinic.CreatedAt,
		UpdatedAt:      clinic.UpdatedAt,
	}
}

// ToClinicDTOs converts a slice of Clinic entities to DTOs
func ToClinicDTOs(clinics []*entities.Clinic) []*ClinicDTO {
	if clinics == nil {
		return nil
	}
	dtos := make([]*ClinicDTO, len(clinics))
	for i, clinic := range clinics {
		dtos[i] = ToClinicDTO(clinic)
	}
	return dtos
}

// ToUnitDTO converts a Unit entity to DTO
func ToUnitDTO(unit *entities.Unit) *UnitDTO {
	if unit == nil {
		return nil
	}
	return &UnitDTO{
		ID:          unit.ID,
		ClinicID:    unit.ClinicID,
		Name:        unit.Name,
		Description: unit.Description,
		IsActive:    unit.IsActive,
		CreatedAt:   unit.CreatedAt,
		UpdatedAt:   unit.UpdatedAt,
	}
}

// ToUnitDTOs converts a slice of Unit entities to DTOs
func ToUnitDTOs(units []*entities.Unit) []*UnitDTO {
	if units == nil {
		return nil
	}
	dtos := make([]*UnitDTO, len(units))
	for i, unit := range units {
		dtos[i] = ToUnitDTO(unit)
	}
	return dtos
}

// ToDoctorDTO converts a Doctor entity to DTO
func ToDoctorDTO(doctor *entities.Doctor) *DoctorDTO {
	if doctor == nil {
		return nil
	}
	return &DoctorDTO{
		ID:             doctor.ID,
		OrganizationID: doctor.OrganizationID,
		Name:           doctor.Name,
		Specialty:      doctor.Specialty,
		Email:          doctor.Email,
		Phone:          doctor.Phone,
		DefaultUnitID:  doctor.DefaultUnitID,
		IsActive:       doctor.IsActive,
		CreatedAt:      doctor.CreatedAt,
		UpdatedAt:      doctor.UpdatedAt,
	}
}

// ToDoctorDTOs converts a slice of Doctor entities to DTOs
func ToDoctorDTOs(doctors []*entities.Doctor) []*DoctorDTO {
	if doctors == nil {
		return nil
	}
	dtos := make([]*DoctorDTO, len(doctors))
	for i, doctor := range doctors {
		dtos[i] = ToDoctorDTO(doctor)
	}
	return dtos
}

// ToAppointmentCalendarDataDTO converts AppointmentCalendarData to DTO
func ToAppointmentCalendarDataDTO(appt *repositories.AppointmentCalendarData) *AppointmentCalendarDataDTO {
	if appt == nil {
		return nil
	}
	return &AppointmentCalendarDataDTO{
		ID:            appt.ID,
		PatientName:   appt.PatientName,
		PatientPhone:  appt.PatientPhone,
		DoctorID:      appt.DoctorID,
		ClinicID:      appt.ClinicID,
		UnitID:        appt.UnitID,
		StartTime:     appt.StartTime,
		EndTime:       appt.EndTime,
		Status:        appt.Status,
		TreatmentType: appt.TreatmentType,
	}
}

// ToAppointmentCalendarDataDTOs converts a slice of AppointmentCalendarData to DTOs
func ToAppointmentCalendarDataDTOs(appointments []*repositories.AppointmentCalendarData) []*AppointmentCalendarDataDTO {
	if appointments == nil {
		return nil
	}
	dtos := make([]*AppointmentCalendarDataDTO, len(appointments))
	for i, appt := range appointments {
		dtos[i] = ToAppointmentCalendarDataDTO(appt)
	}
	return dtos
}
