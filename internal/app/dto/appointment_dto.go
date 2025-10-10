package dto

import (
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreateAppointmentRequest represents the request to create an appointment
type CreateAppointmentRequest struct {
	PatientID uuid.UUID `json:"patient_id" binding:"required"`
	DoctorID  uuid.UUID `json:"doctor_id" binding:"required"`
	UnitID    uuid.UUID `json:"unit_id" binding:"required"`
	ServiceID string    `json:"service_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Notes     *string   `json:"notes,omitempty"`
}

// UpdateAppointmentRequest represents the request to update an appointment (partial updates)
type UpdateAppointmentRequest struct {
	PatientID *uuid.UUID                  `json:"patient_id,omitempty"`
	DoctorID  *uuid.UUID                  `json:"doctor_id,omitempty"`
	UnitID    *uuid.UUID                  `json:"unit_id,omitempty"`
	ServiceID *string                     `json:"service_id,omitempty"`
	Status    *entities.AppointmentStatus `json:"status,omitempty"`
	StartTime *time.Time                  `json:"start_time,omitempty"`
	EndTime   *time.Time                  `json:"end_time,omitempty"`
	Notes     *string                     `json:"notes,omitempty"`
}

// AppointmentResponse represents the response for an appointment
type AppointmentResponse struct {
	ID           uuid.UUID                  `json:"id"`
	PatientID    uuid.UUID                  `json:"patient_id"`
	PatientName  string                     `json:"patient_name"`
	DoctorID     uuid.UUID                  `json:"doctor_id"`
	UnitID       uuid.UUID                  `json:"unit_id"`
	ServiceID    *string                    `json:"service_id,omitempty"`
	ServiceName  *string                    `json:"service_name,omitempty"`
	Status       entities.AppointmentStatus `json:"status"`
	StartTime    time.Time                  `json:"start_time"`
	EndTime      time.Time                  `json:"end_time"`
	Notes        *string                    `json:"notes,omitempty"`
	IsFirstVisit bool                       `json:"is_first_visit"`
	CreatedAt    time.Time                  `json:"created_at"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}

// AppointmentWithDetailsResponse represents the response for an appointment with related entity details
type AppointmentWithDetailsResponse struct {
	ID          uuid.UUID                  `json:"id"`
	Patient     *PatientResponse           `json:"patient"`
	Doctor      *DoctorResponse            `json:"doctor"`
	Unit        *UnitResponse              `json:"unit"`
	ServiceID   *string                    `json:"service_id,omitempty"`
	ServiceName *string                    `json:"service_name,omitempty"`
	Status      entities.AppointmentStatus `json:"status"`
	StartTime   time.Time                  `json:"start_time"`
	EndTime     time.Time                  `json:"end_time"`
	Notes       *string                    `json:"notes,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

// RescheduleAppointmentRequest represents the request to reschedule an appointment
type RescheduleAppointmentRequest struct {
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

// GetAppointmentsRequest represents the request to get appointments with filters
type GetAppointmentsRequest struct {
	OrgID     string `form:"-"` // This will be set from context, not from form parameters
	StartDate string `form:"startDate" binding:"required"`
	EndDate   string `form:"endDate" binding:"required"`
	ClinicID  string `form:"clinicId,omitempty"`
	Status    string `form:"status,omitempty"`
	DoctorID  string `form:"doctorId,omitempty"`
	Page      int    `form:"page,omitempty"`
	Limit     int    `form:"limit,omitempty"`
}

// PatientListDataDTO represents minimal patient data for appointment lists
type PatientListDataDTO struct {
	ID        string  `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  *string `json:"last_name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	Email     *string `json:"email,omitempty"`
}

// AppointmentListResponse represents an appointment with all related details for listing
type AppointmentListResponse struct {
	ID           string              `json:"id"`
	Patient      *PatientListDataDTO `json:"patient"`
	DoctorID     string              `json:"doctor_id"`
	DoctorName   string              `json:"doctor_name"`
	ClinicID     string              `json:"clinic_id"`
	ClinicName   string              `json:"clinic_name"`
	UnitID       *string             `json:"unit_id,omitempty"`
	UnitName     *string             `json:"unit_name,omitempty"`
	StartTime    time.Time           `json:"start_time"`
	EndTime      time.Time           `json:"end_time"`
	Status       string              `json:"status"`
	ServiceID    string              `json:"service_id,omitempty"`
	ServiceName  string              `json:"service_name,omitempty"`
	Notes        string              `json:"notes,omitempty"`
	IsFirstVisit bool                `json:"is_first_visit"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// AppointmentSummary provides summary statistics for the appointments
type AppointmentSummary struct {
	TotalAppointments int                    `json:"total_appointments"`
	ByClinic          map[string]ClinicStats `json:"by_clinic"`
	ByStatus          map[string]int         `json:"by_status"`
	ByDate            map[string]int         `json:"by_date"`
}

// ClinicStats provides statistics per clinic
type ClinicStats struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
}

// GetAppointmentsResponse represents the complete response for appointment listing
type GetAppointmentsResponse struct {
	Appointments []AppointmentListResponse `json:"appointments"`
	Summary      AppointmentSummary        `json:"summary"`
	Pagination   PaginationInfo            `json:"pagination"`
}

// PaginationInfo provides pagination details
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ToEntity converts CreateAppointmentRequest to entities.Appointment
func (req *CreateAppointmentRequest) ToEntity() *entities.Appointment {
	serviceID := req.ServiceID
	return &entities.Appointment{
		ID:        uuid.New(),
		PatientID: req.PatientID,
		DoctorID:  req.DoctorID,
		UnitID:    req.UnitID,
		ServiceID: &serviceID,
		Status:    entities.AppointmentStatusScheduled,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Notes:     req.Notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ToAppointmentResponse converts entities.Appointment to AppointmentResponse
func ToAppointmentResponse(a *entities.Appointment) *AppointmentResponse {
	return &AppointmentResponse{
		ID:           a.ID,
		PatientID:    a.PatientID,
		PatientName:  "", // Will be empty when patient name is not available
		DoctorID:     a.DoctorID,
		UnitID:       a.UnitID,
		ServiceID:    a.ServiceID,
		ServiceName:  nil, // Will be nil when service name is not available
		Status:       a.Status,
		StartTime:    a.StartTime,
		EndTime:      a.EndTime,
		Notes:        a.Notes,
		IsFirstVisit: false, // Default to false when patient info not available
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

// ToAppointmentResponseWithPatientName converts entities.Appointment to AppointmentResponse with patient name
func ToAppointmentResponseWithPatientName(a *entities.Appointment, patientName string) *AppointmentResponse {
	return &AppointmentResponse{
		ID:           a.ID,
		PatientID:    a.PatientID,
		PatientName:  patientName,
		DoctorID:     a.DoctorID,
		UnitID:       a.UnitID,
		ServiceID:    a.ServiceID,
		ServiceName:  nil, // Will be nil when service name is not available
		Status:       a.Status,
		StartTime:    a.StartTime,
		EndTime:      a.EndTime,
		Notes:        a.Notes,
		IsFirstVisit: false, // Default to false, use WithPatientNameAndFirstVisit for accurate flag
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

// ToAppointmentResponseWithPatientNameAndFirstVisit converts entities.Appointment to AppointmentResponse with all patient details
func ToAppointmentResponseWithPatientNameAndFirstVisit(a *entities.Appointment, patientName string, isFirstVisit bool) *AppointmentResponse {
	return &AppointmentResponse{
		ID:           a.ID,
		PatientID:    a.PatientID,
		PatientName:  patientName,
		DoctorID:     a.DoctorID,
		UnitID:       a.UnitID,
		ServiceID:    a.ServiceID,
		ServiceName:  nil, // Will be nil when service name is not available
		Status:       a.Status,
		StartTime:    a.StartTime,
		EndTime:      a.EndTime,
		Notes:        a.Notes,
		IsFirstVisit: isFirstVisit,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

// ToEntityUpdate converts UpdateAppointmentRequest to updated entities.Appointment (partial updates)
func (req *UpdateAppointmentRequest) ToEntityUpdate(existing *entities.Appointment) *entities.Appointment {
	// Only update fields that are provided (not nil)
	if req.PatientID != nil {
		existing.PatientID = *req.PatientID
	}
	if req.DoctorID != nil {
		existing.DoctorID = *req.DoctorID
	}
	if req.UnitID != nil {
		existing.UnitID = *req.UnitID
	}
	if req.ServiceID != nil {
		existing.ServiceID = req.ServiceID
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.StartTime != nil {
		existing.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		existing.EndTime = *req.EndTime
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}
	existing.UpdatedAt = time.Now()
	return existing
}
