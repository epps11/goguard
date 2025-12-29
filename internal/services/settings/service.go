package settings

import (
	"context"
	"sync"

	"github.com/epps11/goguard/internal/database"
	"github.com/rs/zerolog/log"
)

// Service manages application settings with database persistence
type Service struct {
	repo  *database.Repository
	cache map[string]interface{}
	mu    sync.RWMutex
}

// LLMSettings holds LLM configuration
type LLMSettings struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	// AWS Bedrock specific
	AWSRegion string `json:"aws_region"`
}

// SecuritySettings holds security configuration
type SecuritySettings struct {
	InjectionDetectionEnabled bool `json:"injection_detection_enabled"`
	BlockOnDetection          bool `json:"block_on_detection"`
	PIIMaskingEnabled         bool `json:"pii_masking_enabled"`
	RateLimitPerMinute        int  `json:"rate_limit_per_minute"`
}

// NotificationSettings holds notification configuration
type NotificationSettings struct {
	WebhookURL      string   `json:"webhook_url"`
	EmailRecipients []string `json:"email_recipients"`
}

// NewService creates a new settings service
func NewService(repo *database.Repository) *Service {
	return &Service{
		repo:  repo,
		cache: make(map[string]interface{}),
	}
}

// GetLLMConfig implements the llm.SettingsProvider interface
// Returns provider, model, apiKey, baseURL for dynamic LLM configuration
func (s *Service) GetLLMConfig(ctx context.Context) (provider, model, apiKey, baseURL string, err error) {
	settings, err := s.GetLLMSettings(ctx)
	if err != nil {
		return "", "", "", "", err
	}
	return settings.Provider, settings.Model, settings.APIKey, settings.BaseURL, nil
}

