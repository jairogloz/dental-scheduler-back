package handlers

import (
	"net/http"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/app/usecases"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/http/middleware"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AppointmentHandler handles appointment-related HTTP requests
type AppointmentHandler struct {
	appointmentUseCase *usecases.AppointmentUseCase
	logger             *logger.Logger
}

// NewAppointmentHandler creates a new appointment handler
func NewAppointmentHandler(appointmentUseCase *usecases.AppointmentUseCase, logger *logger.Logger) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentUseCase: appointmentUseCase,
		logger:             logger,
	}
}

// GetAppointments retrieves appointments for an organization with filters
// @Summary Get appointments by organization
// @Description Retrieves appointments for an organization within a date range with optional filters
// @Tags appointments
// @Accept json
// @Produce json
// @Param startDate query string true "Start date (YYYY-MM-DD)"
// @Param endDate query string true "End date (YYYY-MM-DD)"
// @Param clinicId query string false "Filter by clinic ID"
// @Param doctorId query string false "Filter by doctor ID"
// @Param status query string false "Filter by status (scheduled, confirmed, completed, cancelled)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 50, max: 100)"
// @Success 200 {object} dto.GetAppointmentsResponse
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 403 {object} ErrorResponse "No organization access"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /appointments [get]
func (h *AppointmentHandler) GetAppointments(c *gin.Context) {
	// Get organization from context (from middleware)
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Warn("GetAppointments called without organization context")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NO_ORGANIZATION",
				"message": "User not associated with an organization",
			},
		})
		return
	}

	// Bind query parameters
	var req dto.GetAppointmentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid query parameters for GetAppointments")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PARAMETERS",
				"message": err.Error(),
			},
		})
		return
	}

	// Override orgId from context (security measure)
	req.OrgID = orgID

	// Validate and set pagination limits
	if req.Limit > 100 {
		req.Limit = 100 // Max limit
	}

	h.logger.Logger.WithFields(map[string]interface{}{
		"org_id":     orgID,
		"start_date": req.StartDate,
		"end_date":   req.EndDate,
		"clinic_id":  req.ClinicID,
		"doctor_id":  req.DoctorID,
		"status":     req.Status,
		"page":       req.Page,
		"limit":      req.Limit,
	}).Info("Fetching appointments for organization")

	// Execute use case
	response, err := h.appointmentUseCase.GetAppointmentsByOrganization(c.Request.Context(), &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get appointments")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve appointments",
			},
		})
		return
	}

	h.logger.Logger.WithFields(map[string]interface{}{
		"org_id":             orgID,
		"appointments_count": len(response.Appointments),
		"total_count":        response.Pagination.Total,
	}).Info("Successfully retrieved appointments")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreateAppointment creates a new appointment
