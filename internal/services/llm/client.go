package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/agentplexus/omnillm"
	"github.com/epps11/goguard/internal/config"
	"github.com/epps11/goguard/internal/models"
)

// Ensure context is used (for settings provider)
var _ = context.Background

// Client wraps the OmniLLM client for LLM interactions
type Client struct {
	client      *omnillm.ChatClient
	config      config.LLMConfig
	initialized bool
}

// NewClient creates a new LLM client
func NewClient(cfg config.LLMConfig) (*Client, error) {
	providerName, err := mapProviderName(cfg.Provider)
	if err != nil {
		return nil, err
	}

	clientConfig := omnillm.ClientConfig{
		Provider: providerName,
		APIKey:   cfg.APIKey,
	}

	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}

	client, err := omnillm.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return &Client{
		client:      client,
		config:      cfg,
		initialized: true,
	}, nil
}

// Chat sends a chat completion request
func (c *Client) Chat(ctx context.Context, messages []models.Message) (*models.LLMResponse, error) {
	if !c.initialized {
		return nil, errors.New("LLM client not initialized")
	}

	// Convert messages to OmniLLM format
	omnillmMessages := make([]omnillm.Message, len(messages))
	for i, msg := range messages {
		omnillmMessages[i] = omnillm.Message{
			Role:    mapRole(msg.Role),
			Content: msg.Content,
		}
	}

	// Build request
	req := &omnillm.ChatCompletionRequest{
		Model:    c.config.Model,
		Messages: omnillmMessages,
	}

	if c.config.MaxTokens > 0 {
		req.MaxTokens = &c.config.MaxTokens
	}

	if c.config.Temperature > 0 {
		req.Temperature = &c.config.Temperature
	}

	// Make request
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM request failed: %w", err)
	}

	// Convert response
	llmResp := &models.LLMResponse{
		Model: resp.Model,
	}

	if len(resp.Choices) > 0 {
		llmResp.Content = resp.Choices[0].Message.Content
		if resp.Choices[0].FinishReason != nil {
			llmResp.FinishReason = *resp.Choices[0].FinishReason
		}
	}

	if resp.Usage.TotalTokens > 0 {
		llmResp.Usage = &models.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		}
	}

	return llmResp, nil
}

// ChatStream sends a streaming chat completion request
func (c *Client) ChatStream(ctx context.Context, messages []models.Message, handler func(chunk string) error) (*models.LLMResponse, error) {
	if !c.initialized {
		return nil, errors.New("LLM client not initialized")
	}

	// Convert messages to OmniLLM format
	omnillmMessages := make([]omnillm.Message, len(messages))
	for i, msg := range messages {
		omnillmMessages[i] = omnillm.Message{
			Role:    mapRole(msg.Role),
			Content: msg.Content,
		}
	}

	// Build request
	req := &omnillm.ChatCompletionRequest{
		Model:    c.config.Model,
		Messages: omnillmMessages,
	}

	if c.config.MaxTokens > 0 {
		req.MaxTokens = &c.config.MaxTokens
	}

	if c.config.Temperature > 0 {
		req.Temperature = &c.config.Temperature
	}

	// Create stream
	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	var fullContent string
	var finishReason string

	// Process stream
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				fullContent += content
				if err := handler(content); err != nil {
					return nil, err
				}
			}
			if chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != "" {
				finishReason = *chunk.Choices[0].FinishReason
			}
		}
	}

	return &models.LLMResponse{
		Content:      fullContent,
		Model:        c.config.Model,
		FinishReason: finishReason,
	}, nil
}

// Close closes the LLM client
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// IsInitialized returns whether the client is ready
func (c *Client) IsInitialized() bool {
	return c.initialized
}

// mapProviderName maps config provider name to OmniLLM provider
func mapProviderName(provider string) (omnillm.ProviderName, error) {
	switch provider {
	case "openai":
		return omnillm.ProviderNameOpenAI, nil
	case "anthropic", "claude":
		return omnillm.ProviderNameAnthropic, nil
	case "gemini", "google":
		return omnillm.ProviderNameGemini, nil
	case "ollama":
		return omnillm.ProviderNameOllama, nil
	case "xai", "grok":
		return omnillm.ProviderNameXAI, nil
	case "bedrock", "aws":
		// Note: OmniLLM may not support Bedrock directly
		// For now, map to OpenAI-compatible endpoint (user must provide base_url)
		return omnillm.ProviderNameOpenAI, nil
	default:
		return "", fmt.Errorf("unsupported provider: %s (supported: openai, anthropic, google, ollama, xai, bedrock)", provider)
	}
}

