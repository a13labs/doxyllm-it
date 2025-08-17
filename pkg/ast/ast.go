// Package ast defines the Abstract Syntax Tree structures for C++ documentable entities
// This version ONLY handles C++ language syntax, not documentation semantics
package ast

import (
	"strings"
)

// EntityType represents the type of documentable entity
type EntityType int

const (
	EntityRoot EntityType = iota
	EntityNamespace
	EntityClass
	EntityStruct
	EntityCallable
	EntityName
	EntityEnum
	EntityTypedef
	EntityUsing
	EntityAccessSpecifier
	EntityUnion
	EntityPreprocessor
	EntityComment
	EntityIdentifier
)

func (et EntityType) String() string {
	switch et {
	case EntityRoot:
		return "root"
	case EntityNamespace:
		return "namespace"
	case EntityClass:
		return "class"
	case EntityStruct:
		return "struct"
	case EntityEnum:
		return "enum"
	case EntityCallable:
		return "function"
	case EntityName:
		return "variable"
	case EntityTypedef:
		return "typedef"
	case EntityUsing:
		return "using"
	case EntityPreprocessor:
		return "preprocessor"
	case EntityComment:
		return "comment"
	case EntityAccessSpecifier:
		return "access"
	case EntityIdentifier:
		return "identifier"
	default:
		return "unknown"
	}
}

// AccessLevel represents C++ access levels
type AccessLevel int

const (
	AccessUnknown AccessLevel = iota
	AccessPublic
	AccessProtected
	AccessPrivate
)

func (al AccessLevel) String() string {
	switch al {
	case AccessPublic:
		return "public"
	case AccessProtected:
		return "protected"
	case AccessPrivate:
		return "private"
	default:
		return "unknown"
	}
}

// Entity represents a C++ language entity with raw comments
type Entity struct {
	// Core entity information
	Type        EntityType  // Type of C++ entity
	Name        string      // Entity name
	FullName    string      // Fully qualified name
	Signature   string      // Complete signature/declaration
	AccessLevel AccessLevel // C++ access level (for class members)

	// C++ language attributes
	Defines              []string // List of macro definitions
	IsForwardDeclaration bool     // Whether this is a forward declaration
	IsTemplate           bool     // Whether entity is templated

	// Tree structure
	Children []*Entity // Child entities
	Parent   *Entity   // Parent entity

	// Source information
	Body        []string // Inner body content for functions (without comments)
	LineComment *Entity  // Line comment associated with the entity
}

// GetPath returns the hierarchical path to this entity
func (e *Entity) GetPath() []string {
	var path []string
	current := e
	for current != nil {
		if current.Name != "" {
			path = append([]string{current.Name}, path...)
		}
		current = current.Parent
	}
	return path
}

// GetFullPath returns the full path as a string
func (e *Entity) GetFullPath() string {
	path := e.GetPath()
	if len(path) == 0 {
		return ""
	}
	return strings.Join(path, "::")
}

// GetScope returns the scope this entity belongs to
func (e *Entity) GetScope() string {
	if e.Parent == nil || e.Parent.Name == "" {
		return "::" // Global scope
	}
	return e.Parent.GetFullPath()
}

// AddChild adds a child entity
func (e *Entity) AddChild(child *Entity) {
	child.Parent = e
	e.Children = append(e.Children, child)
}

// FindChild finds a direct child by name
func (e *Entity) FindChild(name string) *Entity {
	for _, child := range e.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

// FindByPath finds an entity by its path (recursive search)
func (e *Entity) FindByPath(path []string) *Entity {
	if len(path) == 0 {
		return e
	}

	child := e.FindChild(path[0])
	if child == nil {
		return nil
	}

	if len(path) == 1 {
		return child
	}

	return child.FindByPath(path[1:])
}

// GetAllEntities returns all entities in the tree (depth-first)
func (e *Entity) GetAllEntities() []*Entity {
	var entities []*Entity
	entities = append(entities, e)

	for _, child := range e.Children {
		entities = append(entities, child.GetAllEntities()...)
	}

	return entities
}

// GetEntitiesByType returns all entities of a specific type
func (e *Entity) GetEntitiesByType(entityType EntityType) []*Entity {
	var entities []*Entity

	if e.Type == entityType {
		entities = append(entities, e)
	}

	for _, child := range e.Children {
		entities = append(entities, child.GetEntitiesByType(entityType)...)
	}

	return entities
}

func (e *Entity) Depth() int {
	if e.Parent == nil {
		return 0
	}
	return 1 + e.Parent.Depth()
}

func (e *Entity) IsRoot() bool {
	return e.Parent == nil
}

func (e *Entity) GetParentAtDepth(depth int) *Entity {
	if depth >= e.Depth() {
		return e
	}
	if e.Parent != nil && e.Parent.Depth() == depth {
		return e.Parent
	}
	if e.Parent == nil {
		return nil
	}
	return e.Parent.GetParentAtDepth(depth - 1)
}

// ScopeTree represents the complete parsed tree of a C++ file
type ScopeTree struct {
	Root     *Entity   // Root entity (represents the file)
	Filename string    // Source filename
	Content  string    // Original file content
	Entities []*Entity // Flat list of all entities for quick access
}

// NewScopeTree creates a new scope tree
func NewScopeTree(filename, content string) *ScopeTree {
	root := &Entity{
		Type:     EntityRoot,
		Name:     "",
		Children: make([]*Entity, 0),
		Body:     strings.Split(content, "\n"),
	}

	return &ScopeTree{
		Root:     root,
		Filename: filename,
		Content:  content,
		Entities: make([]*Entity, 0),
	}
}

// AddEntity adds an entity to the tree and flat list
func (st *ScopeTree) AddEntity(entity *Entity) {
	st.Entities = append(st.Entities, entity)
}

// FindEntity finds an entity by its full path
func (st *ScopeTree) FindEntity(path string) *Entity {
	if path == "" || path == "::" {
		return st.Root
	}

	parts := strings.Split(strings.Trim(path, ":"), "::")
	return st.Root.FindByPath(parts)
}

// GetEntitiesByType returns all entities of a specific type
func (st *ScopeTree) GetEntitiesByType(entityType EntityType) []*Entity {
	return st.Root.GetEntitiesByType(entityType)
}

func (st *ScopeTree) InsertBefore(entity *Entity, newEntity *Entity) {
	parent := entity.Parent
	if parent == nil {
		return
	}

	// Find the index of the entity to insert before
	index := -1
	for i, child := range parent.Children {
		if child == entity {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	// Insert the new entity before the existing one
	parent.Children = append(parent.Children[:index], append([]*Entity{newEntity}, parent.Children[index:]...)...)
}

func (st *ScopeTree) RemoveEntity(entity *Entity) {
	parent := entity.Parent
	if parent == nil {
		return
	}

	// Remove the entity from its parent's children
	for i, child := range parent.Children {
		if child == entity {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			break
		}
	}
}

func (st *ScopeTree) GetPrecedingEntity(entity *Entity) *Entity {
	if entity == nil {
		return nil
	}

	// Find the parent entity
	parent := entity.Parent
	if parent == nil {
		return nil
	}

	// If the entity is the first child, return the parent
	if parent.Children[0] == entity {
		return parent
	}

	// Otherwise, find the previous sibling
	for i, child := range parent.Children {
		if child == entity {
			if i > 0 {
				return parent.Children[i-1]
			}
			break
		}
	}

	return nil
}
