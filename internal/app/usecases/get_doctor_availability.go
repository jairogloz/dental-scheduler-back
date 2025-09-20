package usecases

import (
	"context"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// GetDoctorAvailabilityUseCase handles getting doctor availability
type GetDoctorAvailabilityUseCase struct {
	availabilityRepo repositories.DoctorAvailabilityRepository
	doctorRepo       repositories.DoctorRepository
}

// NewGetDoctorAvailabilityUseCase creates a new instance of GetDoctorAvailabilityUseCase
func NewGetDoctorAvailabilityUseCase(
	availabilityRepo repositories.DoctorAvailabilityRepository,
	doctorRepo repositories.DoctorRepository,
) *GetDoctorAvailabilityUseCase {
	return &GetDoctorAvailabilityUseCase{
		availabilityRepo: availabilityRepo,
		doctorRepo:       doctorRepo,
	}
}

// Execute retrieves doctor availability with optional date range filtering
func (uc *GetDoctorAvailabilityUseCase) Execute(ctx context.Context, doctorID uuid.UUID, orgID uuid.UUID, req *dto.GetDoctorAvailabilityRequest) (*dto.GetDoctorAvailabilityResponse, error) {
	// Verify doctor exists and belongs to the organization
	doctor, err := uc.doctorRepo.GetByID(ctx, doctorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor: %w", err)
	}
	if doctor == nil {
		return nil, entities.ErrDoctorNotFound
	}
	if doctor.OrganizationID != orgID {
		return nil, entities.ErrDoctorNotFound // Don't reveal that doctor exists in different org
	}

	var availabilities []*entities.DoctorAvailability

	// If date range is provided, use filtered query
	if req.StartDate != "" && req.EndDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD: %w", err)
		}

		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD: %w", err)
		}

		// Validate date range
		if endDate.Before(startDate) {
			return nil, fmt.Errorf("end_date cannot be before start_date")
		}

		// Check if date range is too large (more than 1 year)
		if endDate.Sub(startDate) > 365*24*time.Hour {
			return nil, fmt.Errorf("date range cannot exceed 365 days")
		}

		// Extend end date to include the entire end day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		availabilities, err = uc.availabilityRepo.GetByDoctorIDAndDateRange(ctx, doctorID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get doctor availability by date range: %w", err)
		}
	} else if req.StartDate != "" || req.EndDate != "" {
		// If only one date is provided, return error
		return nil, fmt.Errorf("both start_date and end_date must be provided, or neither")
	} else {
		// No date filtering, get all availability
		availabilities, err = uc.availabilityRepo.GetByDoctorID(ctx, doctorID)
		if err != nil {
			return nil, fmt.Errorf("failed to get doctor availability: %w", err)
		}
	}

	// Convert to response DTOs
	response := &dto.GetDoctorAvailabilityResponse{
		Availabilities: dto.ToDoctorAvailabilityResponses(availabilities),
	}

	return response, nil
}
