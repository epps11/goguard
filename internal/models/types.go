package models

import "time"

// GuardRequest represents an incoming request to be processed
type GuardRequest struct {
	RequestID   string            `json:"request_id"`
	Messages    []Message         `json:"messages"`
	Provider    string            `json:"provider,omitempty"` // openai, anthropic, google, bedrock, ollama, xai
	Model       string            `json:"model,omitempty"`
	APIKey      string            `json:"api_key,omitempty"`  // Optional per-request API key
	BaseURL     string            `json:"base_url,omitempty"` // Optional custom base URL
	MaxTokens   *int              `json:"max_tokens,omitempty"`
	Temperature *float64          `json:"temperature,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// GuardResponse represents the response after processing
type GuardResponse struct {
	RequestID      string          `json:"request_id"`
	Allowed        bool            `json:"allowed"`
	ProcessedInput *ProcessedInput `json:"processed_input,omitempty"`
	LLMResponse    *LLMResponse    `json:"llm_response,omitempty"`
	SecurityReport *SecurityReport `json:"security_report,omitempty"`
	PIIReport      *PIIReport      `json:"pii_report,omitempty"`
	ProcessingTime time.Duration   `json:"processing_time_ms"`
	Error          string          `json:"error,omitempty"`
}

// ProcessedInput contains the sanitized input
type ProcessedInput struct {
	OriginalMessages []Message `json:"original_messages,omitempty"`
	MaskedMessages   []Message `json:"masked_messages"`
	PIIMasked        bool      `json:"pii_masked"`
}

// LLMResponse contains the response from the LLM provider
type LLMResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
	Usage        *Usage `json:"usage,omitempty"`
}

// Usage contains token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// SecurityReport contains injection detection results
type SecurityReport struct {
	InjectionDetected bool        `json:"injection_detected"`
	ThreatLevel       string      `json:"threat_level"` // none, low, medium, high, critical
	Detections        []Detection `json:"detections,omitempty"`
	BlockedReason     string      `json:"blocked_reason,omitempty"`
	Recommendations   []string    `json:"recommendations,omitempty"`
}

// Detection represents a single security detection
type Detection struct {
	Type        string  `json:"type"` // prompt_injection, jailbreak, data_exfil, etc.
	Pattern     string  `json:"pattern"`
	Location    string  `json:"location"`   // which message/field
	Confidence  float64 `json:"confidence"` // 0.0 to 1.0
	Description string  `json:"description"`
}

// PIIReport contains PII detection and masking results
type PIIReport struct {
	PIIDetected bool       `json:"pii_detected"`
	PIICount    int        `json:"pii_count"`
	PIITypes    []PIIMatch `json:"pii_types,omitempty"`
	MaskedCount int        `json:"masked_count"`
}

// PIIMatch represents a detected PII instance
type PIIMatch struct {
	Type          string `json:"type"`                     // email, phone, ssn, etc.
	OriginalValue string `json:"original_value,omitempty"` // only in debug mode
	MaskedValue   string `json:"masked_value"`
	Location      string `json:"location"`
	StartPosition int    `json:"start_position"`
	EndPosition   int    `json:"end_position"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Uptime   string            `json:"uptime"`
	Services map[string]string `json:"services"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
	Details   string `json:"details,omitempty"`
}
