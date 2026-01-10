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

	// Get unit with clinic info (need timezone for conversion)
	unit, clinic, err := uc.unitRepo.GetUnitWithClinic(ctx, req.UnitID)
	if err != nil {
		return nil, err
	}
	if unit == nil || clinic == nil {
		return nil, entities.ErrUnitNotFound
	}

	// Convert appointment times from clinic timezone to UTC
	startTimeUTC := req.StartTime
	endTimeUTC := req.EndTime

	if clinic.Timezone != "" {
		loc, err := time.LoadLocation(clinic.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid clinic timezone %q: %w", clinic.Timezone, err)
		}

		// Parse the incoming times as being in the clinic's timezone
		// The times are naive (no timezone info), so we interpret them in clinic's timezone
		year, month, day := req.StartTime.Date()
		hour, min, sec := req.StartTime.Clock()
		startTimeInClinicTZ := time.Date(year, month, day, hour, min, sec, req.StartTime.Nanosecond(), loc)
		startTimeUTC = startTimeInClinicTZ.UTC()

		year, month, day = req.EndTime.Date()
		hour, min, sec = req.EndTime.Clock()
		endTimeInClinicTZ := time.Date(year, month, day, hour, min, sec, req.EndTime.Nanosecond(), loc)
		endTimeUTC = endTimeInClinicTZ.UTC()
	}

	// Create appointment entity with UTC times
	appointment := req.ToEntity()
	appointment.StartTime = startTimeUTC
	appointment.EndTime = endTimeUTC

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

	// Try to set patient's first_appointment_id if NULL (best-effort, non-critical)
	if err := uc.patientRepo.UpdateFirstAppointmentIfNil(ctx, req.PatientID, appointment.ID); err != nil {
		// Log but don't fail - this is a non-critical operation
		fmt.Printf("Warning: failed to set patient's first_appointment_id: %v\n", err)
	}

	// Fetch patient data to include patient name and is_first_visit flag in response
	patient, err := uc.patientRepo.GetByID(ctx, req.PatientID)
	if err != nil {
		// If we can't get patient data, return response without patient name
		return dto.ToAppointmentResponse(appointment), nil
	}

	patientName := ""
	if patient != nil {
		patientName = patient.FirstName
		if patient.LastName != nil && *patient.LastName != "" {
			patientName += " " + *patient.LastName
		}
	}

	// Determine if this is the patient's first visit
	isFirstVisit := false
	if patient != nil && patient.FirstAppointmentID != nil && *patient.FirstAppointmentID == appointment.ID {
		isFirstVisit = true
	}

	return dto.ToAppointmentResponseWithPatientNameAndFirstVisit(appointment, patientName, isFirstVisit), nil
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

	// Validate status if provided
	if req.Status != nil {
		if !entities.IsValidAppointmentStatus(entities.AppointmentStatus(*req.Status)) {
			return nil, entities.ErrInvalidAppointmentStatus
		}
	}

	// Verify entities exist only if they're being updated
	if req.PatientID != nil {
		patientExists, err := uc.patientRepo.Exists(ctx, *req.PatientID)
		if err != nil {
			return nil, err
		}
		if !patientExists {
			return nil, entities.ErrPatientNotFound
		}
	}

	if req.DoctorID != nil {
		doctorExists, err := uc.doctorRepo.Exists(ctx, *req.DoctorID)
		if err != nil {
			return nil, err
		}
		if !doctorExists {
			return nil, entities.ErrDoctorNotFound
		}
	}

	// Get clinic timezone - use new unit if provided, otherwise use existing
	var clinic *entities.Clinic
	unitIDToCheck := existing.UnitID
	if req.UnitID != nil {
		unitIDToCheck = req.UnitID
	}

	if unitIDToCheck != nil {
		unit, clinicData, err := uc.unitRepo.GetUnitWithClinic(ctx, *unitIDToCheck)
		if err != nil {
			return nil, err
		}
		if unit == nil || clinicData == nil {
			return nil, entities.ErrUnitNotFound
		}
		clinic = clinicData
	}

	// Convert times from clinic timezone to UTC if times are being updated
	if req.StartTime != nil && clinic != nil && clinic.Timezone != "" {
		loc, err := time.LoadLocation(clinic.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid clinic timezone %q: %w", clinic.Timezone, err)
		}

		// Parse the incoming time as being in the clinic's timezone
		year, month, day := req.StartTime.Date()
		hour, min, sec := req.StartTime.Clock()
		startTimeInClinicTZ := time.Date(year, month, day, hour, min, sec, req.StartTime.Nanosecond(), loc)
		startTimeUTC := startTimeInClinicTZ.UTC()
		req.StartTime = &startTimeUTC
	}

	if req.EndTime != nil && clinic != nil && clinic.Timezone != "" {
		loc, err := time.LoadLocation(clinic.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid clinic timezone %q: %w", clinic.Timezone, err)
		}

		// Parse the incoming time as being in the clinic's timezone
		year, month, day := req.EndTime.Date()
		hour, min, sec := req.EndTime.Clock()
		endTimeInClinicTZ := time.Date(year, month, day, hour, min, sec, req.EndTime.Nanosecond(), loc)
		endTimeUTC := endTimeInClinicTZ.UTC()
		req.EndTime = &endTimeUTC
	}

	// Check if date/time is being changed to automatically set status to rescheduled
	dateChanged := false
	if req.StartTime != nil && !req.StartTime.Equal(existing.StartTime) {
		dateChanged = true
	}
	if req.EndTime != nil && !req.EndTime.Equal(existing.EndTime) {
		dateChanged = true
	}

	updated := req.ToEntityUpdate(existing)

	// If date changed and no explicit status provided, automatically set to rescheduled
	if dateChanged && req.Status == nil {
		updated.Status = entities.AppointmentStatusRescheduled
	}

	// Basic validation: if both start and end time are provided, validate the time logic
	if req.StartTime != nil && req.EndTime != nil {
		if updated.EndTime.Before(updated.StartTime) || updated.EndTime.Equal(updated.StartTime) {
			return nil, fmt.Errorf("end time must be after start time")
		}
	}

	if err := updated.Validate(); err != nil {
		return nil, err
	}

	if err := uc.appointmentRepo.Update(ctx, updated); err != nil {
		return nil, err
	}

	// Fetch patient data to include patient name in response
	patientName := ""
	isFirstVisit := false
	if updated.PatientID != nil {
		patient, err := uc.patientRepo.GetByID(ctx, *updated.PatientID)
		if err == nil && patient != nil {
			patientName = patient.FirstName
			if patient.LastName != nil && *patient.LastName != "" {
				patientName += " " + *patient.LastName
			}
			// Determine if this is the patient's first visit
			if patient.FirstAppointmentID != nil && *patient.FirstAppointmentID == updated.ID {
				isFirstVisit = true
			}
		}
	}

	return dto.ToAppointmentResponseWithPatientNameAndFirstVisit(updated, patientName, isFirstVisit), nil
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

	// Validate date range (max 90 days)
	if endDate.Sub(startDate) > 90*24*time.Hour {
		return nil, fmt.Errorf("date range cannot exceed 90 days")
	}

	// Set default pagination
	if req.Limit <= 0 {
		req.Limit = 2500 // Default limit
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
		// Determine if this is the patient's first visit
		isFirstVisit := false
		if appt.Patient != nil && appt.Patient.FirstAppointmentID != nil && *appt.Patient.FirstAppointmentID == appt.Appointment.ID {
			isFirstVisit = true
		}

		// Build patient DTO
		var patient *dto.PatientListDataDTO
		if appt.Patient != nil && appt.Appointment.PatientID != nil {
			patient = &dto.PatientListDataDTO{
				ID:        appt.Appointment.PatientID.String(),
				FirstName: appt.Patient.FirstName,
				LastName:  appt.Patient.LastName,
				Phone:     appt.Patient.Phone,
				Email:     appt.Patient.Email,
			}
		}

		// Handle nullable foreign keys safely
		var doctorID, doctorName string
		if appt.Appointment.DoctorID != nil {
			doctorID = appt.Appointment.DoctorID.String()
		}
		if appt.Doctor != nil {
			doctorName = appt.Doctor.Name
		}

		var clinicID, clinicName string
		if appt.Clinic != nil {
			clinicID = appt.Clinic.ID.String()
			clinicName = appt.Clinic.Name
		}

		// Convert times to clinic timezone and format as strings
		startTime := appt.Appointment.StartTime
		endTime := appt.Appointment.EndTime
		if appt.Clinic != nil && appt.Clinic.Timezone != "" {
			loc, err := time.LoadLocation(appt.Clinic.Timezone)
			if err == nil {
				startTime = appt.Appointment.StartTime.In(loc)
				endTime = appt.Appointment.EndTime.In(loc)
			}
			// If error loading timezone, just use UTC times
		}

		// Format times as naive datetime strings (without timezone offset)
		const layout = "2006-01-02T15:04:05"
		startTimeStr := startTime.Format(layout)
		endTimeStr := endTime.Format(layout)

		var unitID *string
		var unitName *string
		if appt.Unit != nil {
			unitID = getStringPtrFromUUID(&appt.Unit.ID)
			unitName = &appt.Unit.Name
		}

		appointmentDTOs[i] = dto.AppointmentListResponse{
			ID:           appt.Appointment.ID.String(),
			Patient:      patient,
			DoctorID:     doctorID,
			DoctorName:   doctorName,
			ClinicID:     clinicID,
			ClinicName:   clinicName,
			UnitID:       unitID,
			UnitName:     unitName,
			StartTime:    startTimeStr,
			EndTime:      endTimeStr,
			Status:       string(appt.Appointment.Status),
			ServiceID:    getStringPtr(appt.Appointment.ServiceID),
			ServiceName:  getStringPtr(appt.ServiceName),
			Notes:        getStringPtr(appt.Appointment.Notes),
			IsFirstVisit: isFirstVisit,
			CreatedAt:    appt.Appointment.CreatedAt,
			UpdatedAt:    appt.Appointment.UpdatedAt,
		}

		// Build summary data (only if clinic exists)
		if appt.Clinic != nil {
			clinicIDForStats := appt.Clinic.ID.String()
			if stats, exists := clinicMap[clinicIDForStats]; exists {
				stats.Count++
				clinicMap[clinicIDForStats] = stats
			} else {
				clinicMap[clinicIDForStats] = dto.ClinicStats{
					Count: 1,
					Name:  appt.Clinic.Name,
				}
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

// GetReschedulingQueue retrieves appointments in rescheduling queue with pagination
func (uc *AppointmentUseCase) GetReschedulingQueue(ctx context.Context, orgID uuid.UUID, req *dto.ReschedulingQueueRequest) (*dto.ReschedulingQueueResponse, error) {
	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}

	// Determine sort order
	sortOldest := true // Default to oldest first
	if req.Sort == "newest" {
		sortOldest = false
	}

	// Build filters
	filters := repositories.ReschedulingQueueFilters{
		OrganizationID: orgID,
		Search:         req.Search,
		Page:           req.Page,
		Limit:          req.Limit,
		SortOldest:     sortOldest,
	}

	// Parse optional clinic ID
	if req.ClinicID != nil && *req.ClinicID != "" {
		clinicID, err := uuid.Parse(*req.ClinicID)
		if err != nil {
			return nil, fmt.Errorf("invalid clinic_id: %w", err)
		}
		filters.ClinicID = &clinicID
	}

	// Parse optional doctor ID
	if req.DoctorID != nil && *req.DoctorID != "" {
		doctorID, err := uuid.Parse(*req.DoctorID)
		if err != nil {
			return nil, fmt.Errorf("invalid doctor_id: %w", err)
		}
		filters.DoctorID = &doctorID
	}

	// Fetch appointments from repository
	appointments, totalCount, err := uc.appointmentRepo.GetReschedulingQueue(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get rescheduling queue: %w", err)
	}

	// Convert to DTOs
	items := make([]dto.ReschedulingQueueItem, 0, len(appointments))
	for _, appt := range appointments {
		// Build patient DTO
		var patientDTO *dto.PatientListDataDTO
		if appt.Patient != nil {
			phone := ""
			if appt.Patient.Phone != nil {
				phone = *appt.Patient.Phone
			}
			email := ""
			if appt.Patient.Email != nil {
				email = *appt.Patient.Email
			}
			lastName := ""
			if appt.Patient.LastName != nil {
				lastName = *appt.Patient.LastName
			}

			patientDTO = &dto.PatientListDataDTO{
				ID:        appt.Patient.ID.String(),
				FirstName: appt.Patient.FirstName,
				LastName:  &lastName,
				Phone:     &phone,
				Email:     &email,
			}
		}

		// Get clinic timezone for time conversion
		timezone := "UTC"
		if appt.Clinic != nil && appt.Clinic.Timezone != "" {
			timezone = appt.Clinic.Timezone
		}

		// Load timezone location
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			loc = time.UTC
		}

		// Convert times to clinic timezone
		startTimeInClinicTZ := appt.Appointment.StartTime.In(loc)
		endTimeInClinicTZ := appt.Appointment.EndTime.In(loc)

		// Calculate days in queue
		daysInQueue := 0
		if appt.Appointment.MovedToNeedsReschedulingAt != nil {
			daysInQueue = int(time.Since(*appt.Appointment.MovedToNeedsReschedulingAt).Hours() / 24)
		}

		// Last action timestamp (use moved_to_needs_rescheduling_at or updated_at)
		lastActionTimestamp := appt.Appointment.UpdatedAt
		if appt.Appointment.MovedToNeedsReschedulingAt != nil {
			lastActionTimestamp = *appt.Appointment.MovedToNeedsReschedulingAt
		}

		item := dto.ReschedulingQueueItem{
			ID:                         appt.Appointment.ID.String(),
			Patient:                    patientDTO,
			DoctorID:                   "",
			DoctorName:                 "",
			ClinicID:                   "",
			ClinicName:                 "",
			OriginalStart:              startTimeInClinicTZ.Format(time.RFC3339),
			OriginalEnd:                endTimeInClinicTZ.Format(time.RFC3339),
			ServiceName:                "",
			Notes:                      "",
			MovedToNeedsReschedulingAt: "",
			DaysInQueue:                daysInQueue,
			LastActionTimestamp:        lastActionTimestamp.Format(time.RFC3339),
		}

		// Add doctor info
		if appt.Doctor != nil {
			item.DoctorID = appt.Doctor.ID.String()
			item.DoctorName = appt.Doctor.Name
		}

		// Add clinic info
		if appt.Clinic != nil {
			item.ClinicID = appt.Clinic.ID.String()
			item.ClinicName = appt.Clinic.Name
		}

		// Add unit info
		if appt.Unit != nil {
			unitID := appt.Unit.ID.String()
			item.UnitID = &unitID
			unitName := appt.Unit.Name
			item.UnitName = &unitName
		}

		// Add service name
		if appt.ServiceName != nil {
			item.ServiceName = *appt.ServiceName
		}

		// Add notes
		if appt.Appointment.Notes != nil {
			item.Notes = *appt.Appointment.Notes
		}

		// Add moved timestamp
		if appt.Appointment.MovedToNeedsReschedulingAt != nil {
			item.MovedToNeedsReschedulingAt = appt.Appointment.MovedToNeedsReschedulingAt.Format(time.RFC3339)
		}

		items = append(items, item)
	}

	// Calculate total pages
	totalPages := (totalCount + req.Limit - 1) / req.Limit

	return &dto.ReschedulingQueueResponse{
		Items:      items,
		Total:      totalCount,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

// CancelFromQueue cancels an appointment from the rescheduling queue
func (uc *AppointmentUseCase) CancelFromQueue(ctx context.Context, appointmentID uuid.UUID, orgID uuid.UUID, req *dto.CancelAppointmentRequest) error {
	// Get appointment to verify it exists and belongs to organization
	appointment, err := uc.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return err
	}
	if appointment == nil {
		return entities.ErrAppointmentNotFound
	}

	// Verify appointment status is needs-rescheduling
	if appointment.Status != entities.AppointmentStatusNeedsRescheduling {
		return entities.ErrAppointmentNotInQueue
	}

	// Verify appointment belongs to organization (via unit -> clinic -> organization)
	if appointment.UnitID != nil {
		unit, clinic, err := uc.unitRepo.GetUnitWithClinic(ctx, *appointment.UnitID)
		if err != nil {
			return err
		}
		if unit == nil || clinic == nil || clinic.OrganizationID != orgID {
			return fmt.Errorf("appointment does not belong to organization")
		}
	}

	// Combine reason and notes
	fullReason := req.Reason
	if req.Notes != nil && *req.Notes != "" {
		fullReason = fmt.Sprintf("%s - %s", req.Reason, *req.Notes)
	}

	// Cancel with reason
	return uc.appointmentRepo.CancelWithReason(ctx, appointmentID, fullReason)
}

// RescheduleFromQueue reschedules an appointment from the queue by creating a new one
func (uc *AppointmentUseCase) RescheduleFromQueue(ctx context.Context, appointmentID uuid.UUID, orgID uuid.UUID, req *dto.RescheduleFromQueueRequest) (*dto.AppointmentResponse, error) {
	// Get original appointment
	original, err := uc.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, entities.ErrAppointmentNotFound
	}

	// Verify appointment status is needs-rescheduling
	if original.Status != entities.AppointmentStatusNeedsRescheduling {
		return nil, entities.ErrAppointmentNotInQueue
	}

	// Verify appointment belongs to organization
	if original.UnitID != nil {
		unit, clinic, err := uc.unitRepo.GetUnitWithClinic(ctx, *original.UnitID)
		if err != nil {
			return nil, err
		}
		if unit == nil || clinic == nil || clinic.OrganizationID != orgID {
			return nil, fmt.Errorf("appointment does not belong to organization")
		}
	}

	// Verify entities exist
	doctorExists, err := uc.doctorRepo.Exists(ctx, req.DoctorID)
	if err != nil {
		return nil, err
	}
	if !doctorExists {
		return nil, entities.ErrDoctorNotFound
	}

	// Get unit with clinic for timezone conversion
	unit, clinic, err := uc.unitRepo.GetUnitWithClinic(ctx, req.UnitID)
	if err != nil {
		return nil, err
	}
	if unit == nil || clinic == nil {
		return nil, entities.ErrUnitNotFound
	}

	// Verify new unit belongs to same organization
	if clinic.OrganizationID != orgID {
		return nil, fmt.Errorf("unit does not belong to organization")
	}

	// Convert times from clinic timezone to UTC
	startTimeUTC := req.StartTime
	endTimeUTC := req.EndTime

	if clinic.Timezone != "" {
		loc, err := time.LoadLocation(clinic.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid clinic timezone %q: %w", clinic.Timezone, err)
		}

		// Parse times as being in clinic's timezone
		year, month, day := req.StartTime.Date()
		hour, min, sec := req.StartTime.Clock()
		startTimeInClinicTZ := time.Date(year, month, day, hour, min, sec, req.StartTime.Nanosecond(), loc)
		startTimeUTC = startTimeInClinicTZ.UTC()

		year, month, day = req.EndTime.Date()
		hour, min, sec = req.EndTime.Clock()
		endTimeInClinicTZ := time.Date(year, month, day, hour, min, sec, req.EndTime.Nanosecond(), loc)
		endTimeUTC = endTimeInClinicTZ.UTC()
	}

	// Validate times
	if endTimeUTC.Before(startTimeUTC) || endTimeUTC.Equal(startTimeUTC) {
		return nil, fmt.Errorf("end time must be after start time")
	}

	// Create new appointment (will be saved in transaction)
	newAppointment := &entities.Appointment{
		ID:        uuid.New(),
		PatientID: original.PatientID,
		DoctorID:  &req.DoctorID,
		UnitID:    &req.UnitID,
		ServiceID: &req.ServiceID,
		Status:    entities.AppointmentStatusScheduled,
		StartTime: startTimeUTC,
		EndTime:   endTimeUTC,
		Notes:     req.Notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Validate the new appointment
	if err := newAppointment.Validate(); err != nil {
		return nil, err
	}

	// Check for conflicts with existing appointments
	hasConflict, err := uc.appointmentRepo.CheckConflict(
		ctx,
		req.DoctorID,
		req.UnitID,
		startTimeUTC,
		endTimeUTC,
		nil, // No appointment to exclude
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check for conflicts: %w", err)
	}
	if hasConflict {
		return nil, entities.ErrAppointmentConflict
	}

	// Create new appointment
	if err := uc.appointmentRepo.Create(ctx, newAppointment); err != nil {
		return nil, fmt.Errorf("failed to create new appointment: %w", err)
	}

	// Update original appointment to link to new one
	original.LinkToRescheduledAppointment(newAppointment.ID)
	if err := uc.appointmentRepo.Update(ctx, original); err != nil {
		// If this fails, we should ideally rollback the creation
		// For now, we'll return the error
		return nil, fmt.Errorf("failed to update original appointment: %w", err)
	}

	// Build response
	patientName := ""
	isFirstVisit := false
	if newAppointment.PatientID != nil {
		patient, err := uc.patientRepo.GetByID(ctx, *newAppointment.PatientID)
		if err == nil && patient != nil {
			patientName = patient.FirstName
			if patient.LastName != nil && *patient.LastName != "" {
				patientName += " " + *patient.LastName
			}
			if patient.FirstAppointmentID != nil && *patient.FirstAppointmentID == newAppointment.ID {
				isFirstVisit = true
			}
		}
	}

	return dto.ToAppointmentResponseWithPatientNameAndFirstVisit(newAppointment, patientName, isFirstVisit), nil
}
