package handlers

import (
	"net/http"
	"time"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/app/usecases"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/http/middleware"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReconciliationHandler handles reconciliation HTTP requests
type ReconciliationHandler struct {
	reconciliationUseCase *usecases.ReconciliationUseCase
	logger                *logger.Logger
}

// NewReconciliationHandler creates a new handler instance
func NewReconciliationHandler(
	reconciliationUseCase *usecases.ReconciliationUseCase,
	logger *logger.Logger,
) *ReconciliationHandler {
	return &ReconciliationHandler{
		reconciliationUseCase: reconciliationUseCase,
		logger:                logger,
	}
}

// GetReconciliationPreview calculates expected amounts for a cash session
// GET /cash-sessions/:id/reconciliation-preview
func (h *ReconciliationHandler) GetReconciliationPreview(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_SESSION_ID",
				"message": "Session ID is required",
			},
		})
		return
	}

	preview, err := h.reconciliationUseCase.GetReconciliationPreview(c.Request.Context(), sessionID)
	if err != nil {
		if err == entities.ErrCashSessionNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SESSION_NOT_FOUND",
					"message": "Cash session not found",
				},
			})
			return
		}

		h.logger.Logger.WithError(err).Error("Failed to get reconciliation preview")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "GET_PREVIEW_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toReconciliationPreviewResponse(preview)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreateReconciliation creates a reconciliation record
// POST /cash-sessions/:id/reconcile
func (h *ReconciliationHandler) CreateReconciliation(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_SESSION_ID",
				"message": "Session ID is required",
			},
		})
		return
	}

	var req dto.CreateReconciliationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": err.Error(),
			},
		})
		return
	}

	// Get user profile from context
	userProfile, exists := middleware.GetUserProfileFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "USER_PROFILE_REQUIRED",
				"message": "User profile not found in context",
			},
		})
		return
	}

	// Get organization ID
	organizationID, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ORGANIZATION_ID_REQUIRED",
				"message": "Organization ID not found in context",
			},
		})
		return
	}

	// Get clinic ID - use organization ID as fallback
	// In production, this should come from request or authenticated context
	clinicID := organizationID

	userID := userProfile.Profile.ID.String()

	// Validate amounts before creating
	if err := h.reconciliationUseCase.ValidateReconciliationAmounts(
		req.ExpectedCents,
		req.ActualCents,
		req.FloatLeftCents,
		req.DepositedCents,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_AMOUNTS",
				"message": err.Error(),
			},
		})
		return
	}

	input := usecases.CreateReconciliationInput{
		CashSessionID:      sessionID,
		OrganizationID:     organizationID,
		ClinicID:           clinicID,
		PaymentMethod:      req.PaymentMethod,
		Currency:           req.Currency,
		ExpectedCents:      req.ExpectedCents,
		ActualCents:        req.ActualCents,
		FloatLeftCents:     req.FloatLeftCents,
		DepositedCents:     req.DepositedCents,
		ReconciledByUserID: userID,
		Notes:              req.Notes,
	}

	reconciliation, err := h.reconciliationUseCase.CreateReconciliation(c.Request.Context(), input)
	if err != nil {
		if err == entities.ErrCashSessionNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SESSION_NOT_FOUND",
					"message": "Cash session not found",
				},
			})
			return
		}

		if err == entities.ErrCashSessionAlreadyClosed {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SESSION_ALREADY_CLOSED",
					"message": "Cannot reconcile a closed cash session",
				},
			})
			return
		}

		if err == entities.ErrReconciliationAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RECONCILIATION_EXISTS",
					"message": "Reconciliation already exists for this payment method and currency",
				},
			})
			return
		}

		h.logger.Logger.WithError(err).Error("Failed to create reconciliation")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CREATE_RECONCILIATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toReconciliationResponse(reconciliation)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetReconciliation retrieves a reconciliation by ID
// GET /reconciliations/:id
func (h *ReconciliationHandler) GetReconciliation(c *gin.Context) {
	reconciliationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_RECONCILIATION_ID",
				"message": "Invalid reconciliation ID format",
			},
		})
		return
	}

	reconciliation, err := h.reconciliationUseCase.GetReconciliation(c.Request.Context(), reconciliationID)
	if err != nil {
		if err == entities.ErrReconciliationNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RECONCILIATION_NOT_FOUND",
					"message": "Reconciliation not found",
				},
			})
			return
		}

		h.logger.Logger.WithError(err).Error("Failed to get reconciliation")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "GET_RECONCILIATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toReconciliationResponse(reconciliation)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ListReconciliations lists reconciliations with filters
