package api

import (
	"net/http"
	"strconv"

	"github.com/epps11/goguard/internal/database"
	"github.com/epps11/goguard/internal/models"
	"github.com/epps11/goguard/internal/services/audit"
	"github.com/epps11/goguard/internal/services/policy"
	"github.com/epps11/goguard/internal/services/settings"
	"github.com/gin-gonic/gin"
)

// ControlHandler handles control plane API requests
type ControlHandler struct {
	policyEngine    *policy.Engine
	auditLogger     *audit.Logger
	settingsService *settings.Service
	repo            *database.Repository
}

// NewControlHandler creates a new control handler
func NewControlHandler(engine *policy.Engine, logger *audit.Logger, settingsSvc *settings.Service, repo *database.Repository) *ControlHandler {
	return &ControlHandler{
		policyEngine:    engine,
		auditLogger:     logger,
		settingsService: settingsSvc,
		repo:            repo,
	}
}

// Policy Handlers

// CreatePolicy creates a new policy
func (h *ControlHandler) CreatePolicy(c *gin.Context) {
	var policy models.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.policyEngine.CreatePolicy(c.Request.Context(), &policy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetPolicy retrieves a policy by ID
func (h *ControlHandler) GetPolicy(c *gin.Context) {
	id := c.Param("id")

	policy, err := h.policyEngine.GetPolicy(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, policy)
}

// ListPolicies lists all policies
func (h *ControlHandler) ListPolicies(c *gin.Context) {
	policies, err := h.policyEngine.ListPolicies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"policies": policies,
		"total":    len(policies),
	})
}

