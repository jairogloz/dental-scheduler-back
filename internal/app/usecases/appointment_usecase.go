package usecases

import (
	"context"
	"fmt"
	"time"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/domain/services"

	"github.com/google/uuid"
)

// AppointmentUseCase handles appointment-related business logic
type AppointmentUseCase struct {
	appointmentRepo   repositories.AppointmentRepository
	patientRepo       repositories.PatientRepository
	doctorRepo        repositories.DoctorRepository
	unitRepo          repositories.UnitRepository
	schedulingService *services.SchedulingService
}

// NewAppointmentUseCase creates a new instance of AppointmentUseCase
func NewAppointmentUseCase(
	appointmentRepo repositories.AppointmentRepository,
	patientRepo repositories.PatientRepository,
	doctorRepo repositories.DoctorRepository,
	unitRepo repositories.UnitRepository,
	schedulingService *services.SchedulingService,
) *AppointmentUseCase {
	return &AppointmentUseCase{
		appointmentRepo:   appointmentRepo,
		patientRepo:       patientRepo,
		doctorRepo:        doctorRepo,
		unitRepo:          unitRepo,
		schedulingService: schedulingService,
	}
}

// CreateAppointment creates a new appointment with basic validation (no conflict checking)
func (uc *AppointmentUseCase) CreateAppointment(ctx context.Context, orgID uuid.UUID, req *dto.CreateAppointmentRequest) (*dto.AppointmentResponse, error) {
	// Validate date logic: end date can't be before start date
	if req.EndTime.Before(req.StartTime) {
		return nil, fmt.Errorf("end time cannot be before start time")
	}

	// Allow appointments in the past (no validation against past dates)

	// Verify patient exists
	patientExists, err := uc.patientRepo.Exists(ctx, req.PatientID)
	if err != nil {
		return nil, err
	}
	if !patientExists {
		return nil, entities.ErrPatientNotFound
	}

	// Verify doctor exists
	doctorExists, err := uc.doctorRepo.Exists(ctx, req.DoctorID)
	if err != nil {
		return nil, err
	}
	if !doctorExists {
		return nil, entities.ErrDoctorNotFound
	}

	// Verify unit exists
	unitExists, err := uc.unitRepo.Exists(ctx, req.UnitID)
	if err != nil {
		return nil, err
	}
	if !unitExists {
		return nil, entities.ErrUnitNotFound
	}

	// Create appointment entity
	appointment := req.ToEntity()

	// Create appointment directly in repository (no conflict checking)
	if err := uc.appointmentRepo.Create(ctx, appointment); err != nil {
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}

	// Link patient to organization (ignore errors if already linked)
	if err := uc.patientRepo.AddPatientToOrganization(ctx, req.PatientID, orgID); err != nil {
		// Log the error but don't fail the appointment creation
		// The ON CONFLICT DO NOTHING in the query will handle duplicates gracefully
		fmt.Printf("Warning: failed to link patient to organization: %v\n", err)
	}

	return dto.ToAppointmentResponse(appointment), nil
}

// GetAppointmentByID retrieves an appointment by its ID
func (uc *AppointmentUseCase) GetAppointmentByID(ctx context.Context, id uuid.UUID) (*dto.AppointmentResponse, error) {
	appointment, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if appointment == nil {
		return nil, entities.ErrAppointmentNotFound
	}

	return dto.ToAppointmentResponse(appointment), nil
}

// GetAllAppointments retrieves all appointments
func (uc *AppointmentUseCase) GetAllAppointments(ctx context.Context) ([]*dto.AppointmentResponse, error) {
	appointments, err := uc.appointmentRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		responses[i] = dto.ToAppointmentResponse(appointment)
	}

	return responses, nil
}

// GetUpcomingAppointments retrieves all upcoming appointments
func (uc *AppointmentUseCase) GetUpcomingAppointments(ctx context.Context) ([]*dto.AppointmentResponse, error) {
	appointments, err := uc.appointmentRepo.GetUpcoming(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AppointmentResponse, len(appointments))
	for i, appointment := range appointments {
		responses[i] = dto.ToAppointmentResponse(appointment)
	}

	return responses, nil
}

