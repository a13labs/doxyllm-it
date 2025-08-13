// Package document_new provides a high-level abstraction for manipulating C++ header files
// with Doxygen documentation using the new architecture that separates C++ syntax from documentation semantics
package document_new

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
	parser "doxyllm-it/pkg/parser_new"
)

// Document represents a C++ header file with its parsed AST and provides
// high-level operations for manipulating Doxygen documentation
type Document struct {
	filename         string                          // Original filename (if loaded from file)
	content          string                          // Current content
	tree             *ast.ScopeTree                  // Parsed AST
	parser           *parser.Parser                  // Parser instance
	modified         bool                            // Whether document has been modified
	entityCache      map[string]*ast.Entity          // Cache for quick entity lookup by path
	prependedContent string                          // Content prepended to the file (e.g., defgroup comments)
	commentCache     map[*ast.Entity]*DoxygenComment // Cache for parsed Doxygen comments
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
		filename:         name,
		content:          content,
		tree:             tree,
		parser:           p,
		modified:         false,
		entityCache:      make(map[string]*ast.Entity),
		prependedContent: "", // Initialize empty
		commentCache:     make(map[*ast.Entity]*DoxygenComment),
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
	// Return the full content including any prepended content
	if d.prependedContent != "" {
		return d.prependedContent + d.content
	}
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
	// Determine which entity types can be documented
	// Comments and access specifiers should not be documented
	var documentable []*ast.Entity

	allEntities := d.tree.Root.GetAllEntities()
	for _, entity := range allEntities {
		if d.isDocumentableEntityType(entity.Type) {
			documentable = append(documentable, entity)
		}
	}

	return documentable
}

// isDocumentableEntityType determines if an entity type can have documentation
func (d *Document) isDocumentableEntityType(entityType ast.EntityType) bool {
	switch entityType {
	case ast.EntityNamespace, ast.EntityClass, ast.EntityStruct, ast.EntityEnum,
		ast.EntityFunction, ast.EntityConstructor, ast.EntityDestructor,
		ast.EntityVariable, ast.EntityField, ast.EntityTypedef, ast.EntityUsing,
		ast.EntityMacro, ast.EntityTemplate:
		return true
	case ast.EntityComment, ast.EntityAccessSpecifier, ast.EntityPreprocessor, ast.EntityUnknown:
		return false
	default:
		return false
	}
}

// GetUndocumentedEntities returns entities that lack Doxygen documentation
func (d *Document) GetUndocumentedEntities() []*ast.Entity {
	var undocumented []*ast.Entity
	documentable := d.GetDocumentableEntities()

	for _, entity := range documentable {
		if !d.HasDoxygenComment(entity) {
			undocumented = append(undocumented, entity)
		}
	}

	return undocumented
}

// Doxygen Comment Methods

// HasDoxygenComment checks if an entity has a Doxygen comment
func (d *Document) HasDoxygenComment(entity *ast.Entity) bool {
	// Check cache first
	if comment, exists := d.commentCache[entity]; exists {
		return comment != nil && comment.HasDoxygenContent()
	}

	// Find associated comment entity
	commentEntity := d.findCommentForEntity(entity)
	if commentEntity == nil {
		d.commentCache[entity] = nil
		return false
	}

	// Parse and cache the comment
	doxyComment := ParseDoxygenComment(commentEntity.Signature)
	d.commentCache[entity] = doxyComment
	return doxyComment.HasDoxygenContent()
}

// GetDoxygenComment gets the Doxygen comment for an entity
func (d *Document) GetDoxygenComment(entity *ast.Entity) *DoxygenComment {
	// Check cache first
	if comment, exists := d.commentCache[entity]; exists {
		return comment
	}

	// Find associated comment entity
	commentEntity := d.findCommentForEntity(entity)
	if commentEntity == nil {
		d.commentCache[entity] = nil
		return nil
	}

	// Parse and cache the comment
	doxyComment := ParseDoxygenComment(commentEntity.Signature)
	d.commentCache[entity] = doxyComment
	return doxyComment
}

