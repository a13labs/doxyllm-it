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

// LLMService defines the interface for LLM-based documentation generation
type LLMService interface {
	Generate(ctx context.Context, req llm.DocumentationRequest) (*llm.DocumentationResult, error)
	TestConnection(ctx context.Context) error
	GetModelInfo() llm.ModelInfo
}

// DocumentationService provides high-level document processing with LLM integration
type DocumentationService struct {
	llmService LLMService
}

// NewDocumentationService creates a new document documentation service
func NewDocumentationService(llmService LLMService) *DocumentationService {
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
	GroupConfig       *GroupConfig     // Group configuration for @ingroup tags
	AdditionalContext string           // Additional context for LLM generation
}

// GroupConfig defines configuration for Doxygen groups
type GroupConfig struct {
	Name             string   `yaml:"name"`             // Group name (for @defgroup/@ingroup)
	Title            string   `yaml:"title"`            // Group title/brief description
	Description      string   `yaml:"description"`      // Detailed group description
	Files            []string `yaml:"files"`            // Files that belong to this group
	GenerateDefGroup bool     `yaml:"generateDefGroup"` // Whether to generate @defgroup in header files
}

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
		err := s.generateEntityDocumentation(ctx, doc, entity, opts.GroupConfig, opts.AdditionalContext)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to document %s: %w", entityPath, err))
			continue
		}

		result.EntitiesUpdated++
		result.UpdatedEntities = append(result.UpdatedEntities, entityPath)
	}

	return result, nil
}

// AddDefgroupToDocument adds a @defgroup comment to the beginning of a document
func (s *DocumentationService) AddDefgroupToDocument(doc *doxygen.DocLayer, group *GroupConfig) error {
	return nil
}

// generateEntityDocumentation generates documentation for a single entity
func (s *DocumentationService) generateEntityDocumentation(ctx context.Context, doc *doxygen.DocLayer, entity *doxygen.Entity, group *GroupConfig, additionalContext string) error {
	// Extract context for the entity
	context := entity.Context(true, true)
	if context == "" {
		return fmt.Errorf("failed to extract context")
	}

	// Determine entity type for LLM prompt
	entityType := s.getEntityTypeDescription(entity.GetInstruction())

	// Create documentation request
	docRequest := llm.DocumentationRequest{
		EntityName:        entity.GetInstruction().GetFullPath(),
		EntityType:        entityType,
		Context:           context,
		AdditionalContext: additionalContext,
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

// getEntityTypeDescription returns a description of the entity type for LLM prompts
func (s *DocumentationService) getEntityTypeDescription(entity *ast.Entity) string {
	switch entity.Type {
	case ast.EntityNamespace:
		return "namespace"
	case ast.EntityClass:
		return "class"
	case ast.EntityStruct:
		return "struct"
	case ast.EntityEnum:
		return "enum"
	case ast.EntityCallable:
		return "callable"
	case ast.EntityName:
		return "name"
	case ast.EntityTypedef:
		return "typedef"
	case ast.EntityUsing:
		return "using declaration"
	default:
		return "entity"
	}
}

// ShouldSkipEntity determines if an entity should be skipped during processing
func (s *DocumentationService) ShouldSkipEntity(entity *ast.Entity) bool {
	// Skip single-letter entities (likely template parameters)
	if len(entity.Name) == 1 {
		return true
	}

	// Skip system entities
	systemEntities := map[string]bool{
		"std":       true,
		"__gnu_cxx": true,
		"__detail":  true,
	}
	if systemEntities[entity.Name] {
		return true
	}

	// Skip common template parameters
	commonTemplateParams := map[string]bool{
		"T": true, "U": true, "V": true, "E": true, "N": true, "S": true,
		"Container": true, "ElementType": true, "OtherElementType": true,
	}
	if commonTemplateParams[entity.Name] {
		return true
	}

	// Skip local variables for functions
	if entity.Type == ast.EntityName {
		localVarNames := map[string]bool{
			"msg": true, "result": true, "temp": true, "i": true, "j": true, "k": true,
			"it": true, "iter": true, "val": true, "value": true, "ret": true,
		}
		if localVarNames[entity.Name] {
			return true
		}
	}

	return false
}
