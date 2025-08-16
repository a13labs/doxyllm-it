// Package document_new provides a high-level abstraction for manipulating C++ header files
// with Doxygen documentation using the new architecture that separates C++ syntax from documentation semantics
package source

import (
	"fmt"
	"os"
	"path/filepath"

	ast "doxyllm-it/pkg/ast"
	parser "doxyllm-it/pkg/cppparser"
)

// SourceFile represents a C++ header file with its parsed AST and provides
// high-level operations for manipulating Doxygen documentation
type SourceFile struct {
	filename    string                 // Original filename (if loaded from file)
	tree        *ast.ScopeTree         // Parsed AST
	modified    bool                   // Whether document has been modified
	entityCache map[string]*ast.Entity // Cache for quick entity lookup by path
}

// NewSourceFromFile creates a new document by loading and parsing a file
func NewSourceFromFile(filename string) (*SourceFile, error) {
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

	return NewSourceFromContent(absPath, string(content))
}

// NewSourceFromContent creates a new document from content with a given name
func NewSourceFromContent(name, content string) (*SourceFile, error) {
	// Create parser instance
	p := parser.New()

	// Parse the content
	tree, err := p.Parse(name, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	doc := &SourceFile{
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
func (d *SourceFile) buildEntityCache() {
	entities := d.tree.Root.GetAllEntities()
	for _, entity := range entities {
		if entity.Name != "" { // Skip root entity
			path := entity.GetFullPath()
			d.entityCache[path] = entity
		}
	}
}

// GetFilename returns the document's filename
func (d *SourceFile) GetFilename() string {
	return d.filename
}

// IsModified returns whether the document has been modified
func (d *SourceFile) IsModified() bool {
	return d.modified
}

// GetTree returns the underlying AST tree (for advanced use cases)
func (d *SourceFile) GetTree() *ast.ScopeTree {
	return d.tree
}

// Entity Lookup Methods

// FindEntity finds an entity by its full path (e.g., "MyNamespace::MyClass::myMethod")
func (d *SourceFile) FindEntity(path string) *ast.Entity {
	return d.entityCache[path]
}

// FindEntitiesByName finds all entities with a given name (regardless of scope)
func (d *SourceFile) FindEntitiesByName(name string) []*ast.Entity {
	var found []*ast.Entity
	for _, entity := range d.entityCache {
		if entity.Name == name {
			found = append(found, entity)
		}
	}
	return found
}

// FindEntitiesByType returns all entities of a specific type
func (d *SourceFile) FindEntitiesByType(entityType ast.EntityType) []*ast.Entity {
	return d.tree.GetEntitiesByType(entityType)
}

// ListInstructions returns entities that are code instructions
func (d *SourceFile) ListInstructions() []string {
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
func (d *SourceFile) GetInstructions() []*ast.Entity {
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
func (d *SourceFile) isInstruction(entityType ast.EntityType) bool {
	switch entityType {
	case ast.EntityNamespace, ast.EntityClass, ast.EntityStruct, ast.EntityEnum,
		ast.EntityCallable, ast.EntityName, ast.EntityTypedef, ast.EntityUsing:
		return true
	case ast.EntityComment, ast.EntityAccessSpecifier, ast.EntityPreprocessor:
		return false
	default:
		return false
	}
}

func (d *SourceFile) GetInstruction(path string) *ast.Entity {
	e := d.FindEntity(path)
	if !d.isInstruction(e.Type) {
		return nil
	}
	return e
}
