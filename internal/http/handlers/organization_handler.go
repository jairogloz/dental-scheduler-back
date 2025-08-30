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

// OrganizationHandler handles organization-related HTTP requests
type OrganizationHandler struct {
	getOrgDataUseCase *usecases.GetOrganizationDataUseCase
	logger            *logger.Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(getOrgDataUseCase *usecases.GetOrganizationDataUseCase, logger *logger.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		getOrgDataUseCase: getOrgDataUseCase,
		logger:            logger,
	}
}

// GetOrganizationData handles GET /organization requests for loading complete organization data
func (h *OrganizationHandler) GetOrganizationData(c *gin.Context) {
	// Get organization ID from context (set by middleware)
	orgIDValue, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		h.logger.Logger.Error("Organization ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Organization context not found",
			},
		})
		return
	}

	orgID, err := uuid.Parse(orgIDValue)
	if err != nil {
		h.logger.Logger.Error("Invalid organization ID type in context", "error", err)
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
	var req dto.OrganizationDataRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind organization data request")
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
		"organization_id": orgID,
		"start_date":      req.StartDate,
		"end_date":        req.EndDate,
		"limit":           req.Limit,
	}).Info("Getting organization data")

	// Execute use case
	result, err := h.getOrgDataUseCase.Execute(c.Request.Context(), orgID, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get organization data")

		// Check for specific domain errors
		if err == entities.ErrOrganizationNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "ORGANIZATION_NOT_FOUND",
					"message": "Organization not found",
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
		"organization_id":    orgID,
		"clinics_count":      len(result.Clinics),
		"units_count":        len(result.Units),
		"doctors_count":      len(result.Doctors),
		"appointments_count": len(result.Appointments),
	}).Info("Successfully retrieved organization data")

	// Return successful response with caching headers
	c.Header("Cache-Control", "public, max-age=300") // Cache for 5 minutes
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
