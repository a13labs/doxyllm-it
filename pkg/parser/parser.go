// Package parser implements a token-driven C++ header file parser
package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
)

// Parser implements a token-driven parser for C++ headers
type Parser struct {
	tokens         []Token
	current        int
	tree           *ast.ScopeTree
	scopeStack     []*ast.Entity
	accessStack    []ast.AccessLevel
	defines        map[string]string
	pendingComment *ast.DoxygenComment // Comment waiting to be associated with next entity
}

// NewTokenParser creates a new token-driven parser
func NewTokenParser() *Parser {
	return &Parser{
		defines: make(map[string]string),
	}
}

// Parse parses tokens into an AST
func (p *Parser) Parse(filename, content string) (*ast.ScopeTree, error) {
	// Initialize the tree
	p.tree = ast.NewScopeTree(filename, content)
	p.scopeStack = []*ast.Entity{p.tree.Root}
	p.accessStack = []ast.AccessLevel{ast.AccessPublic} // Global scope is public

	// Tokenize the input
	tokenizer := NewTokenizer(content)
	p.tokens = tokenizer.Tokenize()
	p.current = 0

	// Check for tokenizer errors
	if tokenizer.HasErrors() {
		errors := tokenizer.GetErrors()
		if len(errors) > 0 {
			return nil, fmt.Errorf("tokenizer error: %s", errors[0].Value)
		}
	}

	// Parse the tokens
	for !p.isAtEnd() {
		if err := p.parseTopLevel(); err != nil {
			return nil, err
		}
	}

	return p.tree, nil
}

// parseTopLevel parses top-level declarations
func (p *Parser) parseTopLevel() error {
	// Skip whitespace and newlines
	p.skipWhitespaceAndNewlines()

	if p.isAtEnd() {
		return nil
	}

	// Handle different token types
	token := p.peek()

	switch token.Type {
	case TokenHash:
		return p.parsePreprocessor()
	case TokenLineComment, TokenBlockComment, TokenDoxygenComment:
		return p.parseComment()
	case TokenTemplate:
		return p.parseTemplate()
	case TokenNamespace:
		return p.parseNamespace()
	case TokenClass:
		return p.parseClass()
	case TokenStruct:
		return p.parseStruct()
	case TokenEnum:
		return p.parseEnum()
	case TokenTypedef:
		return p.parseTypedef()
	case TokenUsing:
		return p.parseUsing()
	case TokenPublic, TokenPrivate, TokenProtected:
		return p.parseAccessSpecifier()
	case TokenRightBrace:
		return p.parseCloseBrace()
	case TokenIdentifier:
		// Check if this identifier is a macro that should be resolved
		if _, exists := p.defines[token.Value]; exists {
			// Look ahead to see if after macro resolution we have a keyword
			nextTokenIndex := p.current + 1
			for nextTokenIndex < len(p.tokens) && p.tokens[nextTokenIndex].Type == TokenWhitespace {
				nextTokenIndex++
			}
			if nextTokenIndex < len(p.tokens) {
				nextToken := p.tokens[nextTokenIndex]
				switch nextToken.Type {
				case TokenClass:
					return p.parseClassWithMacro()
				case TokenStruct:
					return p.parseStructWithMacro()
				case TokenEnum:
					return p.parseEnumWithMacro()
				}
			}
		}
		// Fall through to default if not a macro or not followed by keyword
		return p.parseFunctionOrVariable()
	default:
		// Try to parse as function or variable
		return p.parseFunctionOrVariable()
	}
}

// parsePreprocessor handles preprocessor directives
func (p *Parser) parsePreprocessor() error {
	start := p.current
	p.advance() // consume '#'

	p.skipWhitespace()

	if p.isAtEnd() {
		return nil
	}

	directive := p.peek()

	if directive.Type == TokenIdentifier && directive.Value == "define" {
		return p.parseDefine(start)
	}

	// Other preprocessor directives
	return p.parseOtherPreprocessor(start)
}

