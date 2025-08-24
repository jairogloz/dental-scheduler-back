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

// UnitHandler handles unit-related HTTP requests
type UnitHandler struct {
	unitUseCase *usecases.UnitUseCase
	logger      *logger.Logger
}

// NewUnitHandler creates a new unit handler
func NewUnitHandler(unitUseCase *usecases.UnitUseCase, logger *logger.Logger) *UnitHandler {
	return &UnitHandler{
		unitUseCase: unitUseCase,
		logger:      logger,
	}
}

// CreateUnit handles POST /units
func (h *UnitHandler) CreateUnit(c *gin.Context) {
	var req dto.CreateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind unit creation request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unit, err := h.unitUseCase.CreateUnit(c.Request.Context(), &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to create unit")
		if err == entities.ErrInvalidUnitName || err == entities.ErrInvalidClinicID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": unit})
}

// GetUnit handles GET /units/:id
func (h *UnitHandler) GetUnit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID format"})
		return
	}

	unit, err := h.unitUseCase.GetUnitByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get unit")
		if err == entities.ErrUnitNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": unit})
}

// GetUnits handles GET /units
func (h *UnitHandler) GetUnits(c *gin.Context) {
	units, err := h.unitUseCase.GetAllUnits(c.Request.Context())
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get units")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get units"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": units})
}

// UpdateUnit handles PUT /units/:id
func (h *UnitHandler) UpdateUnit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID format"})
		return
	}

	var req dto.UpdateUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind unit update request")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unit, err := h.unitUseCase.UpdateUnit(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to update unit")
		if err == entities.ErrUnitNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
			return
		}
		if err == entities.ErrInvalidUnitName {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update unit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": unit})
}

// DeleteUnit handles DELETE /units/:id
func (h *UnitHandler) DeleteUnit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID format"})
		return
	}

	err = h.unitUseCase.DeleteUnit(c.Request.Context(), id)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to delete unit")
		if err == entities.ErrUnitNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Unit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete unit"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUnitsByClinic handles GET /clinics/:clinicId/units
func (h *UnitHandler) GetUnitsByClinic(c *gin.Context) {
	clinicIDStr := c.Param("clinicId")
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clinic ID format"})
		return
	}

	units, err := h.unitUseCase.GetUnitsByClinicID(c.Request.Context(), clinicID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get units by clinic")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get units"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": units})
}
