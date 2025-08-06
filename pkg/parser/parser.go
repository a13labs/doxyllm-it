// Package parser implements a C++ header file parser for extracting documentable entities
package parser

import (
	"fmt"
	"regexp"
	"strings"

	"doxyllm-it/pkg/ast"
)

// Parser represents the C++ parser
type Parser struct {
	content        string
	lines          []string
	current        int
	position       ast.Position
	scopeStack     []*ast.Entity
	tree           *ast.ScopeTree
	accessStack    []ast.AccessLevel   // Track access levels for each scope
	pendingComment *ast.DoxygenComment // Comment waiting to be associated with next entity
}

// New creates a new parser instance
func New() *Parser {
	return &Parser{
		scopeStack:  make([]*ast.Entity, 0),
		accessStack: make([]ast.AccessLevel, 0),
	}
}

// Parse parses a C++ header file and returns a scope tree
func (p *Parser) Parse(filename, content string) (*ast.ScopeTree, error) {
	p.content = content
	p.lines = strings.Split(content, "\n")
	p.current = 0
	p.position = ast.Position{Line: 1, Column: 1, Offset: 0}
	p.tree = ast.NewScopeTree(filename, content)
	p.scopeStack = []*ast.Entity{p.tree.Root}

	// Parse the content line by line
	for p.current < len(p.lines) {
		line := p.lines[p.current]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			p.nextLine()
			continue
		}

		// Handle Doxygen comments
		if strings.HasPrefix(trimmed, "/**") || strings.HasPrefix(trimmed, "///") || strings.HasPrefix(trimmed, "//!") {
			comment, err := p.parseDoxygenComment()
			if err != nil {
				return nil, fmt.Errorf("error parsing doxygen comment at line %d: %w", p.current+1, err)
			}
			// Store comment to associate with next entity
			p.pendingComment = comment
			continue
		}

		// Skip other comments
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			p.nextLine()
			continue
		}

		// Handle different C++ constructs
		if err := p.parseLine(line); err != nil {
			return nil, fmt.Errorf("error parsing line %d: %w", p.current+1, err)
		}

		p.nextLine()
	}

	return p.tree, nil
}

// parseLine parses a single line and identifies C++ constructs
func (p *Parser) parseLine(line string) error {
	trimmed := strings.TrimSpace(line)

	// Skip preprocessor directives and comments
	if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return nil
	}

	// Parse different constructs
	if p.isAccessSpecifier(trimmed) {
		return p.parseAccessSpecifier(trimmed)
	}
	if p.isNamespace(trimmed) {
		return p.parseNamespace(line)
	}
	if p.isClass(trimmed) {
		return p.parseClass(line)
	}
	if p.isStruct(trimmed) {
		return p.parseStruct(line)
	}
	if p.isEnum(trimmed) {
		return p.parseEnum(line)
	}
	if p.isFunction(trimmed) {
		return p.parseFunction(line)
	}
	if p.isVariable(trimmed) {
		return p.parseVariable(line)
	}
	if p.isTypedef(trimmed) {
		return p.parseTypedef(line)
	}
	if p.isUsing(trimmed) {
		return p.parseUsing(line)
	}

	// Handle scope closers
	if trimmed == "}" || strings.HasPrefix(trimmed, "}") {
		return p.closeScope()
	}

	return nil
}

// parseDoxygenComment parses a multi-line or single-line Doxygen comment
func (p *Parser) parseDoxygenComment() (*ast.DoxygenComment, error) {
	startLine := p.current
	var commentLines []string

	line := strings.TrimSpace(p.lines[p.current])

	if strings.HasPrefix(line, "/**") {
		// Multi-line comment starting with /**
		commentLines = append(commentLines, p.lines[p.current])
		p.nextLine()

		// Continue until we find the closing */
		for p.current < len(p.lines) {
			line = p.lines[p.current]
			commentLines = append(commentLines, line)

			if strings.Contains(strings.TrimSpace(line), "*/") {
				p.nextLine()
				break
			}
			p.nextLine()
		}
	} else if strings.HasPrefix(line, "///") || strings.HasPrefix(line, "//!") {
		// Single-line Doxygen comments - collect consecutive ones
		for p.current < len(p.lines) {
			line = strings.TrimSpace(p.lines[p.current])
			if strings.HasPrefix(line, "///") || strings.HasPrefix(line, "//!") {
				commentLines = append(commentLines, p.lines[p.current])
				p.nextLine()
			} else if line == "" {
				// Allow empty lines within single-line comment blocks
				commentLines = append(commentLines, p.lines[p.current])
				p.nextLine()
			} else {
				break
			}
		}
	}

	if len(commentLines) == 0 {
		return nil, nil
	}

	commentText := strings.Join(commentLines, "\n")
	comment := ParseDoxygenComment(commentText)
	if comment != nil {
		comment.Range = ast.Range{
			Start: ast.Position{Line: startLine + 1, Column: 1, Offset: 0},
			End:   ast.Position{Line: p.current, Column: 1, Offset: 0},
		}
	}

	return comment, nil
}

