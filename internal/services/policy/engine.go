package policy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/epps11/goguard/internal/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Engine manages policy evaluation and storage
type Engine struct {
	policies       map[string]*models.Policy
	spendingLimits map[string]*models.SpendingLimit
	users          map[string]*models.User
	groups         map[string]*models.Group
	mu             sync.RWMutex
}

// NewEngine creates a new policy engine
func NewEngine() *Engine {
	return &Engine{
		policies:       make(map[string]*models.Policy),
		spendingLimits: make(map[string]*models.SpendingLimit),
		users:          make(map[string]*models.User),
		groups:         make(map[string]*models.Group),
	}
}

// CreatePolicy creates a new policy
func (e *Engine) CreatePolicy(ctx context.Context, policy *models.Policy) (*models.Policy, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if policy.ID == "" {
		policy.ID = uuid.New().String()
	}
	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	e.policies[policy.ID] = policy

	log.Info().
		Str("policy_id", policy.ID).
		Str("name", policy.Name).
		Str("type", string(policy.Type)).
		Msg("Policy created")

	return policy, nil
}

// GetPolicy retrieves a policy by ID
func (e *Engine) GetPolicy(ctx context.Context, id string) (*models.Policy, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policy, exists := e.policies[id]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", id)
	}
	return policy, nil
}

// ListPolicies returns all policies
func (e *Engine) ListPolicies(ctx context.Context) ([]*models.Policy, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies := make([]*models.Policy, 0, len(e.policies))
	for _, p := range e.policies {
		policies = append(policies, p)
	}
	return policies, nil
}

// UpdatePolicy updates an existing policy
func (e *Engine) UpdatePolicy(ctx context.Context, policy *models.Policy) (*models.Policy, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	existing, exists := e.policies[policy.ID]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", policy.ID)
	}

	policy.CreatedAt = existing.CreatedAt
	policy.UpdatedAt = time.Now()
	e.policies[policy.ID] = policy

	log.Info().
		Str("policy_id", policy.ID).
		Str("name", policy.Name).
		Msg("Policy updated")

	return policy, nil
}

// DeletePolicy deletes a policy
func (e *Engine) DeletePolicy(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.policies[id]; !exists {
		return fmt.Errorf("policy not found: %s", id)
	}

	delete(e.policies, id)

	log.Info().Str("policy_id", id).Msg("Policy deleted")
	return nil
}

// EvaluateRequest evaluates all policies against a request
func (e *Engine) EvaluateRequest(ctx context.Context, req *EvaluationRequest) (*EvaluationResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := &EvaluationResult{
		Allowed:     true,
		Evaluations: []models.PolicyEvaluation{},
	}

	// Get active policies sorted by priority
	activePolicies := e.getActivePolicies()

	for _, policy := range activePolicies {
		eval := e.evaluatePolicy(policy, req)
		result.Evaluations = append(result.Evaluations, eval)

		if eval.Matched {
			switch eval.Action {
			case models.ActionDeny:
				result.Allowed = false
				result.BlockedBy = policy.ID
				result.BlockReason = eval.Message
			case models.ActionWarn:
				result.Warnings = append(result.Warnings, eval.Message)
			case models.ActionThrottle:
				result.Throttled = true
			}
		}
	}

	return result, nil
}

// EvaluationRequest represents a request to be evaluated
type EvaluationRequest struct {
	UserID      string
	Model       string
	Provider    string
	TokenCount  int
	Cost        float64
	ContentType string
	Metadata    map[string]interface{}
}

// EvaluationResult represents the result of policy evaluation
type EvaluationResult struct {
	Allowed     bool
	BlockedBy   string
	BlockReason string
	Warnings    []string
	Throttled   bool
	Evaluations []models.PolicyEvaluation
}

func (e *Engine) getActivePolicies() []*models.Policy {
	var active []*models.Policy
	for _, p := range e.policies {
		if p.Status == models.PolicyStatusActive {
			active = append(active, p)
		}
	}
	return active
}

func (e *Engine) evaluatePolicy(policy *models.Policy, req *EvaluationRequest) models.PolicyEvaluation {
	eval := models.PolicyEvaluation{
		PolicyID:    policy.ID,
		PolicyName:  policy.Name,
		Matched:     false,
		Action:      policy.Actions.Action,
		EvaluatedAt: time.Now(),
	}

	// Check if policy targets this user
	if !e.policyTargetsUser(policy, req.UserID) {
		return eval
	}

	// Evaluate all rules
	matched := e.evaluateRules(policy.Rules, req)
	eval.Matched = matched

	if matched {
		eval.Message = policy.Actions.Message
		if eval.Message == "" {
			eval.Message = fmt.Sprintf("Policy '%s' triggered", policy.Name)
		}
	}

	return eval
}

func (e *Engine) policyTargetsUser(policy *models.Policy, userID string) bool {
	if policy.Targets.AllUsers {
		return true
	}

	for _, u := range policy.Targets.Users {
		if u == userID {
			return true
		}
	}

	// Check groups
	user, exists := e.users[userID]
	if exists {
		for _, groupID := range user.Groups {
			for _, targetGroup := range policy.Targets.Groups {
				if groupID == targetGroup {
					return true
				}
			}
		}
	}

	return len(policy.Targets.Users) == 0 && len(policy.Targets.Groups) == 0
}

func (e *Engine) evaluateRules(rules []models.PolicyRule, req *EvaluationRequest) bool {
	if len(rules) == 0 {
		return true
	}

	for i, rule := range rules {
		matched := e.evaluateRule(rule, req)

		if i == 0 {
			if !matched {
				return false
			}
			continue
		}

		switch rule.Condition {
		case models.ConditionAnd:
			if !matched {
				return false
			}
		case models.ConditionOr:
			if matched {
				return true
			}
		}
	}

	return true
}

