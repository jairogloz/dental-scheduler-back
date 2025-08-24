package handlers

import (
	"net/http"

	"dental-scheduler-backend/internal/app/dto"
	"dental-scheduler-backend/internal/app/usecases"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
)

// DoctorHandler handles doctor-related HTTP requests
type DoctorHandler struct {
	doctorUseCase *usecases.DoctorUseCase
	logger        *logger.Logger
}

// NewDoctorHandler creates a new doctor handler
func NewDoctorHandler(doctorUseCase *usecases.DoctorUseCase, logger *logger.Logger) *DoctorHandler {
	return &DoctorHandler{
		doctorUseCase: doctorUseCase,
		logger:        logger,
	}
}

// CreateDoctor handles POST /doctors (placeholder)
func (h *DoctorHandler) CreateDoctor(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

// GetDoctorsByOrganization handles GET /doctors?orgId=...&clinicId=...
func (h *DoctorHandler) GetDoctorsByOrganization(c *gin.Context) {
	var req dto.GetDoctorsByOrgRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Logger.WithError(err).Error("Failed to bind query parameters")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	// Parse the organization ID
	orgID, err := req.ParsedOrgID()
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid organization ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID format"})
		return
	}

	// Parse the clinic ID if provided
	clinicID, err := req.ParsedClinicID()
	if err != nil {
		h.logger.Logger.WithError(err).Error("Invalid clinic ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clinic ID format"})
		return
	}

	doctors, err := h.doctorUseCase.GetDoctorsByOrganizationID(c.Request.Context(), orgID, clinicID)
	if err != nil {
		h.logger.Logger.WithError(err).Error("Failed to get doctors by organization")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get doctors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": doctors})
}

// PatientHandler handles patient-related HTTP requests
type PatientHandler struct {
	patientUseCase *usecases.PatientUseCase
	logger         *logger.Logger
}

// NewPatientHandler creates a new patient handler
func NewPatientHandler(patientUseCase *usecases.PatientUseCase, logger *logger.Logger) *PatientHandler {
	return &PatientHandler{
		patientUseCase: patientUseCase,
		logger:         logger,
	}
}

// CreatePatient handles POST /patients (placeholder)
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}

// AppointmentHandler handles appointment-related HTTP requests
type AppointmentHandler struct {
	appointmentUseCase *usecases.AppointmentUseCase
	logger             *logger.Logger
}

// NewAppointmentHandler creates a new appointment handler
func NewAppointmentHandler(appointmentUseCase *usecases.AppointmentUseCase, logger *logger.Logger) *AppointmentHandler {
	return &AppointmentHandler{
		appointmentUseCase: appointmentUseCase,
		logger:             logger,
	}
}

// CreateAppointment handles POST /appointments (placeholder)
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	c.JSON(501, gin.H{"error": "Not implemented yet"})
}
