package dto

import (
	"time"

	"dental-scheduler-backend/internal/domain/entities"

	"github.com/google/uuid"
)

// CreateDoctorAvailabilityRequest represents the request to create doctor availability
type CreateDoctorAvailabilityRequest struct {
	DoctorID       uuid.UUID `json:"doctor_id" binding:"required"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required"`
	RecurrenceRule *string   `json:"recurrence_rule,omitempty"`
	IsAvailable    *bool     `json:"is_available,omitempty"`
}

// UpdateDoctorAvailabilityRequest represents the request to update doctor availability
type UpdateDoctorAvailabilityRequest struct {
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required"`
	RecurrenceRule *string   `json:"recurrence_rule,omitempty"`
	IsAvailable    *bool     `json:"is_available,omitempty"`
}

// DoctorAvailabilityResponse represents the response for doctor availability
type DoctorAvailabilityResponse struct {
	ID             uuid.UUID `json:"id"`
	DoctorID       uuid.UUID `json:"doctor_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	RecurrenceRule *string   `json:"recurrence_rule,omitempty"`
	IsAvailable    bool      `json:"is_available"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AvailableSlotResponse represents an available time slot
type AvailableSlotResponse struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ToEntity converts CreateDoctorAvailabilityRequest to entities.DoctorAvailability
func (req *CreateDoctorAvailabilityRequest) ToEntity() *entities.DoctorAvailability {
	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	return &entities.DoctorAvailability{
		ID:             uuid.New(),
		DoctorID:       req.DoctorID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		RecurrenceRule: req.RecurrenceRule,
		IsAvailable:    isAvailable,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// ToDoctorAvailabilityResponse converts entities.DoctorAvailability to DoctorAvailabilityResponse
func ToDoctorAvailabilityResponse(da *entities.DoctorAvailability) *DoctorAvailabilityResponse {
	return &DoctorAvailabilityResponse{
		ID:             da.ID,
		DoctorID:       da.DoctorID,
		StartTime:      da.StartTime,
		EndTime:        da.EndTime,
		RecurrenceRule: da.RecurrenceRule,
		IsAvailable:    da.IsAvailable,
		CreatedAt:      da.CreatedAt,
		UpdatedAt:      da.UpdatedAt,
	}
}

// ToEntityUpdate converts UpdateDoctorAvailabilityRequest to updated entities.DoctorAvailability
func (req *UpdateDoctorAvailabilityRequest) ToEntityUpdate(existing *entities.DoctorAvailability) *entities.DoctorAvailability {
	existing.StartTime = req.StartTime
	existing.EndTime = req.EndTime
	existing.RecurrenceRule = req.RecurrenceRule
	if req.IsAvailable != nil {
		existing.IsAvailable = *req.IsAvailable
	}
	existing.UpdatedAt = time.Now()
	return existing
}
