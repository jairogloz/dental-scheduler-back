package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dental-scheduler-backend/internal/app/usecases"
	"dental-scheduler-backend/internal/domain/services"
	"dental-scheduler-backend/internal/http/handlers"
	"dental-scheduler-backend/internal/http/middleware"
	"dental-scheduler-backend/internal/http/routes"
	"dental-scheduler-backend/internal/infra/config"
	"dental-scheduler-backend/internal/infra/database/postgres"
	postgresRepos "dental-scheduler-backend/internal/infra/database/postgres/repositories"
	"dental-scheduler-backend/internal/infra/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.NewLogger(cfg.Log.Level)
	appLogger.Logger.Info("Starting Dental Scheduler Backend API")

	// Initialize database connection
	dbConn, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		appLogger.Logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer dbConn.Close()

	appLogger.Logger.Info("Database connection established")

	// Initialize repositories
	clinicRepo := postgresRepos.NewClinicPostgresRepository(dbConn.GetDB())
	unitRepo := postgresRepos.NewUnitPostgresRepository(dbConn.GetDB())
	doctorRepo := postgresRepos.NewDoctorPostgresRepository(dbConn.GetDB())
	patientRepo := postgresRepos.NewPatientPostgresRepository(dbConn.GetDB())
	appointmentRepo := postgresRepos.NewAppointmentPostgresRepository(dbConn.GetDB())
	availabilityRepo := postgresRepos.NewDoctorAvailabilityPostgresRepository(dbConn.GetDB())
	userRepo := postgresRepos.NewUserPostgresRepository(dbConn.GetDB())
	organizationRepo := postgresRepos.NewOrganizationPostgresRepository(dbConn.GetDB())

	// Initialize domain services
	conflictChecker := services.NewAppointmentConflictChecker(appointmentRepo, availabilityRepo)
	schedulingService := services.NewSchedulingService(
		appointmentRepo,
		availabilityRepo,
		doctorRepo,
		unitRepo,
		conflictChecker,
	)

	// Initialize use cases
	clinicUseCase := usecases.NewClinicUseCase(clinicRepo)
	unitUseCase := usecases.NewUnitUseCase(unitRepo, clinicRepo)
	doctorUseCase := usecases.NewDoctorUseCase(doctorRepo, unitRepo, appointmentRepo)
	patientUseCase := usecases.NewPatientUseCase(patientRepo)
	// userUseCase := usecases.NewUserUseCase(userRepo, appLogger) // Available when needed
	appointmentUseCase := usecases.NewAppointmentUseCase(
		appointmentRepo,
		patientRepo,
		doctorRepo,
		unitRepo,
		schedulingService,
	)
	getOrgDataUseCase := usecases.NewGetOrganizationDataUseCase(organizationRepo)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	clinicHandler := handlers.NewClinicHandler(clinicUseCase, appLogger)
	unitHandler := handlers.NewUnitHandler(unitUseCase, appLogger)
	doctorHandler := handlers.NewDoctorHandler(doctorUseCase, appLogger)
	patientHandler := handlers.NewPatientHandler(patientUseCase, appLogger)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentUseCase, appLogger)
	organizationHandler := handlers.NewOrganizationHandler(getOrgDataUseCase, appLogger)

	// Set Gin mode
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.RequestLogger(appLogger))
	router.Use(middleware.Recovery(appLogger))
	router.Use(middleware.CORS(cfg.CORS.AllowedOrigins))

	// Setup routes
	routes.SetupRoutes(
		router,
		healthHandler,
		clinicHandler,
		unitHandler,
		doctorHandler,
		patientHandler,
		appointmentHandler,
		organizationHandler,
		userRepo,
		appLogger,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.GetAddress(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Logger.WithField("address", cfg.Server.GetAddress()).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Logger.Info("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		appLogger.Logger.WithError(err).Error("Server forced to shutdown")
	}

	appLogger.Logger.Info("Server exited")
}
