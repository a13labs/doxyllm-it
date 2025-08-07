// Package document provides a high-level abstraction for manipulating C++ header files
// with Doxygen documentation. It encapsulates the complexity of parsing and AST manipulation
// behind a simple, intuitive API.
package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/formatter"
	"doxyllm-it/pkg/parser"
)

// Document represents a C++ header file with its parsed AST and provides
// high-level operations for manipulating Doxygen documentation
type Document struct {
	filename    string                 // Original filename (if loaded from file)
	content     string                 // Current content
	tree        *ast.ScopeTree         // Parsed AST
	parser      *parser.Parser         // Parser instance
	formatter   *formatter.Formatter   // Formatter instance for code reconstruction
	modified    bool                   // Whether document has been modified
	entityCache map[string]*ast.Entity // Cache for quick entity lookup by path
}

// NewFromFile creates a new document by loading and parsing a file
func NewFromFile(filename string) (*Document, error) {
	// Read file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", filename, err)
	}

	return NewFromContent(absPath, string(content))
}

// NewFromContent creates a new document from content with a given name
func NewFromContent(name, content string) (*Document, error) {
	// Create parser instance
	p := parser.New()

	// Parse the content
	tree, err := p.Parse(name, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	doc := &Document{
		filename:    name,
		content:     content,
		tree:        tree,
		parser:      p,
		formatter:   formatter.New(),
		modified:    false,
		entityCache: make(map[string]*ast.Entity),
	}

	// Build entity cache
	doc.buildEntityCache()

	return doc, nil
}

// buildEntityCache builds a cache of entities by their full path for quick lookup
func (d *Document) buildEntityCache() {
	entities := d.tree.Root.GetAllEntities()
	for _, entity := range entities {
		if entity.Name != "" { // Skip root entity
			path := entity.GetFullPath()
			d.entityCache[path] = entity
		}
	}
}

// GetFilename returns the document's filename
func (d *Document) GetFilename() string {
	return d.filename
}

// GetContent returns the current content of the document
func (d *Document) GetContent() string {
	return d.content
}

// IsModified returns whether the document has been modified
func (d *Document) IsModified() bool {
	return d.modified
}

// GetTree returns the underlying AST tree (for advanced use cases)
func (d *Document) GetTree() *ast.ScopeTree {
	return d.tree
}

// Entity Lookup Methods

// FindEntity finds an entity by its full path (e.g., "MyNamespace::MyClass::myMethod")
func (d *Document) FindEntity(path string) *ast.Entity {
	return d.entityCache[path]
}

// FindEntitiesByName finds all entities with a given name (regardless of scope)
func (d *Document) FindEntitiesByName(name string) []*ast.Entity {
	var found []*ast.Entity
	for _, entity := range d.entityCache {
		if entity.Name == name {
			found = append(found, entity)
		}
	}
	return found
}

// FindEntitiesByType returns all entities of a specific type
func (d *Document) FindEntitiesByType(entityType ast.EntityType) []*ast.Entity {
	return d.tree.GetEntitiesByType(entityType)
}

// GetAllEntities returns all entities in the document
func (d *Document) GetAllEntities() []*ast.Entity {
	var entities []*ast.Entity
	for _, entity := range d.entityCache {
		entities = append(entities, entity)
	}
	return entities
}

// GetDocumentableEntities returns entities that can have Doxygen comments
func (d *Document) GetDocumentableEntities() []*ast.Entity {
	return d.tree.GetDocumentableEntities()
}

// GetUndocumentedEntities returns entities that lack Doxygen documentation
func (d *Document) GetUndocumentedEntities() []*ast.Entity {
	var undocumented []*ast.Entity
	documentable := d.GetDocumentableEntities()

	for _, entity := range documentable {
		if !entity.HasDoxygenComment() {
			undocumented = append(undocumented, entity)
		}
	}

	return undocumented
}

// Documentation Manipulation Methods

// SetEntityComment sets or updates the Doxygen comment for an entity
func (d *Document) SetEntityComment(entityPath string, comment *ast.DoxygenComment) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	entity.Comment = comment
	d.modified = true
	return nil
}

