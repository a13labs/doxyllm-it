package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestShouldUseInlineStyle tests the inline style detection logic
func TestShouldUseInlineStyle(t *testing.T) {
	service := &DocumentationService{}

	tests := []struct {
		name        string
		entityType  string
		description string
		expected    bool
	}{
		{
			name:        "Field with short description",
			entityType:  "field",
			description: "The x-coordinate.",
			expected:    true,
		},
		{
			name:        "Field with description starting with <",
			entityType:  "field",
			description: "< The x-coordinate.",
			expected:    true,
		},
		{
			name:        "Field with long description",
			entityType:  "field",
			description: "This is a very long description that exceeds the character limit for inline comments and should be formatted as a block comment instead.",
			expected:    false,
		},
		{
			name:        "Field with multiple sentences",
			entityType:  "field",
			description: "The x-coordinate. It represents horizontal position.",
			expected:    false,
		},
		{
			name:        "Non-field entity",
			entityType:  "class",
			description: "Short description.",
			expected:    false,
		},
		{
			name:        "Member entity with short description",
			entityType:  "member",
			description: "The value.",
			expected:    true,
		},
		{
			name:        "Field with multiline description",
			entityType:  "field",
			description: "First line.\nSecond line.",
			expected:    false,
		},
		{
			name:        "Field with exclamation",
			entityType:  "field",
			description: "Important! Use carefully.",
			expected:    false,
		},
		{
			name:        "Field with question",
			entityType:  "field",
			description: "Is this correct? Maybe.",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldUseInlineStyle(tt.entityType, tt.description)
			if result != tt.expected {
				t.Errorf("shouldUseInlineStyle(%q, %q) = %v, want %v",
					tt.entityType, tt.description, result, tt.expected)
			}
		})
	}
}

// TestGenerateDocumentationWithInlineStyle tests that inline style is applied correctly
func TestGenerateDocumentationWithInlineStyle(t *testing.T) {
	// Mock provider that returns short field descriptions
	mockProvider := &MockProvider{
		generateCommentFunc: func(ctx context.Context, request CommentRequest) (*CommentResponse, error) {
			if request.EntityType == "field" {
				return &CommentResponse{
					Comment:  "The coordinate value.",
					Metadata: map[string]string{"provider": "mock"},
				}, nil
			}
			return &CommentResponse{
				Comment:  "A longer description for non-field entities that should use block comment style.",
				Metadata: map[string]string{"provider": "mock"},
			}, nil
		},
	}

	service := NewDocumentationService(mockProvider)

	tests := []struct {
		name                string
		entityType          string
		expectInlineKeyword string
	}{
		{
			name:                "Field should use inline style",
			entityType:          "field",
			expectInlineKeyword: "The coordinate value",
		},
		{
			name:                "Class should use block style",
			entityType:          "class",
			expectInlineKeyword: "A longer description for", // Block style uses @brief
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := DocumentationRequest{
				EntityName:        "testEntity",
				EntityType:        tt.entityType,
				Context:           "int testEntity;",
				AdditionalContext: "",
			}

			result, err := service.Generate(context.Background(), req)
			if err != nil {
				t.Fatalf("GenerateDocumentation failed: %v", err)
			}

			if !strings.Contains(result.Response, tt.expectInlineKeyword) {
				t.Errorf("Expected comment to contain %q, got: %s",
					tt.expectInlineKeyword, result.Response)
			}
		})
	}
}

// MockProvider implements Provider interface for testing
type MockProvider struct {
	generateCommentFunc func(ctx context.Context, request CommentRequest) (*CommentResponse, error)
	testConnectionFunc  func(ctx context.Context) error
	getModelInfoFunc    func() ModelInfo
}

func (m *MockProvider) Generate(ctx context.Context, request CommentRequest) (*CommentResponse, error) {
	if m.generateCommentFunc != nil {
		return m.generateCommentFunc(ctx, request)
	}
	return &CommentResponse{
		Comment:  "Mock description for " + request.EntityName,
		Metadata: map[string]string{"provider": "mock"},
	}, nil
}

func (m *MockProvider) TestConnection(ctx context.Context) error {
	if m.testConnectionFunc != nil {
		return m.testConnectionFunc(ctx)
	}
	return nil
}

func (m *MockProvider) GetModelInfo() ModelInfo {
	if m.getModelInfoFunc != nil {
		return m.getModelInfoFunc()
	}
	return ModelInfo{
		Name:        "mock-model",
		Provider:    "mock",
		Version:     "1.0.0",
		ContextSize: 2048,
	}
}

