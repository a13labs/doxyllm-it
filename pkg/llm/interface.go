package llm

import (
	"context"
	"time"
)

// Provider represents a generic LLM provider interface
type Provider interface {
	// GenerateComment generates a documentation comment for the given entity
	GenerateComment(ctx context.Context, request CommentRequest) (*CommentResponse, error)

	// TestConnection verifies the LLM provider is accessible
	TestConnection(ctx context.Context) error

	// GetModelInfo returns information about the current model
	GetModelInfo() ModelInfo
}

// CommentRequest represents a request for comment generation
type CommentRequest struct {
	EntityName        string                 // Name of the entity to document
	EntityType        string                 // Type of entity (function, class, namespace, etc.)
	Context           string                 // Code context around the entity
	AdditionalContext string                 // Additional project context
	GroupInfo         *GroupInfo             // Group membership information
	Options           map[string]interface{} // Provider-specific options
}

// CommentResponse represents the response from comment generation
type CommentResponse struct {
	Description string            // Generated description text
	Metadata    map[string]string // Additional metadata from the LLM
}

// GroupInfo contains information about Doxygen groups
type GroupInfo struct {
	Name        string // Group name for @ingroup
	Title       string // Group title
	Description string // Group description
}

// ModelInfo contains information about the LLM model
type ModelInfo struct {
	Name        string
	Provider    string
	Version     string
	ContextSize int
}

// Config represents configuration for LLM providers
type Config struct {
	Provider    string                 `yaml:"provider"`    // Provider type (ollama, openai, etc.)
	URL         string                 `yaml:"url"`         // Provider URL
	Model       string                 `yaml:"model"`       // Model name
	Temperature float64                `yaml:"temperature"` // Generation temperature
	TopP        float64                `yaml:"top_p"`       // Top-p sampling
	NumCtx      int                    `yaml:"num_ctx"`     // Context window size
	Timeout     time.Duration          `yaml:"timeout"`     // Request timeout
	Options     map[string]interface{} `yaml:"options"`     // Provider-specific options
}

// Error types for better error handling
type ProviderError struct {
	Provider string
	Message  string
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return e.Provider + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Provider + ": " + e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}