// UpdateAppointment updates an existing appointment
func (uc *AppointmentUseCase) UpdateAppointment(ctx context.Context, id uuid.UUID, req *dto.UpdateAppointmentRequest) (*dto.AppointmentResponse, error) {
	existing, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		return nil, entities.ErrAppointmentNotFound
	}

	// Verify patient exists
	patientExists, err := uc.patientRepo.Exists(ctx, req.PatientID)
	if err != nil {
		return nil, err
	}
	if !patientExists {
		return nil, entities.ErrPatientNotFound
	}

	// Verify doctor exists
	doctorExists, err := uc.doctorRepo.Exists(ctx, req.DoctorID)
	if err != nil {
		return nil, err
	}
	if !doctorExists {
		return nil, entities.ErrDoctorNotFound
	}

	// Verify unit exists
	unitExists, err := uc.unitRepo.Exists(ctx, req.UnitID)
	if err != nil {
		return nil, err
	}
	if !unitExists {
		return nil, entities.ErrUnitNotFound
	}

	updated := req.ToEntityUpdate(existing)

	if err := updated.Validate(); err != nil {
		return nil, err
	}

	// Check for conflicts if time has changed
	if !updated.StartTime.Equal(existing.StartTime) || !updated.EndTime.Equal(existing.EndTime) {
		hasConflict, err := uc.appointmentRepo.CheckConflict(
			ctx,
			updated.DoctorID,
			updated.UnitID,
			updated.StartTime,
			updated.EndTime,
			&updated.ID,
		)
		if err != nil {
			return nil, err
		}
		if hasConflict {
			return nil, entities.ErrAppointmentConflict
		}
	}

	if err := uc.appointmentRepo.Update(ctx, updated); err != nil {
		return nil, err
	}

	return dto.ToAppointmentResponse(updated), nil
}

// RescheduleAppointment reschedules an existing appointment
func (uc *AppointmentUseCase) RescheduleAppointment(ctx context.Context, id uuid.UUID, req *dto.RescheduleAppointmentRequest) (*dto.AppointmentResponse, error) {
	if err := uc.schedulingService.RescheduleAppointment(ctx, id, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}

	// Get the updated appointment
	appointment, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.ToAppointmentResponse(appointment), nil
}

// CancelAppointment cancels an appointment
func (uc *AppointmentUseCase) CancelAppointment(ctx context.Context, id uuid.UUID) error {
	appointment, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if appointment == nil {
		return entities.ErrAppointmentNotFound
	}

	appointment.Cancel()

	return uc.appointmentRepo.Update(ctx, appointment)
}

// CompleteAppointment marks an appointment as completed
func (uc *AppointmentUseCase) CompleteAppointment(ctx context.Context, id uuid.UUID) error {
	appointment, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if appointment == nil {
		return entities.ErrAppointmentNotFound
	}

	appointment.Complete()

	return uc.appointmentRepo.Update(ctx, appointment)
}

// DeleteAppointment deletes an appointment by its ID
func (uc *AppointmentUseCase) DeleteAppointment(ctx context.Context, id uuid.UUID) error {
	exists, err := uc.appointmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if exists == nil {
		return entities.ErrAppointmentNotFound
	}

	return uc.appointmentRepo.Delete(ctx, id)
}

// GetAvailableSlots returns available time slots for a doctor on a specific date
func (uc *AppointmentUseCase) GetAvailableSlots(ctx context.Context, doctorID uuid.UUID, date time.Time, slotDurationMinutes int) ([]*dto.AvailableSlotResponse, error) {
	slotDuration := time.Duration(slotDurationMinutes) * time.Minute

	slots, err := uc.schedulingService.GetAvailableSlots(ctx, doctorID, date, slotDuration)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.AvailableSlotResponse, len(slots))
	for i, slot := range slots {
		responses[i] = &dto.AvailableSlotResponse{
			StartTime: slot,
			EndTime:   slot.Add(slotDuration),
		}
	}

	return responses, nil
}

