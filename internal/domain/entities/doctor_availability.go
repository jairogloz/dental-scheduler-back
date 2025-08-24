package entities

import (
	"time"

	"github.com/google/uuid"
)

// DoctorAvailability represents a doctor's availability entity
type DoctorAvailability struct {
	ID             uuid.UUID `json:"id" db:"id"`
	DoctorID       uuid.UUID `json:"doctor_id" db:"doctor_id"`
	StartTime      time.Time `json:"start_time" db:"start_time"`
	EndTime        time.Time `json:"end_time" db:"end_time"`
	RecurrenceRule *string   `json:"recurrence_rule,omitempty" db:"recurrence_rule"`
	IsAvailable    bool      `json:"is_available" db:"is_available"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Validate checks if the doctor availability entity is valid
func (da *DoctorAvailability) Validate() error {
	if da.DoctorID == uuid.Nil {
		return ErrInvalidDoctorID
	}

	if da.StartTime.IsZero() || da.EndTime.IsZero() {
		return ErrInvalidAvailabilityTime
	}

	if da.EndTime.Before(da.StartTime) || da.EndTime.Equal(da.StartTime) {
		return ErrInvalidAvailabilityTime
	}

	return nil
}

// IsValid checks if the doctor availability has valid data
func (da *DoctorAvailability) IsValid() bool {
	return da.Validate() == nil
}

// Duration returns the duration of the availability period
func (da *DoctorAvailability) Duration() time.Duration {
	return da.EndTime.Sub(da.StartTime)
}

// ConflictsWith checks if this availability conflicts with another time range
func (da *DoctorAvailability) ConflictsWith(startTime, endTime time.Time) bool {
	return da.StartTime.Before(endTime) && da.EndTime.After(startTime)
}
