package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOllamaProvider_GenerateComment(t *testing.T) {
	tests := []struct {
		name         string
		request      CommentRequest
		responseBody string
		expectedDesc string
		expectError  bool
		httpStatus   int
	}{
		{
			name: "successful comment generation",
			request: CommentRequest{
				EntityName: "testFunction",
				EntityType: "function",
				Context:    "void testFunction(int param);",
			},
			responseBody: `{"response": "A simple test function that performs basic operations.", "done": true}`,
			expectedDesc: "A simple test function that performs basic operations.",
			expectError:  false,
			httpStatus:   http.StatusOK,
		},
		{
			name: "response with code blocks",
			request: CommentRequest{
				EntityName: "TestClass",
				EntityType: "class",
				Context:    "class TestClass {};",
			},
			responseBody: `{"response": "` + "```cpp\\nA test class for validation.\\n```" + `", "done": true}`,
			expectedDesc: "A test class for validation.",
			expectError:  false,
			httpStatus:   http.StatusOK,
		},
		{
			name: "HTTP error response",
			request: CommentRequest{
				EntityName: "testFunction",
				EntityType: "function",
				Context:    "void testFunction();",
			},
			responseBody: `{"error": "Model not found"}`,
			expectError:  true,
			httpStatus:   http.StatusNotFound,
		},
		{
			name: "response with unwanted prefix",
			request: CommentRequest{
				EntityName: "TestClass",
				EntityType: "class",
				Context:    "class TestClass {};",
			},
			responseBody: `{"response": "This is a C++ documentation expert. A comprehensive test class.", "done": true}`,
			expectedDesc: "A comprehensive test class.",
			expectError:  false,
			httpStatus:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create provider with test config
			config := &Config{
				Provider:    "ollama",
				URL:         server.URL,
				Model:       "test-model",
				Temperature: 0.1,
				TopP:        0.9,
				NumCtx:      2048,
				Timeout:     5 * time.Second,
			}

			provider := NewOllamaProvider(config)
			ctx := context.Background()

			// Generate comment
			response, err := provider.GenerateComment(ctx, tt.request)

			// Check results
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if response.Description != tt.expectedDesc {
				t.Errorf("expected description %q, got %q", tt.expectedDesc, response.Description)
			}

			// Check metadata
			if response.Metadata["provider"] != "ollama" {
				t.Errorf("expected provider metadata to be 'ollama', got %q", response.Metadata["provider"])
			}

			if response.Metadata["model"] != "test-model" {
				t.Errorf("expected model metadata to be 'test-model', got %q", response.Metadata["model"])
			}
		})
	}
}

func TestOllamaProvider_TestConnection(t *testing.T) {
	tests := []struct {
		name        string
		httpStatus  int
		expectError bool
	}{
		{
			name:        "successful connection",
			httpStatus:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "connection failed",
			httpStatus:  http.StatusInternalServerError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/tags" {
					w.WriteHeader(tt.httpStatus)
					w.Write([]byte(`{"models": []}`))
				}
			}))
			defer server.Close()

			// Create provider with test config
			config := &Config{
				Provider: "ollama",
				URL:      server.URL + "/api/generate", // Will be replaced with /api/tags
				Timeout:  5 * time.Second,
			}

			provider := NewOllamaProvider(config)
			ctx := context.Background()

			// Test connection
			err := provider.TestConnection(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestOllamaProvider_GetModelInfo(t *testing.T) {
	config := &Config{
		Provider: "ollama",
		Model:    "test-model:7b",
		NumCtx:   4096,
	}

	provider := NewOllamaProvider(config)
	info := provider.GetModelInfo()

	if info.Name != "test-model:7b" {
		t.Errorf("expected model name 'test-model:7b', got %q", info.Name)
	}

	if info.Provider != "ollama" {
		t.Errorf("expected provider 'ollama', got %q", info.Provider)
	}

	if info.ContextSize != 4096 {
		t.Errorf("expected context size 4096, got %d", info.ContextSize)
	}
}

func TestOllamaProvider_CleanResponse(t *testing.T) {
	provider := &OllamaProvider{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean text",
			input:    "A simple function description.",
			expected: "A simple function description.",
		},
		{
			name:     "text with code blocks",
			input:    "```cpp\nA function with code blocks.\n```",
			expected: "A function with code blocks.",
		},
		{
			name:     "text with unwanted prefix",
			input:    "This is a C++ documentation expert. Here is the description.",
			expected: "Here is the description.",
		},
		{
			name:     "text with brief description prefix",
			input:    "Brief Description: A concise description of the function.",
			expected: "A concise description of the function.",
		},
		{
			name:     "multiple code block types",
			input:    "```c++\nDetailed function description.\n```",
			expected: "Detailed function description.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.cleanResponse(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		expectError  bool
		expectedType string
	}{
		{
			name: "valid ollama config",
			config: &Config{
				Provider: "ollama",
				Model:    "test-model",
			},
			expectError:  false,
			expectedType: "*llm.OllamaProvider",
		},
		{
			name: "empty provider defaults to ollama",
			config: &Config{
				Model: "test-model",
			},
			expectError:  false,
			expectedType: "*llm.OllamaProvider",
		},
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "unsupported provider",
			config: &Config{
				Provider: "unsupported",
				Model:    "test-model",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Errorf("expected provider but got nil")
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Provider != "ollama" {
		t.Errorf("expected provider 'ollama', got %q", config.Provider)
	}

	if config.Model != "deepseek-coder:6.7b" {
		t.Errorf("expected model 'deepseek-coder:6.7b', got %q", config.Model)
	}

	if config.Temperature != 0.1 {
		t.Errorf("expected temperature 0.1, got %f", config.Temperature)
	}

	if config.NumCtx != 4096 {
		t.Errorf("expected context size 4096, got %d", config.NumCtx)
	}
}
