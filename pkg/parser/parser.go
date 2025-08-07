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
	content         string
	lines           []string
	current         int
	position        ast.Position
	scopeStack      []*ast.Entity
	tree            *ast.ScopeTree
	accessStack     []ast.AccessLevel   // Track access levels for each scope
	pendingComment  *ast.DoxygenComment // Comment waiting to be associated with next entity
	defines         map[string]string   // Preprocessor defines
	pendingTemplate string              // Template declaration waiting for next entity
}

// New creates a new parser instance
func New() *Parser {
	return &Parser{
		scopeStack:  make([]*ast.Entity, 0),
		accessStack: make([]ast.AccessLevel, 0),
		defines:     make(map[string]string),
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

		// Handle regular comments as entities
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			if err := p.parseComment(line); err != nil {
				return nil, fmt.Errorf("error parsing comment at line %d: %w", p.current+1, err)
			}
			p.nextLine()
			continue
		}

		// Handle different C++ constructs
		if err := p.parseLine(line); err != nil {
			return nil, fmt.Errorf("error parsing line %d: %w", p.current+1, err)
		}

		p.nextLine()
	}

	// Handle any remaining pending comment as a file-level comment
	if p.pendingComment != nil {
		entity := &ast.Entity{
			Type:         ast.EntityComment, // Use comment type for file-level comments
			Name:         "file-comment",
			FullName:     "file-comment",
			Signature:    "",
			SourceRange:  p.pendingComment.Range,
			HeaderRange:  p.pendingComment.Range,
			OriginalText: p.pendingComment.Raw,
			Children:     make([]*ast.Entity, 0),
			Comment:      p.pendingComment,
		}
		p.tree.Root.AddChild(entity)
		p.pendingComment = nil
	}

	return p.tree, nil
}

// parseLine parses a single line and identifies C++ constructs
func (p *Parser) parseLine(line string) error {
	trimmed := strings.TrimSpace(line)

	// Handle #define directives first (before resolution)
	if p.isDefine(trimmed) {
		return p.parseDefine(line)
	}

	// Handle other preprocessor directives
	if p.isPreprocessor(trimmed) {
		return p.parsePreprocessor(line)
	}

	// Parse file-level comments as entities
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
		return p.parseComment(line)
	}

	// Resolve defines in the line before parsing
	resolvedLine := p.resolveDefines(line)
	resolvedTrimmed := strings.TrimSpace(resolvedLine)

	// Parse different constructs with resolved content
	if p.isAccessSpecifier(resolvedTrimmed) {
		return p.parseAccessSpecifier(resolvedTrimmed)
	}
	if p.isTemplate(resolvedTrimmed) {
		return p.handleTemplate(resolvedLine)
	}
	if p.isNamespace(resolvedTrimmed) {
		return p.parseNamespace(resolvedLine)
	}
	if p.isEnum(resolvedTrimmed) {
		return p.parseEnum(resolvedLine)
	}
	if p.isClass(resolvedTrimmed) {
		return p.parseClass(resolvedLine)
	}
	if p.isStruct(resolvedTrimmed) {
		return p.parseStruct(resolvedLine)
	}
	if p.isFunction(resolvedTrimmed) {
		return p.parseFunction(resolvedLine)
	}
	if p.isVariable(resolvedTrimmed) {
		return p.parseVariable(resolvedLine)
	}
	if p.isTypedef(resolvedTrimmed) {
		return p.parseTypedef(resolvedLine)
	}
	if p.isUsing(resolvedTrimmed) {
		return p.parseUsing(resolvedLine)
	}

	// Handle scope closers
	if resolvedTrimmed == "}" || strings.HasPrefix(resolvedTrimmed, "}") {
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

// parseDefine parses a #define directive (including multiline defines)
func (p *Parser) parseDefine(line string) error {
	var defineContent strings.Builder

	// Start with the first line, removing trailing backslash if present
	current := strings.TrimSpace(line)
	current = strings.TrimSuffix(current, "\\")
	defineContent.WriteString(current)

	// Check if this is a multiline define (original line ends with backslash)
	for strings.HasSuffix(strings.TrimSpace(line), "\\") {
		// Move to next line
		p.nextLine()
		if p.current >= len(p.lines) {
			break
		}

		line = p.lines[p.current]
		trimmed := strings.TrimSpace(line)

		// Add the continuation line
		defineContent.WriteString(" ")
		defineContent.WriteString(strings.TrimSuffix(trimmed, "\\"))
	}

	// Parse the complete define
	fullDefine := defineContent.String()
	matches := defineRegex.FindStringSubmatch(fullDefine)
	if len(matches) >= 2 {
		name := matches[1]
		value := ""
		if len(matches) >= 3 {
			value = strings.TrimSpace(matches[2])
		}

		// Store in defines map for resolution
		p.defines[name] = value

		// Create an entity for the define to preserve it in reconstruction
		entity := &ast.Entity{
			Type:         ast.EntityPreprocessor,
			Name:         name,
			FullName:     name,
			Signature:    fullDefine,
			SourceRange:  p.getCurrentRange(),
			HeaderRange:  p.getCurrentRange(),
			OriginalText: line,
			Children:     make([]*ast.Entity, 0),
		}

		p.addEntity(entity)
	}

	return nil
}

// parseComment parses a file-level comment
func (p *Parser) parseComment(line string) error {
	trimmed := strings.TrimSpace(line)

	// Extract a meaningful name for the comment
	name := "comment"

	// Try to extract first meaningful words from comment content
	content := ""
	if strings.HasPrefix(trimmed, "//") {
		content = strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
	} else if strings.HasPrefix(trimmed, "/*") {
		content = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(trimmed, "*/"), "/*"))
	}

	// Use first few words as name if available
	if content != "" {
		words := strings.Fields(content)
		if len(words) > 0 {
			if len(words) == 1 {
				name = words[0]
			} else {
				name = strings.Join(words[:min(3, len(words))], "_")
			}
			// Clean up name to be a valid identifier
			name = strings.ReplaceAll(name, " ", "_")
			name = strings.ReplaceAll(name, ".", "")
			name = strings.ReplaceAll(name, ",", "")
			if name == "" {
				name = "comment"
			}
		}
	}

	entity := &ast.Entity{
		Type:         ast.EntityComment,
		Name:         name,
		FullName:     name, // Comments are global
		Signature:    trimmed,
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)
	return nil
}