// GetAppointmentsByOrganization retrieves appointments for an organization with filters
func (uc *AppointmentUseCase) GetAppointmentsByOrganization(ctx context.Context, req *dto.GetAppointmentsRequest) (*dto.GetAppointmentsResponse, error) {
	// Parse and validate dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	// Validate date range (max 30 days)
	if endDate.Sub(startDate) > 30*24*time.Hour {
		return nil, fmt.Errorf("date range cannot exceed 30 days")
	}

	// Set default pagination
	if req.Limit <= 0 {
		req.Limit = 50 // Default limit
	}
	if req.Page <= 0 {
		req.Page = 1 // Default page
	}

	// Build filters
	filters := repositories.AppointmentFilters{
		Page:  req.Page,
		Limit: req.Limit,
	}

	if req.ClinicID != "" {
		clinicUUID, err := uuid.Parse(req.ClinicID)
		if err != nil {
			return nil, fmt.Errorf("invalid clinic ID: %w", err)
		}
		filters.ClinicID = &clinicUUID
	}

	if req.DoctorID != "" {
		doctorUUID, err := uuid.Parse(req.DoctorID)
		if err != nil {
			return nil, fmt.Errorf("invalid doctor ID: %w", err)
		}
		filters.DoctorID = &doctorUUID
	}

	if req.Status != "" {
		status := entities.AppointmentStatus(req.Status)
		filters.Status = &status
	}

	// Get organization UUID
	orgUUID, err := uuid.Parse(req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}

	// Get appointments with details
	appointments, totalCount, err := uc.appointmentRepo.GetByOrganizationAndDateRange(ctx, orgUUID, startDate, endDate, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch appointments: %w", err)
	}

	// Convert to DTOs and build response
	return uc.buildAppointmentResponse(appointments, totalCount, req.Page, req.Limit), nil
}

// buildAppointmentResponse converts appointments to DTOs and builds summary
func (uc *AppointmentUseCase) buildAppointmentResponse(appointments []*repositories.AppointmentWithDetails, totalCount, page, limit int) *dto.GetAppointmentsResponse {
	// Convert appointments to DTOs
	appointmentDTOs := make([]dto.AppointmentListResponse, len(appointments))
	clinicMap := make(map[string]dto.ClinicStats)
	statusMap := make(map[string]int)
	dateMap := make(map[string]int)

	for i, appt := range appointments {
		// Build appointment DTO
		appointmentDTOs[i] = dto.AppointmentListResponse{
			ID:            appt.Appointment.ID.String(),
			PatientID:     appt.Appointment.PatientID.String(),
			PatientName:   appt.Patient.Name,
			PatientPhone:  getStringPtr(appt.Patient.Phone),
			DoctorID:      appt.Appointment.DoctorID.String(),
			DoctorName:    appt.Doctor.Name,
			ClinicID:      appt.Clinic.ID.String(),
			ClinicName:    appt.Clinic.Name,
			UnitID:        getStringPtrFromUUID(&appt.Unit.ID),
			UnitName:      &appt.Unit.Name,
			StartTime:     appt.Appointment.StartTime,
			EndTime:       appt.Appointment.EndTime,
			Status:        string(appt.Appointment.Status),
			TreatmentType: getStringPtr(appt.Appointment.TreatmentType),
			Notes:         getStringPtr(appt.Appointment.Notes),
			CreatedAt:     appt.Appointment.CreatedAt,
			UpdatedAt:     appt.Appointment.UpdatedAt,
		}

		// Build summary data
		clinicID := appt.Clinic.ID.String()
		if stats, exists := clinicMap[clinicID]; exists {
			stats.Count++
			clinicMap[clinicID] = stats
		} else {
			clinicMap[clinicID] = dto.ClinicStats{
				Count: 1,
				Name:  appt.Clinic.Name,
			}
		}

		// Status summary
		status := string(appt.Appointment.Status)
		statusMap[status]++

		// Date summary
		dateKey := appt.Appointment.StartTime.Format("2006-01-02")
		dateMap[dateKey]++
	}

	// Calculate pagination
	totalPages := (totalCount + limit - 1) / limit

	return &dto.GetAppointmentsResponse{
		Appointments: appointmentDTOs,
		Summary: dto.AppointmentSummary{
			TotalAppointments: totalCount,
			ByClinic:          clinicMap,
			ByStatus:          statusMap,
			ByDate:            dateMap,
		},
		Pagination: dto.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      totalCount,
			TotalPages: totalPages,
		},
	}
}

// Helper functions
func getStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getStringPtrFromUUID(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	str := u.String()
	return &str
}
