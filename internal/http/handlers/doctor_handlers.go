package handlers

import (
	"net/http"
	"strings"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/app/usecases"
	"dental-scheduler-backend/internal/domain/entities"
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

// CreatePatient handles POST /patients
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	var req dto.CreatePatientWithOrgRequest

	// Bind JSON body
	if err := c.ShouldBindJSON(&req.CreatePatientRequest); err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid JSON for CreatePatient")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	// Bind query parameters (organization_id)
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid query parameters for CreatePatient")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_QUERY_PARAMS",
				"message": err.Error(),
			},
		})
		return
	}

	h.logger.Logger.WithFields(map[string]interface{}{
		"patient_first_name": req.FirstName,
		"patient_last_name":  req.LastName,
		"organization_id":    req.OrganizationIDStr,
	}).Info("Creating new patient")

	// Call use case
	response, err := h.patientUseCase.CreatePatientWithOrganization(c.Request.Context(), &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to create patient")

		// Handle specific error types
		if err == entities.ErrOrganizationNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ORGANIZATION_NOT_FOUND",
					"message": "Organization not found",
				},
			})
			return
		}

		// Check for UUID parsing errors
		if strings.Contains(err.Error(), "invalid organization_id format") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_ORGANIZATION_ID",
					"message": "Invalid organization_id format. Must be a valid UUID.",
				},
			})
			return
		}

		// Check for patient validation errors
		if err == entities.ErrInvalidPatientName {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_PATIENT_NAME",
					"message": "Patient name is required",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to create patient",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}
