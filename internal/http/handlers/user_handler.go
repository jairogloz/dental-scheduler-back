package handlers

import (
	"net/http"

	"dental-scheduler-backend/internal/http/middleware"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	logger *logger.Logger
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(logger *logger.Logger) *UserHandler {
	return &UserHandler{
		logger: logger,
	}
}

// GetProfile returns the current user's profile with organization information
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Check if user is authenticated
	_, exists := middleware.GetUserIDFromContext(c)
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

	// Try to get full user profile from database (if middleware fetched it)
	if userProfile, exists := middleware.GetUserProfileFromContext(c); exists {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"profile":      userProfile.Profile,
				"organization": userProfile.Organization,
			},
		})
		return
	}

	// Fallback: Get basic user info from JWT if database lookup wasn't available
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Logger.Error("User not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Unable to retrieve user information",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user": gin.H{
				"id":    user.ID,
				"email": user.Email,
				"roles": user.Roles,
			},
			"organization": nil, // No organization data available from JWT only
		},
	})
}

// GetOrganization returns the current user's organization information
func (h *UserHandler) GetOrganization(c *gin.Context) {
	// Try to get organization from context (set by middleware)
	organization, exists := middleware.GetOrganizationFromContext(c)
	if !exists {
		h.logger.Logger.Debug("Organization not found in context")
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Organization not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    organization,
	})
}
