package handlers

import (
	"net/http"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/app/usecases"
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
	orgID, exists := c.Get("organization_id")
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

	orgUUID, ok := orgID.(uuid.UUID)
	if !ok {
		h.logger.Logger.Error("Invalid organization ID format in context")
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

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool `json:"success"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
