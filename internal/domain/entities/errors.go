package entities

import "errors"

// Domain errors
var (
	// Organization errors
	ErrInvalidOrganizationName = errors.New("organization name is required")
	ErrOrganizationNotFound    = errors.New("organization not found")
	ErrInvalidOrganizationID   = errors.New("organization ID is required")

	// Profile errors
	ErrInvalidProfileID = errors.New("profile ID is required")
	ErrInvalidRoles     = errors.New("at least one role is required")
	ErrProfileNotFound  = errors.New("profile not found")

	// Clinic errors
	ErrInvalidClinicName = errors.New("clinic name is required")
	ErrClinicNotFound    = errors.New("clinic not found")

	// Unit errors
	ErrInvalidUnitName = errors.New("unit name is required")
	ErrInvalidClinicID = errors.New("clinic ID is required")
	ErrUnitNotFound    = errors.New("unit not found")

	// Doctor errors
	ErrInvalidDoctorName              = errors.New("doctor name is required")
	ErrDoctorNotFound                 = errors.New("doctor not found")
	ErrInvalidEmail                   = errors.New("invalid email format")
	ErrInvalidColor                   = errors.New("invalid color format (must be hex color like #3B82F6)")
	ErrDoctorUnitOrganizationMismatch = errors.New("doctor's default unit must belong to the same organization")

	// Patient errors
	ErrInvalidPatientName = errors.New("patient name is required")
	ErrPatientNotFound    = errors.New("patient not found")

	// Appointment errors
	ErrInvalidPatientID           = errors.New("patient ID is required")
	ErrInvalidDoctorID            = errors.New("doctor ID is required")
	ErrInvalidUnitID              = errors.New("unit ID is required")
	ErrInvalidAppointmentTime     = errors.New("invalid appointment time")
	ErrEndTimeBeforeStartTime     = errors.New("end time must be after start time")
	ErrAppointmentNotFound        = errors.New("appointment not found")
	ErrAppointmentConflict        = errors.New("appointment conflicts with existing appointment")
	ErrPastAppointmentTime        = errors.New("appointment time cannot be in the past")
	ErrInvalidAppointmentStatus   = errors.New("invalid appointment status")
	ErrAppointmentNotInQueue      = errors.New("appointment is not in rescheduling queue")
	ErrInvalidStatusTransition    = errors.New("invalid appointment status transition")
	ErrCancellationReasonRequired = errors.New("cancellation reason is required")
	ErrUnauthorizedAccess         = errors.New("unauthorized access to resource")

	// Doctor Availability errors
	ErrInvalidAvailabilityTime = errors.New("invalid availability time")
	ErrAvailabilityNotFound    = errors.New("availability not found")
	ErrDoctorNotAvailable      = errors.New("doctor is not available at the requested time")

	// General errors
	ErrInvalidID = errors.New("invalid ID format")
)
