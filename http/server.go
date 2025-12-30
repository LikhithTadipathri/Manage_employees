package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"employee-service/config"
	"employee-service/errors"
	"employee-service/http/handlers"
	"employee-service/http/middlewares"
	"employee-service/repositories/postgres"
	emailService "employee-service/services/email"
	employeeService "employee-service/services/employee"
	leaveService "employee-service/services/leave"
	userService "employee-service/services/user"
	"employee-service/utils/jwt"
)

// Server holds all the dependencies for the HTTP server
type Server struct {
	router     *chi.Mux
	config     *config.Config
	db         *sql.DB
	PostgresDB *sql.DB
	SQLiteDB   *sql.DB
	httpServer *http.Server
	emailQueue *emailService.EmailQueue
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, db *sql.DB) *Server {
	return &Server{
		router: chi.NewRouter(),
		config: cfg,
		db:     db,
	}
}

// Setup sets up all routes and middleware
func (s *Server) Setup() {
	// Add global middleware in order
	// 1. Panic recovery
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal server error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	})

	// 2. Correlation ID
	s.router.Use(middlewares.CorrelationIDMiddleware())

	// 3. CORS
	s.router.Use(middlewares.CORSMiddleware())

	// 4. Security Headers
	s.router.Use(middlewares.SecurityHeadersMiddleware())

	// 5. Rate Limiting
	s.router.Use(middlewares.RateLimitMiddleware())

	// 6. Request Logger
	s.router.Use(middlewares.LoggerMiddleware)

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(&s.config.JWT)

	// Initialize repositories
	employeeRepo := postgres.NewEmployeeRepository(s.db)
	userRepo := postgres.NewUserRepository(s.db)
	leaveRepo := postgres.NewLeaveRepository(s.db)
	notificationRepo := postgres.NewNotificationRepository(s.db)

	// Initialize SMTP email service with environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPortStr := os.Getenv("SMTP_PORT")
	if smtpPortStr == "" {
		smtpPortStr = "587"
	}
	smtpPort, _ := strconv.Atoi(smtpPortStr)
	smtpUsername := os.Getenv("SMTP_USERNAME")
	if smtpUsername == "" {
		smtpUsername = "no-reply@company.com"
	}
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFromAddr := os.Getenv("SMTP_FROM_ADDR")
	if smtpFromAddr == "" {
		smtpFromAddr = "no-reply@company.com"
	}
	smtpFromName := os.Getenv("SMTP_FROM_NAME")
	if smtpFromName == "" {
		smtpFromName = "HR Management System"
	}

	// Create SMTP email service
	emailSvc := emailService.NewEmailService(
		smtpHost,
		smtpPort,
		smtpUsername,
		smtpPassword,
		smtpFromAddr,
		smtpFromName,
	)

	// Create and start email queue
	s.emailQueue = emailService.NewEmailQueue(notificationRepo, emailSvc)
	numWorkers := 3
	if workersStr := os.Getenv("EMAIL_QUEUE_WORKERS"); workersStr != "" {
		if w, err := strconv.Atoi(workersStr); err == nil {
			numWorkers = w
		}
	}
	err := s.emailQueue.Start(numWorkers)
	if err != nil {
		errors.LogError("Failed to start email queue", err)
	} else {
		errors.LogInfo(fmt.Sprintf("✅ Email queue started with %d workers", numWorkers))
	}

	// Initialize services
	leaveServiceInstance := leaveService.NewService(leaveRepo, employeeRepo, userRepo, notificationRepo, s.emailQueue)
	userServiceInstance := userService.NewUserService(userRepo)

	// Initialize employee service with user service for creating login credentials
	employeeServiceInstance := employeeService.NewServiceWithUser(employeeRepo, userRepo, userServiceInstance)

	// Initialize handlers
	employeeHandler := handlers.NewEmployeeHandlerWithLeave(employeeServiceInstance, leaveServiceInstance)
	leaveHandler := handlers.NewLeaveHandler(leaveServiceInstance)
	authHandler := handlers.NewAuthHandlerWithServices(userServiceInstance, jwtManager, employeeRepo, leaveServiceInstance)
	dashboardHandler := handlers.NewDashboardHandler(employeeServiceInstance, userServiceInstance)

	// Health check endpoints (no auth required)
	s.router.Get("/health", s.healthCheck)
	s.router.Get("/readiness", s.healthCheck)

	// Auth routes (no auth required)
	s.router.Post("/auth/login", authHandler.Login)
	
	// Auth routes (auth required)
	s.router.Route("/auth", func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware(jwtManager))
		r.Post("/register", authHandler.Register)  // Admin only - register new users
		r.Post("/logout", authHandler.Logout)
		r.Get("/me", authHandler.GetMe)
	})

	// Dashboard routes with JWT auth
	s.router.Route("/employee", func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware(jwtManager))
		r.Get("/records", dashboardHandler.GetUserRecords)
		r.Get("/overview", dashboardHandler.GetUserOverview)
	})

	s.router.Route("/admin", func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware(jwtManager))
		r.Get("/records", dashboardHandler.GetAdminRecords)
		r.Get("/overview", dashboardHandler.GetAdminOverview)
		r.Get("/logs", dashboardHandler.GetAdminLogs)
	})

	// Leave routes with JWT auth
	s.router.Route("/leave", func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware(jwtManager))
		// Employee routes
		r.Post("/apply", leaveHandler.ApplyLeave)
		r.Get("/my-requests", leaveHandler.GetMyLeaveRequests)
		r.Delete("/cancel/{id}", leaveHandler.CancelLeave)
		r.Get("/balance", leaveHandler.GetMyLeaveBalance)
		r.Get("/balance/{type}", leaveHandler.GetMyLeaveBalanceByType)
		
		// Admin routes
		r.Get("/review", leaveHandler.ReviewLeaveRequests)
		r.Get("/all", leaveHandler.GetAllLeaveRequests)
		r.Post("/approve/{id}", leaveHandler.ApproveLeave)
		r.Post("/reject/{id}", leaveHandler.RejectLeave)
	})