// parsePreprocessor parses a preprocessor directive (except #define)
func (p *Parser) parsePreprocessor(line string) error {
	trimmed := strings.TrimSpace(line)

	// Extract directive name for entity naming
	parts := strings.Fields(trimmed)
	name := trimmed // Default to full line
	if len(parts) > 0 {
		name = parts[0] // Use the directive name (e.g., "#pragma", "#include")
	}

	entity := &ast.Entity{
		Type:         ast.EntityPreprocessor,
		Name:         name,
		FullName:     name, // Preprocessor directives are global
		Signature:    trimmed,
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)
	return nil
}

// Regular expressions for identifying C++ constructs
var (
	defineRegex       = regexp.MustCompile(`^\s*#\s*define\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(.*)$`)
	preprocessorRegex = regexp.MustCompile(`^\s*#\s*(pragma|include|ifndef|ifdef|if|else|elif|endif|define|undef).*$`)
	templateRegex     = regexp.MustCompile(`^\s*template\s*<.*`)
	namespaceRegex    = regexp.MustCompile(`^\s*namespace\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\{?`)
	classRegex        = regexp.MustCompile(`^\s*.*?class\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?::\s*[^{]*?)?\s*\{?`)
	structRegex       = regexp.MustCompile(`^\s*.*?struct\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?::\s*[^{]*?)?\s*\{?`)
	enumRegex         = regexp.MustCompile(`^\s*.*?enum\s+(?:class\s+)?([a-zA-Z_][a-zA-Z0-9_]*)\s*(?::\s*[^{]*?)?\s*\{?`)
	// Function regex that captures the function name including template specializations
	functionRegex       = regexp.MustCompile(`(~?[a-zA-Z_][a-zA-Z0-9_]*(?:<[^>]*>)?)\s*\([^)]*\)\s*(?:->\s*[^{;]*)?(?:const\s*)?(?:override\s*)?(?:final\s*)?(?:noexcept\s*)?(?:\{|;)`)
	variableRegex       = regexp.MustCompile(`^\s*(?:(?:static|const|constexpr|mutable|extern)\s+)*[a-zA-Z_][a-zA-Z0-9_:<>*&\s]+\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:=.*?)?;`)
	typedefRegex        = regexp.MustCompile(`^\s*typedef\s+.*?\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*;`)
	usingRegex          = regexp.MustCompile(`^\s*using\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=`)
	usingNamespaceRegex = regexp.MustCompile(`^\s*using\s+namespace\s+([a-zA-Z_][a-zA-Z0-9_:]*)\s*;`)
)

// isDefine checks if line contains a #define directive
func (p *Parser) isDefine(line string) bool {
	return defineRegex.MatchString(line)
}

// isPreprocessor checks if line contains a preprocessor directive (except #define)
func (p *Parser) isPreprocessor(line string) bool {
	return preprocessorRegex.MatchString(line)
}

// isTemplate checks if line contains a template declaration
func (p *Parser) isTemplate(line string) bool {
	return templateRegex.MatchString(line)
}

