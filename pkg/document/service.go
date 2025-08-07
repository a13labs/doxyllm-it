// Package document provides a high-level service for processing documentation
// requests using the document abstraction
package document

import (
	"context"
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/llm"
)

// LLMService defines the interface for LLM-based documentation generation
type LLMService interface {
	GenerateDocumentation(ctx context.Context, req llm.DocumentationRequest) (*llm.DocumentationResult, error)
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
	MaxEntities  int              // Maximum entities to process (0 = unlimited)
	DryRun       bool             // Don't make actual changes
	BackupFiles  bool             // Create backup files
	FormatOutput bool             // Apply clang-format after processing
	ExcludeTypes []ast.EntityType // Entity types to exclude
	GroupConfig  *GroupConfig     // Group configuration for @ingroup tags
}

// GroupConfig defines configuration for Doxygen groups
type GroupConfig struct {
	Name             string   `yaml:"name"`             // Group name (for @defgroup/@ingroup)
	Title            string   `yaml:"title"`            // Group title/brief description
	Description      string   `yaml:"description"`      // Detailed group description
	Files            []string `yaml:"files"`            // Files that belong to this group
	GenerateDefgroup bool     `yaml:"generateDefgroup"` // Whether to generate @defgroup in header files
}

// ProcessingResult contains the result of document processing
type ProcessingResult struct {
	EntitiesProcessed int      // Number of entities processed
	EntitiesUpdated   int      // Number of entities actually updated
	UpdatedEntities   []string // List of updated entity paths
	DefgroupAdded     bool     // Whether @defgroup was added
	Errors            []error  // Non-fatal errors encountered
}

// ProcessUndocumentedEntities processes all undocumented entities in a document
func (s *DocumentationService) ProcessUndocumentedEntities(ctx context.Context, doc *Document, opts ProcessingOptions) (*ProcessingResult, error) {
	result := &ProcessingResult{
		UpdatedEntities: make([]string, 0),
		Errors:          make([]error, 0),
	}

	// Get undocumented entities
	undocumented := doc.GetUndocumentedEntities()

	// Filter by excluded types
	if len(opts.ExcludeTypes) > 0 {
		filtered := make([]*ast.Entity, 0)
		excludeMap := make(map[ast.EntityType]bool)
		for _, t := range opts.ExcludeTypes {
			excludeMap[t] = true
		}

		for _, entity := range undocumented {
			if !excludeMap[entity.Type] {
				filtered = append(filtered, entity)
			}
		}
		undocumented = filtered
	}

	// Apply entity limit
	if opts.MaxEntities > 0 && len(undocumented) > opts.MaxEntities {
		undocumented = undocumented[:opts.MaxEntities]
	}

	result.EntitiesProcessed = len(undocumented)

	// Process each entity
	for _, entity := range undocumented {
		entityPath := entity.GetFullPath()

		if opts.DryRun {
			result.UpdatedEntities = append(result.UpdatedEntities, entityPath)
			continue
		}

		// Generate documentation for the entity
		err := s.generateEntityDocumentation(ctx, doc, entity, opts.GroupConfig)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to document %s: %w", entityPath, err))
			continue
		}

		result.EntitiesUpdated++
		result.UpdatedEntities = append(result.UpdatedEntities, entityPath)
	}

	return result, nil
}

// ProcessEntitiesNeedingGroupUpdate processes entities that need @ingroup tags
func (s *DocumentationService) ProcessEntitiesNeedingGroupUpdate(ctx context.Context, doc *Document, group *GroupConfig) (*ProcessingResult, error) {
	result := &ProcessingResult{
		UpdatedEntities: make([]string, 0),
		Errors:          make([]error, 0),
	}

	if group == nil {
		return result, nil
	}

	// Find entities that are documented but missing @ingroup
	documentedEntities := s.getDocumentedEntitiesWithoutGroup(doc, group.Name)
	result.EntitiesProcessed = len(documentedEntities)

	for _, entity := range documentedEntities {
		entityPath := entity.GetFullPath()

		// Add @ingroup to existing documentation
		err := doc.AddEntityGroup(entityPath, group.Name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to add group to %s: %w", entityPath, err))
			continue
		}

		result.EntitiesUpdated++
		result.UpdatedEntities = append(result.UpdatedEntities, entityPath)
	}

	return result, nil
}

// AddDefgroupToDocument adds a @defgroup comment to the beginning of a document
func (s *DocumentationService) AddDefgroupToDocument(doc *Document, group *GroupConfig) error {
	if group == nil || !group.GenerateDefgroup {
		return nil
	}

	// Check if @defgroup already exists
	content := doc.GetContent()
	if strings.Contains(content, "@defgroup") && strings.Contains(content, group.Name) {
		return nil // Already exists
	}

	// Generate @defgroup comment
	_ = s.generateDefgroupComment(group)

	// TODO: Implement insertion of defgroup at file beginning
	// This would require extending the Document interface to support
	// inserting content at arbitrary positions

	return fmt.Errorf("defgroup insertion not yet implemented in document abstraction")
}

