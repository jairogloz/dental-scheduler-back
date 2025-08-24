package routes

import (
	"dental-scheduler-backend/internal/http/handlers"

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
) {
	// Health check routes
	router.GET("/health", healthHandler.Check)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Clinic routes
		clinics := v1.Group("/clinics")
		{
			clinics.POST("", clinicHandler.CreateClinic)
			clinics.GET("", clinicHandler.GetClinics)
			clinics.GET("/:id", clinicHandler.GetClinic)
			clinics.PUT("/:id", clinicHandler.UpdateClinic)
			clinics.DELETE("/:id", clinicHandler.DeleteClinic)
		}

		// Unit routes
		units := v1.Group("/units")
		{
			units.POST("", unitHandler.CreateUnit)
			units.GET("", unitHandler.GetUnits) // Supports ?clinic_id=uuid query param
			units.GET("/:id", unitHandler.GetUnit)
			units.PUT("/:id", unitHandler.UpdateUnit)
			units.DELETE("/:id", unitHandler.DeleteUnit)
		}

		// Doctor routes (placeholder - handlers to be implemented)
		doctors := v1.Group("/doctors")
		{
			doctors.POST("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			doctors.GET("", doctorHandler.GetDoctorsByOrganization) // Implemented: GET /doctors?orgId=...&clinicId=...
			doctors.GET("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			doctors.GET("/:id/availability", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			doctors.PUT("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			doctors.DELETE("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
		}

		// Patient routes (placeholder - handlers to be implemented)
		patients := v1.Group("/patients")
		{
			patients.POST("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			patients.GET("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			patients.GET("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			patients.PUT("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			patients.DELETE("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
		}

		// Appointment routes (placeholder - handlers to be implemented)
		appointments := v1.Group("/appointments")
		{
			appointments.POST("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			appointments.GET("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			appointments.GET("/upcoming", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			appointments.GET("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			appointments.PUT("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			appointments.DELETE("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
		}

		// Doctor availability routes (placeholder - handlers to be implemented)
		availability := v1.Group("/doctor-availability")
		{
			availability.POST("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			availability.GET("", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			availability.GET("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			availability.PUT("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
			availability.DELETE("/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "Not implemented"}) })
		}
	}
}