// GET /reconciliations
func (h *ReconciliationHandler) ListReconciliations(c *gin.Context) {
	var query dto.ListReconciliationsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_QUERY",
				"message": err.Error(),
			},
		})
		return
	}

	orgIDStr, exists := middleware.GetOrganizationIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ORGANIZATION_ID_REQUIRED",
				"message": "Organization ID not found in context",
			},
		})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid organization ID format")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ORGANIZATION_ID",
				"message": "Invalid organization ID format",
			},
		})
		return
	}

	// Set defaults
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	// Build filters
	filters := repositories.ReconciliationFilters{
		OrganizationID: &orgID,
		ClinicID:       query.ClinicID,
		CashSessionID:  query.CashSessionID,
		PaymentMethod:  query.PaymentMethod,
		Currency:       query.Currency,
		Status:         query.Status,
		HasDiscrepancy: query.HasDiscrepancy,
		Page:           query.Page,
		Limit:          query.Limit,
	}

	// Parse dates if provided
	if query.StartDate != nil {
		startDate, err := time.Parse("2006-01-02", *query.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_START_DATE",
					"message": "start_date must be in YYYY-MM-DD format",
				},
			})
			return
		}
		filters.StartDate = &startDate
	}

	if query.EndDate != nil {
		endDate, err := time.Parse("2006-01-02", *query.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_END_DATE",
					"message": "end_date must be in YYYY-MM-DD format",
				},
			})
			return
		}
		filters.EndDate = &endDate
	}

	// If cash session ID is provided, use a different method
	if query.CashSessionID != nil {
		reconciliations, err := h.reconciliationUseCase.GetReconciliationsByCashSession(c.Request.Context(), *query.CashSessionID)
		if err != nil {
			h.logger.Logger.WithError(err).Error("Failed to list reconciliations")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "LIST_RECONCILIATIONS_FAILED",
					"message": err.Error(),
				},
			})
			return
		}

		responses := make([]dto.ReconciliationResponse, len(reconciliations))
		for i, rec := range reconciliations {
			responses[i] = h.toReconciliationResponse(rec)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    responses,
		})
		return
	}

	// TODO: Implement general list with filters when repository method is available
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "NOT_IMPLEMENTED",
			"message": "General reconciliation listing not yet implemented",
		},
	})
}

// GetDiscrepancies retrieves reconciliations with discrepancies
// GET /reconciliations/discrepancies?clinic_id=xxx&start_date=xxx&end_date=xxx
func (h *ReconciliationHandler) GetDiscrepancies(c *gin.Context) {
	clinicIDStr := c.Query("clinic_id")
	if clinicIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CLINIC_ID_REQUIRED",
				"message": "clinic_id query parameter is required",
			},
		})
		return
	}

	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_CLINIC_ID",
				"message": "Invalid clinic ID format",
			},
		})
		return
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DATE_RANGE_REQUIRED",
				"message": "start_date and end_date query parameters are required",
			},
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_START_DATE",
				"message": "start_date must be in YYYY-MM-DD format",
			},
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_END_DATE",
				"message": "end_date must be in YYYY-MM-DD format",
			},
		})
		return
	}

	reconciliations, err := h.reconciliationUseCase.GetDiscrepancies(c.Request.Context(), clinicID, startDate, endDate)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get discrepancies")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "GET_DISCREPANCIES_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	responses := make([]dto.ReconciliationResponse, len(reconciliations))
	for i, rec := range reconciliations {
		responses[i] = h.toReconciliationResponse(rec)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
	})
}

// Helper functions
func (h *ReconciliationHandler) toReconciliationResponse(rec *entities.Reconciliation) dto.ReconciliationResponse {
	return dto.ReconciliationResponse{
		ID:                  rec.ID,
		CashSessionID:       rec.CashSessionID,
		OrganizationID:      rec.OrganizationID,
		ClinicID:            rec.ClinicID,
		PaymentMethod:       rec.PaymentMethod,
		Currency:            rec.Currency,
		ReconciledAt:        rec.ReconciledAt.Format("2006-01-02T15:04:05Z07:00"),
		ReconciledByUserID:  rec.ReconciledByUserID,
		ExpectedAmountCents: rec.ExpectedAmountCents,
		ActualAmountCents:   rec.ActualAmountCents,
		FloatLeftCents:      rec.FloatLeftCents,
		DepositedCents:      rec.DepositedCents,
		DiscrepancyCents:    rec.DiscrepancyCents,
		Status:              rec.Status,
		Notes:               rec.Notes,
	}
}

func (h *ReconciliationHandler) toReconciliationPreviewResponse(preview *usecases.ReconciliationPreviewOutput) dto.ReconciliationPreviewResponse {
	existingRecs := make([]dto.ReconciliationResponse, len(preview.ExistingReconciliations))
	for i, rec := range preview.ExistingReconciliations {
		existingRecs[i] = h.toReconciliationResponse(rec)
	}

	var closedAt *string
	if preview.Session.ClosedAt != nil {
		formatted := preview.Session.ClosedAt.Format("2006-01-02T15:04:05Z07:00")
		closedAt = &formatted
	}

	return dto.ReconciliationPreviewResponse{
		Session: dto.CashSessionResponse{
			ID:                 preview.Session.ID,
			OrganizationID:     preview.Session.OrganizationID,
			ClinicID:           preview.Session.ClinicID,
			UserID:             preview.Session.UserID,
			OpenedAt:           preview.Session.OpenedAt.Format("2006-01-02T15:04:05Z07:00"),
			ClosedAt:           closedAt,
			StartingFloatCents: preview.Session.StartingFloatCents,
			Status:             preview.Session.Status,
			OpeningType:        preview.Session.OpeningType,
			Notes:              preview.Session.Notes,
		},
		ExpectedAmountsByCurrency: preview.ExpectedAmountsByCurrency,
		ExistingReconciliations:   existingRecs,
	}
}