// @Summary Create a new appointment
// @Description Creates a new appointment with conflict checking
// @Tags appointments
// @Accept json
// @Produce json
// @Param appointment body dto.CreateAppointmentRequest true "Appointment data"
// @Success 201 {object} dto.AppointmentResponse
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 409 {object} ErrorResponse "Schedule conflict"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /appointments [post]
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	// Get organization ID from context (set by auth middleware)
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Error("Organization ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Organization context required",
			},
		})
		return
	}

	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.logger.Logger.Error("Invalid organization ID format in context", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_CONTEXT",
				"message": "Invalid organization context",
			},
		})
		return
	}

	var req dto.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid JSON for CreateAppointment")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	h.logger.Logger.WithFields(map[string]interface{}{
		"organization_id": orgUUID,
		"patient_id":      req.PatientID,
		"doctor_id":       req.DoctorID,
		"unit_id":         req.UnitID,
		"start_time":      req.StartTime,
		"end_time":        req.EndTime,
	}).Info("Creating new appointment")

	response, err := h.appointmentUseCase.CreateAppointment(c.Request.Context(), orgUUID, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to create appointment")

		// Handle specific error types
		if err.Error() == "schedule conflict detected" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SCHEDULE_CONFLICT",
					"message": "The requested time slot conflicts with existing appointments",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to create appointment",
			},
		})
		return
	}

	h.logger.Logger.WithField("appointment_id", response.ID).Info("Successfully created appointment")

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetAppointmentByID retrieves a specific appointment by ID
// @Summary Get appointment by ID
// @Description Retrieves a specific appointment by its ID
// @Tags appointments
// @Accept json
// @Produce json
// @Param id path string true "Appointment ID"
// @Success 200 {object} dto.AppointmentWithDetailsResponse
// @Failure 400 {object} ErrorResponse "Invalid appointment ID"
// @Failure 404 {object} ErrorResponse "Appointment not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /appointments/{id} [get]
func (h *AppointmentHandler) GetAppointmentByID(c *gin.Context) {
	appointmentID := c.Param("id")
	if appointmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Appointment ID is required",
			},
		})
		return
	}
	appointmentIDUUID, err := uuid.Parse(appointmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid appointment ID format",
			},
		})
		return
	}

	h.logger.Logger.WithField("appointment_id", appointmentID).Info("Fetching appointment by ID")

	response, err := h.appointmentUseCase.GetAppointmentByID(c.Request.Context(), appointmentIDUUID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get appointment by ID")

		if err.Error() == "appointment not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "NOT_FOUND",
					"message": "Appointment not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve appointment",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateAppointment handles PATCH /appointments/{appointment_id}
func (h *AppointmentHandler) UpdateAppointment(c *gin.Context) {
	// Get appointment ID from path parameter
	appointmentIDStr := c.Param("appointment_id")
	appointmentID, err := uuid.Parse(appointmentIDStr)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid appointment ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_APPOINTMENT_ID",
				"message": "Invalid appointment ID format. Must be a valid UUID.",
			},
		})
		return
	}

	// Get organization ID from context (set by auth middleware)
	orgIDStr, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Error("Organization ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Organization context required",
			},
		})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.logger.Logger.Error("Invalid organization ID format in context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Invalid organization context",
			},
		})
		return
	}

	// Bind and validate request body
	var req dto.UpdateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind update appointment request")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Log the request
	logFields := map[string]interface{}{
		"appointment_id":  appointmentID,
		"organization_id": orgID,
	}
	if req.PatientID != nil {
		logFields["patient_id"] = *req.PatientID
	}
	if req.DoctorID != nil {
		logFields["doctor_id"] = *req.DoctorID
	}
	if req.UnitID != nil {
		logFields["unit_id"] = *req.UnitID
	}
	h.logger.Logger.WithFields(logFields).Info("Updating appointment")

	// Execute use case
	result, err := h.appointmentUseCase.UpdateAppointment(c.Request.Context(), appointmentID, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to update appointment")

		// Handle specific domain errors
		switch err {
		case entities.ErrAppointmentNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "APPOINTMENT_NOT_FOUND",
					"message": "Appointment not found",
				},
			})
		case entities.ErrPatientNotFound:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "PATIENT_NOT_FOUND",
					"message": "Patient not found",
				},
			})
		case entities.ErrDoctorNotFound:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DOCTOR_NOT_FOUND",
					"message": "Doctor not found",
				},
			})
		case entities.ErrUnitNotFound:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNIT_NOT_FOUND",
					"message": "Unit not found",
				},
			})
		default:
			// Handle validation errors and other errors
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_REQUEST",
					"message": err.Error(),
				},
			})
		}
		return
	}

	// Log success
	h.logger.Logger.WithFields(map[string]interface{}{
		"appointment_id": appointmentID,
		"patient_id":     result.PatientID,
		"doctor_id":      result.DoctorID,
		"unit_id":        result.UnitID,
	}).Info("Successfully updated appointment")

	// Return successful response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetReschedulingQueue retrieves appointments in rescheduling queue with pagination
// @Summary Get rescheduling queue
// @Description Retrieves paginated list of appointments that need rescheduling
// @Tags appointments
// @Accept json
// @Produce json
// @Param clinic_id query string false "Filter by clinic ID"
// @Param doctor_id query string false "Filter by doctor ID"
// @Param search query string false "Search in patient name, phone, or email"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param sort query string false "Sort order: 'oldest' or 'newest' (default: 'oldest')"
// @Success 200 {object} dto.ReschedulingQueueResponse
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 403 {object} ErrorResponse "No organization access"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /appointments/rescheduling-queue [get]
func (h *AppointmentHandler) GetReschedulingQueue(c *gin.Context) {
	// Get organization from context
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Warn("GetReschedulingQueue called without organization context")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NO_ORGANIZATION",
				"message": "User not associated with an organization",
			},
		})
		return
	}

	// Bind query parameters
	var req dto.ReschedulingQueueRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid query parameters for GetReschedulingQueue")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_PARAMETERS",
				"message": err.Error(),
			},
		})
		return
	}

	h.logger.Logger.WithFields(map[string]interface{}{
		"org_id":    orgID,
		"clinic_id": req.ClinicID,
		"doctor_id": req.DoctorID,
		"search":    req.Search,
		"page":      req.Page,
		"limit":     req.Limit,
		"sort":      req.Sort,
	}).Info("Fetching rescheduling queue for organization")

	// Parse organization ID
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid organization ID")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ORGANIZATION",
				"message": "Invalid organization ID",
			},
		})
		return
	}

	// Execute use case
	response, err := h.appointmentUseCase.GetReschedulingQueue(c.Request.Context(), orgUUID, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get rescheduling queue")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve rescheduling queue",
			},
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// CancelFromQueue cancels an appointment from the rescheduling queue
// @Summary Cancel appointment from queue
// @Description Cancels an appointment that is in the rescheduling queue
// @Tags appointments
// @Accept json
// @Produce json
// @Param appointment_id path string true "Appointment ID"
// @Param request body dto.CancelAppointmentRequest true "Cancellation details"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 403 {object} ErrorResponse "No organization access"
// @Failure 404 {object} ErrorResponse "Appointment not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /appointments/{appointment_id}/cancel [post]
func (h *AppointmentHandler) CancelFromQueue(c *gin.Context) {
	// Get appointment ID from path
	appointmentIDStr := c.Param("appointment_id")
	appointmentID, err := uuid.Parse(appointmentIDStr)
	if err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid appointment ID")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_APPOINTMENT_ID",
				"message": "Invalid appointment ID format",
			},
		})
		return
	}

	// Get organization from context
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Warn("CancelFromQueue called without organization context")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NO_ORGANIZATION",
				"message": "User not associated with an organization",
			},
		})
		return
	}

	// Parse organization ID
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid organization ID")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ORGANIZATION",
				"message": "Invalid organization ID",
			},
		})
		return
	}

	// Bind request body
	var req dto.CancelAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid request body for CancelFromQueue")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	h.logger.Logger.WithFields(map[string]interface{}{
		"appointment_id": appointmentID,
		"org_id":         orgID,
		"reason":         req.Reason,
	}).Info("Cancelling appointment from queue")

	// Execute use case
	err = h.appointmentUseCase.CancelFromQueue(c.Request.Context(), appointmentID, orgUUID, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to cancel appointment from queue")

		// Handle specific domain errors
		switch err {
		case entities.ErrAppointmentNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "APPOINTMENT_NOT_FOUND",
					"message": "Appointment not found",
				},
			})
		case entities.ErrAppointmentNotInQueue:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "APPOINTMENT_NOT_IN_QUEUE",
					"message": "Appointment is not in rescheduling queue",
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": err.Error(),
				},
			})
		}
		return
	}

	// Log success
	h.logger.Logger.WithField("appointment_id", appointmentID).Info("Successfully cancelled appointment from queue")

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Appointment cancelled successfully",
	})
}

