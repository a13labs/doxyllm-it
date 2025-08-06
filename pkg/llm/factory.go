package llm

import (
	"fmt"
	"strings"
)

// NewProvider creates a new LLM provider based on the configuration
func NewProvider(config *Config) (Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	provider := strings.ToLower(config.Provider)

	switch provider {
	case "ollama", "":
		// Default to Ollama if not specified
		return NewOllamaProvider(config), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// DefaultConfig returns a default configuration for Ollama
func DefaultConfig() *Config {
	return &Config{
		Provider:    "ollama",
		URL:         "http://localhost:11434/api/generate",
		Model:       "deepseek-coder:6.7b",
		Temperature: 0.1,
		TopP:        0.9,
		NumCtx:      4096,
		Timeout:     120000000000, // 120 seconds in nanoseconds
	}
}
