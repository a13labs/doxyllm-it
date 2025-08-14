// Package document_new provides a high-level abstraction for manipulating C++ header files
// with Doxygen documentation using the new architecture that separates C++ syntax from documentation semantics
package document_new

import (
	"fmt"
	"os"
	"path/filepath"

	ast "doxyllm-it/pkg/ast_new"
	parser "doxyllm-it/pkg/parser_new"
)

// Document represents a C++ header file with its parsed AST and provides
// high-level operations for manipulating Doxygen documentation
type Document struct {
	filename    string                 // Original filename (if loaded from file)
	tree        *ast.ScopeTree         // Parsed AST
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
		tree:        tree,
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

// ListInstructions returns entities that are code instructions
func (d *Document) ListInstructions() []string {
	// Determine which entity types can be documented
	// Comments and access specifiers should not be documented
	var instructions []string

	allEntities := d.tree.Root.GetAllEntities()
	for _, entity := range allEntities {
		if d.isInstruction(entity.Type) {
			instructions = append(instructions, entity.GetFullPath())
		}
	}

	return instructions
}

// GetInstructions returns entities that are code instructions
func (d *Document) GetInstructions() []*ast.Entity {
	// Determine which entity types can be documented
	// Comments and access specifiers should not be documented
	var instructions []*ast.Entity

	allEntities := d.tree.Root.GetAllEntities()
	for _, entity := range allEntities {
		if d.isInstruction(entity.Type) {
			instructions = append(instructions, entity)
		}
	}

	return instructions
}

// isInstruction determines if an entity type can have documentation
func (d *Document) isInstruction(entityType ast.EntityType) bool {
	switch entityType {
	case ast.EntityNamespace, ast.EntityClass, ast.EntityStruct, ast.EntityEnum,
		ast.EntityFunction, ast.EntityConstructor, ast.EntityDestructor,
		ast.EntityVariable, ast.EntityField, ast.EntityTypedef, ast.EntityUsing,
		ast.EntityMacro:
		return true
	case ast.EntityComment, ast.EntityAccessSpecifier, ast.EntityPreprocessor:
		return false
	default:
		return false
	}
}

func (d *Document) GetInstruction(path string) *ast.Entity {
	e := d.FindEntity(path)
	if !d.isInstruction(e.Type) {
		return nil
	}
	return e
}