// updateCommentRaw updates the Raw field of a comment using the formatter
func (d *Document) updateCommentRaw(comment *ast.DoxygenComment) {
	// Use the formatter to generate a properly formatted comment representation
	if comment.Raw == "" {
		// Create a temporary entity to format the comment
		tempEntity := &ast.Entity{
			Comment: comment,
		}
		// Extract just the comment part using the formatter's logic
		formattedComment := d.formatter.ReconstructScope(tempEntity)
		// Extract just the comment block (remove any extra formatting)
		lines := strings.Split(formattedComment, "\n")
		var commentLines []string
		inComment := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "/**") || strings.HasPrefix(trimmed, "/*") {
				inComment = true
			}
			if inComment {
				commentLines = append(commentLines, line)
			}
			if strings.HasSuffix(trimmed, "*/") {
				inComment = false
			}
		}
		if len(commentLines) > 0 {
			comment.Raw = strings.Join(commentLines, "\n")
		} else {
			comment.Raw = "/** Programmatically generated comment */"
		}
	}
}

// SetEntityBrief sets the brief description for an entity
func (d *Document) SetEntityBrief(entityPath, brief string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Comment == nil {
		entity.Comment = &ast.DoxygenComment{
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
		}
	}

	entity.Comment.Brief = brief
	// Update Raw field to indicate this entity has documentation
	d.updateCommentRaw(entity.Comment)
	d.modified = true
	return nil
}

// SetEntityDetailed sets the detailed description for an entity
func (d *Document) SetEntityDetailed(entityPath, detailed string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Comment == nil {
		entity.Comment = &ast.DoxygenComment{
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
		}
	}

	entity.Comment.Detailed = detailed
	d.updateCommentRaw(entity.Comment)
	d.modified = true
	return nil
}

// AddEntityParam adds or updates a parameter description for a function/method
func (d *Document) AddEntityParam(entityPath, paramName, description string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Type != ast.EntityFunction && entity.Type != ast.EntityMethod &&
		entity.Type != ast.EntityConstructor {
		return fmt.Errorf("entity %s is not a function/method", entityPath)
	}

	if entity.Comment == nil {
		entity.Comment = &ast.DoxygenComment{
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
		}
	}

	entity.Comment.Params[paramName] = description
	d.updateCommentRaw(entity.Comment)
	d.modified = true
	return nil
}

// SetEntityReturn sets the return description for a function/method
func (d *Document) SetEntityReturn(entityPath, description string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Type != ast.EntityFunction && entity.Type != ast.EntityMethod {
		return fmt.Errorf("entity %s is not a function/method", entityPath)
	}

	if entity.Comment == nil {
		entity.Comment = &ast.DoxygenComment{
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
		}
	}

	entity.Comment.Returns = description
	d.updateCommentRaw(entity.Comment)
	d.modified = true
	return nil
}

// AddEntityGroup adds an entity to a Doxygen group
func (d *Document) AddEntityGroup(entityPath, groupName string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Comment == nil {
		entity.Comment = &ast.DoxygenComment{
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
		}
	}

	// Add to ingroup list if not already present
	for _, group := range entity.Comment.Ingroup {
		if group == groupName {
			return nil // Already in group
		}
	}

	entity.Comment.Ingroup = append(entity.Comment.Ingroup, groupName)
	d.updateCommentRaw(entity.Comment)
	d.modified = true
	return nil
}

// SetEntityDeprecated marks an entity as deprecated with an optional message
func (d *Document) SetEntityDeprecated(entityPath, message string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Comment == nil {
		entity.Comment = &ast.DoxygenComment{
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
		}
	}

	entity.Comment.Deprecated = message
	d.updateCommentRaw(entity.Comment)
	d.modified = true
	return nil
}

// SetEntityCustomTag sets a custom Doxygen tag for an entity
func (d *Document) SetEntityCustomTag(entityPath, tagName, value string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Comment == nil {
		entity.Comment = &ast.DoxygenComment{
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
		}
	}

	entity.Comment.CustomTags[tagName] = value
	d.updateCommentRaw(entity.Comment)
	d.modified = true
	return nil
}

