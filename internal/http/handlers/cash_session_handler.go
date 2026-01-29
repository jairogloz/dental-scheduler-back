package handlers

import (
	"net/http"
	"time"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/app/usecases"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CashSessionHandler handles cash session HTTP requests
type CashSessionHandler struct {
	cashSessionUseCase *usecases.CashSessionUseCase
	logger             *logger.Logger
}

// NewCashSessionHandler creates a new handler instance
func NewCashSessionHandler(
	cashSessionUseCase *usecases.CashSessionUseCase,
	logger *logger.Logger,
) *CashSessionHandler {
	return &CashSessionHandler{
		cashSessionUseCase: cashSessionUseCase,
		logger:             logger,
	}
}

// OpenSession opens a new cash session
// POST /cash-sessions/open
func (h *CashSessionHandler) OpenSession(c *gin.Context) {
	var req dto.OpenCashSessionRequest
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

	organizationID, _ := c.Get("organization_id")
	userID, _ := c.Get("user_id")

	input := usecases.OpenSessionInput{
		OrganizationID:     organizationID.(uuid.UUID),
		ClinicID:           req.ClinicID,
		UserID:             userID.(uuid.UUID),
		OpeningType:        req.OpeningType,
		StartingFloatCents: req.StartingFloatCents,
		Notes:              req.Notes,
	}

	session, err := h.cashSessionUseCase.OpenSession(c.Request.Context(), input)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to open cash session")
		
		// Handle specific business errors
		if err == entities.ErrCashSessionAlreadyOpen {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "SESSION_ALREADY_OPEN",
					"message": "You already have an open cash session for this clinic",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "OPEN_SESSION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toSessionResponse(session)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetCurrentSession retrieves the current open session for a user at a clinic
// GET /cash-sessions/current?clinic_id=xxx
func (h *CashSessionHandler) GetCurrentSession(c *gin.Context) {
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

	userID, _ := c.Get("user_id")

	session, err := h.cashSessionUseCase.GetCurrentSession(c.Request.Context(), userID.(uuid.UUID), clinicID)
	if err != nil {
		if err == entities.ErrNoCashSessionOpen {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "NO_OPEN_SESSION",
					"message": "No open cash session found",
				},
			})
			return
		}

		h.logger.Logger.WithError(err).Error("Failed to get current session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "GET_SESSION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toSessionResponse(session)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetSessionDetails retrieves full details of a cash session
// GET /cash-sessions/:id
func (h *CashSessionHandler) GetSessionDetails(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_SESSION_ID",
				"message": "Invalid session ID format",
			},
		})
		return
	}

	details, err := h.cashSessionUseCase.GetSessionDetails(c.Request.Context(), sessionID)
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

		h.logger.Logger.WithError(err).Error("Failed to get session details")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "GET_SESSION_DETAILS_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toSessionDetailsResponse(details)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CloseSession closes a cash session
// POST /cash-sessions/:id/close
func (h *CashSessionHandler) CloseSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_SESSION_ID",
				"message": "Invalid session ID format",
			},
		})
		return
	}

	if err := h.cashSessionUseCase.CloseSession(c.Request.Context(), sessionID); err != nil {
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
					"message": "Cash session is already closed",
				},
			})
			return
		}

		h.logger.Logger.WithError(err).Error("Failed to close session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CLOSE_SESSION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cash session closed successfully",
	})
}

// ListSessions lists cash sessions with filters
// GET /cash-sessions
func (h *CashSessionHandler) ListSessions(c *gin.Context) {
	var query dto.ListCashSessionsQuery
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

	organizationID, _ := c.Get("organization_id")
	orgID := organizationID.(uuid.UUID)

	// Set defaults
	if query.Page == 0 {
		query.Page = 1
	}
	if query.Limit == 0 {
		query.Limit = 20
	}

	// Build filters
	filters := repositories.CashSessionFilters{
		OrganizationID: &orgID,
		ClinicID:       query.ClinicID,
		UserID:         query.UserID,
		Status:         query.Status,
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

	sessions, err := h.cashSessionUseCase.ListSessions(c.Request.Context(), filters)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to list sessions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "LIST_SESSIONS_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	responses := make([]dto.CashSessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = h.toSessionResponse(session)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
	})
}

// Helper functions
func (h *CashSessionHandler) toSessionResponse(session *entities.CashSession) dto.CashSessionResponse {
	var closedAt *string
	if session.ClosedAt != nil {
		formatted := session.ClosedAt.Format("2006-01-02T15:04:05Z07:00")
		closedAt = &formatted
	}

	return dto.CashSessionResponse{
		ID:                 session.ID,
		OrganizationID:     session.OrganizationID,
		ClinicID:           session.ClinicID,
		UserID:             session.UserID,
		OpenedAt:           session.OpenedAt.Format("2006-01-02T15:04:05Z07:00"),
		ClosedAt:           closedAt,
		StartingFloatCents: session.StartingFloatCents,
		Status:             session.Status,
		OpeningType:        session.OpeningType,
		Notes:              session.Notes,
	}
}

func (h *CashSessionHandler) toSessionDetailsResponse(details *usecases.CashSessionDetailsOutput) dto.CashSessionDetailsResponse {
	entries := make([]dto.AppointmentAccountEntryResponse, len(details.Entries))
	for i, entry := range details.Entries {
		entries[i] = dto.AppointmentAccountEntryResponse{
			ID:                     entry.ID,
			AppointmentAccountID:   entry.AppointmentAccountID,
			Type:                   entry.Type,
			Currency:               entry.Currency,
			AmountCents:            entry.AmountCents,
			Description:            entry.Description,
			CreatedByUserID:        entry.CreatedByUserID,
			CreatedAt:              entry.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			PaymentMethod:          entry.PaymentMethod,
			ExchangeRateUsed:       entry.ExchangeRateUsed,
			DoctorID:               entry.DoctorID,
			DoctorType:             entry.DoctorType,
			CommissionPct:          entry.CommissionPct,
			ExternalDoctorFeeCents: entry.ExternalDoctorFeeCents,
			ServiceID:              entry.ServiceID,
			CashSessionID:          entry.CashSessionID,
			CorrectsEntryID:        entry.CorrectsEntryID,
			Notes:                  entry.Notes,
			Quantity:               entry.Quantity,
		}
	}

	return dto.CashSessionDetailsResponse{
		Session:         h.toSessionResponse(details.Session),
		Entries:         entries,
		ExpectedAmounts: details.ExpectedAmounts,
		PaymentSummary:  details.PaymentSummary,
	}
}