func (e *Engine) evaluateRule(rule models.PolicyRule, req *EvaluationRequest) bool {
	var fieldValue interface{}

	switch rule.Field {
	case "user_id":
		fieldValue = req.UserID
	case "model":
		fieldValue = req.Model
	case "provider":
		fieldValue = req.Provider
	case "token_count":
		fieldValue = req.TokenCount
	case "cost":
		fieldValue = req.Cost
	default:
		if req.Metadata != nil {
			fieldValue = req.Metadata[rule.Field]
		}
	}

	return e.compareValues(fieldValue, rule.Operator, rule.Value)
}

func (e *Engine) compareValues(fieldValue interface{}, operator models.RuleOperator, ruleValue interface{}) bool {
	switch operator {
	case models.OperatorEquals:
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", ruleValue)
	case models.OperatorNotEquals:
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", ruleValue)
	case models.OperatorGreaterThan:
		return toFloat(fieldValue) > toFloat(ruleValue)
	case models.OperatorLessThan:
		return toFloat(fieldValue) < toFloat(ruleValue)
	case models.OperatorContains:
		return contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", ruleValue))
	case models.OperatorNotContains:
		return !contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", ruleValue))
	default:
		return false
	}
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	default:
		return 0
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Spending Limit Methods

// CreateSpendingLimit creates a new spending limit
func (e *Engine) CreateSpendingLimit(ctx context.Context, limit *models.SpendingLimit) (*models.SpendingLimit, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if limit.ID == "" {
		limit.ID = uuid.New().String()
	}
	limit.CreatedAt = time.Now()
	limit.UpdatedAt = time.Now()
	limit.CurrentSpend = 0

	e.spendingLimits[limit.ID] = limit

	log.Info().
		Str("limit_id", limit.ID).
		Str("user_id", limit.UserID).
		Float64("amount", limit.LimitAmount).
		Msg("Spending limit created")

	return limit, nil
}

// GetSpendingLimit retrieves a spending limit by ID
func (e *Engine) GetSpendingLimit(ctx context.Context, id string) (*models.SpendingLimit, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	limit, exists := e.spendingLimits[id]
	if !exists {
		return nil, fmt.Errorf("spending limit not found: %s", id)
	}
	return limit, nil
}

// GetUserSpendingLimits retrieves all spending limits for a user
func (e *Engine) GetUserSpendingLimits(ctx context.Context, userID string) ([]*models.SpendingLimit, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var limits []*models.SpendingLimit
	for _, l := range e.spendingLimits {
		if l.UserID == userID {
			limits = append(limits, l)
		}
	}
	return limits, nil
}

// ListSpendingLimits returns all spending limits
func (e *Engine) ListSpendingLimits(ctx context.Context) ([]*models.SpendingLimit, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	limits := make([]*models.SpendingLimit, 0, len(e.spendingLimits))
	for _, l := range e.spendingLimits {
		limits = append(limits, l)
	}
	return limits, nil
}

// UpdateSpendingLimit updates a spending limit
func (e *Engine) UpdateSpendingLimit(ctx context.Context, limit *models.SpendingLimit) (*models.SpendingLimit, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	existing, exists := e.spendingLimits[limit.ID]
	if !exists {
		return nil, fmt.Errorf("spending limit not found: %s", limit.ID)
	}

	limit.CreatedAt = existing.CreatedAt
	limit.UpdatedAt = time.Now()
	e.spendingLimits[limit.ID] = limit

	return limit, nil
}

// RecordSpending records spending against a limit
func (e *Engine) RecordSpending(ctx context.Context, userID string, amount float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, limit := range e.spendingLimits {
		if limit.UserID == userID {
			limit.CurrentSpend += amount
			limit.UpdatedAt = time.Now()

			// Check if alert threshold reached
			if limit.AlertAt > 0 {
				percentage := (limit.CurrentSpend / limit.LimitAmount) * 100
				if percentage >= limit.AlertAt {
					log.Warn().
						Str("user_id", userID).
						Float64("current_spend", limit.CurrentSpend).
						Float64("limit", limit.LimitAmount).
						Float64("percentage", percentage).
						Msg("Spending alert threshold reached")
				}
			}
		}
	}

	return nil
}

// CheckSpendingLimit checks if a user has exceeded their spending limit
func (e *Engine) CheckSpendingLimit(ctx context.Context, userID string, additionalAmount float64) (bool, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, limit := range e.spendingLimits {
		if limit.UserID == userID {
			if limit.CurrentSpend+additionalAmount > limit.LimitAmount {
				return false, fmt.Sprintf("Spending limit exceeded: $%.2f of $%.2f used",
					limit.CurrentSpend, limit.LimitAmount)
			}
		}
	}

	return true, ""
}

// User Management Methods

// CreateUser creates a new user
func (e *Engine) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()

	e.users[user.ID] = user

	log.Info().
		Str("user_id", user.ID).
		Str("email", user.Email).
		Msg("User created")

	return user, nil
}

// GetUser retrieves a user by ID
func (e *Engine) GetUser(ctx context.Context, id string) (*models.User, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	user, exists := e.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return user, nil
}

// ListUsers returns all users
func (e *Engine) ListUsers(ctx context.Context) ([]*models.User, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	users := make([]*models.User, 0, len(e.users))
	for _, u := range e.users {
		users = append(users, u)
	}
	return users, nil
}

// UpdateUser updates a user
func (e *Engine) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	existing, exists := e.users[user.ID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", user.ID)
	}

	user.CreatedAt = existing.CreatedAt
	e.users[user.ID] = user

	return user, nil
}

// DeleteUser deletes a user
func (e *Engine) DeleteUser(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}

	delete(e.users, id)
	return nil
}