// Regular expressions for identifying C++ constructs
var (
	namespaceRegex      = regexp.MustCompile(`^\s*namespace\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\{?`)
	classRegex          = regexp.MustCompile(`^\s*(?:template\s*<[^>]*>\s*)?class\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?::\s*[^{]*?)?\s*\{?`)
	structRegex         = regexp.MustCompile(`^\s*(?:template\s*<[^>]*>\s*)?struct\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?::\s*[^{]*?)?\s*\{?`)
	enumRegex           = regexp.MustCompile(`^\s*enum\s+(?:class\s+)?([a-zA-Z_][a-zA-Z0-9_]*)\s*(?::\s*[^{]*?)?\s*\{?`)
	functionRegex       = regexp.MustCompile(`^\s*(?:(?:template\s*<[^>]*>\s*)?(?:inline|static|virtual|explicit|constexpr|friend|TCB_SPAN_CONSTEXPR11|TCB_SPAN_ARRAY_CONSTEXPR|TCB_SPAN_NODISCARD)\s+)*(?:(?:[a-zA-Z_][a-zA-Z0-9_:<>*&\s]*\s+)+)?([a-zA-Z_~][a-zA-Z0-9_]*)\s*\([^{]*\)\s*(?:const\s*)?(?:override\s*)?(?:final\s*)?(?:noexcept\s*)?(?:\{|;)`)
	variableRegex       = regexp.MustCompile(`^\s*(?:(?:static|const|constexpr|mutable|extern)\s+)*[a-zA-Z_][a-zA-Z0-9_:<>*&\s]+\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:=.*?)?;`)
	typedefRegex        = regexp.MustCompile(`^\s*typedef\s+.*?\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*;`)
	usingRegex          = regexp.MustCompile(`^\s*using\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
	usingNamespaceRegex = regexp.MustCompile(`^\s*using\s+namespace\s+([a-zA-Z_][a-zA-Z0-9_:]*)\s*;`)
)

// isNamespace checks if line contains a namespace declaration
func (p *Parser) isNamespace(line string) bool {
	return namespaceRegex.MatchString(line)
}

// isClass checks if line contains a class declaration
func (p *Parser) isClass(line string) bool {
	return classRegex.MatchString(line)
}

// isStruct checks if line contains a struct declaration
func (p *Parser) isStruct(line string) bool {
	return structRegex.MatchString(line)
}

// isEnum checks if line contains an enum declaration
func (p *Parser) isEnum(line string) bool {
	return enumRegex.MatchString(line)
}

// isFunction checks if line contains a function declaration
func (p *Parser) isFunction(line string) bool {
	// Skip if it's a class/struct declaration
	if p.isClass(line) || p.isStruct(line) {
		return false
	}

	// Skip variable declarations with initialization
	if strings.Contains(line, "=") && !strings.Contains(line, "==") && !strings.Contains(line, "!=") {
		return false
	}

	// Skip lines that are clearly not function declarations
	if strings.Contains(line, "return ") || strings.Contains(line, "throw ") {
		return false
	}

	return functionRegex.MatchString(line)
}

// isVariable checks if line contains a variable declaration
func (p *Parser) isVariable(line string) bool {
	// Skip other types of declarations
	if p.isFunction(line) || p.isClass(line) || p.isStruct(line) || p.isEnum(line) || p.isUsing(line) || p.isTypedef(line) {
		return false
	}

	// Skip lines that don't end with semicolon
	if !strings.HasSuffix(strings.TrimSpace(line), ";") {
		return false
	}

	// Skip lines that look like function calls or statements
	if strings.Contains(line, "return ") || strings.Contains(line, "throw ") || strings.Contains(line, "if ") {
		return false
	}

	return variableRegex.MatchString(line)
}

// isTypedef checks if line contains a typedef declaration
func (p *Parser) isTypedef(line string) bool {
	return typedefRegex.MatchString(line)
}

// isUsing checks if line contains a using declaration
func (p *Parser) isUsing(line string) bool {
	return usingRegex.MatchString(line) || usingNamespaceRegex.MatchString(line)
}

// isAccessSpecifier checks if line contains an access specifier
func (p *Parser) isAccessSpecifier(line string) bool {
	return line == "public:" || line == "private:" || line == "protected:"
}

// parseNamespace parses a namespace declaration
func (p *Parser) parseNamespace(line string) error {
	matches := namespaceRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse namespace: %s", line)
	}

	name := matches[1]
	entity := &ast.Entity{
		Type:         ast.EntityNamespace,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    strings.TrimSpace(line),
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// If line contains opening brace, enter scope
	if strings.Contains(line, "{") {
		p.enterScope(entity)
	}

	return nil
}

// parseClass parses a class declaration
func (p *Parser) parseClass(line string) error {
	matches := classRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse class: %s", line)
	}

	name := matches[1]
	entity := &ast.Entity{
		Type:         ast.EntityClass,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    strings.TrimSpace(line),
		AccessLevel:  ast.AccessPrivate, // Default for class
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// If line contains opening brace, enter scope
	if strings.Contains(line, "{") {
		p.enterScope(entity)
	}

	return nil
}

// parseStruct parses a struct declaration
func (p *Parser) parseStruct(line string) error {
	matches := structRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse struct: %s", line)
	}

	name := matches[1]
	entity := &ast.Entity{
		Type:         ast.EntityStruct,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    strings.TrimSpace(line),
		AccessLevel:  ast.AccessPublic, // Default for struct
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// If line contains opening brace, enter scope
	if strings.Contains(line, "{") {
		p.enterScope(entity)
	}

	return nil
}

// parseEnum parses an enum declaration
func (p *Parser) parseEnum(line string) error {
	matches := enumRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse enum: %s", line)
	}

	name := matches[1]
	entity := &ast.Entity{
		Type:         ast.EntityEnum,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    strings.TrimSpace(line),
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// If line contains opening brace, enter scope
	if strings.Contains(line, "{") {
		p.enterScope(entity)
	}

	return nil
}

// parseFunction parses a function declaration
func (p *Parser) parseFunction(line string) error {
	matches := functionRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse function: %s", line)
	}

	name := matches[1]
	entityType := ast.EntityFunction

	// Determine if it's a method (inside a class/struct)
	if p.getCurrentScope().Type == ast.EntityClass || p.getCurrentScope().Type == ast.EntityStruct {
		entityType = ast.EntityMethod

		// Check if it's a constructor or destructor
		if name == p.getCurrentScope().Name {
			entityType = ast.EntityConstructor
		} else if strings.HasPrefix(name, "~") {
			entityType = ast.EntityDestructor
			name = strings.TrimPrefix(name, "~")
		}
	}

	entity := &ast.Entity{
		Type:         entityType,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    strings.TrimSpace(line),
		IsStatic:     strings.Contains(line, "static"),
		IsVirtual:    strings.Contains(line, "virtual"),
		IsInline:     strings.Contains(line, "inline"),
		IsConst:      strings.Contains(line, ") const"),
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	// Set access level for methods
	if entityType == ast.EntityMethod || entityType == ast.EntityConstructor || entityType == ast.EntityDestructor {
		entity.AccessLevel = p.getCurrentAccessLevel()
	}

	p.addEntity(entity)

	return nil
}

// parseVariable parses a variable declaration
func (p *Parser) parseVariable(line string) error {
	matches := variableRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse variable: %s", line)
	}

	name := matches[1]
	entityType := ast.EntityVariable

	// If inside a class/struct, it's a field
	if p.getCurrentScope().Type == ast.EntityClass || p.getCurrentScope().Type == ast.EntityStruct {
		entityType = ast.EntityField
	}

	entity := &ast.Entity{
		Type:         entityType,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    strings.TrimSpace(line),
		IsStatic:     strings.Contains(line, "static"),
		IsConst:      strings.Contains(line, "const"),
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	// Set access level for fields
	if entityType == ast.EntityField {
		entity.AccessLevel = p.getCurrentAccessLevel()
	}

	p.addEntity(entity)

	return nil
}

// parseTypedef parses a typedef declaration
func (p *Parser) parseTypedef(line string) error {
	matches := typedefRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse typedef: %s", line)
	}

	name := matches[1]
	entity := &ast.Entity{
		Type:         ast.EntityTypedef,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    strings.TrimSpace(line),
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	return nil
}

// parseUsing parses a using declaration
func (p *Parser) parseUsing(line string) error {
	// Try regular using declaration first
	matches := usingRegex.FindStringSubmatch(line)
	if len(matches) >= 2 {
		name := matches[1]
		entity := &ast.Entity{
			Type:         ast.EntityUsing,
			Name:         name,
			FullName:     p.buildFullName(name),
			Signature:    strings.TrimSpace(line),
			SourceRange:  p.getCurrentRange(),
			HeaderRange:  p.getCurrentRange(),
			OriginalText: line,
			Children:     make([]*ast.Entity, 0),
		}

		p.addEntity(entity)
		return nil
	}

	// Try using namespace directive
	matches = usingNamespaceRegex.FindStringSubmatch(line)
	if len(matches) >= 2 {
		name := matches[1]
		entity := &ast.Entity{
			Type:         ast.EntityUsing,
			Name:         name,
			FullName:     p.buildFullName(name),
			Signature:    strings.TrimSpace(line),
			SourceRange:  p.getCurrentRange(),
			HeaderRange:  p.getCurrentRange(),
			OriginalText: line,
			Children:     make([]*ast.Entity, 0),
		}

		p.addEntity(entity)
		return nil
	}

	return fmt.Errorf("failed to parse using: %s", line)
}

// parseAccessSpecifier parses an access specifier (public:, private:, protected:)
func (p *Parser) parseAccessSpecifier(line string) error {
	var accessLevel ast.AccessLevel
	switch line {
	case "public:":
		accessLevel = ast.AccessPublic
	case "private:":
		accessLevel = ast.AccessPrivate
	case "protected:":
		accessLevel = ast.AccessProtected
	default:
		return fmt.Errorf("unknown access specifier: %s", line)
	}

	// Update the current access level for this scope
	if len(p.accessStack) > 0 {
		p.accessStack[len(p.accessStack)-1] = accessLevel
	}

	return nil
}

// Helper methods

// nextLine advances to the next line
func (p *Parser) nextLine() {
	if p.current < len(p.lines) {
		p.position.Line++
		p.position.Column = 1
		p.position.Offset += len(p.lines[p.current]) + 1 // +1 for newline
		p.current++
	}
}

// getCurrentRange returns the current position as a range
func (p *Parser) getCurrentRange() ast.Range {
	line := ""
	if p.current < len(p.lines) {
		line = p.lines[p.current]
	}

	start := p.position
	end := ast.Position{
		Line:   p.position.Line,
		Column: p.position.Column + len(line),
		Offset: p.position.Offset + len(line),
	}

	return ast.Range{Start: start, End: end}
}

// getCurrentScope returns the current scope
func (p *Parser) getCurrentScope() *ast.Entity {
	if len(p.scopeStack) == 0 {
		return p.tree.Root
	}
	return p.scopeStack[len(p.scopeStack)-1]
}

// getCurrentAccessLevel returns the current access level (for class members)
func (p *Parser) getCurrentAccessLevel() ast.AccessLevel {
	// Use the access stack if available
	if len(p.accessStack) > 0 {
		return p.accessStack[len(p.accessStack)-1]
	}

	// Fall back to default based on scope type
	scope := p.getCurrentScope()
	if scope.Type == ast.EntityClass {
		return ast.AccessPrivate // Default for class
	} else if scope.Type == ast.EntityStruct {
		return ast.AccessPublic // Default for struct
	}
	return ast.AccessUnknown
}

// buildFullName builds the fully qualified name for an entity
func (p *Parser) buildFullName(name string) string {
	var parts []string
	for _, scope := range p.scopeStack {
		if scope.Name != "" {
			parts = append(parts, scope.Name)
		}
	}
	parts = append(parts, name)
	return strings.Join(parts, "::")
}

// addEntity adds an entity to the current scope
func (p *Parser) addEntity(entity *ast.Entity) {
	// Associate pending comment with this entity
	if p.pendingComment != nil {
		entity.Comment = p.pendingComment
		p.pendingComment = nil // Clear the pending comment
	}

	currentScope := p.getCurrentScope()
	currentScope.AddChild(entity)
	// Don't add nested entities to the flat tree list - they should only exist in the hierarchy
}

// enterScope enters a new scope
func (p *Parser) enterScope(entity *ast.Entity) {
	p.scopeStack = append(p.scopeStack, entity)

	// Initialize access level for the new scope
	var defaultAccess ast.AccessLevel
	if entity.Type == ast.EntityClass {
		defaultAccess = ast.AccessPrivate
	} else if entity.Type == ast.EntityStruct {
		defaultAccess = ast.AccessPublic
	} else {
		defaultAccess = ast.AccessUnknown
	}
	p.accessStack = append(p.accessStack, defaultAccess)
}

// closeScope closes the current scope
func (p *Parser) closeScope() error {
	if len(p.scopeStack) <= 1 {
		return nil // Don't pop the root scope
	}

	p.scopeStack = p.scopeStack[:len(p.scopeStack)-1]

	// Also pop the access stack
	if len(p.accessStack) > 0 {
		p.accessStack = p.accessStack[:len(p.accessStack)-1]
	}

	return nil
}

// ParseDoxygenComment parses a doxygen comment block
func ParseDoxygenComment(comment string) *ast.DoxygenComment {
	if comment == "" {
		return nil
	}

	doc := &ast.DoxygenComment{
		Raw:        comment,
		Params:     make(map[string]string),
		CustomTags: make(map[string]string),
	}

	// Clean up the comment (remove /** */ and leading *)
	lines := strings.Split(comment, "\n")
	var cleanLines []string

	for i, line := range lines {
		clean := strings.TrimSpace(line)

		// Remove comment markers
		if i == 0 && strings.HasPrefix(clean, "/**") {
			clean = strings.TrimPrefix(clean, "/**")
		}
		if i == len(lines)-1 && strings.HasSuffix(clean, "*/") {
			clean = strings.TrimSuffix(clean, "*/")
		}
		clean = strings.TrimPrefix(clean, "*")

		clean = strings.TrimSpace(clean)
		if clean != "" {
			cleanLines = append(cleanLines, clean)
		}
	}

	// Parse doxygen tags
	var currentTag string
	var currentContent []string

	for _, line := range cleanLines {
		if strings.HasPrefix(line, "@") || strings.HasPrefix(line, "\\") {
			// Save previous tag
			if currentTag != "" {
				setDoxygenTag(doc, currentTag, strings.Join(currentContent, " "))
			} // Start new tag
			parts := strings.SplitN(line[1:], " ", 2)
			currentTag = parts[0]
			currentContent = []string{}

			if len(parts) > 1 {
				currentContent = append(currentContent, parts[1])
			}
		} else {
			if currentTag == "" {
				// This is part of the main description
				if doc.Brief == "" {
					doc.Brief = line
				} else {
					if doc.Detailed == "" {
						doc.Detailed = line
					} else {
						doc.Detailed += " " + line
					}
				}
			} else {
				currentContent = append(currentContent, line)
			}
		}
	}

	// Save last tag
	if currentTag != "" {
		setDoxygenTag(doc, currentTag, strings.Join(currentContent, " "))
	}

	return doc
}

// setDoxygenTag sets a doxygen tag value
func setDoxygenTag(doc *ast.DoxygenComment, tag, content string) {
	switch tag {
	case "brief":
		doc.Brief = content
	case "details", "detailed":
		doc.Detailed = content
	case "param":
		parts := strings.SplitN(content, " ", 2)
		if len(parts) == 2 {
			doc.Params[parts[0]] = parts[1]
		}
	case "return", "returns":
		doc.Returns = content
	case "throw", "throws", "exception":
		doc.Throws = append(doc.Throws, content)
	case "since":
		doc.Since = content
	case "deprecated":
		doc.Deprecated = content
	case "see":
		doc.See = append(doc.See, content)
	case "author":
		doc.Author = content
	case "version":
		doc.Version = content
	// Group-related tags
	case "defgroup":
		doc.Defgroup = content
	case "ingroup":
		doc.Ingroup = append(doc.Ingroup, content)
	case "addtogroup":
		doc.Addtogroup = content
	// Structural tags
	case "file":
		doc.File = content
	case "namespace":
		doc.Namespace = content
	case "class":
		doc.Class = content
	default:
		doc.CustomTags[tag] = content
	}
}
