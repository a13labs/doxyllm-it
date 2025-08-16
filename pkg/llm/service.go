package llm

import (
	"context"
	"fmt"
	"strings"
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
	}

	// Generate comment using LLM
	response, err := s.provider.GenerateDescription(ctx, llmRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to generate comment: %w", err)
	}

	return &DocumentationResult{
		Description: response.Description,
		Metadata:    response.Metadata,
	}, nil
}

// shouldUseInlineStyle determines if inline comment style should be used
func (s *DocumentationService) shouldUseInlineStyle(entityType, description string) bool {
	// Use inline style for struct/class fields (members) with simple descriptions
	if entityType != "field" && entityType != "member" {
		return false
	}

	// For debugging: temporarily make this very permissive
	firstLine := strings.TrimSpace(description)
	// Remove any leading < character that might have been added by LLM
	firstLine = strings.TrimPrefix(firstLine, "<")
	firstLine = strings.TrimSpace(firstLine)

	// Check if description is simple enough for inline style
	lines := strings.Split(firstLine, "\n")
	if len(lines) > 1 {
		return false // Multi-line descriptions should use block style
	}

	// Be more strict for field descriptions - allow up to 80 characters for single sentence
	isSimpleSentence := !strings.Contains(firstLine, ". ") &&
		!strings.Contains(firstLine, "! ") &&
		!strings.Contains(firstLine, "? ")

	return len(firstLine) < 80 && isSimpleSentence && firstLine != ""
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

// DocumentationResult represents the result of documentation generation
type DocumentationResult struct {
	Description string            // Raw description from LLM
	Metadata    map[string]string // Additional metadata
}