// findCommentForEntity finds the comment entity associated with a given entity
// This implements the logic to associate comments with their documented entities
func (d *Document) findCommentForEntity(entity *ast.Entity) *ast.Entity {
	if entity.Parent == nil {
		return nil
	}

	// Look for a comment entity immediately before this entity in the parent's children
	siblings := entity.Parent.Children
	entityIndex := -1

	for i, child := range siblings {
		if child == entity {
			entityIndex = i
			break
		}
	}

	if entityIndex <= 0 {
		return nil
	}

	// Check the previous sibling
	prevSibling := siblings[entityIndex-1]
	if prevSibling.Type == ast.EntityComment {
		// Check if this comment appears to be a Doxygen comment
		if strings.Contains(prevSibling.Signature, "/**") ||
			strings.Contains(prevSibling.Signature, "///") {
			return prevSibling
		}
	}

	return nil
}

// Documentation Manipulation Methods

// SetEntityComment sets or updates the Doxygen comment for an entity by path
func (d *Document) SetEntityComment(entityPath string, comment *DoxygenComment) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	return d.SetEntityCommentDirect(entity, comment)
}

// SetEntityCommentDirect sets or updates the Doxygen comment for an entity directly
func (d *Document) SetEntityCommentDirect(entity *ast.Entity, comment *DoxygenComment) error {
	// Cache the comment
	d.commentCache[entity] = comment
	d.modified = true

	// Find existing comment entity or create a new one
	commentEntity := d.findCommentForEntity(entity)

	if commentEntity == nil && comment != nil && comment.HasDoxygenContent() {
		// Create a new comment entity and insert it before the target entity
		commentEntity = d.createCommentEntity(comment)
		err := d.insertCommentEntityBeforeTarget(commentEntity, entity)
		if err != nil {
			return fmt.Errorf("failed to insert comment entity: %w", err)
		}
	} else if commentEntity != nil {
		// Update existing comment entity
		if comment != nil && comment.HasDoxygenContent() {
			// Update the existing comment's raw text
			commentEntity.Signature = d.generateCommentText(comment)
		} else {
			// Remove the comment entity if comment is empty/nil
			d.removeCommentEntity(commentEntity)
		}
	}

	return nil
}

// createCommentEntity creates a new comment entity from a DoxygenComment
func (d *Document) createCommentEntity(comment *DoxygenComment) *ast.Entity {
	commentText := d.generateCommentText(comment)

	return &ast.Entity{
		Type:        ast.EntityComment,
		Name:        "comment",
		Signature:   commentText,
		AccessLevel: ast.AccessUnknown,
	}
}

// generateCommentText generates formatted comment text from a DoxygenComment
func (d *Document) generateCommentText(comment *DoxygenComment) string {
	if comment.Raw != "" {
		return comment.Raw
	}

	// Generate basic Doxygen comment format
	var lines []string
	lines = append(lines, "/**")

	if comment.Brief != "" {
		lines = append(lines, " * @brief "+comment.Brief)
	}

	if comment.Detailed != "" {
		lines = append(lines, " * @details "+comment.Detailed)
	}

	for paramName, paramDesc := range comment.Params {
		lines = append(lines, " * @param "+paramName+" "+paramDesc)
	}

	for tparamName, tparamDesc := range comment.TParams {
		lines = append(lines, " * @tparam "+tparamName+" "+tparamDesc)
	}

	if comment.Returns != "" {
		lines = append(lines, " * @return "+comment.Returns)
	}

	for _, group := range comment.Groups {
		lines = append(lines, " * @ingroup "+group)
	}

	if comment.Since != "" {
		lines = append(lines, " * @since "+comment.Since)
	}

	if comment.Deprecated != "" {
		lines = append(lines, " * @deprecated "+comment.Deprecated)
	}

	for tagName, tagValue := range comment.CustomTags {
		lines = append(lines, " * @"+tagName+" "+tagValue)
	}

	lines = append(lines, " */")

	return strings.Join(lines, "\n")
}

