package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaProvider implements the Provider interface for Ollama
type OllamaProvider struct {
	config *Config
	client *http.Client
}

const defaultOllamaPromptTemplate = `You are a C++ documentation expert. Generate a Doxygen comment for the specific entity requested.

CRITICAL INSTRUCTIONS:
- Generate ONLY the Doxygen comment, no code is allowed
- You must include Doxygen tags (@brief, @param, @return, etc.)
- You must include comment markers (/** */) - the system will format these
- Document ONLY the target entity.
- Focus on describing the purpose, behavior, and usage
- For functions: Describe what it does, not parameters/return (those will be handled separately)
- For classes: Describe the class responsibility and main purpose
- For namespaces: Describe the purpose and scope
- For templates: Describe the template parameters and their usage
- These are the only allowed Doxygen tags, do not use any other tags @brief, @param, @tparam, @return.
%s

%s



Generate focused doxygen content for this entity (no source code).`

const fieldPromptTemplate = `You are a C++ documentation expert. Generate a very concise description for a struct/class field.

CRITICAL REQUIREMENTS:
- Generate ONLY a brief description (ONE sentence, under 60 characters)
- Do NOT include Doxygen tags, comment markers, or < symbols
- Start with "The" (e.g., "The x-coordinate of the point")
- Be specific and concise
- Examples of good responses:
  * "The width of the rectangle"
  * "The player's health points"  
  * "The file handle for reading"

%s

%s

Generate ONE concise sentence (under 60 characters, no < symbols).`

// getPromptTemplate returns the appropriate prompt template based on entity type
func (p *OllamaProvider) getPromptTemplate(entityType string) string {
	switch entityType {
	case "name":
		return fieldPromptTemplate
	default:
		// Use custom template if provided, otherwise default
		if p.config.PromptTemplate != "" {
			return p.config.PromptTemplate
		}
		return defaultOllamaPromptTemplate
	}
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider(config *Config) *OllamaProvider {
	return &OllamaProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// OllamaRequest represents the request structure for Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents the response structure from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// Generate generates a documentation comment using Ollama
func (p *OllamaProvider) Generate(ctx context.Context, request CommentRequest) (*CommentResponse, error) {
	// Build additional context section
	var contextSection string
	if request.AdditionalContext != "" {
		contextSection = fmt.Sprintf("ADDITIONAL PROJECT CONTEXT:\n%s\n", request.AdditionalContext)
	}

	// Get the appropriate prompt template based on entity type
	promptTemplate := p.getPromptTemplate(request.EntityType)

	prompt := fmt.Sprintf(
		promptTemplate,
		contextSection,
		request.Context,
	)

	// Prepare request options
	options := make(map[string]interface{})
	options["temperature"] = p.config.Temperature
	options["top_p"] = p.config.TopP
	options["num_ctx"] = p.config.NumCtx

	// Add any additional options from request
	for k, v := range request.Options {
		options[k] = v
	}

	// Create Ollama request
	ollamaReq := OllamaRequest{
		Model:   p.config.Model,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, &ProviderError{
			Provider: "ollama",
			Message:  "failed to marshal request",
			Err:      err,
		}
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &ProviderError{
			Provider: "ollama",
			Message:  "failed to create HTTP request",
			Err:      err,
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &ProviderError{
			Provider: "ollama",
			Message:  "HTTP request failed",
			Err:      err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &ProviderError{
			Provider: "ollama",
			Message:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, &ProviderError{
			Provider: "ollama",
			Message:  "failed to decode response",
			Err:      err,
		}
	}

	return &CommentResponse{
		Comment: ollamaResp.Response,
		Metadata: map[string]string{
			"model":    p.config.Model,
			"provider": "ollama",
		},
	}, nil
}

// TestConnection verifies Ollama is accessible
func (p *OllamaProvider) TestConnection(ctx context.Context) error {
	// Test with /api/tags endpoint
	tagsURL := strings.Replace(p.config.URL, "/api/generate", "/api/tags", 1)

	req, err := http.NewRequestWithContext(ctx, "GET", tagsURL, nil)
	if err != nil {
		return &ProviderError{
			Provider: "ollama",
			Message:  "failed to create test request",
			Err:      err,
		}
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return &ProviderError{
			Provider: "ollama",
			Message:  "connection test failed",
			Err:      err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ProviderError{
			Provider: "ollama",
			Message:  fmt.Sprintf("connection test returned HTTP %d", resp.StatusCode),
		}
	}

	return nil
}

// GetModelInfo returns information about the current model
func (p *OllamaProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:        p.config.Model,
		Provider:    "ollama",
		Version:     "unknown", // Ollama doesn't provide version info easily
		ContextSize: p.config.NumCtx,
	}
}
