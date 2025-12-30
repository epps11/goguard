package api

import (
	"github.com/gin-gonic/gin"

	"github.com/epps11/goguard/internal/config"
	"github.com/epps11/goguard/internal/database"
	"github.com/epps11/goguard/internal/services/audit"
	"github.com/epps11/goguard/internal/services/injection"
	"github.com/epps11/goguard/internal/services/llm"
	"github.com/epps11/goguard/internal/services/pii"
	"github.com/epps11/goguard/internal/services/policy"
	"github.com/epps11/goguard/internal/services/settings"
	"github.com/epps11/goguard/internal/services/spending"
)

// Router manages the API routes
type Router struct {
	engine         *gin.Engine
	handler        *Handler
	controlHandler *ControlHandler
	config         *config.Config
	policyEngine   *policy.Engine
	auditLogger    *audit.Logger
}

// NewRouter creates a new router with all routes configured
// repo is optional - if nil, settings will use defaults from config
func NewRouter(cfg *config.Config, llmClient *llm.Client, repo ...*database.Repository) *Router {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Create services
	detector := injection.NewDetector(
		cfg.Security.InjectionPatterns,
		cfg.Security.EnableInjectionDetection,
		cfg.Security.BlockOnDetection,
	)

	masker := pii.NewMasker(
		cfg.PII.PIITypes,
		cfg.PII.MaskCharacter,
		cfg.PII.PreserveDomain,
		cfg.PII.EnableMasking,
	)

	// Create control plane services
	policyEngine := policy.NewEngine()
	auditLogger := audit.NewLogger(10000)

	// Initialize settings service and spending tracker with database if provided
	var settingsSvc *settings.Service
	var spendingTracker *spending.Tracker
	if len(repo) > 0 && repo[0] != nil {
		settingsSvc = settings.NewService(repo[0])
		spendingTracker = spending.NewTracker(repo[0])
	}

	// Create LLM client factory for per-request provider support
	llmFactory, err := llm.NewClientFactory(cfg.LLM)
	var handler *Handler
	if err != nil || llmFactory == nil {
		// Fall back to legacy handler if factory creation fails
		handler = NewHandler(detector, masker, llmClient, auditLogger)
	} else {
		// Wire up settings service to factory for dynamic configuration from dashboard
		if settingsSvc != nil {
			llmFactory.SetSettingsProvider(settingsSvc)
		}
		handler = NewHandlerWithFactory(detector, masker, llmFactory, auditLogger, spendingTracker)
	}

	// Get repository for control handler (may be nil if no database)
	var dbRepo *database.Repository
	if len(repo) > 0 && repo[0] != nil {
		dbRepo = repo[0]
	}
	controlHandler := NewControlHandler(policyEngine, auditLogger, settingsSvc, dbRepo)

	// Create engine
	engine := gin.New()

	// Apply global middleware
	engine.Use(Recovery())
	engine.Use(RequestLogger())
	engine.Use(CORS())
	engine.Use(SecurityHeaders())
	engine.Use(MaxBodySize(10 * 1024 * 1024)) // 10MB max

	// Apply rate limiting if configured
	if cfg.Security.RateLimitPerMinute > 0 {
		rateLimiter := NewRateLimiter(cfg.Security.RateLimitPerMinute)
		engine.Use(rateLimiter.RateLimit())
	}

	router := &Router{
		engine:         engine,
		handler:        handler,
		controlHandler: controlHandler,
		config:         cfg,
		policyEngine:   policyEngine,
		auditLogger:    auditLogger,
	}

	router.setupRoutes()

	return router
}

func (r *Router) setupRoutes() {
	// Health endpoints
	r.engine.GET("/health", r.handler.Health)
	r.engine.GET("/ready", r.handler.Ready)

	// API v1 routes - Data Plane
	v1 := r.engine.Group("/api/v1")
	{
		// Main guard endpoint - full pipeline
		v1.POST("/guard", r.handler.Guard)

		// Individual service endpoints
		v1.POST("/analyze", r.handler.Analyze)
		v1.POST("/mask", r.handler.MaskPII)
		v1.POST("/detect", r.handler.DetectInjection)
	}

	// Control Plane API routes
	control := r.engine.Group("/api/v1/control")
	{
		// Policy management
		policies := control.Group("/policies")
		{
			policies.POST("", r.controlHandler.CreatePolicy)
			policies.GET("", r.controlHandler.ListPolicies)
			policies.GET("/:id", r.controlHandler.GetPolicy)
			policies.PUT("/:id", r.controlHandler.UpdatePolicy)
			policies.DELETE("/:id", r.controlHandler.DeletePolicy)
		}

		// Spending limits
		spending := control.Group("/spending-limits")
		{
			spending.POST("", r.controlHandler.CreateSpendingLimit)
			spending.GET("", r.controlHandler.ListSpendingLimits)
			spending.GET("/:id", r.controlHandler.GetSpendingLimit)
			spending.PUT("/:id", r.controlHandler.UpdateSpendingLimit)
		}

		// User management
		users := control.Group("/users")
		{
			users.POST("", r.controlHandler.CreateUser)
			users.GET("", r.controlHandler.ListUsers)
			users.GET("/:id", r.controlHandler.GetUser)
			users.PUT("/:id", r.controlHandler.UpdateUser)
			users.DELETE("/:id", r.controlHandler.DeleteUser)
		}

		// Audit logs
		audit := control.Group("/audit")
		{
			audit.GET("/logs", r.controlHandler.QueryAuditLogs)
			audit.GET("/stats", r.controlHandler.GetAuditStats)
		}

		// Dashboard
		control.GET("/dashboard", r.controlHandler.GetDashboardMetrics)

		// Alerts
		alerts := control.Group("/alerts")
		{
			alerts.GET("", r.controlHandler.GetAlerts)
			alerts.POST("/:id/ack", r.controlHandler.AckAlert)
		}

		// Settings
		settingsGroup := control.Group("/settings")
		{
			settingsGroup.GET("", r.controlHandler.GetSettings)
			settingsGroup.GET("/llm", r.controlHandler.GetLLMSettings)
			settingsGroup.PUT("/llm", r.controlHandler.UpdateLLMSettings)
			settingsGroup.GET("/security", r.controlHandler.GetSecuritySettings)
			settingsGroup.PUT("/security", r.controlHandler.UpdateSecuritySettings)
			settingsGroup.GET("/storage", r.controlHandler.GetStorageInfo)
		}
	}
}

// Engine returns the underlying gin engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

// PolicyEngine returns the policy engine for external use
func (r *Router) PolicyEngine() *policy.Engine {
	return r.policyEngine
}

// AuditLogger returns the audit logger for external use
func (r *Router) AuditLogger() *audit.Logger {
	return r.auditLogger
}
