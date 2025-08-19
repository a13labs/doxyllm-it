// Package document provides a high-level service for processing documentation
// requests using the document abstraction
package document

import (
	"context"
	"fmt"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/doxygen"
	"doxyllm-it/pkg/llm"
)

// DocumentationService provides high-level document processing with LLM integration
type DocumentationService struct {
	llmService llm.Provider
}

// NewDocumentationService creates a new document documentation service
func NewDocumentationService(llmService llm.Provider) *DocumentationService {
	return &DocumentationService{
		llmService: llmService,
	}
}

// ProcessingOptions contains options for document processing
type ProcessingOptions struct {
	MaxEntities       int              // Maximum entities to process (0 = unlimited)
	DryRun            bool             // Don't make actual changes
	BackupFiles       bool             // Create backup files
	FormatOutput      bool             // Apply clang-format after processing
	ExcludeTypes      []ast.EntityType // Entity types to exclude
	AdditionalContext string           // Additional context for LLM generation
}

// GroupConfig defines configuration for Doxygen groups

// ProcessingResult contains the result of document processing
type ProcessingResult struct {
	EntitiesProcessed int      // Number of entities processed
	EntitiesUpdated   int      // Number of entities actually updated
	UpdatedEntities   []string // List of updated entity paths
	Errors            []error  // Non-fatal errors encountered
}

// ProcessUndocumentedEntities processes all undocumented entities in a document
func (s *DocumentationService) ProcessUndocumentedEntities(ctx context.Context, doc *doxygen.DocLayer, opts ProcessingOptions) (*ProcessingResult, error) {
	// Get undocumented entities
	undocumented := doc.GetUndocumentedEntities()
	return s.processEntities(ctx, doc, opts, undocumented)
}

func (s *DocumentationService) ProcessAllEntities(ctx context.Context, doc *doxygen.DocLayer, opts ProcessingOptions) (*ProcessingResult, error) {
	// Get all entities
	allEntities := doc.GetAllEntities()
	return s.processEntities(ctx, doc, opts, allEntities)
}

func (s *DocumentationService) processEntities(ctx context.Context, doc *doxygen.DocLayer, opts ProcessingOptions, entities []*doxygen.Entity) (*ProcessingResult, error) {
	result := &ProcessingResult{
		UpdatedEntities: make([]string, 0),
		Errors:          make([]error, 0),
	}

	// Filter by excluded types
	if len(opts.ExcludeTypes) > 0 {
		filtered := make([]*doxygen.Entity, 0)
		excludeMap := make(map[ast.EntityType]bool)
		for _, t := range opts.ExcludeTypes {
			excludeMap[t] = true
		}

		for _, entity := range entities {
			if !excludeMap[entity.GetInstruction().Type] {
				filtered = append(filtered, entity)
			}
		}
		entities = filtered
	}

	// Apply entity limit
	if opts.MaxEntities > 0 && len(entities) > opts.MaxEntities {
		entities = entities[:opts.MaxEntities]
	}

	result.EntitiesProcessed = len(entities)

	// Process each entity
	for _, entity := range entities {
		entityPath := entity.GetInstruction().GetFullPath()

		if opts.DryRun {
			result.UpdatedEntities = append(result.UpdatedEntities, entityPath)
			continue
		}

		// Generate documentation for the entity
		err := s.generateEntityDocumentation(ctx, doc, entity, opts.AdditionalContext)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to document %s: %w", entityPath, err))
			continue
		}

		result.EntitiesUpdated++
		result.UpdatedEntities = append(result.UpdatedEntities, entityPath)
	}

	return result, nil
}

// generateEntityDocumentation generates documentation for a single entity
func (s *DocumentationService) generateEntityDocumentation(ctx context.Context, doc *doxygen.DocLayer, entity *doxygen.Entity, additionalContext string) error {
	// Extract context for the entity
	context := entity.Context(true, true)
	if context == "" {
		return fmt.Errorf("failed to extract context")
	}

	// Determine entity type for LLM prompt
	entityType := entity.GetInstruction().Type.String()

	// Get the appropriate prompt template based on entity type
	promptTemplate := s.getPromptTemplate(entityType)

	prompt := fmt.Sprintf(
		promptTemplate,
		additionalContext,
		context,
	)

	// Create documentation request
	docRequest := llm.LLMRequest{
		Prompt: prompt,
	}

	// Generate documentation using LLM
	result, err := s.llmService.Generate(ctx, docRequest)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	err = entity.ApplyRaw(fmt.Sprintf("/** @brief %s */", result.Response))
	if err != nil {
		return fmt.Errorf("failed to apply generated comment: %w", err)
	}

	return nil
}