// SettingsProvider interface for fetching dynamic LLM settings
type SettingsProvider interface {
	GetLLMConfig(ctx context.Context) (provider, model, apiKey, baseURL string, err error)
}

// ClientFactory creates LLM clients dynamically based on request parameters
type ClientFactory struct {
	defaultConfig    config.LLMConfig
	defaultClient    *Client
	settingsProvider SettingsProvider
}

// NewClientFactory creates a new client factory with default configuration
func NewClientFactory(cfg config.LLMConfig) (*ClientFactory, error) {
	var defaultClient *Client
	var err error

	// Create default client if API key is provided
	if cfg.APIKey != "" {
		defaultClient, err = NewClient(cfg)
		if err != nil {
			return nil, err
		}
	}

	return &ClientFactory{
		defaultConfig: cfg,
		defaultClient: defaultClient,
	}, nil
}

// NewClientFactoryWithSettings creates a factory that can fetch settings dynamically
func NewClientFactoryWithSettings(cfg config.LLMConfig, provider SettingsProvider) (*ClientFactory, error) {
	factory, err := NewClientFactory(cfg)
	if err != nil {
		return nil, err
	}
	factory.settingsProvider = provider
	return factory, nil
}

// SetSettingsProvider sets the settings provider for dynamic configuration
func (f *ClientFactory) SetSettingsProvider(provider SettingsProvider) {
	f.settingsProvider = provider
}

// GetClient returns an LLM client based on request parameters
// If request specifies provider/model/apikey, creates a new client
// Otherwise checks settings provider, then falls back to default client
func (f *ClientFactory) GetClient(req *models.GuardRequest) (*Client, bool, error) {
	// If no override specified in request, check settings provider
	if req.Provider == "" && req.APIKey == "" && req.BaseURL == "" {
		// Try to get dynamic settings from database
		if f.settingsProvider != nil {
			ctx := context.Background()
			provider, model, apiKey, baseURL, err := f.settingsProvider.GetLLMConfig(ctx)
			if err == nil && apiKey != "" {
				// Use settings from database
				cfg := config.LLMConfig{
					Provider:    provider,
					Model:       model,
					APIKey:      apiKey,
					BaseURL:     baseURL,
					MaxTokens:   f.defaultConfig.MaxTokens,
					Temperature: f.defaultConfig.Temperature,
				}
				client, err := NewClient(cfg)
				if err != nil {
					return nil, false, fmt.Errorf("failed to create client from settings: %w", err)
				}
				return client, true, nil // true = close after use
			}
		}

		// Fall back to default client
		if f.defaultClient == nil {
			return nil, false, errors.New("no LLM client configured and no provider specified in request")
		}
		return f.defaultClient, false, nil // false = don't close after use
	}

	// Build config from request, falling back to defaults
	cfg := config.LLMConfig{
		Provider:    req.Provider,
		APIKey:      req.APIKey,
		BaseURL:     req.BaseURL,
		Model:       req.Model,
		MaxTokens:   f.defaultConfig.MaxTokens,
		Temperature: f.defaultConfig.Temperature,
	}

	// Use defaults if not specified in request
	if cfg.Provider == "" {
		cfg.Provider = f.defaultConfig.Provider
	}
	if cfg.APIKey == "" {
		cfg.APIKey = f.defaultConfig.APIKey
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = f.defaultConfig.BaseURL
	}
	if cfg.Model == "" {
		cfg.Model = f.defaultConfig.Model
	}
	if req.MaxTokens != nil {
		cfg.MaxTokens = *req.MaxTokens
	}
	if req.Temperature != nil {
		cfg.Temperature = *req.Temperature
	}

	// Create new client for this request
	client, err := NewClient(cfg)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create LLM client for request: %w", err)
	}

	return client, true, nil // true = close after use
}

// GetDefaultClient returns the default client
func (f *ClientFactory) GetDefaultClient() *Client {
	return f.defaultClient
}

// Close closes the default client
func (f *ClientFactory) Close() error {
	if f.defaultClient != nil {
		return f.defaultClient.Close()
	}
	return nil
}

// mapRole maps message role to OmniLLM role
func mapRole(role string) omnillm.Role {
	switch role {
	case "system":
		return omnillm.RoleSystem
	case "user":
		return omnillm.RoleUser
	case "assistant":
		return omnillm.RoleAssistant
	default:
		return omnillm.RoleUser
	}
}
