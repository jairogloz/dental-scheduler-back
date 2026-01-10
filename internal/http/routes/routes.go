package routes

import (
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/http/handlers"
	"dental-scheduler-backend/internal/http/middleware"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the application
func SetupRoutes(
	router *gin.Engine,
	healthHandler *handlers.HealthHandler,
	clinicHandler *handlers.ClinicHandler,
	unitHandler *handlers.UnitHandler,
	doctorHandler *handlers.DoctorHandler,
	patientHandler *handlers.PatientHandler,
	appointmentHandler *handlers.AppointmentHandler,
	organizationHandler *handlers.OrganizationHandler,
	doctorAvailabilityHandler *handlers.DoctorAvailabilityHandler,
	userRepo repositories.UserRepository,
	logger *logger.Logger,
) {
	// Health check routes (public)
	router.GET("/health", healthHandler.Check)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.SupabaseAuth(logger, userRepo))
		{
			// Clinic routes
			clinics := protected.Group("/clinics")
			{
				clinics.POST("", clinicHandler.CreateClinic)
				clinics.GET("", clinicHandler.GetClinics)
				clinics.GET("/:id", clinicHandler.GetClinic)
				clinics.PUT("/:id", clinicHandler.UpdateClinic)
				clinics.DELETE("/:id", clinicHandler.DeleteClinic)
			}

			// Unit routes
			units := protected.Group("/units")
			{
				units.POST("", unitHandler.CreateUnit)
				units.GET("", unitHandler.GetUnits) // Supports ?clinic_id=uuid query param
				units.GET("/:id", unitHandler.GetUnit)
				units.PUT("/:id", unitHandler.UpdateUnit)
				units.DELETE("/:id", unitHandler.DeleteUnit)
			}

			// Doctor routes
			doctors := protected.Group("/doctors")
			{
				doctors.POST("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				doctors.GET("", doctorHandler.GetDoctorsByOrganization) // Implemented: GET /doctors?orgId=...&clinicId=...
				doctors.GET("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				doctors.GET("/:id/availability", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				doctors.PUT("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				doctors.DELETE("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			}

			// Patient routes
			patients := protected.Group("/patients")
			{
				patients.GET("/search", patientHandler.SearchPatients) // Patient search for autocomplete
				patients.POST("", patientHandler.CreatePatient)        // Create patient and link to organization from auth context
				patients.PATCH("/:id", patientHandler.UpdatePatient)   // Update patient
				patients.GET("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				patients.GET("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				patients.PUT("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				patients.DELETE("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			}

			// Appointment routes
			appointments := protected.Group("/appointments")
			{
				appointments.GET("/rescheduling-queue", appointmentHandler.GetReschedulingQueue)         // Get rescheduling queue
				appointments.POST("", appointmentHandler.CreateAppointment)                              // This needs to be implemented for conflict detection
				appointments.GET("", appointmentHandler.GetAppointments)                                 // Get appointments by organization with filters
				appointments.PATCH("/:appointment_id", appointmentHandler.UpdateAppointment)             // Update appointment
				appointments.POST("/:appointment_id/cancel", appointmentHandler.CancelFromQueue)         // Cancel from queue
				appointments.POST("/:appointment_id/reschedule", appointmentHandler.RescheduleFromQueue) // Reschedule from queue
				appointments.POST("/:appointment_id/snooze", appointmentHandler.SnoozeFromQueue)         // Snooze from queue
				appointments.GET("/upcoming", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				appointments.GET("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				appointments.PUT("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				appointments.DELETE("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			}

			// Doctor availability routes
			availability := protected.Group("/doctor-availability")
			{
				availability.POST("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				availability.GET("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				availability.GET("/:doctor_id", doctorAvailabilityHandler.GetDoctorAvailability) // Get availability for specific doctor
				availability.PUT("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
				availability.DELETE("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			}

			// Organization data route for calendar loading
			protected.GET("/organization", organizationHandler.GetOrganizationData)
		}

		// Optional authentication routes (user info is available if authenticated)
		optionalAuth := v1.Group("/")
		optionalAuth.Use(middleware.OptionalAuth(logger))
		{
			// Add routes here that benefit from user context but don't require authentication
		}
	}
}
