package usecases

import (
	"context"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/ports/repositories"

	"github.com/google/uuid"
)

// GetOrganizationDataUseCase handles getting complete organization data for calendar loading
type GetOrganizationDataUseCase struct {
	orgRepo repositories.OrganizationRepository
}

// NewGetOrganizationDataUseCase creates a new instance of GetOrganizationDataUseCase
func NewGetOrganizationDataUseCase(orgRepo repositories.OrganizationRepository) *GetOrganizationDataUseCase {
	return &GetOrganizationDataUseCase{
		orgRepo: orgRepo,
	}
}

// Execute retrieves complete organization data for calendar view
func (uc *GetOrganizationDataUseCase) Execute(ctx context.Context, orgID uuid.UUID, req *dto.OrganizationDataRequest) (*dto.OrganizationDataResponse, error) {
	// Validate and parse dates
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

	// Set default limit if not provided
	limit := req.Limit
	if limit == 0 {
		limit = 500 // Default limit for appointments
	}

	// Extend end date to include the entire end day
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Get organization data
	orgData, err := uc.orgRepo.GetOrganizationData(ctx, orgID, startDate, endDate, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization data: %w", err)
	}

	// Convert to DTOs
	response := &dto.OrganizationDataResponse{
		Organization: dto.ToOrganizationDTO(orgData.Organization),
		Clinics:      dto.ToClinicDTOs(orgData.Clinics),
		Units:        dto.ToUnitDTOs(orgData.Units),
		Doctors:      dto.ToDoctorDTOs(orgData.Doctors),
		Appointments: dto.ToAppointmentCalendarDataDTOs(orgData.Appointments),
		Services:     dto.ToServiceDTOs(orgData.Services),
	}

	return response, nil
}
