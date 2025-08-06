// Package ast defines the Abstract Syntax Tree structures for C++ documentable entities
package ast

import (
	"strings"
)

// Position represents a position in the source file
type Position struct {
	Line   int
	Column int
	Offset int
}

// Range represents a range in the source file
type Range struct {
	Start Position
	End   Position
}

// EntityType represents the type of documentable entity
type EntityType int

const (
	EntityUnknown EntityType = iota
	EntityNamespace
	EntityClass
	EntityStruct
	EntityEnum
	EntityFunction
	EntityMethod
	EntityConstructor
	EntityDestructor
	EntityVariable
	EntityField
	EntityTypedef
	EntityUsing
	EntityMacro
	EntityTemplate
)

func (et EntityType) String() string {
	switch et {
	case EntityNamespace:
		return "namespace"
	case EntityClass:
		return "class"
	case EntityStruct:
		return "struct"
	case EntityEnum:
		return "enum"
	case EntityFunction:
		return "function"
	case EntityMethod:
		return "method"
	case EntityConstructor:
		return "constructor"
	case EntityDestructor:
		return "destructor"
	case EntityVariable:
		return "variable"
	case EntityField:
		return "field"
	case EntityTypedef:
		return "typedef"
	case EntityUsing:
		return "using"
	case EntityMacro:
		return "macro"
	case EntityTemplate:
		return "template"
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

// DoxygenComment represents a parsed doxygen comment
type DoxygenComment struct {
	Raw        string            // Original comment text
	Brief      string            // Brief description
	Detailed   string            // Detailed description
	Params     map[string]string // Parameter documentation
	Returns    string            // Return value documentation
	Throws     []string          // Exception documentation
	Since      string            // Since version
	Deprecated string            // Deprecation notice
	See        []string          // See also references
	Author     string            // Author information
	Version    string            // Version information
	// Group-related tags
	Defgroup   string   // @defgroup tag (for group definitions)
	Ingroup    []string // @ingroup tags (group memberships)
	Addtogroup string   // @addtogroup tag
	// Structural tags
	File       string            // @file tag
	Namespace  string            // @namespace tag
	Class      string            // @class tag
	CustomTags map[string]string // Custom doxygen tags
	Range      Range             // Position in source
}

// Entity represents a documentable C++ entity
type Entity struct {
	Type           EntityType      // Type of entity
	Name           string          // Entity name
	FullName       string          // Fully qualified name
	Signature      string          // Complete signature/declaration
	AccessLevel    AccessLevel     // Access level (for class members)
	IsStatic       bool            // Whether entity is static
	IsConst        bool            // Whether entity is const
	IsVirtual      bool            // Whether method is virtual
	IsPure         bool            // Whether method is pure virtual
	IsInline       bool            // Whether function is inline
	IsTemplate     bool            // Whether entity is templated
	TemplateParams []string        // Template parameters
	Namespace      string          // Containing namespace
	Class          string          // Containing class (for methods/fields)
	Comment        *DoxygenComment // Associated doxygen comment
	Children       []*Entity       // Child entities
	Parent         *Entity         // Parent entity
	SourceRange    Range           // Range in source file
	HeaderRange    Range           // Range of just the declaration/header
	BodyRange      *Range          // Range of body (for functions/classes with implementation)
	OriginalText   string          // Original text including whitespace and comments
	LeadingWS      string          // Leading whitespace/comments before entity
	TrailingWS     string          // Trailing whitespace/comments after entity
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

// IsGlobal returns true if entity is at global scope
func (e *Entity) IsGlobal() bool {
	return e.Parent == nil || (e.Parent.Type == EntityUnknown && e.Parent.Name == "")
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

// HasDoxygenComment returns true if entity has doxygen documentation
func (e *Entity) HasDoxygenComment() bool {
	return e.Comment != nil && e.Comment.Raw != ""
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
		Type:         EntityUnknown,
		Name:         "",
		FullName:     "",
		Children:     make([]*Entity, 0),
		SourceRange:  Range{Start: Position{Line: 1, Column: 1, Offset: 0}, End: Position{Line: strings.Count(content, "\n") + 1, Column: 1, Offset: len(content)}},
		OriginalText: content,
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

// GetDocumentableEntities returns entities that can have doxygen comments
func (st *ScopeTree) GetDocumentableEntities() []*Entity {
	var entities []*Entity
	documentableTypes := []EntityType{
		EntityNamespace, EntityClass, EntityStruct, EntityEnum,
		EntityFunction, EntityMethod, EntityConstructor, EntityDestructor,
		EntityVariable, EntityField, EntityTypedef, EntityUsing,
	}

	for _, entityType := range documentableTypes {
		entities = append(entities, st.GetEntitiesByType(entityType)...)
	}

	return entities
}
