package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// Check handles health check requests
func (h *HealthHandler) Check(c *gin.Context) {
	healthResponse := HealthResponse{
		Status:  "healthy",
		Service: "dental-scheduler-backend",
		Version: "1.0.0",
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    healthResponse,
	})
}

// Readiness handles readiness check requests
func (h *HealthHandler) Readiness(c *gin.Context) {
	// In a real application, you would check database connectivity, etc.
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// Liveness handles liveness check requests
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