func TestDocumentationService_GenerateDocumentation(t *testing.T) {
	tests := []struct {
		name         string
		request      DocumentationRequest
		mockResponse *CommentResponse
		mockError    error
		expectError  bool
		checkFunc    func(t *testing.T, result *DocumentationResult)
	}{
		{
			name: "successful generation",
			request: DocumentationRequest{
				EntityName: "testFunction",
				EntityType: "function",
				Context:    "void testFunction(int param);",
			},
			mockResponse: &CommentResponse{
				Comment:  "A test function that does something useful.",
				Metadata: map[string]string{"provider": "mock"},
			},
			expectError: false,
			checkFunc: func(t *testing.T, result *DocumentationResult) {
				if !strings.Contains(result.Response, "A test function") {
					t.Errorf("comment should contain brief description")
				}
				if result.Response != "A test function that does something useful." {
					t.Errorf("description should match mock response")
				}
				if result.Metadata["provider"] != "mock" {
					t.Errorf("metadata should be preserved")
				}
			},
		},
		{
			name: "LLM provider error",
			request: DocumentationRequest{
				EntityName: "testFunction",
				EntityType: "function",
				Context:    "void testFunction();",
			},
			mockError:   errors.New("LLM service unavailable"),
			expectError: true,
		},
		{
			name: "invalid request - empty entity name",
			request: DocumentationRequest{
				EntityName: "",
				EntityType: "function",
				Context:    "void testFunction();",
			},
			expectError: true,
		},
		{
			name: "invalid request - empty entity type",
			request: DocumentationRequest{
				EntityName: "testFunction",
				EntityType: "",
				Context:    "void testFunction();",
			},
			expectError: true,
		},
		{
			name: "invalid request - empty context",
			request: DocumentationRequest{
				EntityName: "testFunction",
				EntityType: "function",
				Context:    "",
			},
			expectError: true,
		},
		{
			name: "with additional context",
			request: DocumentationRequest{
				EntityName:        "TestClass",
				EntityType:        "class",
				Context:           "class TestClass {};",
				AdditionalContext: "This is a test class for validation.",
			},
			mockResponse: &CommentResponse{
				Comment:  "A test class for validation.",
				Metadata: map[string]string{"provider": "mock"},
			},
			expectError: false,
			checkFunc: func(t *testing.T, result *DocumentationResult) {
				if !strings.Contains(result.Response, "A test class") {
					t.Errorf("comment should contain brief description")
				}
				// Note: @ingroup is now handled by post-processor in cmd layer, not here
			},
		},
		{
			name: "with additional project context",
			request: DocumentationRequest{
				EntityName:        "helper",
				EntityType:        "function",
				Context:           "int helper();",
				AdditionalContext: "This is a utility function for testing.",
			},
			mockResponse: &CommentResponse{
				Comment:  "A helper function with context.",
				Metadata: map[string]string{"provider": "mock"},
			},
			expectError: false,
			checkFunc: func(t *testing.T, result *DocumentationResult) {
				if !strings.Contains(result.Response, "A helper function") {
					t.Errorf("comment should contain description")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock provider
			mockProvider := &MockProvider{
				generateCommentFunc: func(ctx context.Context, request CommentRequest) (*CommentResponse, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			// Create service
			service := NewDocumentationService(mockProvider)
			ctx := context.Background()

			// Generate documentation
			result, err := service.Generate(ctx, tt.request)

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

			if result == nil {
				t.Errorf("expected result but got nil")
				return
			}

			// Run specific test checks
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestDocumentationService_TestConnection(t *testing.T) {
	tests := []struct {
		name        string
		mockError   error
		expectError bool
	}{
		{
			name:        "successful connection",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "connection failed",
			mockError:   errors.New("connection failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock provider
			mockProvider := &MockProvider{
				testConnectionFunc: func(ctx context.Context) error {
					return tt.mockError
				},
			}

			// Create service
			service := NewDocumentationService(mockProvider)
			ctx := context.Background()

			// Test connection
			err := service.TestConnection(ctx)

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

func TestDocumentationService_GetModelInfo(t *testing.T) {
	expectedInfo := ModelInfo{
		Name:        "test-model",
		Provider:    "test-provider",
		Version:     "2.0.0",
		ContextSize: 8192,
	}

	// Create mock provider
	mockProvider := &MockProvider{
		getModelInfoFunc: func() ModelInfo {
			return expectedInfo
		},
	}

	// Create service
	service := NewDocumentationService(mockProvider)

	// Get model info
	info := service.GetModelInfo()

	// Check results
	if info.Name != expectedInfo.Name {
		t.Errorf("expected name %q, got %q", expectedInfo.Name, info.Name)
	}
	if info.Provider != expectedInfo.Provider {
		t.Errorf("expected provider %q, got %q", expectedInfo.Provider, info.Provider)
	}
	if info.Version != expectedInfo.Version {
		t.Errorf("expected version %q, got %q", expectedInfo.Version, info.Version)
	}
	if info.ContextSize != expectedInfo.ContextSize {
		t.Errorf("expected context size %d, got %d", expectedInfo.ContextSize, info.ContextSize)
	}
}

func TestDocumentationService_ValidateRequest(t *testing.T) {
	service := &DocumentationService{}

	tests := []struct {
		name        string
		request     DocumentationRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: DocumentationRequest{
				EntityName: "testFunction",
				EntityType: "function",
				Context:    "void testFunction();",
			},
			expectError: false,
		},
		{
			name: "empty entity name",
			request: DocumentationRequest{
				EntityName: "",
				EntityType: "function",
				Context:    "void testFunction();",
			},
			expectError: true,
			errorMsg:    "entity name cannot be empty",
		},
		{
			name: "empty entity type",
			request: DocumentationRequest{
				EntityName: "testFunction",
				EntityType: "",
				Context:    "void testFunction();",
			},
			expectError: true,
			errorMsg:    "entity type cannot be empty",
		},
		{
			name: "empty context",
			request: DocumentationRequest{
				EntityName: "testFunction",
				EntityType: "function",
				Context:    "",
			},
			expectError: true,
			errorMsg:    "context cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRequest(tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