// handleTemplate handles multi-line template declarations
func (p *Parser) handleTemplate(line string) error {
	templateLine := strings.TrimSpace(line)

	// Check if this is already a complete template (single line)
	openCount := strings.Count(templateLine, "<")
	closeCount := strings.Count(templateLine, ">")

	if openCount <= closeCount {
		// Template is complete on this line
		p.pendingTemplate = templateLine
		return nil
	}

	// Template spans multiple lines, collect until complete
	for p.current+1 < len(p.lines) {
		p.nextLine()
		nextLine := strings.TrimSpace(p.lines[p.current])
		templateLine += " " + nextLine
		closeCount += strings.Count(nextLine, ">")

		// Stop once we have balanced brackets
		if openCount <= closeCount {
			break
		}
	}

	p.pendingTemplate = templateLine
	return nil
}

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

	// Build complete signature, looking ahead for opening brace if needed
	signature := strings.TrimSpace(line)
	hasOpeningBrace := strings.Contains(line, "{")

	// If no opening brace on current line, check next line
	if !hasOpeningBrace && p.current+1 < len(p.lines) {
		nextLine := strings.TrimSpace(p.lines[p.current+1])
		if nextLine == "{" {
			signature += "\n{"
			hasOpeningBrace = true
		}
	}

	entity := &ast.Entity{
		Type:         ast.EntityNamespace,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    signature,
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// If opening brace was found, enter scope
	if hasOpeningBrace {
		p.enterScope(entity)
		// If brace was on next line, skip it during parsing
		if !strings.Contains(line, "{") && p.current+1 < len(p.lines) && strings.TrimSpace(p.lines[p.current+1]) == "{" {
			p.nextLine() // Skip the brace line
		}
	}

	return nil
}