// Convenience Methods

// GetEntitySummary returns a summary of an entity's documentation status
type EntitySummary struct {
	Path        string
	Type        ast.EntityType
	HasDoc      bool
	HasBrief    bool
	HasDetailed bool
	ParamCount  int
	HasReturn   bool
}

// GetEntitySummary returns documentation summary for an entity
func (d *Document) GetEntitySummary(entityPath string) (*EntitySummary, error) {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return nil, fmt.Errorf("entity not found: %s", entityPath)
	}

	summary := &EntitySummary{
		Path:   entityPath,
		Type:   entity.Type,
		HasDoc: entity.HasDoxygenComment(),
	}

	if entity.Comment != nil {
		summary.HasBrief = entity.Comment.Brief != ""
		summary.HasDetailed = entity.Comment.Detailed != ""
		summary.ParamCount = len(entity.Comment.Params)
		summary.HasReturn = entity.Comment.Returns != ""
	}

	return summary, nil
}

// GetDocumentationStats returns overall documentation statistics
type DocumentationStats struct {
	TotalEntities         int
	DocumentedEntities    int
	UndocumentedEntities  int
	DocumentationCoverage float64
}

// GetDocumentationStats returns documentation statistics for the document
func (d *Document) GetDocumentationStats() *DocumentationStats {
	documentable := d.GetDocumentableEntities()
	undocumented := d.GetUndocumentedEntities()

	total := len(documentable)
	documented := total - len(undocumented)

	coverage := 0.0
	if total > 0 {
		coverage = float64(documented) / float64(total) * 100.0
	}

	return &DocumentationStats{
		TotalEntities:         total,
		DocumentedEntities:    documented,
		UndocumentedEntities:  len(undocumented),
		DocumentationCoverage: coverage,
	}
}

// Batch Operations

// ApplyBatchUpdates applies multiple documentation updates in a single operation
type BatchUpdate struct {
	EntityPath string
	Brief      *string
	Detailed   *string
	Params     map[string]string
	Return     *string
	Groups     []string
	CustomTags map[string]string
	Deprecated *string
}

// ApplyBatchUpdates applies multiple updates efficiently
func (d *Document) ApplyBatchUpdates(updates []BatchUpdate) error {
	for _, update := range updates {
		entity := d.FindEntity(update.EntityPath)
		if entity == nil {
			return fmt.Errorf("entity not found: %s", update.EntityPath)
		}

		// Ensure comment exists
		if entity.Comment == nil {
			entity.Comment = &ast.DoxygenComment{
				Params:     make(map[string]string),
				CustomTags: make(map[string]string),
			}
		}

		// Apply updates
		if update.Brief != nil {
			entity.Comment.Brief = *update.Brief
		}
		if update.Detailed != nil {
			entity.Comment.Detailed = *update.Detailed
		}
		if update.Return != nil {
			entity.Comment.Returns = *update.Return
		}
		if update.Deprecated != nil {
			entity.Comment.Deprecated = *update.Deprecated
		}

		// Update parameters
		for paramName, description := range update.Params {
			entity.Comment.Params[paramName] = description
		}

		// Add groups
		for _, group := range update.Groups {
			// Check if already in group
			found := false
			for _, existing := range entity.Comment.Ingroup {
				if existing == group {
					found = true
					break
				}
			}
			if !found {
				entity.Comment.Ingroup = append(entity.Comment.Ingroup, group)
			}
		}

		// Set custom tags
		for tagName, value := range update.CustomTags {
			entity.Comment.CustomTags[tagName] = value
		}

		// Update raw field to indicate this entity has documentation
		d.updateCommentRaw(entity.Comment)
	}

	if len(updates) > 0 {
		d.modified = true
	}
	return nil
}

// File Operations

// Save saves the document back to its original file (if loaded from file)
func (d *Document) Save() error {
	if d.filename == "" {
		return fmt.Errorf("cannot save: document was not loaded from a file")
	}
	return d.SaveAs(d.filename)
}

