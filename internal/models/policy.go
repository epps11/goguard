package models

import "time"

// Policy represents an AI governance policy
type Policy struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        PolicyType        `json:"type"`
	Status      PolicyStatus      `json:"status"`
	Priority    int               `json:"priority"`
	Config      PolicyConfig      `json:"config"`
	Rules       []PolicyRule      `json:"rules"`
	Targets     PolicyTargets     `json:"targets"`
	Actions     PolicyActions     `json:"actions"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by"`
}

// PolicyConfig holds type-specific configuration for policies
type PolicyConfig struct {
	// Spending Limit
	DailyLimit   float64 `json:"daily_limit,omitempty"`
	MonthlyLimit float64 `json:"monthly_limit,omitempty"`
	Currency     string  `json:"currency,omitempty"`

	// Rate Limit
	RequestsPerMinute int `json:"requests_per_minute,omitempty"`
	RequestsPerHour   int `json:"requests_per_hour,omitempty"`
	BurstLimit        int `json:"burst_limit,omitempty"`

	// Content Filter
	BlockedKeywords string `json:"blocked_keywords,omitempty"`
	AllowedModels   string `json:"allowed_models,omitempty"`
	MaxTokens       int    `json:"max_tokens,omitempty"`

	// Access Control
	AllowedRoles string `json:"allowed_roles,omitempty"`
	AllowedUsers string `json:"allowed_users,omitempty"`
	DeniedUsers  string `json:"denied_users,omitempty"`

	// Compliance
	RequireAudit      bool   `json:"require_audit,omitempty"`
	DataRetentionDays int    `json:"data_retention_days,omitempty"`
	PIIHandling       string `json:"pii_handling,omitempty"`
}

// PolicyType defines the type of policy
type PolicyType string

const (
	PolicyTypeSpending   PolicyType = "spending"
	PolicyTypeRateLimit  PolicyType = "rate_limit"
	PolicyTypeContent    PolicyType = "content"
	PolicyTypeAccess     PolicyType = "access"
	PolicyTypeCompliance PolicyType = "compliance"
)

// PolicyStatus defines the status of a policy
type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusInactive PolicyStatus = "inactive"
	PolicyStatusDraft    PolicyStatus = "draft"
)

// PolicyRule defines a single rule within a policy
type PolicyRule struct {
	ID        string        `json:"id"`
	Field     string        `json:"field"`    // e.g., "user_id", "model", "token_count"
	Operator  RuleOperator  `json:"operator"` // e.g., "equals", "greater_than"
	Value     interface{}   `json:"value"`
	Condition RuleCondition `json:"condition"` // AND, OR
}

// RuleOperator defines comparison operators
type RuleOperator string

const (
	OperatorEquals      RuleOperator = "equals"
	OperatorNotEquals   RuleOperator = "not_equals"
	OperatorGreaterThan RuleOperator = "greater_than"
	OperatorLessThan    RuleOperator = "less_than"
	OperatorContains    RuleOperator = "contains"
	OperatorNotContains RuleOperator = "not_contains"
	OperatorIn          RuleOperator = "in"
	OperatorNotIn       RuleOperator = "not_in"
)

// RuleCondition defines logical conditions
type RuleCondition string

const (
	ConditionAnd RuleCondition = "and"
	ConditionOr  RuleCondition = "or"
)

// PolicyTargets defines who/what the policy applies to
type PolicyTargets struct {
	Users     []string `json:"users,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	Models    []string `json:"models,omitempty"`
	Providers []string `json:"providers,omitempty"`
	AllUsers  bool     `json:"all_users,omitempty"`
}

// PolicyActions defines what happens when policy is triggered
type PolicyActions struct {
	Action     ActionType `json:"action"`
	Notify     []string   `json:"notify,omitempty"` // email addresses
	WebhookURL string     `json:"webhook_url,omitempty"`
	LogLevel   string     `json:"log_level,omitempty"`
	Message    string     `json:"message,omitempty"`
}

// ActionType defines the action to take
type ActionType string

const (
	ActionAllow    ActionType = "allow"
	ActionDeny     ActionType = "deny"
	ActionWarn     ActionType = "warn"
	ActionAudit    ActionType = "audit"
	ActionThrottle ActionType = "throttle"
)

// SpendingLimit represents a spending limit policy
type SpendingLimit struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id,omitempty"`
	GroupID      string    `json:"group_id,omitempty"`
	LimitType    string    `json:"limit_type"` // daily, weekly, monthly
	LimitAmount  float64   `json:"limit_amount"`
	CurrentSpend float64   `json:"current_spend"`
	Currency     string    `json:"currency"`
	ResetAt      time.Time `json:"reset_at"`
	AlertAt      float64   `json:"alert_at"` // percentage to alert at
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID          string            `json:"id"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Role        UserRole          `json:"role"`
	Groups      []string          `json:"groups"`
	Status      string            `json:"status"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	LastLoginAt *time.Time        `json:"last_login_at,omitempty"`
}

// UserRole defines user roles with RBAC
type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin" // Full system access, can manage other admins
	RoleAdmin      UserRole = "admin"       // Can manage policies, users, and view all data
	RoleUser       UserRole = "user"        // Standard user, can use AI features within limits
	RoleViewer     UserRole = "viewer"      // Read-only access to dashboards and reports
)

// Group represents a group of users
type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Members     []string  `json:"members"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PolicyEvaluation represents the result of evaluating a policy
type PolicyEvaluation struct {
	PolicyID    string     `json:"policy_id"`
	PolicyName  string     `json:"policy_name"`
	Matched     bool       `json:"matched"`
	Action      ActionType `json:"action"`
	Message     string     `json:"message,omitempty"`
	EvaluatedAt time.Time  `json:"evaluated_at"`
}