// API routes with JWT auth
	s.router.Route("/api/v1/employees", func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware(jwtManager))
		

		// Create employee
		r.Post("/", employeeHandler.CreateEmployee)

		// List employees
		r.Get("/", employeeHandler.ListEmployees)

		// Search employees
		r.Get("/search", employeeHandler.SearchEmployees)

		// Get specific employee
		r.Get("/{id}", employeeHandler.GetEmployee)

		// Update employee
		r.Put("/{id}", employeeHandler.UpdateEmployee)

		// Delete employee
		r.Delete("/{id}", employeeHandler.DeleteEmployee)
	})
}

// healthCheck handles health check endpoint
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if this is a readiness probe
	isReadinessProbe := r.URL.Path == "/readiness"

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"checks": map[string]interface{}{
			"database": "ok",
			"email_queue": "ok",
		},
	}

	checks := health["checks"].(map[string]interface{})

	// Check database
	if s.db != nil {
		if err := s.db.Ping(); err != nil {
			checks["database"] = "error"
			health["status"] = "unhealthy"
		} else {
			// Get database stats
			stats := s.db.Stats()
			checks["database"] = map[string]interface{}{
				"status": "ok",
				"open_connections": stats.OpenConnections,
				"in_use": stats.InUse,
				"idle": stats.Idle,
				"max_open_conns": stats.MaxOpenConnections,
			}
		}
	} else {
		checks["database"] = "unavailable"
		health["status"] = "unhealthy"
	}

	// Check email queue
	if s.emailQueue != nil {
		if !s.emailQueue.IsRunning() {
			checks["email_queue"] = "stopped"
			if isReadinessProbe {
				health["status"] = "unhealthy"
			}
		}
	}

	// Decide status code based on components
	statusCode := http.StatusOK
	if health["status"] == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}
// GetRouter returns the chi router
func (s *Server) GetRouter() *chi.Mux {
	return s.router
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Set up TLS if enabled
	if s.config.TLS.IsEnabled() {
		tlsConfig, err := s.config.TLS.GetTLSConfig()
		if err != nil {
			return err
		}
		s.httpServer.TLSConfig = tlsConfig
		errors.LogInfo("Starting HTTPS server on port " + addr)
		return s.httpServer.ListenAndServeTLS(s.config.TLS.CertFile, s.config.TLS.KeyFile)
	}

	errors.LogInfo("Starting HTTP server on port " + addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server and email queue
func (s *Server) Shutdown(ctx context.Context) error {
	errors.LogInfo("Shutting down server...")
	
	// Stop email queue first
	if s.emailQueue != nil {
		err := s.emailQueue.Stop()
		if err != nil {
			errors.LogError("Error shutting down email queue", err)
		} else {
			errors.LogInfo("✅ Email queue shut down gracefully")
		}
	}
	
	return s.httpServer.Shutdown(ctx)
}
