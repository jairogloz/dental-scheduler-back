package handlers

import (
	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/http/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SearchPatients handles GET /patients/search
func (h *PatientHandler) SearchPatients(c *gin.Context) {
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

	// Bind query parameters
	var req dto.PatientSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Logger.WithError(err).Warn("Invalid query parameters for SearchPatients")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	// Validate query parameter
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "MISSING_QUERY",
				"message": "Search query is required",
			},
		})
		return
	}

	h.logger.Logger.WithFields(map[string]interface{}{
		"organization_id": orgUUID,
		"query":           req.Query,
		"limit":           req.Limit,
	}).Info("Searching patients")

	// Call use case
	result, err := h.patientUseCase.SearchPatients(c.Request.Context(), orgUUID, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to search patients")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "SEARCH_FAILED",
				"message": "Failed to search patients",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
