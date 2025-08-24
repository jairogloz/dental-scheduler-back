package handlers

import (
	"net/http"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/app/usecases"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Response helpers
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

func notFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

func internalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

func validationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// ClinicHandler handles clinic-related HTTP requests
type ClinicHandler struct {
	clinicUseCase *usecases.ClinicUseCase
	logger        *logger.Logger
}

// NewClinicHandler creates a new clinic handler
func NewClinicHandler(clinicUseCase *usecases.ClinicUseCase, logger *logger.Logger) *ClinicHandler {
	return &ClinicHandler{
		clinicUseCase: clinicUseCase,
		logger:        logger,
	}
}

// CreateClinic handles POST /clinics
func (h *ClinicHandler) CreateClinic(c *gin.Context) {
	var req dto.CreateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind clinic creation request")
		validationError(c, err.Error())
		return
	}

	clinic, err := h.clinicUseCase.CreateClinic(c.Request.Context(), &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to create clinic")
		if err == entities.ErrInvalidClinicName {
			badRequest(c, err.Error())
			return
		}
		internalServerError(c, "Failed to create clinic")
		return
	}

	created(c, clinic)
}

// GetClinic handles GET /clinics/:id
func (h *ClinicHandler) GetClinic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		badRequest(c, "Invalid clinic ID format")
		return
	}

	clinic, err := h.clinicUseCase.GetClinicByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get clinic")
		if err == entities.ErrClinicNotFound {
			notFound(c, "Clinic not found")
			return
		}
		internalServerError(c, "Failed to get clinic")
		return
	}

	success(c, clinic)
}

// GetClinics handles GET /clinics
func (h *ClinicHandler) GetClinics(c *gin.Context) {
	clinics, err := h.clinicUseCase.GetAllClinics(c.Request.Context())
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get clinics")
		internalServerError(c, "Failed to get clinics")
		return
	}

	success(c, clinics)
}

// UpdateClinic handles PUT /clinics/:id
func (h *ClinicHandler) UpdateClinic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		badRequest(c, "Invalid clinic ID format")
		return
	}

	var req dto.UpdateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind clinic update request")
		validationError(c, err.Error())
		return
	}

	clinic, err := h.clinicUseCase.UpdateClinic(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to update clinic")
		if err == entities.ErrClinicNotFound {
			notFound(c, "Clinic not found")
			return
		}
		if err == entities.ErrInvalidClinicName {
			badRequest(c, err.Error())
			return
		}
		internalServerError(c, "Failed to update clinic")
		return
	}

	success(c, clinic)
}

// DeleteClinic handles DELETE /clinics/:id
func (h *ClinicHandler) DeleteClinic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		badRequest(c, "Invalid clinic ID format")
		return
	}

	err = h.clinicUseCase.DeleteClinic(c.Request.Context(), id)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to delete clinic")
		if err == entities.ErrClinicNotFound {
			notFound(c, "Clinic not found")
			return
		}
		internalServerError(c, "Failed to delete clinic")
		return
	}

	c.Status(http.StatusNoContent)
}