// RescheduleFromQueue reschedules an appointment from the queue by creating a new one
// @Summary Reschedule appointment from queue
// @Description Creates a new appointment and marks the original as rescheduled
// @Tags appointments
// @Accept json
// @Produce json
// @Param appointment_id path string true "Appointment ID"
// @Param request body dto.RescheduleFromQueueRequest true "New appointment details"
// @Success 200 {object} map[string]interface{} "Success response with new appointment"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 403 {object} ErrorResponse "No organization access"
// @Failure 404 {object} ErrorResponse "Appointment not found"
// @Failure 409 {object} ErrorResponse "Time slot conflict"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /appointments/{appointment_id}/reschedule [post]
func (h *AppointmentHandler) RescheduleFromQueue(c *gin.Context) {
	// Get appointment ID from path
	appointmentIDStr := c.Param("appointment_id")
	appointmentID, err := uuid.Parse(appointmentIDStr)
	if err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid appointment ID")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_APPOINTMENT_ID",
				"message": "Invalid appointment ID format",
			},
		})
		return
	}

	// Get organization from context
	orgID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Warn("RescheduleFromQueue called without organization context")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NO_ORGANIZATION",
				"message": "User not associated with an organization",
			},
		})
		return
	}

	// Parse organization ID
	orgUUID, err := uuid.Parse(orgID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid organization ID")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ORGANIZATION",
				"message": "Invalid organization ID",
			},
		})
		return
	}

	// Bind request body
	var req dto.RescheduleFromQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid request body for RescheduleFromQueue")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	h.logger.Logger.WithFields(map[string]interface{}{
		"appointment_id": appointmentID,
		"org_id":         orgID,
		"doctor_id":      req.DoctorID,
		"unit_id":        req.UnitID,
		"start_time":     req.StartTime,
		"end_time":       req.EndTime,
	}).Info("Rescheduling appointment from queue")

	// Execute use case
	newAppointment, err := h.appointmentUseCase.RescheduleFromQueue(c.Request.Context(), appointmentID, orgUUID, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to reschedule appointment from queue")

		// Handle specific domain errors
		switch err {
		case entities.ErrAppointmentNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "APPOINTMENT_NOT_FOUND",
					"message": "Appointment not found",
				},
			})
		case entities.ErrAppointmentNotInQueue:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "APPOINTMENT_NOT_IN_QUEUE",
					"message": "Appointment is not in rescheduling queue",
				},
			})
		case entities.ErrAppointmentConflict:
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TIME_SLOT_CONFLICT",
					"message": "The selected time slot conflicts with an existing appointment",
				},
			})
		case entities.ErrDoctorNotFound:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DOCTOR_NOT_FOUND",
					"message": "Doctor not found",
				},
			})
		case entities.ErrUnitNotFound:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNIT_NOT_FOUND",
					"message": "Unit not found",
				},
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_REQUEST",
					"message": err.Error(),
				},
			})
		}
		return
	}

	// Log success
	h.logger.Logger.WithFields(map[string]interface{}{
		"original_appointment_id": appointmentID,
		"new_appointment_id":      newAppointment.ID,
	}).Info("Successfully rescheduled appointment from queue")

	// Return success response with new appointment
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    newAppointment,
	})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool `json:"success"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
