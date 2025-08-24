package errors

import "errors"

// Custom error types for the application
var (
	// Common errors
	ErrInternalServer = errors.New("internal server error")
	ErrBadRequest     = errors.New("bad request")
	ErrNotFound       = errors.New("resource not found")
	ErrConflict       = errors.New("resource conflict")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")

	// Validation errors
	ErrInvalidUUID   = errors.New("invalid UUID format")
	ErrInvalidEmail  = errors.New("invalid email format")
	ErrInvalidDate   = errors.New("invalid date format")
	ErrRequiredField = errors.New("required field is missing")
)

// AppError represents an application error with additional context
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(code, message, details string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: details,
		Err:     err,
	}
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}