// parseDefine handles #define directives
func (p *Parser) parseDefine(start int) error {
	p.advance() // consume 'define'
	p.skipWhitespace()

	if p.isAtEnd() {
		return fmt.Errorf("expected identifier after #define")
	}

	nameToken := p.peek()
	if nameToken.Type != TokenIdentifier {
		return fmt.Errorf("expected identifier after #define, got %v", nameToken.Type)
	}
	p.advance()

	// Collect the definition value until end of line or end of file
	var value strings.Builder
	depth := 0
	lastWasSpace := false

	for !p.isAtEnd() {
		token := p.peek()

		// Handle multiline macros with backslash continuation
		if token.Type == TokenBackslash {
			p.advance()
			// Skip the backslash and any following whitespace/newline
			p.skipWhitespace()
			if !p.isAtEnd() && p.peek().Type == TokenNewline {
				p.advance()
			}
			value.WriteString(" ") // Replace backslash-newline with space
			lastWasSpace = true
			continue
		}

		if token.Type == TokenNewline && depth == 0 {
			break
		}

		// Track brace depth for complex macros
		if token.Type == TokenLeftBrace || token.Type == TokenLeftParen {
			depth++
		} else if token.Type == TokenRightBrace || token.Type == TokenRightParen {
			depth--
		}

		// Normalize whitespace - collapse multiple spaces into one
		if token.Type == TokenWhitespace {
			if !lastWasSpace {
				value.WriteString(" ")
				lastWasSpace = true
			}
		} else {
			value.WriteString(token.Value)
			lastWasSpace = false
		}
		p.advance()
	}

	// Store the define
	defineName := nameToken.Value
	defineValue := strings.TrimSpace(value.String())
	p.defines[defineName] = defineValue

	// Create preprocessor entity
	entity := &ast.Entity{
		Type:        ast.EntityPreprocessor,
		Name:        defineName,
		FullName:    defineName,
		Signature:   fmt.Sprintf("#define %s %s", defineName, defineValue),
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// parseOtherPreprocessor handles other preprocessor directives
func (p *Parser) parseOtherPreprocessor(start int) error {
	// Consume until end of line
	var content strings.Builder
	content.WriteString("#")

	for !p.isAtEnd() && p.peek().Type != TokenNewline {
		content.WriteString(p.peek().Value)
		p.advance()
	}

	entity := &ast.Entity{
		Type:        ast.EntityPreprocessor,
		Name:        strings.TrimSpace(content.String()),
		FullName:    strings.TrimSpace(content.String()),
		Signature:   content.String(),
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// resolveDefine recursively resolves defines to their final values
func (p *Parser) resolveDefine(name string) string {
	visited := make(map[string]bool)
	return p.resolveDefineRecursive(name, visited)
}

// resolveDefineRecursive does the recursive resolution with cycle detection
func (p *Parser) resolveDefineRecursive(name string, visited map[string]bool) string {
	// Prevent infinite loops in circular definitions
	if visited[name] {
		return name
	}

	defineValue, exists := p.defines[name]
	if !exists {
		return name
	}

	visited[name] = true

	// Check if the define value is also a define
	if _, isDefine := p.defines[defineValue]; isDefine {
		return p.resolveDefineRecursive(defineValue, visited)
	}

	return defineValue
}

// parseComment handles comment parsing
func (p *Parser) parseComment() error {
	start := p.current
	token := p.advance()

	// Check if this is a Doxygen comment
	content := strings.TrimSpace(token.Value)
	if p.isDoxygenComment(content) {
		// Parse as Doxygen comment and store as pending
		p.pendingComment = p.parseDoxygenComment(content)
		return nil
	}

	// For regular comments, create comment entities
	name := "comment"

	// Try to extract first few words as name
	if len(content) > 2 {
		// Remove comment markers
		if strings.HasPrefix(content, "//") {
			content = strings.TrimSpace(content[2:])
		} else if strings.HasPrefix(content, "/*") && strings.HasSuffix(content, "*/") {
			content = strings.TrimSpace(content[2 : len(content)-2])
		}

		words := strings.Fields(content)
		if len(words) > 0 {
			if len(words) == 1 {
				name = words[0]
			} else {
				name = strings.Join(words[:min(3, len(words))], " ")
			}
		}
	}

	entity := &ast.Entity{
		Type:        ast.EntityComment,
		Name:        name,
		FullName:    name,
		Signature:   token.Value,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// isDoxygenComment checks if a comment is a Doxygen comment
func (p *Parser) isDoxygenComment(content string) bool {
	// Check for Doxygen comment patterns
	return strings.HasPrefix(content, "/**") ||
		strings.HasPrefix(content, "///") ||
		strings.HasPrefix(content, "//!")
}

// parseDoxygenComment parses a Doxygen comment string
func (p *Parser) parseDoxygenComment(content string) *ast.DoxygenComment {
	// Reuse the existing ParseDoxygenComment function from parser.go
	return ParseDoxygenComment(content)
}

// parseTemplate handles template declarations
func (p *Parser) parseTemplate() error {
	start := p.current
	p.advance() // consume 'template'

	// Skip whitespace and newlines
	p.skipWhitespaceAndNewlines()

	// Parse template parameters
	if !p.match(TokenLess) {
		return fmt.Errorf("expected '<' after template")
	}

	depth := 1
	var templateParams strings.Builder
	templateParams.WriteString("<")

	for !p.isAtEnd() && depth > 0 {
		token := p.peek()
		if token.Type == TokenLess {
			depth++
		} else if token.Type == TokenGreater {
			depth--
		}

		templateParams.WriteString(token.Value)
		p.advance()
	}

	p.skipWhitespaceAndNewlines()

	// Now parse the templated entity
	return p.parseTemplatedEntity(start, templateParams.String())
}

// parseTemplatedEntity parses the entity that follows a template declaration
func (p *Parser) parseTemplatedEntity(templateStart int, templateParams string) error {
	if p.isAtEnd() {
		return fmt.Errorf("expected entity after template")
	}

	token := p.peek()

	switch token.Type {
	case TokenClass:
		return p.parseTemplatedClass(templateStart, templateParams)
	case TokenStruct:
		return p.parseTemplatedStruct(templateStart, templateParams)
	case TokenUsing:
		return p.parseTemplatedUsing(templateStart, templateParams)
	default:
		// Template function
		return p.parseTemplatedFunction(templateStart, templateParams)
	}
}

// parseNamespace handles namespace declarations
func (p *Parser) parseNamespace() error {
	start := p.current
	p.advance() // consume 'namespace'

	p.skipWhitespace()

	if p.isAtEnd() || !p.isValidIdentifierToken(p.peek()) {
		return fmt.Errorf("expected namespace name")
	}

	// Parse namespace name (could be nested like mgl::io)
	var nameBuilder strings.Builder
	nameBuilder.WriteString(p.advance().Value) // first identifier

	// Check for :: followed by more identifiers (nested namespace)
	for !p.isAtEnd() {
		p.skipWhitespace()
		if p.peek().Type == TokenDoubleColon {
			nameBuilder.WriteString(p.advance().Value) // add ::
			p.skipWhitespace()
			if p.isValidIdentifierToken(p.peek()) {
				nameBuilder.WriteString(p.advance().Value) // add next identifier
			} else {
				break
			}
		} else {
			break
		}
	}

	namespaceName := nameBuilder.String()

	p.skipWhitespace()

	// Build signature
	signature := fmt.Sprintf("namespace %s", namespaceName)

	// Look for opening brace, which might be on the same line or next line
	if p.match(TokenLeftBrace) {
		signature += " {"
	} else {
		// The brace might be on the next line, so we need to look ahead
		// Save current position to check for brace
		checkpoint := p.current

		// Skip any whitespace/newlines to find the brace
		for !p.isAtEnd() && (p.peek().Type == TokenWhitespace || p.peek().Type == TokenNewline) {
			p.advance()
		}

		if !p.isAtEnd() && p.peek().Type == TokenLeftBrace {
			p.advance() // consume the brace
			signature += " {"
		} else {
			// No brace found, restore position
			p.current = checkpoint
		}
	}

	entity := &ast.Entity{
		Type:        ast.EntityNamespace,
		Name:        namespaceName,
		FullName:    p.buildFullName(namespaceName),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// Enter scope if we found opening brace
	if strings.Contains(signature, "{") {
		p.enterScope(entity)
	}

	return nil
}

// parseClass handles class declarations
func (p *Parser) parseClass() error {
	return p.parseClassOrStruct(ast.EntityClass)
}

// parseStruct handles struct declarations
func (p *Parser) parseStruct() error {
	return p.parseClassOrStruct(ast.EntityStruct)
}

// parseClassOrStruct handles both class and struct declarations
func (p *Parser) parseClassOrStruct(entityType ast.EntityType) error {
	start := p.current
	keyword := p.advance() // consume 'class' or 'struct'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected %s name", keyword.Value)
	}

	nameToken := p.advance()

	// Parse inheritance if present
	inheritance := ""
	p.skipWhitespace()
	if p.match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.skipWhitespace()

	// Build signature
	signature := fmt.Sprintf("%s %s", keyword.Value, nameToken.Value)
	if inheritance != "" {
		signature += " : " + inheritance
	}
	if p.match(TokenLeftBrace) {
		signature += " {"
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// Enter scope if we found opening brace
	if strings.Contains(signature, "{") {
		p.enterScope(entity)
		// Set default access level for the new scope
		if entityType == ast.EntityClass {
			p.accessStack = append(p.accessStack, ast.AccessPrivate) // class default
		} else {
			p.accessStack = append(p.accessStack, ast.AccessPublic) // struct default
		}
	}

	return nil
}

// parseTemplatedClass handles template class declarations
func (p *Parser) parseTemplatedClass(templateStart int, templateParams string) error {
	p.advance() // consume 'class'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected class name")
	}

	nameToken := p.advance()

	// Parse inheritance if present
	inheritance := ""
	p.skipWhitespace()
	if p.match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.skipWhitespace()

	// Build signature with template
	signature := fmt.Sprintf("template %s class %s", templateParams, nameToken.Value)
	if inheritance != "" {
		signature += " : " + inheritance
	}
	if p.match(TokenLeftBrace) {
		signature += " {"
	} else if p.match(TokenSemicolon) {
		// Forward declaration - consume the semicolon but don't add it to signature
		// The semicolon is implied for class declarations
	}

	entity := &ast.Entity{
		Type:        ast.EntityClass,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		IsTemplate:  true,
		SourceRange: p.getRangeFromTokens(templateStart, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// Enter scope if we found opening brace
	if strings.Contains(signature, "{") {
		p.enterScope(entity)
		p.accessStack = append(p.accessStack, ast.AccessPrivate) // class default
	}

	return nil
}

// parseTemplatedStruct handles template struct declarations
func (p *Parser) parseTemplatedStruct(templateStart int, templateParams string) error {
	p.advance() // consume 'struct'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected struct name")
	}

	nameToken := p.advance()

	// Parse inheritance if present
	inheritance := ""
	p.skipWhitespace()
	if p.match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.skipWhitespace()

	// Build signature with template
	signature := fmt.Sprintf("template %s struct %s", templateParams, nameToken.Value)
	if inheritance != "" {
		signature += " : " + inheritance
	}
	if p.match(TokenLeftBrace) {
		signature += " {"
	} else if p.match(TokenSemicolon) {
		// Forward declaration - consume the semicolon but don't add it to signature
	}

	entity := &ast.Entity{
		Type:        ast.EntityStruct,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		IsTemplate:  true,
		SourceRange: p.getRangeFromTokens(templateStart, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// Enter scope if we found opening brace
	if strings.Contains(signature, "{") {
		p.enterScope(entity)
		p.accessStack = append(p.accessStack, ast.AccessPublic) // struct default
	}

	return nil
}

// parseTemplatedUsing handles template using declarations
func (p *Parser) parseTemplatedUsing(templateStart int, templateParams string) error {
	p.advance() // consume 'using'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected identifier after using")
	}

	nameToken := p.advance()

	p.skipWhitespace()

	if !p.match(TokenEquals) {
		return fmt.Errorf("expected '=' in using declaration")
	}

	// Parse the rest until semicolon
	var typeValue strings.Builder
	for !p.isAtEnd() && p.peek().Type != TokenSemicolon {
		typeValue.WriteString(p.peek().Value)
		p.advance()
	}

	if p.match(TokenSemicolon) {
		// consumed semicolon
	}

	signature := fmt.Sprintf("template %s using %s = %s", templateParams, nameToken.Value, typeValue.String())

	entity := &ast.Entity{
		Type:        ast.EntityUsing,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		IsTemplate:  true,
		SourceRange: p.getRangeFromTokens(templateStart, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// parseTemplatedFunction handles template function declarations
func (p *Parser) parseTemplatedFunction(templateStart int, templateParams string) error {
	// Parse the function signature
	signature, name, isMethod, err := p.parseFunctionSignature()
	if err != nil {
		return err
	}

	// Check if there's a function body after the signature
	var bodyRange *ast.Range
	var bodyText string
	p.skipWhitespace()
	if !p.isAtEnd() && p.peek().Type == TokenLeftBrace {
		// This function has a body - track its range and content for the formatter
		bodyStart := p.current
		braceDepth := 1
		var bodyTokens []Token
		bodyTokens = append(bodyTokens, p.advance()) // consume opening brace

		for !p.isAtEnd() && braceDepth > 0 {
			token := p.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			bodyTokens = append(bodyTokens, p.advance())
		}

		if braceDepth == 0 {
			rangeValue := p.getRangeFromTokens(bodyStart, p.current-1)
			bodyRange = &rangeValue

			// Reconstruct body text from tokens
			var bodyBuilder strings.Builder
			for _, token := range bodyTokens {
				bodyBuilder.WriteString(token.Value)
			}
			bodyText = bodyBuilder.String()
		}
	}

	// Add template to signature
	fullSignature := fmt.Sprintf("template %s %s", templateParams, signature)

	entityType := ast.EntityFunction
	if isMethod {
		entityType = ast.EntityMethod
	}

	entity := &ast.Entity{
		Type:         entityType,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    fullSignature,
		AccessLevel:  p.getCurrentAccessLevel(),
		IsTemplate:   true,
		SourceRange:  p.getRangeFromTokens(templateStart, p.current-1),
		BodyRange:    bodyRange,
		OriginalText: bodyText,
	}

	p.addEntity(entity)
	return nil
}

// parseEnum handles enum declarations
func (p *Parser) parseEnum() error {
	start := p.current
	p.advance() // consume 'enum'

	p.skipWhitespace()

	// Check for 'class' or 'struct' after enum
	isScoped := false
	if !p.isAtEnd() && (p.peek().Type == TokenClass || p.peek().Type == TokenStruct) {
		isScoped = true
		p.advance()
		p.skipWhitespace()
	}

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected enum name")
	}

	nameToken := p.advance()

	p.skipWhitespace()

	// Parse underlying type if present
	underlyingType := ""
	if p.match(TokenColon) {
		underlyingType = p.parseType()
	}

	p.skipWhitespace()

	// Build signature and handle body
	signature := "enum"
	if isScoped {
		signature += " class"
	}
	signature += " " + nameToken.Value
	if underlyingType != "" {
		signature += " : " + underlyingType
	}

	// Check if this enum has a body (opening brace)
	if p.match(TokenLeftBrace) {
		signature += " {"

		// Consume the entire enum body
		braceDepth := 1
		for !p.isAtEnd() && braceDepth > 0 {
			token := p.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			signature += token.Value
			p.advance()
		}

		// Consume optional semicolon after enum body
		if !p.isAtEnd() && p.peek().Type == TokenSemicolon {
			signature += ";"
			p.advance()
		}
	}

	entity := &ast.Entity{
		Type:        ast.EntityEnum,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	return nil
}

// parseTypedef handles typedef declarations
func (p *Parser) parseTypedef() error {
	start := p.current
	p.advance() // consume 'typedef'

	// Parse until we find the identifier and semicolon
	var signature strings.Builder
	signature.WriteString("typedef ")

	var name string
	lastIdentifier := ""

	for !p.isAtEnd() && p.peek().Type != TokenSemicolon {
		token := p.peek()
		signature.WriteString(token.Value)

		if token.Type == TokenIdentifier {
			lastIdentifier = token.Value
		}

		p.advance()
	}

	if p.match(TokenSemicolon) {
		signature.WriteString(";")
	}

	name = lastIdentifier // The last identifier is typically the typedef name

	entity := &ast.Entity{
		Type:        ast.EntityTypedef,
		Name:        name,
		FullName:    p.buildFullName(name),
		Signature:   signature.String(),
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// parseUsing handles using declarations
func (p *Parser) parseUsing() error {
	start := p.current
	p.advance() // consume 'using'

	p.skipWhitespace()

	if p.isAtEnd() {
		return fmt.Errorf("expected identifier after using")
	}

	// Check for 'namespace' keyword
	if p.peek().Type == TokenNamespace {
		return p.parseUsingNamespace(start)
	}

	// Regular using declaration
	if p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected identifier after using")
	}

	nameToken := p.advance()

	p.skipWhitespace()

	if !p.match(TokenEquals) {
		return fmt.Errorf("expected '=' in using declaration")
	}

	// Parse the rest until semicolon
	var typeValue strings.Builder
	for !p.isAtEnd() && p.peek().Type != TokenSemicolon {
		typeValue.WriteString(p.peek().Value)
		p.advance()
	}

	if p.match(TokenSemicolon) {
		// consumed semicolon
	}

	signature := fmt.Sprintf("using %s = %s", nameToken.Value, typeValue.String())

	entity := &ast.Entity{
		Type:        ast.EntityUsing,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// parseUsingNamespace handles using namespace declarations
func (p *Parser) parseUsingNamespace(start int) error {
	p.advance() // consume 'namespace'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected namespace name after using namespace")
	}

	nameToken := p.advance()

	// Parse qualified namespace name
	var namespaceName strings.Builder
	namespaceName.WriteString(nameToken.Value)

	for !p.isAtEnd() && p.peek().Type == TokenDoubleColon {
		namespaceName.WriteString("::")
		p.advance()

		if p.isAtEnd() || p.peek().Type != TokenIdentifier {
			break
		}

		namespaceName.WriteString(p.advance().Value)
	}

	if p.match(TokenSemicolon) {
		// consumed semicolon
	}

	signature := fmt.Sprintf("using namespace %s", namespaceName.String())

	entity := &ast.Entity{
		Type:        ast.EntityUsing,
		Name:        namespaceName.String(),
		FullName:    namespaceName.String(),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// parseAccessSpecifier handles access specifier declarations
func (p *Parser) parseAccessSpecifier() error {
	start := p.current
	accessToken := p.advance()

	if !p.match(TokenColon) {
		return fmt.Errorf("expected ':' after access specifier")
	}

	// Update current access level
	var accessLevel ast.AccessLevel
	switch accessToken.Value {
	case "public":
		accessLevel = ast.AccessPublic
	case "private":
		accessLevel = ast.AccessPrivate
	case "protected":
		accessLevel = ast.AccessProtected
	}

	// Update the access stack for current scope
	if len(p.accessStack) > 0 {
		p.accessStack[len(p.accessStack)-1] = accessLevel
	}

	// Create access specifier entity
	entity := &ast.Entity{
		Type:        ast.EntityAccessSpecifier,
		Name:        accessToken.Value,
		FullName:    accessToken.Value,
		Signature:   accessToken.Value + ":",
		AccessLevel: accessLevel,
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// parseCloseBrace handles closing braces
func (p *Parser) parseCloseBrace() error {
	p.advance() // consume '}'

	// Check for optional semicolon after brace (for class/struct)
	if !p.isAtEnd() && p.peek().Type == TokenSemicolon {
		p.advance()
	}

	// Exit current scope
	p.exitScope()

	return nil
}

// parseFunctionOrVariable attempts to parse function or variable declarations
func (p *Parser) parseFunctionOrVariable() error {
	// Collect tokens until we can determine what this is
	checkpoint := p.current

	// Skip storage specifiers, cv-qualifiers, etc.
	p.skipSpecifiers()

	// Look for function pattern: type name(args) or name(args)
	if p.isFunction() {
		p.current = checkpoint
		return p.parseFunction()
	}

	// Otherwise, try as variable
	p.current = checkpoint
	return p.parseVariable()
}

// parseFunction handles function declarations
func (p *Parser) parseFunction() error {
	start := p.current

	// Parse specifiers and attributes first
	var isStatic, isInline, isVirtual, isConst bool

	// Parse function specifiers
	for !p.isAtEnd() {
		token := p.peek()
		if token.Type == TokenStatic {
			isStatic = true
			p.advance()
			p.skipWhitespace()
		} else if token.Type == TokenInline {
			isInline = true
			p.advance()
			p.skipWhitespace()
		} else if token.Type == TokenVirtual {
			isVirtual = true
			p.advance()
			p.skipWhitespace()
		} else {
			break
		}
	}

	signature, name, isMethod, err := p.parseFunctionSignature()
	if err != nil {
		return err
	}

	// Check if there's a function body after the signature
	var bodyRange *ast.Range
	var bodyText string
	p.skipWhitespace()
	if !p.isAtEnd() && p.peek().Type == TokenLeftBrace {
		// This function has a body - track its range and content for the formatter
		bodyStart := p.current
		braceDepth := 1
		var bodyTokens []Token
		bodyTokens = append(bodyTokens, p.advance()) // consume opening brace

		for !p.isAtEnd() && braceDepth > 0 {
			token := p.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			bodyTokens = append(bodyTokens, p.advance())
		}

		if braceDepth == 0 {
			rangeValue := p.getRangeFromTokens(bodyStart, p.current-1)
			bodyRange = &rangeValue

			// Reconstruct body text from tokens
			var bodyBuilder strings.Builder
			for _, token := range bodyTokens {
				bodyBuilder.WriteString(token.Value)
			}
			bodyText = bodyBuilder.String()
		}
	}

	// Check for const after function signature
	if strings.Contains(signature, ") const") {
		isConst = true
	}

	entityType := ast.EntityFunction
	if isMethod {
		entityType = ast.EntityMethod

		// Check for special method types
		if strings.Contains(signature, "~") {
			// Detect destructor by checking for ~ in signature
			entityType = ast.EntityDestructor
		} else if name == p.getCurrentScope().Name {
			entityType = ast.EntityConstructor
		}
	}

	entity := &ast.Entity{
		Type:         entityType,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    signature,
		AccessLevel:  p.getCurrentAccessLevel(),
		IsStatic:     isStatic,
		IsInline:     isInline,
		IsVirtual:    isVirtual,
		IsConst:      isConst,
		SourceRange:  p.getRangeFromTokens(start, p.current-1),
		BodyRange:    bodyRange,
		OriginalText: bodyText,
	}

	p.addEntity(entity)
	return nil
}

// parseVariable handles variable declarations
func (p *Parser) parseVariable() error {
	start := p.current

	// Parse specifiers first
	var isStatic, isConst bool

	// Parse variable specifiers
	for !p.isAtEnd() {
		token := p.peek()
		if token.Type == TokenStatic {
			isStatic = true
			p.advance()
			p.skipWhitespace()
		} else if token.Type == TokenConst || token.Type == TokenConstexpr {
			isConst = true
			p.advance()
			p.skipWhitespace()
		} else if token.Type == TokenExtern || token.Type == TokenMutable {
			p.advance() // consume but don't track these for now
			p.skipWhitespace()
		} else {
			break
		}
	}

	// Parse until semicolon
	var signature strings.Builder
	var name string
	lastIdentifier := ""

	for !p.isAtEnd() && p.peek().Type != TokenSemicolon {
		token := p.peek()

		// Resolve defines in token values
		tokenValue := token.Value
		if token.Type == TokenIdentifier {
			tokenValue = p.resolveDefine(token.Value)
		}

		signature.WriteString(tokenValue)

		if token.Type == TokenIdentifier {
			lastIdentifier = token.Value
		}

		p.advance()
	}

	if p.match(TokenSemicolon) {
		signature.WriteString(";")
	}

	name = lastIdentifier

	entityType := ast.EntityVariable
	if p.isInsideClass() {
		entityType = ast.EntityField
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        name,
		FullName:    p.buildFullName(name),
		Signature:   signature.String(),
		AccessLevel: p.getCurrentAccessLevel(),
		IsStatic:    isStatic,
		IsConst:     isConst,
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// Helper methods

// advance returns the current token and moves to the next
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// isAtEnd checks if we're at the end of tokens
func (p *Parser) isAtEnd() bool {
	return p.current >= len(p.tokens) || p.peek().Type == TokenEOF
}

// isValidIdentifierToken checks if a token can be used as an identifier
// This includes actual identifiers and keywords that can be used as names
func (p *Parser) isValidIdentifierToken(token Token) bool {
	// Regular identifiers are always valid
	if token.Type == TokenIdentifier {
		return true
	}

	// Some keywords can also be used as identifiers in certain contexts
	// (like namespace names, class names, etc.)
	switch token.Type {
	case TokenVoid, TokenBool, TokenChar, TokenShort, TokenInt, TokenLong,
		TokenFloat, TokenDouble, TokenSigned, TokenUnsigned, TokenAuto:
		return true
	default:
		return false
	}
}

// peek returns the current token without advancing
func (p *Parser) peek() Token {
	if p.current >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.current]
}

// previous returns the previous token
func (p *Parser) previous() Token {
	if p.current <= 0 {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.current-1]
}

// match checks if current token matches any of the given types
func (p *Parser) match(types ...TokenType) bool {
	for _, tokenType := range types {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}
	return false
}

// check returns true if current token is of given type
func (p *Parser) check(tokenType TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
}

// skipWhitespace skips whitespace tokens
func (p *Parser) skipWhitespace() {
	for !p.isAtEnd() && p.peek().Type == TokenWhitespace {
		p.advance()
	}
}

// skipWhitespaceAndNewlines skips whitespace and newline tokens
func (p *Parser) skipWhitespaceAndNewlines() {
	for !p.isAtEnd() && (p.peek().Type == TokenWhitespace || p.peek().Type == TokenNewline) {
		p.advance()
	}
}

// skipSpecifiers skips storage and cv specifiers
func (p *Parser) skipSpecifiers() {
	specifiers := []TokenType{
		TokenStatic, TokenExtern, TokenInline, TokenVirtual,
		TokenConst, TokenConstexpr, TokenMutable, TokenVolatile,
		TokenExplicit, TokenFriend,
	}

	for !p.isAtEnd() {
		found := false
		for _, spec := range specifiers {
			if p.check(spec) {
				p.advance()
				p.skipWhitespace()
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
}

// isFunction tries to determine if current position is a function
func (p *Parser) isFunction() bool {
	saved := p.current
	defer func() { p.current = saved }()

	// Skip specifiers first
	for !p.isAtEnd() {
		token := p.peek()
		if token.Type == TokenStatic || token.Type == TokenInline || token.Type == TokenVirtual ||
			token.Type == TokenExtern || token.Type == TokenConst || token.Type == TokenConstexpr {
			p.advance()
			p.skipWhitespaceAndNewlines()
		} else {
			break
		}
	}

	// Look for pattern: [return_type] function_name(
	identifiersSeen := 0

	for !p.isAtEnd() && identifiersSeen < 6 { // increased limit to handle more complex types
		token := p.peek()

		if token.Type == TokenVoid || token.Type == TokenInt || token.Type == TokenDouble ||
			token.Type == TokenChar || token.Type == TokenFloat || token.Type == TokenBool ||
			token.Type == TokenIdentifier {
			identifiersSeen++
			p.advance()
			p.skipWhitespaceAndNewlines() // Handle newlines between return type and function name
		} else if token.Type == TokenLeftParen {
			// Found opening parenthesis, this looks like a function
			return identifiersSeen >= 1
		} else if token.Type == TokenDoubleColon {
			p.advance()
			p.skipWhitespaceAndNewlines()
		} else if token.Type == TokenStar || token.Type == TokenAmpersand {
			p.advance() // Skip pointer/reference indicators
			p.skipWhitespaceAndNewlines()
		} else if token.Type == TokenLess {
			// Skip template parameters
			depth := 1
			p.advance()
			for !p.isAtEnd() && depth > 0 {
				t := p.peek()
				if t.Type == TokenLess {
					depth++
				} else if t.Type == TokenGreater {
					depth--
				}
				p.advance()
			}
			p.skipWhitespaceAndNewlines()
		} else if token.Type == TokenTilde {
			// Destructor
			p.advance()
			p.skipWhitespaceAndNewlines()
		} else {
			break
		}
	}

	return false
}

// parseFunctionSignature parses a complete function signature
func (p *Parser) parseFunctionSignature() (signature string, name string, isMethod bool, err error) {
	var sig strings.Builder
	var funcName string

	// Parse until we find the function name and parameters
	depth := 0
	beforeParen := true
	isDestructor := false

	for !p.isAtEnd() {
		token := p.peek()

		if token.Type == TokenLeftParen && depth == 0 && beforeParen {
			beforeParen = false
		}

		if token.Type == TokenLeftParen {
			depth++
		} else if token.Type == TokenRightParen {
			depth--
		}

		// Stop at semicolon (declaration) or opening brace (definition)
		if token.Type == TokenSemicolon && depth == 0 {
			break
		} else if token.Type == TokenLeftBrace && depth == 0 {
			// This is a function definition with a body
			// Don't include the body in the signature - the formatter handles it separately
			break
		}

		// Resolve defines in token values
		tokenValue := token.Value
		if token.Type == TokenIdentifier {
			tokenValue = p.resolveDefine(token.Value)
		}

		sig.WriteString(tokenValue)

		// Track the function name (last identifier before '(')
		if token.Type == TokenIdentifier && beforeParen {
			funcName = token.Value
		} else if token.Type == TokenTilde && beforeParen {
			// Handle destructor names
			isDestructor = true
			if !p.isAtEnd() && p.tokens[p.current+1].Type == TokenIdentifier {
				funcName = p.tokens[p.current+1].Value // Just the class name, not including ~
			}
		}

		p.advance()
	}

	// Consume semicolon if present
	if !p.isAtEnd() && p.peek().Type == TokenSemicolon {
		sig.WriteString(";")
		p.advance()
	}

	isMethod = p.isInsideClass()

	// If we didn't find a function name, try to extract it from the signature
	if funcName == "" && sig.Len() > 0 {
		// Look for the pattern before the opening parenthesis
		sigStr := sig.String()
		parenIndex := strings.Index(sigStr, "(")
		if parenIndex > 0 {
			beforeParen := strings.TrimSpace(sigStr[:parenIndex])
			parts := strings.Fields(beforeParen)
			if len(parts) > 0 {
				funcName = parts[len(parts)-1]
			}
		}
	}

	// For destructors, we need to return additional information
	if isDestructor {
		// We could return this as part of the signature or handle it differently
		// For now, the calling code will detect destructors by checking if the signature contains ~
	}

	return sig.String(), funcName, isMethod, nil
}

// parseInheritance parses class inheritance specification
func (p *Parser) parseInheritance() string {
	var inheritance strings.Builder

	for !p.isAtEnd() && p.peek().Type != TokenLeftBrace && p.peek().Type != TokenSemicolon {
		inheritance.WriteString(p.peek().Value)
		p.advance()
	}

	return strings.TrimSpace(inheritance.String())
}

// parseType parses a type specification
func (p *Parser) parseType() string {
	var typeStr strings.Builder

	for !p.isAtEnd() && p.peek().Type != TokenLeftBrace && p.peek().Type != TokenSemicolon {
		if p.peek().Type == TokenLeftBrace {
			break
		}
		typeStr.WriteString(p.peek().Value)
		p.advance()
	}

	return strings.TrimSpace(typeStr.String())
}

// getCurrentScope returns the current scope entity
func (p *Parser) getCurrentScope() *ast.Entity {
	if len(p.scopeStack) == 0 {
		return p.tree.Root
	}
	return p.scopeStack[len(p.scopeStack)-1]
}

// getCurrentAccessLevel returns the current access level
func (p *Parser) getCurrentAccessLevel() ast.AccessLevel {
	if len(p.accessStack) == 0 {
		return ast.AccessPublic
	}
	return p.accessStack[len(p.accessStack)-1]
}

// isInsideClass returns true if currently inside a class or struct
func (p *Parser) isInsideClass() bool {
	scope := p.getCurrentScope()
	return scope.Type == ast.EntityClass || scope.Type == ast.EntityStruct
}

// buildFullName builds the fully qualified name for an entity
func (p *Parser) buildFullName(name string) string {
	scope := p.getCurrentScope()
	if scope == p.tree.Root || scope.Name == "" {
		return name
	}
	return scope.GetFullPath() + "::" + name
}

// addEntity adds an entity to the current scope
func (p *Parser) addEntity(entity *ast.Entity) {
	// Associate pending comment with this entity
	if p.pendingComment != nil {
		entity.Comment = p.pendingComment
		p.pendingComment = nil // Clear the pending comment
	}

	scope := p.getCurrentScope()
	entity.Parent = scope
	scope.AddChild(entity)
	p.tree.AddEntity(entity)
}

// enterScope enters a new scope
func (p *Parser) enterScope(entity *ast.Entity) {
	p.scopeStack = append(p.scopeStack, entity)
}

// exitScope exits the current scope
func (p *Parser) exitScope() {
	if len(p.scopeStack) > 1 {
		p.scopeStack = p.scopeStack[:len(p.scopeStack)-1]
	}
	if len(p.accessStack) > 1 {
		p.accessStack = p.accessStack[:len(p.accessStack)-1]
	}
}

// getRangeFromTokens creates a range from token indices
func (p *Parser) getRangeFromTokens(start, end int) ast.Range {
	if start >= len(p.tokens) {
		start = len(p.tokens) - 1
	}
	if end >= len(p.tokens) {
		end = len(p.tokens) - 1
	}
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}

	startToken := p.tokens[start]
	endToken := p.tokens[end]

	return ast.Range{
		Start: ast.Position{Line: startToken.Line, Column: startToken.Column, Offset: startToken.Offset},
		End:   ast.Position{Line: endToken.Line, Column: endToken.Column, Offset: endToken.Offset},
	}
}

// parseClassWithMacro handles class declarations preceded by macros
func (p *Parser) parseClassWithMacro() error {
	return p.parseClassOrStructWithMacro(ast.EntityClass)
}

// parseStructWithMacro handles struct declarations preceded by macros
func (p *Parser) parseStructWithMacro() error {
	return p.parseClassOrStructWithMacro(ast.EntityStruct)
}

// parseEnumWithMacro handles enum declarations preceded by macros
func (p *Parser) parseEnumWithMacro() error {
	start := p.current

	// Resolve the macro
	macroToken := p.advance()
	macroValue := p.resolveDefine(macroToken.Value)

	p.skipWhitespace()

	// Now parse the enum normally but include the macro in the signature
	enumToken := p.advance() // consume 'enum'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected enum name")
	}

	nameToken := p.advance()

	// Build signature with macro
	signature := fmt.Sprintf("%s %s %s", macroValue, enumToken.Value, nameToken.Value)

	p.skipWhitespace()

	// Check for enum class or enum struct
	if p.match(TokenClass) || p.match(TokenStruct) {
		signature += fmt.Sprintf(" %s", p.previous().Value)
	}

	// Check for base type
	if p.match(TokenColon) {
		signature += " :"
		for !p.isAtEnd() && p.peek().Type != TokenLeftBrace && p.peek().Type != TokenSemicolon {
			signature += p.advance().Value
		}
	}

	// Handle body if present
	if p.match(TokenLeftBrace) {
		signature += " {"

		// Consume the entire enum body
		braceDepth := 1
		for !p.isAtEnd() && braceDepth > 0 {
			token := p.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			signature += token.Value
			p.advance()
		}
	} else if p.match(TokenSemicolon) {
		signature += ";"
	}

	entity := &ast.Entity{
		Type:        ast.EntityEnum,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)
	return nil
}

// parseClassOrStructWithMacro handles class/struct declarations preceded by macros
func (p *Parser) parseClassOrStructWithMacro(entityType ast.EntityType) error {
	start := p.current

	// Resolve the macro
	macroToken := p.advance()
	macroValue := p.resolveDefine(macroToken.Value)

	p.skipWhitespace()

	// Now parse the class/struct normally but include the macro in the signature
	keyword := p.advance() // consume 'class' or 'struct'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected %s name", keyword.Value)
	}

	nameToken := p.advance()

	// Parse inheritance if present
	inheritance := ""
	p.skipWhitespace()
	if p.match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.skipWhitespace()

	// Build signature with macro
	signature := fmt.Sprintf("%s %s %s", macroValue, keyword.Value, nameToken.Value)
	if inheritance != "" {
		signature += " : " + inheritance
	}
	if p.match(TokenLeftBrace) {
		signature += " {"
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// Enter scope if we found opening brace
	if strings.Contains(signature, "{") {
		p.enterScope(entity)
		// Set default access level for the new scope
		if entityType == ast.EntityClass {
			p.accessStack = append(p.accessStack, ast.AccessPrivate) // class default
		} else {
			p.accessStack = append(p.accessStack, ast.AccessPublic) // struct default
		}
	}

	return nil
}

// New creates a new parser instance (alias for NewTokenParser for compatibility)
func New() *Parser {
	return NewTokenParser()
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
			}
			// Start new tag
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
