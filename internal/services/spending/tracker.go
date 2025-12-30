package spending

import (
	"context"
	"sync"

	"github.com/epps11/goguard/internal/database"
	"github.com/epps11/goguard/internal/models"
	"github.com/rs/zerolog/log"
)

// ModelPricing contains pricing information for a specific model (per 1M tokens)
type ModelPricing struct {
	InputPricePerMillion  float64 // Cost per 1M input tokens
	OutputPricePerMillion float64 // Cost per 1M output tokens
}

// Default pricing for common models (USD per 1M tokens)
// Prices as of late 2024 - should be configurable
var defaultPricing = map[string]ModelPricing{
	// OpenAI models
	"gpt-4o":        {InputPricePerMillion: 2.50, OutputPricePerMillion: 10.00},
	"gpt-4o-mini":   {InputPricePerMillion: 0.15, OutputPricePerMillion: 0.60},
	"gpt-4-turbo":   {InputPricePerMillion: 10.00, OutputPricePerMillion: 30.00},
	"gpt-4":         {InputPricePerMillion: 30.00, OutputPricePerMillion: 60.00},
	"gpt-3.5-turbo": {InputPricePerMillion: 0.50, OutputPricePerMillion: 1.50},

	// Anthropic models
	"claude-3-5-sonnet-latest":   {InputPricePerMillion: 3.00, OutputPricePerMillion: 15.00},
	"claude-3-5-sonnet-20241022": {InputPricePerMillion: 3.00, OutputPricePerMillion: 15.00},
	"claude-3-opus-20240229":     {InputPricePerMillion: 15.00, OutputPricePerMillion: 75.00},
	"claude-3-sonnet-20240229":   {InputPricePerMillion: 3.00, OutputPricePerMillion: 15.00},
	"claude-3-haiku-20240307":    {InputPricePerMillion: 0.25, OutputPricePerMillion: 1.25},

	// Google models
	"gemini-1.5-pro":   {InputPricePerMillion: 1.25, OutputPricePerMillion: 5.00},
	"gemini-1.5-flash": {InputPricePerMillion: 0.075, OutputPricePerMillion: 0.30},
	"gemini-pro":       {InputPricePerMillion: 0.50, OutputPricePerMillion: 1.50},

	// AWS Bedrock Claude models (on-demand pricing)
	"anthropic.claude-3-5-sonnet-20241022-v2:0": {InputPricePerMillion: 3.00, OutputPricePerMillion: 15.00},
	"anthropic.claude-3-sonnet-20240229-v1:0":   {InputPricePerMillion: 3.00, OutputPricePerMillion: 15.00},
	"anthropic.claude-3-haiku-20240307-v1:0":    {InputPricePerMillion: 0.25, OutputPricePerMillion: 1.25},

	// X.AI models
	"grok-beta": {InputPricePerMillion: 5.00, OutputPricePerMillion: 15.00},

	// Default fallback for unknown models
	"default": {InputPricePerMillion: 1.00, OutputPricePerMillion: 3.00},
}

// Tracker tracks spending for users based on LLM usage
type Tracker struct {
	repo          *database.Repository
	customPricing map[string]ModelPricing
	mu            sync.RWMutex
}

// NewTracker creates a new spending tracker
func NewTracker(repo *database.Repository) *Tracker {
	return &Tracker{
		repo:          repo,
		customPricing: make(map[string]ModelPricing),
	}
}

// SetCustomPricing allows setting custom pricing for a model
func (t *Tracker) SetCustomPricing(model string, pricing ModelPricing) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.customPricing[model] = pricing
}

// GetPricing returns the pricing for a model
func (t *Tracker) GetPricing(model string) ModelPricing {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check custom pricing first
	if pricing, ok := t.customPricing[model]; ok {
		return pricing
	}

	// Check default pricing
	if pricing, ok := defaultPricing[model]; ok {
		return pricing
	}

	// Try to match partial model names (e.g., "gpt-4o-2024-08-06" -> "gpt-4o")
	for key, pricing := range defaultPricing {
		if len(model) >= len(key) && model[:len(key)] == key {
			return pricing
		}
	}

	// Return default pricing
	return defaultPricing["default"]
}

// CalculateCost calculates the cost for a given usage
func (t *Tracker) CalculateCost(model string, promptTokens, completionTokens int) float64 {
	pricing := t.GetPricing(model)

	inputCost := float64(promptTokens) * pricing.InputPricePerMillion / 1_000_000
	outputCost := float64(completionTokens) * pricing.OutputPricePerMillion / 1_000_000

	return inputCost + outputCost
}

// RecordUsage records usage for a user and updates their spending limits
func (t *Tracker) RecordUsage(ctx context.Context, userID, model string, usage *models.Usage) error {
	if t.repo == nil || usage == nil {
		return nil
	}

	cost := t.CalculateCost(model, usage.PromptTokens, usage.CompletionTokens)

	log.Debug().
		Str("user_id", userID).
		Str("model", model).
		Int("prompt_tokens", usage.PromptTokens).
		Int("completion_tokens", usage.CompletionTokens).
		Float64("cost", cost).
		Msg("Recording usage")

	// Update all spending limits for this user
	limits, err := t.repo.ListSpendingLimits(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list spending limits")
		return err
	}

	for _, limit := range limits {
		// Update limits that match this user or are global (empty user_id)
		if limit.UserID == userID || limit.UserID == "" || limit.UserID == "*" {
			limit.CurrentSpend += cost
			if err := t.repo.UpdateSpendingLimit(ctx, limit); err != nil {
				log.Warn().Err(err).Str("limit_id", limit.ID).Msg("Failed to update spending limit")
			} else {
				log.Debug().
					Str("limit_id", limit.ID).
					Float64("new_spend", limit.CurrentSpend).
					Float64("limit", limit.LimitAmount).
					Msg("Updated spending limit")

				// Check if alert threshold reached
				if limit.AlertAt > 0 {
					alertThreshold := limit.LimitAmount * (limit.AlertAt / 100)
					if limit.CurrentSpend >= alertThreshold {
						log.Warn().
							Str("limit_id", limit.ID).
							Str("user_id", limit.UserID).
							Float64("current_spend", limit.CurrentSpend).
							Float64("alert_threshold", alertThreshold).
							Msg("Spending alert threshold reached")
					}
				}
			}
		}
	}

	return nil
}

// CheckLimit checks if a user has exceeded their spending limit
func (t *Tracker) CheckLimit(ctx context.Context, userID string) (bool, float64, float64, error) {
	if t.repo == nil {
		return false, 0, 0, nil
	}

	limits, err := t.repo.ListSpendingLimits(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	for _, limit := range limits {
		if limit.UserID == userID || limit.UserID == "" || limit.UserID == "*" {
			if limit.CurrentSpend >= limit.LimitAmount {
				return true, limit.CurrentSpend, limit.LimitAmount, nil
			}
		}
	}

	return false, 0, 0, nil
}

// GetUserSpending returns the current spending for a user
func (t *Tracker) GetUserSpending(ctx context.Context, userID string) (float64, error) {
	if t.repo == nil {
		return 0, nil
	}

	limits, err := t.repo.ListSpendingLimits(ctx)
	if err != nil {
		return 0, err
	}

	var totalSpend float64
	for _, limit := range limits {
		if limit.UserID == userID || limit.UserID == "" || limit.UserID == "*" {
			totalSpend += limit.CurrentSpend
		}
	}

	return totalSpend, nil
}
