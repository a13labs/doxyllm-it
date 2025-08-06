package llm

import (
	"context"
	"fmt"
)

// DocumentationService provides high-level documentation generation functionality
type DocumentationService struct {
	provider Provider
	builder  *CommentBuilder
}

// NewDocumentationService creates a new documentation service
func NewDocumentationService(provider Provider) *DocumentationService {
	return &DocumentationService{
		provider: provider,
		builder:  NewCommentBuilder(),
	}
}

// GenerateDocumentation generates a complete Doxygen comment for an entity
func (s *DocumentationService) GenerateDocumentation(ctx context.Context, req DocumentationRequest) (*DocumentationResult, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create LLM request
	llmRequest := CommentRequest{
		EntityName:        req.EntityName,
		EntityType:        req.EntityType,
		Context:           req.Context,
		AdditionalContext: req.AdditionalContext,
		GroupInfo:         req.GroupInfo,
		Options:           req.Options,
	}

	// Generate comment using LLM
	response, err := s.provider.GenerateComment(ctx, llmRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate comment: %w", err)
	}

	// Build structured comment
	structuredComment := s.builder.BuildStructuredComment(
		response,
		req.EntityName,
		req.EntityType,
		req.GroupInfo,
		req.Context,
	)

	return &DocumentationResult{
		Comment:     structuredComment,
		Description: response.Description,
		Metadata:    response.Metadata,
	}, nil
}

// TestConnection tests the connection to the LLM provider
func (s *DocumentationService) TestConnection(ctx context.Context) error {
	return s.provider.TestConnection(ctx)
}

// GetModelInfo returns information about the current model
func (s *DocumentationService) GetModelInfo() ModelInfo {
	return s.provider.GetModelInfo()
}

// validateRequest validates the documentation request
func (s *DocumentationService) validateRequest(req DocumentationRequest) error {
	if req.EntityName == "" {
		return fmt.Errorf("entity name cannot be empty")
	}
	if req.EntityType == "" {
		return fmt.Errorf("entity type cannot be empty")
	}
	if req.Context == "" {
		return fmt.Errorf("context cannot be empty")
	}
	return nil
}

// DocumentationRequest represents a request for documentation generation
type DocumentationRequest struct {
	EntityName        string                 // Name of the entity to document
	EntityType        string                 // Type of entity (function, class, namespace, etc.)
	Context           string                 // Code context around the entity
	AdditionalContext string                 // Additional project context
	GroupInfo         *GroupInfo             // Group membership information
	Options           map[string]interface{} // Provider-specific options
}

// DocumentationResult represents the result of documentation generation
type DocumentationResult struct {
	Comment     string            // Complete structured Doxygen comment
	Description string            // Raw description from LLM
	Metadata    map[string]string // Additional metadata
}