// insertCommentEntityBeforeTarget inserts a comment entity before the target entity
func (d *Document) insertCommentEntityBeforeTarget(commentEntity, targetEntity *ast.Entity) error {
	if targetEntity.Parent == nil {
		return fmt.Errorf("target entity has no parent")
	}

	parent := targetEntity.Parent

	// Find the index of the target entity in parent's children
	targetIndex := -1
	for i, child := range parent.Children {
		if child == targetEntity {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return fmt.Errorf("target entity not found in parent's children")
	}

	// Set the comment entity's parent
	commentEntity.Parent = parent

	// Insert the comment entity before the target entity
	parent.Children = append(parent.Children[:targetIndex],
		append([]*ast.Entity{commentEntity}, parent.Children[targetIndex:]...)...)

	return nil
}

// removeCommentEntity removes a comment entity from the AST
func (d *Document) removeCommentEntity(commentEntity *ast.Entity) {
	if commentEntity.Parent == nil {
		return
	}

	parent := commentEntity.Parent

	// Find and remove the comment entity from parent's children
	for i, child := range parent.Children {
		if child == commentEntity {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			break
		}
	}
} // SetEntityBrief sets the brief description for an entity
func (d *Document) SetEntityBrief(entityPath, brief string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	if comment == nil {
		comment = NewDoxygenComment()
	}

	comment.Brief = brief
	return d.SetEntityCommentDirect(entity, comment)
}

// SetEntityDetailed sets the detailed description for an entity
func (d *Document) SetEntityDetailed(entityPath, detailed string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	if comment == nil {
		comment = NewDoxygenComment()
	}

	comment.Detailed = detailed
	return d.SetEntityCommentDirect(entity, comment)
}

// AddEntityParam adds or updates a parameter description for a function/method
func (d *Document) AddEntityParam(entityPath, paramName, description string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Type != ast.EntityFunction && entity.Type != ast.EntityConstructor {
		return fmt.Errorf("entity %s is not a function/method", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	if comment == nil {
		comment = NewDoxygenComment()
	}

	comment.Params[paramName] = description
	return d.SetEntityCommentDirect(entity, comment)
}

// AddEntityTParam adds or updates a template parameter description for a templated entity
func (d *Document) AddEntityTParam(entityPath, tparamName, description string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if !entity.IsTemplate {
		return fmt.Errorf("entity %s is not templated", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	if comment == nil {
		comment = NewDoxygenComment()
	}

	comment.TParams[tparamName] = description
	return d.SetEntityCommentDirect(entity, comment)
}

// SetEntityReturn sets the return description for a function/method
func (d *Document) SetEntityReturn(entityPath, description string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if entity.Type != ast.EntityFunction {
		return fmt.Errorf("entity %s is not a function/method", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	if comment == nil {
		comment = NewDoxygenComment()
	}

	comment.Returns = description
	return d.SetEntityCommentDirect(entity, comment)
}

// AddEntityGroup adds an entity to a Doxygen group
func (d *Document) AddEntityGroup(entityPath, groupName string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	if comment == nil {
		comment = NewDoxygenComment()
	}

	// Add to ingroup list if not already present
	for _, group := range comment.Groups {
		if group == groupName {
			return nil // Already in group
		}
	}

	comment.Groups = append(comment.Groups, groupName)
	return d.SetEntityCommentDirect(entity, comment)
}

// SetEntityDeprecated sets the deprecation message for an entity
func (d *Document) SetEntityDeprecated(entityPath, message string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	if comment == nil {
		comment = NewDoxygenComment()
	}

	comment.Deprecated = message
	return d.SetEntityCommentDirect(entity, comment)
}

// SetEntityCustomTag sets a custom Doxygen tag for an entity
func (d *Document) SetEntityCustomTag(entityPath, tagName, value string) error {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	if comment == nil {
		comment = NewDoxygenComment()
	}

	comment.CustomTags[tagName] = value
	return d.SetEntityCommentDirect(entity, comment)
}

// PrependFileComment prepends a comment to the beginning of the file
func (d *Document) PrependFileComment(comment string) error {
	if !strings.HasSuffix(comment, "\n") {
		comment += "\n"
	}
	d.prependedContent = comment + d.prependedContent
	d.modified = true
	return nil
}

// Entity Summary and Statistics

// EntitySummary provides summary information about an entity's documentation status
type EntitySummary struct {
	Path        string
	Type        ast.EntityType
	HasDoc      bool
	HasBrief    bool
	HasDetailed bool
	ParamCount  int
	HasReturn   bool
}

// GetEntitySummary returns a summary of an entity's documentation status
func (d *Document) GetEntitySummary(entityPath string) (*EntitySummary, error) {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return nil, fmt.Errorf("entity not found: %s", entityPath)
	}

	comment := d.GetDoxygenComment(entity)
	hasDoc := comment != nil && comment.HasDoxygenContent()

	summary := &EntitySummary{
		Path:        entityPath,
		Type:        entity.Type,
		HasDoc:      hasDoc,
		HasBrief:    hasDoc && comment.Brief != "",
		HasDetailed: hasDoc && comment.Detailed != "",
		ParamCount:  0,
		HasReturn:   hasDoc && comment.Returns != "",
	}

	if hasDoc {
		summary.ParamCount = len(comment.Params)
	}

	return summary, nil
}

// DocumentationStats provides overall documentation statistics
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

// BatchUpdate represents a batch update operation
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

		comment := d.GetDoxygenComment(entity)
		if comment == nil {
			comment = NewDoxygenComment()
		}

		// Apply updates
		if update.Brief != nil {
			comment.Brief = *update.Brief
		}
		if update.Detailed != nil {
			comment.Detailed = *update.Detailed
		}
		if update.Return != nil {
			comment.Returns = *update.Return
		}
		if update.Deprecated != nil {
			comment.Deprecated = *update.Deprecated
		}

		// Update params
		for paramName, paramDesc := range update.Params {
			comment.Params[paramName] = paramDesc
		}

		// Update groups
		for _, groupName := range update.Groups {
			found := false
			for _, existingGroup := range comment.Groups {
				if existingGroup == groupName {
					found = true
					break
				}
			}
			if !found {
				comment.Groups = append(comment.Groups, groupName)
			}
		}

		// Update custom tags
		for tagName, tagValue := range update.CustomTags {
			comment.CustomTags[tagName] = tagValue
		}

		// Cache the updated comment
		d.commentCache[entity] = comment
	}

	d.modified = true
	return nil
}

// Validation

// ValidationIssue represents a validation issue
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
		comment := d.GetDoxygenComment(entity)

		if comment == nil || !comment.HasDoxygenContent() {
			issues = append(issues, ValidationIssue{
				EntityPath: path,
				IssueType:  "missing_documentation",
				Message:    "Entity has no Doxygen documentation",
				Severity:   "warning",
			})
			continue
		}

		// Check for missing brief
		if comment.Brief == "" {
			issues = append(issues, ValidationIssue{
				EntityPath: path,
				IssueType:  "missing_brief",
				Message:    "Entity documentation lacks brief description",
				Severity:   "info",
			})
		}

		// Check function-specific issues
		if entity.Type == ast.EntityFunction {
			// TODO: We would need to parse function parameters from the signature
			// to validate @param completeness. This will be implemented when
			// we have better signature parsing in the AST.
		}
	}

	return issues
}

// Utility Methods

// String returns a string representation of the document
func (d *Document) String() string {
	stats := d.GetDocumentationStats()
	return fmt.Sprintf("Document{filename: %s, entities: %d, documented: %.1f%%}",
		d.filename, stats.TotalEntities, stats.DocumentationCoverage)
}

// Save saves the document to its original file (requires formatter_new)
func (d *Document) Save() error {
	return fmt.Errorf("Save() is not yet implemented - requires formatter_new package")
}

// SaveAs saves the document to a specific file (requires formatter_new)
func (d *Document) SaveAs(filename string) error {
	return fmt.Errorf("SaveAs() is not yet implemented - requires formatter_new package")
}

// SaveToString returns the document as a string (requires formatter_new)
func (d *Document) SaveToString() (string, error) {
	return "", fmt.Errorf("SaveToString() is not yet implemented - requires formatter_new package")
}

// SaveToStringFormatted returns the formatted document as a string (requires formatter_new)
func (d *Document) SaveToStringFormatted() (string, error) {
	return "", fmt.Errorf("SaveToStringFormatted() is not yet implemented - requires formatter_new package")
}

// Context and reconstruction methods that will need formatter_new

// GetEntityContext gets context for an entity (requires formatter_new for full implementation)
func (d *Document) GetEntityContext(entityPath string, includeParent, includeSiblings bool) (string, error) {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return "", fmt.Errorf("entity not found: %s", entityPath)
	}

	// For now, just return the original text
	// TODO: Implement proper context extraction with formatter_new
	return entity.Signature, nil
}

// GetEntitySummaryFormatted returns a formatted summary of an entity (requires formatter_new)
func (d *Document) GetEntitySummaryFormatted(entityPath string) (string, error) {
	summary, err := d.GetEntitySummary(entityPath)
	if err != nil {
		return "", err
	}

	// Simple text representation for now
	return fmt.Sprintf("Entity: %s\nType: %s\nDocumented: %t\n",
		summary.Path, summary.Type, summary.HasDoc), nil
}

// ReconstructScope reconstructs the scope containing an entity (requires formatter_new)
func (d *Document) ReconstructScope(entityPath string) (string, error) {
	entity := d.FindEntity(entityPath)
	if entity == nil {
		return "", fmt.Errorf("entity not found: %s", entityPath)
	}

	// For now, just return the original text
	// TODO: Implement proper reconstruction with formatter_new
	return entity.Signature, nil
}
