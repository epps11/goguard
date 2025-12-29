package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/epps11/goguard/internal/models"
	"github.com/epps11/goguard/internal/services/audit"
	"github.com/epps11/goguard/internal/services/injection"
	"github.com/epps11/goguard/internal/services/llm"
	"github.com/epps11/goguard/internal/services/pii"
)

// Handler contains all HTTP handlers
type Handler struct {
	injectionDetector *injection.Detector
	piiMasker         *pii.Masker
	llmClient         *llm.Client
	llmFactory        *llm.ClientFactory
	auditLogger       *audit.Logger
	startTime         time.Time
	version           string
}

// NewHandler creates a new handler instance
func NewHandler(detector *injection.Detector, masker *pii.Masker, client *llm.Client, logger *audit.Logger) *Handler {
	return &Handler{
		injectionDetector: detector,
		piiMasker:         masker,
		llmClient:         client,
		auditLogger:       logger,
		startTime:         time.Now(),
		version:           "1.0.0",
	}
}

// NewHandlerWithFactory creates a new handler with LLM client factory for per-request provider support
func NewHandlerWithFactory(detector *injection.Detector, masker *pii.Masker, factory *llm.ClientFactory, logger *audit.Logger) *Handler {
	return &Handler{
		injectionDetector: detector,
		piiMasker:         masker,
		llmClient:         factory.GetDefaultClient(),
		llmFactory:        factory,
		auditLogger:       logger,
		startTime:         time.Now(),
		version:           "1.0.0",
	}
}

// Guard processes a request through the security pipeline
func (h *Handler) Guard(c *gin.Context) {
	startTime := time.Now()

	var req models.GuardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	response := &models.GuardResponse{
		RequestID: req.RequestID,
		Allowed:   true,
	}

	// Step 1: Injection Detection
	securityReport := h.injectionDetector.Analyze(req.Messages)
	response.SecurityReport = securityReport

	if h.injectionDetector.ShouldBlock(securityReport) {
		response.Allowed = false
		response.ProcessingTime = time.Since(startTime)
		c.JSON(http.StatusForbidden, response)
		return
	}

	// Step 2: PII Masking
	maskedMessages, piiReport := h.piiMasker.Mask(req.Messages)
	response.PIIReport = piiReport
	response.ProcessedInput = &models.ProcessedInput{
		OriginalMessages: req.Messages,
		MaskedMessages:   maskedMessages,
		PIIMasked:        piiReport.PIIDetected,
	}

	// Step 3: Forward to LLM (if client is configured)
	// Use factory if available for per-request provider support
	if h.llmFactory != nil {
		client, shouldClose, err := h.llmFactory.GetClient(&req)
		if err != nil {
			response.Error = err.Error()
		} else {
			if shouldClose {
				defer client.Close()
			}
			llmResp, err := client.Chat(c.Request.Context(), maskedMessages)
			if err != nil {
				response.Error = err.Error()
			} else {
				response.LLMResponse = llmResp
			}
		}
	} else if h.llmClient != nil && h.llmClient.IsInitialized() {
		llmResp, err := h.llmClient.Chat(c.Request.Context(), maskedMessages)
		if err != nil {
			response.Error = err.Error()
		} else {
			response.LLMResponse = llmResp
		}
	}

	response.ProcessingTime = time.Since(startTime)

	// Log to audit
	h.logRequest(c, req.RequestID, "guard", response.Allowed, response.SecurityReport, response.PIIReport, time.Since(startTime))

	c.JSON(http.StatusOK, response)
}

// Analyze performs security analysis without forwarding to LLM
func (h *Handler) Analyze(c *gin.Context) {
	startTime := time.Now()

	var req models.GuardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	response := &models.GuardResponse{
		RequestID:      req.RequestID,
		Allowed:        true,
		SecurityReport: h.injectionDetector.Analyze(req.Messages),
		PIIReport:      h.piiMasker.Analyze(req.Messages),
		ProcessingTime: time.Since(startTime),
	}

	if h.injectionDetector.ShouldBlock(response.SecurityReport) {
		response.Allowed = false
	}

	// Log to audit
	h.logRequest(c, req.RequestID, "analyze", response.Allowed, response.SecurityReport, response.PIIReport, time.Since(startTime))

	c.JSON(http.StatusOK, response)
}

// MaskPII masks PII in the provided messages
func (h *Handler) MaskPII(c *gin.Context) {
	startTime := time.Now()

	var req models.GuardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	maskedMessages, piiReport := h.piiMasker.Mask(req.Messages)

	response := &models.GuardResponse{
		RequestID: req.RequestID,
		Allowed:   true,
		ProcessedInput: &models.ProcessedInput{
			MaskedMessages: maskedMessages,
			PIIMasked:      piiReport.PIIDetected,
		},
		PIIReport:      piiReport,
		ProcessingTime: time.Since(startTime),
	}

	// Log to audit
	h.logRequest(c, req.RequestID, "mask", true, nil, piiReport, time.Since(startTime))

	c.JSON(http.StatusOK, response)
}

// DetectInjection checks for injection attempts
func (h *Handler) DetectInjection(c *gin.Context) {
	startTime := time.Now()

	var req models.GuardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request body",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	securityReport := h.injectionDetector.Analyze(req.Messages)

	response := &models.GuardResponse{
		RequestID:      req.RequestID,
		Allowed:        !h.injectionDetector.ShouldBlock(securityReport),
		SecurityReport: securityReport,
		ProcessingTime: time.Since(startTime),
	}

	// Log to audit
	h.logRequest(c, req.RequestID, "detect", response.Allowed, securityReport, nil, time.Since(startTime))

	c.JSON(http.StatusOK, response)
}

// Health returns the health status
func (h *Handler) Health(c *gin.Context) {
	services := map[string]string{
		"injection_detector": "healthy",
		"pii_masker":         "healthy",
	}

	if h.llmClient != nil && h.llmClient.IsInitialized() {
		services["llm_client"] = "healthy"
	} else {
		services["llm_client"] = "not_configured"
	}

	c.JSON(http.StatusOK, models.HealthResponse{
		Status:   "healthy",
		Version:  h.version,
		Uptime:   time.Since(h.startTime).String(),
		Services: services,
	})
}

// Ready returns readiness status
func (h *Handler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}

// logRequest logs a request to the audit logger
func (h *Handler) logRequest(c *gin.Context, requestID, action string, allowed bool, secReport *models.SecurityReport, piiReport *models.PIIReport, duration time.Duration) {
	if h.auditLogger == nil {
		return
	}

	status := models.AuditStatusSuccess
	if !allowed {
		status = models.AuditStatusBlocked
	}

	details := map[string]interface{}{
		"action": action,
	}

	if secReport != nil {
		details["injection_detected"] = secReport.InjectionDetected
		details["threat_level"] = secReport.ThreatLevel
		if secReport.InjectionDetected {
			details["detection_count"] = len(secReport.Detections)
		}
	}

	if piiReport != nil {
		details["pii_detected"] = piiReport.PIIDetected
		details["pii_count"] = piiReport.PIICount
	}

	entry := &models.AuditLog{
		RequestID:    requestID,
		EventType:    models.EventTypeRequest,
		Action:       action,
		ResourceType: "llm",
		Status:       status,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		Duration:     duration,
		Details:      details,
	}

	h.auditLogger.Log(c.Request.Context(), entry)
}