// generateEntityDocumentation generates documentation for a single entity
func (s *DocumentationService) generateEntityDocumentation(ctx context.Context, doc *Document, entity *ast.Entity, group *GroupConfig) error {
	// Extract context for the entity
	context, err := doc.GetEntityContext(entity.GetFullPath(), false, false)
	if err != nil {
		return fmt.Errorf("failed to extract context: %w", err)
	}

	// Determine entity type for LLM prompt
	entityType := s.getEntityTypeDescription(entity)

	// Create documentation request
	docRequest := llm.DocumentationRequest{
		EntityName:        entity.GetFullPath(),
		EntityType:        entityType,
		Context:           context,
		AdditionalContext: "", // TODO: Add support for .doxyllm context
	}

	// Generate documentation using LLM
	result, err := s.llmService.GenerateDocumentation(ctx, docRequest)
	if err != nil {
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse the generated comment into structured format
	comment := s.parseGeneratedComment(result.Comment)

	// Add group information if specified
	if group != nil {
		if comment.Ingroup == nil {
			comment.Ingroup = make([]string, 0)
		}
		// Check if group not already present
		found := false
		for _, g := range comment.Ingroup {
			if g == group.Name {
				found = true
				break
			}
		}
		if !found {
			comment.Ingroup = append(comment.Ingroup, group.Name)
		}
	}

	// Set the comment on the entity
	return doc.SetEntityComment(entity.GetFullPath(), comment)
}

// getDocumentedEntitiesWithoutGroup finds entities with documentation but missing the specified group
func (s *DocumentationService) getDocumentedEntitiesWithoutGroup(doc *Document, groupName string) []*ast.Entity {
	var entities []*ast.Entity

	documentable := doc.GetDocumentableEntities()
	for _, entity := range documentable {
		if entity.HasDoxygenComment() {
			// Check if entity is missing the group
			if entity.Comment == nil {
				continue
			}

			hasGroup := false
			for _, group := range entity.Comment.Ingroup {
				if group == groupName {
					hasGroup = true
					break
				}
			}

			if !hasGroup {
				entities = append(entities, entity)
			}
		}
	}

	return entities
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
	case ast.EntityFunction:
		return "function"
	case ast.EntityMethod:
		return "method"
	case ast.EntityConstructor:
		return "constructor"
	case ast.EntityDestructor:
		return "destructor"
	case ast.EntityVariable:
		return "variable"
	case ast.EntityField:
		return "field"
	case ast.EntityTypedef:
		return "typedef"
	case ast.EntityUsing:
		return "using declaration"
	default:
		return "entity"
	}
}

// parseGeneratedComment parses a generated comment string into a DoxygenComment structure
func (s *DocumentationService) parseGeneratedComment(commentText string) *ast.DoxygenComment {
	// This is a simplified parser - in practice you might want to use
	// the parser.ParseDoxygenComment function or enhance it
	comment := &ast.DoxygenComment{
		Raw:        commentText,
		Params:     make(map[string]string),
		CustomTags: make(map[string]string),
		Ingroup:    make([]string, 0),
	}

	// Extract brief description
	lines := strings.Split(commentText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "/**")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSuffix(line, "*/")
		line = strings.TrimSpace(line)

		if line != "" {
			if strings.HasPrefix(line, "@brief") {
				// Extract the content after @brief
				briefContent := strings.TrimSpace(strings.TrimPrefix(line, "@brief"))
				if briefContent != "" {
					comment.Brief = briefContent
					break
				}
			} else if !strings.HasPrefix(line, "@") {
				// If no @brief tag, use the first non-tag line
				comment.Brief = line
				break
			}
		}
	}

	return comment
}

// generateDefgroupComment generates a @defgroup comment for a group
func (s *DocumentationService) generateDefgroupComment(group *GroupConfig) string {
	var comment strings.Builder
	comment.WriteString("/**\n")
	comment.WriteString(fmt.Sprintf(" * @defgroup %s %s\n", group.Name, group.Title))

	if group.Description != "" {
		comment.WriteString(" * @{\n")
		comment.WriteString(" *\n")
		lines := strings.Split(group.Description, "\n")
		for _, line := range lines {
			if line != "" {
				comment.WriteString(fmt.Sprintf(" * %s\n", line))
			} else {
				comment.WriteString(" *\n")
			}
		}
		comment.WriteString(" * @}\n")
	}

	comment.WriteString(" */")
	return comment.String()
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
	if entity.Type == ast.EntityVariable {
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
