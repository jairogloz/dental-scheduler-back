package handlers

import (
	"net/http"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/app/usecases"
	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/http/middleware"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AppointmentAccountHandler handles appointment accounting HTTP requests
type AppointmentAccountHandler struct {
	createEntryUseCase *usecases.CreateAppointmentEntryUseCase
	getAccountUseCase  *usecases.GetAppointmentAccountUseCase
	logger             *logger.Logger
}

// NewAppointmentAccountHandler creates a new handler instance
func NewAppointmentAccountHandler(
	createEntryUseCase *usecases.CreateAppointmentEntryUseCase,
	getAccountUseCase *usecases.GetAppointmentAccountUseCase,
	logger *logger.Logger,
) *AppointmentAccountHandler {
	return &AppointmentAccountHandler{
		createEntryUseCase: createEntryUseCase,
		getAccountUseCase:  getAccountUseCase,
		logger:             logger,
	}
}

// GetAppointmentAccount retrieves an appointment account with all entries and balance
// GET /appointments/:appointment_id/account
func (h *AppointmentAccountHandler) GetAppointmentAccount(c *gin.Context) {
	appointmentID := c.Param("appointment_id")
	if appointmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_APPOINTMENT_ID",
				"message": "Appointment ID is required",
			},
		})
		return
	}

	// Get organization ID from context using middleware helper
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

	result, err := h.getAccountUseCase.Execute(c.Request.Context(), organizationID, appointmentID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get appointment account")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "GET_ACCOUNT_FAILED",
				"message": "Failed to retrieve appointment account",
			},
		})
		return
	}

	// Convert to response DTO
	response := h.toAppointmentAccountResponse(result)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetAppointmentBalance retrieves just the balance for an appointment
// GET /appointments/:appointment_id/account/balance
func (h *AppointmentAccountHandler) GetAppointmentBalance(c *gin.Context) {
	appointmentID := c.Param("appointment_id")
	if appointmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_APPOINTMENT_ID",
				"message": "Appointment ID is required",
			},
		})
		return
	}

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

	balance, err := h.getAccountUseCase.GetBalanceOnly(c.Request.Context(), organizationID, appointmentID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get appointment balance")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "GET_BALANCE_FAILED",
				"message": "Failed to retrieve appointment balance",
			},
		})
		return
	}

	response := h.toBalanceResponse(balance)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreateServiceCharge creates a service charge entry
// POST /appointments/:appointment_id/account/charges
func (h *AppointmentAccountHandler) CreateServiceCharge(c *gin.Context) {
	appointmentID := c.Param("appointment_id")
	if appointmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_APPOINTMENT_ID",
				"message": "Appointment ID is required",
			},
		})
		return
	}

	var req dto.CreateServiceChargeRequest
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

	// Get user profile from context using middleware helper
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

	// Get user ID from user profile
	userID := userProfile.Profile.ID.String()

	// Get clinic ID from user profile default clinic
	var clinicID *string = nil
	if defaultClinicID, exists := middleware.GetDefaultClinicIDFromContext(c); exists {
		clinicID = &defaultClinicID
	}

	input := usecases.CreateServiceChargeInput{
		OrganizationID:         organizationID,
		AppointmentID:          appointmentID,
		DoctorID:               req.DoctorID,
		DoctorType:             req.DoctorType,
		Currency:               req.Currency,
		AmountCents:            req.AmountCents,
		Description:            req.Description,
		CreatedByUserID:        userID,
		ServiceID:              req.ServiceID,
		CommissionPct:          req.CommissionPct,
		ExternalDoctorFeeCents: req.ExternalDoctorFeeCents,
		ClinicID:               clinicID,
		UserID:                 &userID,
	}

	entry, err := h.createEntryUseCase.CreateServiceCharge(c.Request.Context(), input)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to create service charge")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CREATE_CHARGE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toEntryResponse(entry)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreatePayment creates a payment entry
// POST /appointments/:appointment_id/account/payments
func (h *AppointmentAccountHandler) CreatePayment(c *gin.Context) {
	appointmentID := c.Param("appointment_id")
	if appointmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_APPOINTMENT_ID",
				"message": "Appointment ID is required",
			},
		})
		return
	}

	var req dto.CreatePaymentRequest
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

	// Get user ID from user profile
	userID := userProfile.Profile.ID.String()

	// Get clinic ID from user profile default clinic (validated by middleware)
	clinicID := userProfile.Profile.DefaultClinicID.String()

	input := usecases.CreatePaymentInput{
		OrganizationID:  organizationID,
		AppointmentID:   appointmentID,
		PaymentMethod:   req.PaymentMethod,
		Currency:        req.Currency,
		AmountCents:     req.AmountCents,
		Description:     req.Description,
		CreatedByUserID: userID,
		ExchangeRate:    req.ExchangeRate,
		ClinicID:        clinicID,
		UserID:          userID,
	}

	entry, err := h.createEntryUseCase.CreatePayment(c.Request.Context(), input)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to create payment")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CREATE_PAYMENT_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toEntryResponse(entry)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreateCorrection creates a correction entry
// POST /appointments/:id/account/entries/:entry_id/correct
func (h *AppointmentAccountHandler) CreateCorrection(c *gin.Context) {
	entryID, err := uuid.Parse(c.Param("entry_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ENTRY_ID",
				"message": "Invalid entry ID format",
			},
		})
		return
	}

	var req dto.CreateCorrectionRequest
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

	// Get user profile from context using middleware helper
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

	input := usecases.CreateCorrectionInput{
		OriginalEntryID: entryID,
		Description:     req.Description,
		CreatedByUserID: userProfile.Profile.ID,
		Notes:           req.Notes,
	}

	entry, err := h.createEntryUseCase.CreateCorrection(c.Request.Context(), input)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to create correction")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CREATE_CORRECTION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	response := h.toEntryResponse(entry)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// Helper functions to convert entities to DTOs
func (h *AppointmentAccountHandler) toAppointmentAccountResponse(result *usecases.GetAccountWithEntriesOutput) dto.AppointmentAccountResponse {
	entries := make([]dto.AppointmentAccountEntryResponse, len(result.Entries))
	for i, entry := range result.Entries {
		entries[i] = h.toEntryResponse(entry)
	}

	return dto.AppointmentAccountResponse{
		ID:             result.Account.ID,
		OrganizationID: result.Account.OrganizationID,
		AppointmentID:  result.Account.AppointmentID,
		Entries:        entries,
		Balance:        h.toBalanceResponse(result.Balance),
	}
}

func (h *AppointmentAccountHandler) toEntryResponse(entry *entities.AppointmentAccountEntry) dto.AppointmentAccountEntryResponse {
	return dto.AppointmentAccountEntryResponse{
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

func (h *AppointmentAccountHandler) toBalanceResponse(balance *repositories.AccountBalance) dto.AccountBalanceResponse {
	return dto.AccountBalanceResponse{
		TotalChargesCents:   balance.TotalChargesCents,
		TotalDiscountsCents: balance.TotalDiscountsCents,
		TotalPaymentsCents:  balance.TotalPaymentsCents,
		TotalRefundsCents:   balance.TotalRefundsCents,
		BalanceDueCents:     balance.BalanceDueCents,
		PaymentsByCurrency:  balance.PaymentsByCurrency,
		PaymentsByMethod:    balance.PaymentsByMethod,
	}
}