// parseClass parses a class declaration
func (p *Parser) parseClass(line string) error {
	// Use resolved line for matching but preserve original for OriginalText
	resolvedLine := p.resolveDefines(line)
	matches := classRegex.FindStringSubmatch(resolvedLine)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse class: %s", line)
	}

	name := matches[1]

	// Build complete signature including any pending template
	signature := strings.TrimSpace(resolvedLine)
	if p.pendingTemplate != "" {
		// Trim the template line to remove any existing indentation
		templateLine := strings.TrimSpace(p.pendingTemplate)
		signature = templateLine + "\n" + signature
	}

	entity := &ast.Entity{
		Type:         ast.EntityClass,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    signature,
		AccessLevel:  ast.AccessPrivate, // Default for class
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)
	p.pendingTemplate = "" // Clear the pending template

	// If resolved line contains opening brace, enter scope
	if strings.Contains(resolvedLine, "{") {
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

	// Build complete signature including any pending template
	signature := strings.TrimSpace(line)
	if p.pendingTemplate != "" {
		// Trim the template line to remove any existing indentation
		templateLine := strings.TrimSpace(p.pendingTemplate)
		signature = templateLine + "\n" + signature
	}

	entity := &ast.Entity{
		Type:         ast.EntityStruct,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    signature,
		AccessLevel:  ast.AccessPublic, // Default for struct
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)
	p.pendingTemplate = "" // Clear the pending template

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
	// Use resolved line for matching but preserve original for OriginalText
	resolvedLine := p.resolveDefines(line)
	matches := functionRegex.FindStringSubmatch(resolvedLine)
	if len(matches) < 2 {
		return fmt.Errorf("failed to parse function: %s", line)
	}

	name := matches[1]
	originalName := name // Keep original for destructor check

	// For template specializations, extract just the base function name
	if strings.Contains(name, "<") {
		if idx := strings.Index(name, "<"); idx > 0 {
			name = name[:idx]
		}
	}

	entityType := ast.EntityFunction

	// Determine if it's a method (inside a class/struct)
	if p.getCurrentScope().Type == ast.EntityClass || p.getCurrentScope().Type == ast.EntityStruct {
		entityType = ast.EntityMethod

		// Check if it's a destructor first (check original name with ~)
		if strings.HasPrefix(originalName, "~") {
			entityType = ast.EntityDestructor
			name = strings.TrimPrefix(originalName, "~")
			// Also strip template args from destructor name if any
			if strings.Contains(name, "<") {
				if idx := strings.Index(name, "<"); idx > 0 {
					name = name[:idx]
				}
			}
		} else if name == p.getCurrentScope().Name {
			entityType = ast.EntityConstructor
		}
	}

	// Build complete signature including any pending template
	signature := strings.TrimSpace(resolvedLine)
	if p.pendingTemplate != "" {
		// Trim the template line to remove any existing indentation
		templateLine := strings.TrimSpace(p.pendingTemplate)
		signature = templateLine + "\n" + signature
	}

	entity := &ast.Entity{
		Type:         entityType,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    signature,
		IsStatic:     strings.Contains(resolvedLine, "static"),
		IsVirtual:    strings.Contains(resolvedLine, "virtual"),
		IsInline:     strings.Contains(resolvedLine, "inline"),
		IsConst:      strings.Contains(resolvedLine, ") const"),
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
	p.pendingTemplate = "" // Clear the pending template

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
	// Build complete signature including any pending template
	signature := strings.TrimSpace(line)
	if p.pendingTemplate != "" {
		// Trim the template line to remove any existing indentation
		templateLine := strings.TrimSpace(p.pendingTemplate)
		signature = templateLine + "\n" + signature
	}

	// Try regular using declaration first
	matches := usingRegex.FindStringSubmatch(line)
	if len(matches) >= 2 {
		name := matches[1]
		entity := &ast.Entity{
			Type:         ast.EntityUsing,
			Name:         name,
			FullName:     p.buildFullName(name),
			Signature:    signature,
			SourceRange:  p.getCurrentRange(),
			HeaderRange:  p.getCurrentRange(),
			OriginalText: line,
			Children:     make([]*ast.Entity, 0),
		}

		p.addEntity(entity)
		p.pendingTemplate = "" // Clear the pending template
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
			Signature:    signature,
			SourceRange:  p.getCurrentRange(),
			HeaderRange:  p.getCurrentRange(),
			OriginalText: line,
			Children:     make([]*ast.Entity, 0),
		}

		p.addEntity(entity)
		p.pendingTemplate = "" // Clear the pending template
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

	// Create an entity for the access specifier to preserve it in reconstruction
	entity := &ast.Entity{
		Type:         ast.EntityAccessSpecifier,
		Name:         strings.TrimSuffix(line, ":"), // Remove the colon for the name
		FullName:     line,
		Signature:    line,
		AccessLevel:  accessLevel,
		SourceRange:  p.getCurrentRange(),
		HeaderRange:  p.getCurrentRange(),
		OriginalText: line,
		Children:     make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	return nil
}

// Define resolution methods

// resolveDefines resolves all defines in a line
func (p *Parser) resolveDefines(line string) string {
	resolved := line

	// Sort defines by length (longest first) to avoid partial replacements
	// e.g., if we have MAX_SIZE and MAX_SIZE_LIMIT, we want to replace MAX_SIZE_LIMIT first
	var defineNames []string
	for name := range p.defines {
		defineNames = append(defineNames, name)
	}

	// Sort by length descending
	for i := 0; i < len(defineNames)-1; i++ {
		for j := i + 1; j < len(defineNames); j++ {
			if len(defineNames[i]) < len(defineNames[j]) {
				defineNames[i], defineNames[j] = defineNames[j], defineNames[i]
			}
		}
	}

	// Replace defines in order of length (longest first)
	for _, name := range defineNames {
		value := p.defines[name]
		// Only replace whole words, not partial matches
		resolved = p.replaceWholeWord(resolved, name, value)
	}

	return resolved
}

// replaceWholeWord replaces whole word occurrences only
func (p *Parser) replaceWholeWord(text, oldWord, newWord string) string {
	if oldWord == "" {
		return text
	}

	result := ""
	i := 0
	oldLen := len(oldWord)

	for i < len(text) {
		// Find the next occurrence of oldWord
		index := strings.Index(text[i:], oldWord)
		if index == -1 {
			// No more occurrences, append the rest
			result += text[i:]
			break
		}

		// Adjust index to absolute position
		index += i

		// Check if it's a whole word (not part of another identifier)
		isWholeWord := true

		// Check character before
		if index > 0 {
			prevChar := text[index-1]
			if isAlphaNumericOrUnderscore(prevChar) {
				isWholeWord = false
			}
		}

		// Check character after
		if index+oldLen < len(text) {
			nextChar := text[index+oldLen]
			if isAlphaNumericOrUnderscore(nextChar) {
				isWholeWord = false
			}
		}

		if isWholeWord {
			// Add text before the match
			result += text[i:index]
			// Add the replacement
			result += newWord
			// Move past the replaced word
			i = index + oldLen
		} else {
			// Not a whole word, add the character and continue
			result += text[i : index+1]
			i = index + 1
		}
	}

	return result
}

// Helper functions

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// isAlphaNumericOrUnderscore checks if character is alphanumeric or underscore
func isAlphaNumericOrUnderscore(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
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

	// If the signature was resolved from defines, store both versions
	if entity.Signature != entity.OriginalText {
		// Store the resolved signature in the main signature field
		// The original text is already stored in OriginalText
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
