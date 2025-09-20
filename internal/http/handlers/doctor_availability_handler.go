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

// DoctorAvailabilityHandler handles doctor availability-related HTTP requests
type DoctorAvailabilityHandler struct {
	getDoctorAvailabilityUseCase *usecases.GetDoctorAvailabilityUseCase
	logger                       *logger.Logger
}

// NewDoctorAvailabilityHandler creates a new doctor availability handler
func NewDoctorAvailabilityHandler(
	getDoctorAvailabilityUseCase *usecases.GetDoctorAvailabilityUseCase,
	logger *logger.Logger,
) *DoctorAvailabilityHandler {
	return &DoctorAvailabilityHandler{
		getDoctorAvailabilityUseCase: getDoctorAvailabilityUseCase,
		logger:                       logger,
	}
}

// GetDoctorAvailability handles GET /doctor-availability/{doctor_id}
func (h *DoctorAvailabilityHandler) GetDoctorAvailability(c *gin.Context) {
	// Get doctor ID from path parameter
	doctorIDStr := c.Param("doctor_id")
	doctorID, err := uuid.Parse(doctorIDStr)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid doctor ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_DOCTOR_ID",
				"message": "Invalid doctor ID format. Must be a valid UUID.",
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

	// Bind and validate query parameters
	var req dto.GetDoctorAvailabilityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind doctor availability request")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request parameters",
				"details": err.Error(),
			},
		})
		return
	}

	// Log the request
	h.logger.Logger.WithFields(map[string]interface{}{
		"doctor_id":       doctorID,
		"organization_id": orgID,
		"start_date":      req.StartDate,
		"end_date":        req.EndDate,
	}).Info("Getting doctor availability")

	// Execute use case
	result, err := h.getDoctorAvailabilityUseCase.Execute(c.Request.Context(), doctorID, orgID, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get doctor availability")

		// Check for specific domain errors
		if err == entities.ErrDoctorNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "DOCTOR_NOT_FOUND",
					"message": "Doctor not found",
				},
			})
			return
		}

		// Handle validation errors
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	// Log success
	h.logger.Logger.WithFields(map[string]interface{}{
		"doctor_id":            doctorID,
		"availabilities_count": len(result.Availabilities),
	}).Info("Successfully retrieved doctor availability")

	// Return successful response
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
