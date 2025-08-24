package handlers

import (
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