// GetLLMSettings returns current LLM settings
func (s *Service) GetLLMSettings(ctx context.Context) (*LLMSettings, error) {
	s.mu.RLock()
	if cached, ok := s.cache["llm_settings"]; ok {
		s.mu.RUnlock()
		return cached.(*LLMSettings), nil
	}
	s.mu.RUnlock()

	settings := &LLMSettings{
		Provider:    "openai",
		Model:       "gpt-4o",
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	if s.repo != nil {
		if provider, err := s.repo.GetSetting(ctx, "llm_provider"); err == nil && provider != nil {
			if str, ok := provider.(string); ok {
				settings.Provider = str
			}
		}
		if model, err := s.repo.GetSetting(ctx, "llm_model"); err == nil && model != nil {
			if str, ok := model.(string); ok {
				settings.Model = str
			}
		}
		if apiKey, err := s.repo.GetSetting(ctx, "llm_api_key"); err == nil && apiKey != nil {
			if str, ok := apiKey.(string); ok {
				settings.APIKey = str
			}
		}
		if baseURL, err := s.repo.GetSetting(ctx, "llm_base_url"); err == nil && baseURL != nil {
			if str, ok := baseURL.(string); ok {
				settings.BaseURL = str
			}
		}
		if maxTokens, err := s.repo.GetSetting(ctx, "llm_max_tokens"); err == nil && maxTokens != nil {
			if num, ok := maxTokens.(float64); ok {
				settings.MaxTokens = int(num)
			}
		}
		if temp, err := s.repo.GetSetting(ctx, "llm_temperature"); err == nil && temp != nil {
			if num, ok := temp.(float64); ok {
				settings.Temperature = num
			}
		}
		if region, err := s.repo.GetSetting(ctx, "aws_region"); err == nil && region != nil {
			if str, ok := region.(string); ok {
				settings.AWSRegion = str
			}
		}
	}

	s.mu.Lock()
	s.cache["llm_settings"] = settings
	s.mu.Unlock()

	return settings, nil
}

// UpdateLLMSettings updates LLM settings in the database
func (s *Service) UpdateLLMSettings(ctx context.Context, settings *LLMSettings) error {
	if s.repo == nil {
		return nil
	}

	if err := s.repo.SetSetting(ctx, "llm_provider", settings.Provider); err != nil {
		return err
	}
	if err := s.repo.SetSetting(ctx, "llm_model", settings.Model); err != nil {
		return err
	}
	if settings.APIKey != "" {
		if err := s.repo.SetSetting(ctx, "llm_api_key", settings.APIKey); err != nil {
			return err
		}
	}
	if settings.BaseURL != "" {
		if err := s.repo.SetSetting(ctx, "llm_base_url", settings.BaseURL); err != nil {
			return err
		}
	}
	if err := s.repo.SetSetting(ctx, "llm_max_tokens", settings.MaxTokens); err != nil {
		return err
	}
	if err := s.repo.SetSetting(ctx, "llm_temperature", settings.Temperature); err != nil {
		return err
	}
	if settings.AWSRegion != "" {
		if err := s.repo.SetSetting(ctx, "aws_region", settings.AWSRegion); err != nil {
			return err
		}
	}

	// Invalidate cache
	s.mu.Lock()
	delete(s.cache, "llm_settings")
	s.mu.Unlock()

	log.Info().Str("provider", settings.Provider).Str("model", settings.Model).Msg("LLM settings updated")
	return nil
}

// GetSecuritySettings returns current security settings
func (s *Service) GetSecuritySettings(ctx context.Context) (*SecuritySettings, error) {
	settings := &SecuritySettings{
		InjectionDetectionEnabled: true,
		BlockOnDetection:          true,
		PIIMaskingEnabled:         true,
		RateLimitPerMinute:        100,
	}

	if s.repo != nil {
		if val, err := s.repo.GetSetting(ctx, "injection_detection_enabled"); err == nil && val != nil {
			if b, ok := val.(bool); ok {
				settings.InjectionDetectionEnabled = b
			}
		}
		if val, err := s.repo.GetSetting(ctx, "block_on_detection"); err == nil && val != nil {
			if b, ok := val.(bool); ok {
				settings.BlockOnDetection = b
			}
		}
		if val, err := s.repo.GetSetting(ctx, "pii_masking_enabled"); err == nil && val != nil {
			if b, ok := val.(bool); ok {
				settings.PIIMaskingEnabled = b
			}
		}
		if val, err := s.repo.GetSetting(ctx, "rate_limit_requests_per_minute"); err == nil && val != nil {
			if num, ok := val.(float64); ok {
				settings.RateLimitPerMinute = int(num)
			}
		}
	}

	return settings, nil
}

// UpdateSecuritySettings updates security settings
func (s *Service) UpdateSecuritySettings(ctx context.Context, settings *SecuritySettings) error {
	if s.repo == nil {
		return nil
	}

	if err := s.repo.SetSetting(ctx, "injection_detection_enabled", settings.InjectionDetectionEnabled); err != nil {
		return err
	}
	if err := s.repo.SetSetting(ctx, "block_on_detection", settings.BlockOnDetection); err != nil {
		return err
	}
	if err := s.repo.SetSetting(ctx, "pii_masking_enabled", settings.PIIMaskingEnabled); err != nil {
		return err
	}
	if err := s.repo.SetSetting(ctx, "rate_limit_requests_per_minute", settings.RateLimitPerMinute); err != nil {
		return err
	}

	log.Info().Msg("Security settings updated")
	return nil
}

// GetAllSettings returns all settings as a map
func (s *Service) GetAllSettings(ctx context.Context) (map[string]interface{}, error) {
	if s.repo == nil {
		return map[string]interface{}{
			"llm_provider":                   "openai",
			"llm_model":                      "gpt-4o",
			"injection_detection_enabled":    true,
			"pii_masking_enabled":            true,
			"rate_limit_requests_per_minute": 100,
		}, nil
	}
	return s.repo.GetAllSettings(ctx)
}

// InvalidateCache clears the settings cache
func (s *Service) InvalidateCache() {
	s.mu.Lock()
	s.cache = make(map[string]interface{})
	s.mu.Unlock()
}
