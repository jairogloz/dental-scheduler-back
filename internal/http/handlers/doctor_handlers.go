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

// DoctorHandler handles doctor-related HTTP requests
type DoctorHandler struct {
	doctorUseCase *usecases.DoctorUseCase
	logger        *logger.Logger
}

// NewDoctorHandler creates a new doctor handler
func NewDoctorHandler(doctorUseCase *usecases.DoctorUseCase, logger *logger.Logger) *DoctorHandler {
	return &DoctorHandler{
		doctorUseCase: doctorUseCase,
		logger:        logger,
	}
}

// CreateDoctor handles POST /doctors (placeholder)
func (h *DoctorHandler) CreateDoctor(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

// GetDoctorsByOrganization handles GET /doctors?orgId=...&clinicId=...
func (h *DoctorHandler) GetDoctorsByOrganization(c *gin.Context) {
	var req dto.GetDoctorsByOrgRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	// Try to get the organization ID from context (set by middleware)
	orgIDFromCtx, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Error("Organization ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Organization ID not found"})
		return
	}

	// q: parse orgIDFromCtx to uuid.UUID
	orgID, err := uuid.Parse(orgIDFromCtx)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid organization ID format in context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
		return
	}

	// Parse the clinic ID if provided
	clinicID, err := req.ParsedClinicID()
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid clinic ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clinic ID format"})
		return
	}

	doctors, err := h.doctorUseCase.GetDoctorsByOrganizationID(c.Request.Context(), orgID, clinicID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get doctors by organization")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get doctors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": doctors})
}

// PatientHandler handles patient-related HTTP requests
type PatientHandler struct {
	patientUseCase *usecases.PatientUseCase
	logger         *logger.Logger
}

// NewPatientHandler creates a new patient handler
func NewPatientHandler(patientUseCase *usecases.PatientUseCase, logger *logger.Logger) *PatientHandler {
	return &PatientHandler{
		patientUseCase: patientUseCase,
		logger:         logger,
	}
}

// CreatePatient handles POST /patients (placeholder)
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

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

// CreateAppointment handles POST /appointments with organization-aware functionality
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	// Get user information from context (set by middleware)
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		h.logger.Logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// Get organization ID from context (fetched from database by middleware)
	organizationID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Error("Organization ID not found in context")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "No organization access",
			},
		})
		return
	}

	// Optional: Get full organization details if needed
	organization, orgExists := middleware.GetOrganizationFromContext(c)
	if orgExists {
		h.logger.Logger.WithFields(map[string]interface{}{
			"user_id":         userID,
			"organization_id": organizationID,
			"org_name":        organization.Name,
		}).Info("Creating appointment with organization context")
	}

	// Your appointment creation logic here...
	// You now have:
	// - userID: the authenticated user's ID from JWT
	// - organizationID: the user's organization ID from database
	// - organization: full organization details (optional)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Appointment creation logic would go here",
		"data": gin.H{
			"user_id":         userID,
			"organization_id": organizationID,
		},
	})
}
