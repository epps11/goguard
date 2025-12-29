package models

import "time"

// AuditLog represents an audit log entry
type AuditLog struct {
	ID            string                 `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	EventType     AuditEventType         `json:"event_type"`
	Action        string                 `json:"action"`
	UserID        string                 `json:"user_id,omitempty"`
	UserEmail     string                 `json:"user_email,omitempty"`
	ResourceType  string                 `json:"resource_type"`
	ResourceID    string                 `json:"resource_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Status        AuditStatus            `json:"status"`
	Details       map[string]interface{} `json:"details,omitempty"`
	PolicyResults []PolicyEvaluation     `json:"policy_results,omitempty"`
	Duration      time.Duration          `json:"duration_ms"`
}

// AuditEventType defines the type of audit event
type AuditEventType string

const (
	EventTypeRequest       AuditEventType = "request"
	EventTypePolicyChange  AuditEventType = "policy_change"
	EventTypeUserAction    AuditEventType = "user_action"
	EventTypeSystemEvent   AuditEventType = "system_event"
	EventTypeSecurityAlert AuditEventType = "security_alert"
	EventTypeSpendingAlert AuditEventType = "spending_alert"
)

// AuditStatus defines the status of an audit event
type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailure AuditStatus = "failure"
	AuditStatusBlocked AuditStatus = "blocked"
	AuditStatusWarning AuditStatus = "warning"
)

// AuditQuery represents query parameters for audit logs
type AuditQuery struct {
	StartTime    *time.Time       `json:"start_time,omitempty"`
	EndTime      *time.Time       `json:"end_time,omitempty"`
	EventTypes   []AuditEventType `json:"event_types,omitempty"`
	UserID       string           `json:"user_id,omitempty"`
	ResourceType string           `json:"resource_type,omitempty"`
	Status       AuditStatus      `json:"status,omitempty"`
	Limit        int              `json:"limit,omitempty"`
	Offset       int              `json:"offset,omitempty"`
	SortBy       string           `json:"sort_by,omitempty"`
	SortOrder    string           `json:"sort_order,omitempty"`
}

// AuditStats represents aggregated audit statistics
type AuditStats struct {
	TotalRequests   int64            `json:"total_requests"`
	BlockedRequests int64            `json:"blocked_requests"`
	AllowedRequests int64            `json:"allowed_requests"`
	WarningRequests int64            `json:"warning_requests"`
	UniqueUsers     int64            `json:"unique_users"`
	TotalTokensUsed int64            `json:"total_tokens_used"`
	TotalCost       float64          `json:"total_cost"`
	TopUsers        []UserStats      `json:"top_users"`
	TopModels       []ModelStats     `json:"top_models"`
	RequestsByHour  map[string]int64 `json:"requests_by_hour"`
	EventsByType    map[string]int64 `json:"events_by_type"`
	Period          string           `json:"period"`
}

// UserStats represents usage statistics for a user
type UserStats struct {
	UserID       string  `json:"user_id"`
	UserEmail    string  `json:"user_email"`
	RequestCount int64   `json:"request_count"`
	TokensUsed   int64   `json:"tokens_used"`
	TotalCost    float64 `json:"total_cost"`
}

// ModelStats represents usage statistics for a model
type ModelStats struct {
	Model        string  `json:"model"`
	Provider     string  `json:"provider"`
	RequestCount int64   `json:"request_count"`
	TokensUsed   int64   `json:"tokens_used"`
	TotalCost    float64 `json:"total_cost"`
}

// DashboardMetrics represents metrics for the dashboard
type DashboardMetrics struct {
	Overview     OverviewMetrics `json:"overview"`
	Security     SecurityMetrics `json:"security"`
	Usage        UsageMetrics    `json:"usage"`
	Spending     SpendingMetrics `json:"spending"`
	RecentAlerts []Alert         `json:"recent_alerts"`
	TopPolicies  []PolicyMetric  `json:"top_policies"`
}

// OverviewMetrics represents high-level overview metrics
type OverviewMetrics struct {
	TotalRequests24h   int64   `json:"total_requests_24h"`
	RequestsChange     float64 `json:"requests_change_percent"`
	ActiveUsers24h     int64   `json:"active_users_24h"`
	UsersChange        float64 `json:"users_change_percent"`
	BlockedRequests24h int64   `json:"blocked_requests_24h"`
	BlockedChange      float64 `json:"blocked_change_percent"`
	TotalSpend24h      float64 `json:"total_spend_24h"`
	SpendChange        float64 `json:"spend_change_percent"`
}

// SecurityMetrics represents security-related metrics
type SecurityMetrics struct {
	InjectionAttempts24h int64            `json:"injection_attempts_24h"`
	PIIDetections24h     int64            `json:"pii_detections_24h"`
	ThreatsByLevel       map[string]int64 `json:"threats_by_level"`
	TopThreatTypes       map[string]int64 `json:"top_threat_types"`
}

// UsageMetrics represents usage metrics
type UsageMetrics struct {
	TotalTokens24h      int64            `json:"total_tokens_24h"`
	PromptTokens24h     int64            `json:"prompt_tokens_24h"`
	CompletionTokens24h int64            `json:"completion_tokens_24h"`
	RequestsByModel     map[string]int64 `json:"requests_by_model"`
	RequestsByProvider  map[string]int64 `json:"requests_by_provider"`
}

// SpendingMetrics represents spending metrics
type SpendingMetrics struct {
	TotalSpendToday float64            `json:"total_spend_today"`
	TotalSpendMonth float64            `json:"total_spend_month"`
	BudgetRemaining float64            `json:"budget_remaining"`
	SpendByUser     map[string]float64 `json:"spend_by_user"`
	SpendByModel    map[string]float64 `json:"spend_by_model"`
	ProjectedSpend  float64            `json:"projected_spend_month"`
}

// Alert represents a system alert
type Alert struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Severity  string     `json:"severity"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	UserID    string     `json:"user_id,omitempty"`
	PolicyID  string     `json:"policy_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	AckedAt   *time.Time `json:"acked_at,omitempty"`
	AckedBy   string     `json:"acked_by,omitempty"`
}

// PolicyMetric represents metrics for a policy
type PolicyMetric struct {
	PolicyID     string `json:"policy_id"`
	PolicyName   string `json:"policy_name"`
	TriggerCount int64  `json:"trigger_count"`
	BlockCount   int64  `json:"block_count"`
	WarnCount    int64  `json:"warn_count"`
}
