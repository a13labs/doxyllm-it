package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// MockProvider implements Provider interface for testing
type MockProvider struct {
	generateCommentFunc func(ctx context.Context, request CommentRequest) (*CommentResponse, error)
	testConnectionFunc  func(ctx context.Context) error
	getModelInfoFunc    func() ModelInfo
}

func (m *MockProvider) GenerateComment(ctx context.Context, request CommentRequest) (*CommentResponse, error) {
	if m.generateCommentFunc != nil {
		return m.generateCommentFunc(ctx, request)
	}
	return &CommentResponse{
		Description: "Mock description for " + request.EntityName,
		Metadata:    map[string]string{"provider": "mock"},
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
				Description: "A test function that does something useful.",
				Metadata:    map[string]string{"provider": "mock"},
			},
			expectError: false,
			checkFunc: func(t *testing.T, result *DocumentationResult) {
				if !strings.Contains(result.Comment, "@brief A test function") {
					t.Errorf("comment should contain brief description")
				}
				if result.Description != "A test function that does something useful." {
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
				Description: "A test class for validation.",
				Metadata:    map[string]string{"provider": "mock"},
			},
			expectError: false,
			checkFunc: func(t *testing.T, result *DocumentationResult) {
				if !strings.Contains(result.Comment, "@brief A test class") {
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
				Description: "A helper function with context.",
				Metadata:    map[string]string{"provider": "mock"},
			},
			expectError: false,
			checkFunc: func(t *testing.T, result *DocumentationResult) {
				if !strings.Contains(result.Comment, "@brief A helper function") {
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
			result, err := service.GenerateDocumentation(ctx, tt.request)

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
