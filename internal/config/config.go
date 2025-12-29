package config

import (
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	LLM      LLMConfig      `yaml:"llm"`
	Security SecurityConfig `yaml:"security"`
	PII      PIIConfig      `yaml:"pii"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	Mode         string        `yaml:"mode"` // debug, release, test
}

type LLMConfig struct {
	Provider    string  `yaml:"provider"` // openai, anthropic, gemini, ollama, etc.
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

type SecurityConfig struct {
	EnableInjectionDetection bool     `yaml:"enable_injection_detection"`
	BlockOnDetection         bool     `yaml:"block_on_detection"`
	InjectionPatterns        []string `yaml:"injection_patterns"`
	MaxPromptLength          int      `yaml:"max_prompt_length"`
	RateLimitPerMinute       int      `yaml:"rate_limit_per_minute"`
}

type PIIConfig struct {
	EnableMasking  bool     `yaml:"enable_masking"`
	MaskCharacter  string   `yaml:"mask_character"`
	PIITypes       []string `yaml:"pii_types"`       // email, phone, ssn, credit_card, etc.
	PreserveDomain bool     `yaml:"preserve_domain"` // for emails, keep domain visible
}

type LoggingConfig struct {
	Level      string `yaml:"level"`  // debug, info, warn, error
	Format     string `yaml:"format"` // json, console
	OutputPath string `yaml:"output_path"`
}

func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	cfg.loadFromEnv()
	return cfg, nil
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			Mode:         "release",
		},
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			MaxTokens:   4096,
			Temperature: 0.7,
		},
		Security: SecurityConfig{
			EnableInjectionDetection: true,
			BlockOnDetection:         true,
			MaxPromptLength:          32000,
			RateLimitPerMinute:       60,
		},
		PII: PIIConfig{
			EnableMasking:  true,
			MaskCharacter:  "*",
			PIITypes:       []string{"email", "phone", "ssn", "credit_card", "ip_address"},
			PreserveDomain: false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

func (c *Config) loadFromEnv() {
	if v := os.Getenv("GOGUARD_HOST"); v != "" {
		c.Server.Host = v
	}
	if v := os.Getenv("GOGUARD_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Server.Port = port
		}
	}
	if v := os.Getenv("GOGUARD_MODE"); v != "" {
		c.Server.Mode = v
	}
	if v := os.Getenv("GOGUARD_LLM_PROVIDER"); v != "" {
		c.LLM.Provider = v
	}
	if v := os.Getenv("GOGUARD_LLM_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("GOGUARD_LLM_BASE_URL"); v != "" {
		c.LLM.BaseURL = v
	}
	if v := os.Getenv("GOGUARD_LLM_MODEL"); v != "" {
		c.LLM.Model = v
	}
	if v := os.Getenv("GOGUARD_LOG_LEVEL"); v != "" {
		c.Logging.Level = v
	}
}