// SaveAs saves the document to a specified file
func (d *Document) SaveAs(filename string) error {
	// Reconstruct the code with updated comments using the formatter
	reconstructedCode := d.formatter.ReconstructCode(d.tree)

	// Write the reconstructed code to the file
	err := os.WriteFile(filename, []byte(reconstructedCode), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	// Update the document's content and filename if successful
	d.content = reconstructedCode
	d.filename = filename
	d.modified = false

	return nil
}

// SaveToString returns the document content as a string with all modifications applied
func (d *Document) SaveToString() (string, error) {
	// Use the formatter to reconstruct the code with updated comments
	reconstructedCode := d.formatter.ReconstructCode(d.tree)
	return reconstructedCode, nil
}

// SaveToStringFormatted returns the document content formatted with clang-format
func (d *Document) SaveToStringFormatted() (string, error) {
	// Get the reconstructed code
	reconstructedCode := d.formatter.ReconstructCode(d.tree)

	// Apply clang-format if available
	formattedCode, err := d.formatter.FormatWithClang(reconstructedCode)
	if err != nil {
		// If clang-format fails, return the unformatted code with a warning
		return reconstructedCode, fmt.Errorf("clang-format not available, returning unformatted code: %w", err)
	}

	return formattedCode, nil
}

// GetEntityContext returns formatted context for an entity (useful for LLM workflows)
func (d *Document) GetEntityContext(entityPath string, includeParent, includeSiblings bool) (string, error) {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return "", fmt.Errorf("entity not found: %s", entityPath)
	}

	return d.formatter.ExtractEntityContext(entity, includeParent, includeSiblings), nil
}

// GetEntitySummaryFormatted returns a formatted summary of an entity
func (d *Document) GetEntitySummaryFormatted(entityPath string) (string, error) {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return "", fmt.Errorf("entity not found: %s", entityPath)
	}

	return d.formatter.GetEntitySummary(entity), nil
}

// ReconstructScope returns the reconstructed code for a specific entity scope
func (d *Document) ReconstructScope(entityPath string) (string, error) {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return "", fmt.Errorf("entity not found: %s", entityPath)
	}

	return d.formatter.ReconstructScope(entity), nil
}

// Validation Methods

// Validate checks the document for common documentation issues
type ValidationIssue struct {
	EntityPath string
	IssueType  string
	Message    string
	Severity   string // "error", "warning", "info"
}

// Validate performs validation checks on the document
func (d *Document) Validate() []ValidationIssue {
	var issues []ValidationIssue

	documentable := d.GetDocumentableEntities()

	for _, entity := range documentable {
		path := entity.GetFullPath()

		// Check for missing documentation
		if !entity.HasDoxygenComment() {
			issues = append(issues, ValidationIssue{
				EntityPath: path,
				IssueType:  "missing_documentation",
				Message:    "Entity lacks Doxygen documentation",
				Severity:   "warning",
			})
			continue
		}

		// Check functions/methods for parameter documentation
		if entity.Type == ast.EntityFunction || entity.Type == ast.EntityMethod {
			// Extract parameter names from signature (simplified)
			if entity.Comment.Returns == "" {
				issues = append(issues, ValidationIssue{
					EntityPath: path,
					IssueType:  "missing_return_doc",
					Message:    "Function/method lacks @return documentation",
					Severity:   "info",
				})
			}
		}

		// Check for empty brief descriptions
		if entity.Comment != nil && entity.Comment.Brief == "" {
			issues = append(issues, ValidationIssue{
				EntityPath: path,
				IssueType:  "missing_brief",
				Message:    "Entity has documentation but lacks brief description",
				Severity:   "info",
			})
		}
	}

	return issues
}

// String returns a string representation of the document
func (d *Document) String() string {
	stats := d.GetDocumentationStats()
	return fmt.Sprintf("Document[%s]: %d entities, %.1f%% documented, modified=%t",
		filepath.Base(d.filename), stats.TotalEntities, stats.DocumentationCoverage, d.modified)
}