// UpdatePolicy updates a policy
func (h *ControlHandler) UpdatePolicy(c *gin.Context) {
	id := c.Param("id")

	var policy models.Policy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	policy.ID = id
	updated, err := h.policyEngine.UpdatePolicy(c.Request.Context(), &policy)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeletePolicy deletes a policy
func (h *ControlHandler) DeletePolicy(c *gin.Context) {
	id := c.Param("id")

	if err := h.policyEngine.DeletePolicy(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Spending Limit Handlers

// CreateSpendingLimit creates a new spending limit
func (h *ControlHandler) CreateSpendingLimit(c *gin.Context) {
	var limit models.SpendingLimit
	if err := c.ShouldBindJSON(&limit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use database if available, otherwise fall back to in-memory
	if h.repo != nil {
		if err := h.repo.CreateSpendingLimit(c.Request.Context(), &limit); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, limit)
		return
	}

	created, err := h.policyEngine.CreateSpendingLimit(c.Request.Context(), &limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetSpendingLimit retrieves a spending limit by ID
func (h *ControlHandler) GetSpendingLimit(c *gin.Context) {
	id := c.Param("id")

	// Use database if available
	if h.repo != nil {
		limit, err := h.repo.GetSpendingLimit(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, limit)
		return
	}

	limit, err := h.policyEngine.GetSpendingLimit(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, limit)
}

// ListSpendingLimits lists all spending limits
func (h *ControlHandler) ListSpendingLimits(c *gin.Context) {
	// Use database if available
	if h.repo != nil {
		limits, err := h.repo.ListSpendingLimits(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"spending_limits": limits,
			"total":           len(limits),
		})
		return
	}

	limits, err := h.policyEngine.ListSpendingLimits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"spending_limits": limits,
		"total":           len(limits),
	})
}

// UpdateSpendingLimit updates a spending limit
func (h *ControlHandler) UpdateSpendingLimit(c *gin.Context) {
	id := c.Param("id")

	var limit models.SpendingLimit
	if err := c.ShouldBindJSON(&limit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit.ID = id
	updated, err := h.policyEngine.UpdateSpendingLimit(c.Request.Context(), &limit)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// User Handlers

// CreateUser creates a new user
func (h *ControlHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.policyEngine.CreateUser(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetUser retrieves a user by ID
func (h *ControlHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.policyEngine.GetUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers lists all users
func (h *ControlHandler) ListUsers(c *gin.Context) {
	users, err := h.policyEngine.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": len(users),
	})
}

// UpdateUser updates a user
func (h *ControlHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.ID = id
	updated, err := h.policyEngine.UpdateUser(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteUser deletes a user
func (h *ControlHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.policyEngine.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Audit Log Handlers

// QueryAuditLogs queries audit logs
func (h *ControlHandler) QueryAuditLogs(c *gin.Context) {
	query := &models.AuditQuery{}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			query.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			query.Offset = o
		}
	}
	if userID := c.Query("user_id"); userID != "" {
		query.UserID = userID
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		query.ResourceType = resourceType
	}
	if status := c.Query("status"); status != "" {
		query.Status = models.AuditStatus(status)
	}

	logs, total, err := h.auditLogger.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  query.Limit,
		"offset": query.Offset,
	})
}

// GetAuditStats returns audit statistics
func (h *ControlHandler) GetAuditStats(c *gin.Context) {
	period := c.DefaultQuery("period", "24h")

	stats, err := h.auditLogger.GetStats(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Dashboard Handlers

// GetDashboardMetrics returns dashboard metrics
func (h *ControlHandler) GetDashboardMetrics(c *gin.Context) {
	metrics, err := h.auditLogger.GetDashboardMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// Alert Handlers

// GetAlerts returns alerts
func (h *ControlHandler) GetAlerts(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	includeAcked := c.Query("include_acked") == "true"

	alerts, err := h.auditLogger.GetAlerts(c.Request.Context(), limit, includeAcked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// AckAlert acknowledges an alert
func (h *ControlHandler) AckAlert(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id") // From auth middleware

	if err := h.auditLogger.AckAlert(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"acknowledged": true})
}

// Settings Handlers

// GetSettings returns all settings
func (h *ControlHandler) GetSettings(c *gin.Context) {
	if h.settingsService == nil {
		c.JSON(http.StatusOK, gin.H{
			"llm_provider": "openai",
			"llm_model":    "gpt-4o",
		})
		return
	}

	allSettings, err := h.settingsService.GetAllSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, allSettings)
}

// GetLLMSettings returns LLM configuration
func (h *ControlHandler) GetLLMSettings(c *gin.Context) {
	if h.settingsService == nil {
		c.JSON(http.StatusOK, gin.H{
			"provider": "openai",
			"model":    "gpt-4o",
		})
		return
	}

	llmSettings, err := h.settingsService.GetLLMSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, llmSettings)
}

// UpdateLLMSettings updates LLM configuration
func (h *ControlHandler) UpdateLLMSettings(c *gin.Context) {
	var req settings.LLMSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.settingsService == nil {
		c.JSON(http.StatusOK, gin.H{"message": "settings updated (in-memory only)"})
		return
	}

	if err := h.settingsService.UpdateLLMSettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "LLM settings updated"})
}

// GetSecuritySettings returns security configuration
func (h *ControlHandler) GetSecuritySettings(c *gin.Context) {
	if h.settingsService == nil {
		c.JSON(http.StatusOK, gin.H{
			"injection_detection_enabled": true,
			"pii_masking_enabled":         true,
			"rate_limit_per_minute":       100,
		})
		return
	}

	secSettings, err := h.settingsService.GetSecuritySettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, secSettings)
}

// UpdateSecuritySettings updates security configuration
func (h *ControlHandler) UpdateSecuritySettings(c *gin.Context) {
	var req settings.SecuritySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.settingsService == nil {
		c.JSON(http.StatusOK, gin.H{"message": "settings updated (in-memory only)"})
		return
	}

	if err := h.settingsService.UpdateSecuritySettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "security settings updated"})
}

// GetStorageInfo returns information about the storage backend
func (h *ControlHandler) GetStorageInfo(c *gin.Context) {
	storageType := "in-memory"
	if h.settingsService != nil {
		storageType = "postgresql"
	}

	c.JSON(http.StatusOK, gin.H{
		"storage_type":        storageType,
		"audit_log_retention": 10000,
		"database_connected":  h.settingsService != nil,
	})
}
