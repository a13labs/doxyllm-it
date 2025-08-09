package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
)

// parseTemplate handles template declarations
func (p *Parser) parseTemplate() error {
	start := p.tokenCache.getCurrentPosition()
	p.tokenCache.advance() // consume 'template'

	// Skip whitespace and newlines
	p.tokenCache.skipWhitespaceAndNewlines()

	// Parse template parameters
	if !p.tokenCache.match(TokenLess) {
		return fmt.Errorf("expected '<' after template")
	}

	depth := 1
	var templateParams strings.Builder
	templateParams.WriteString("<")

	for !p.tokenCache.isAtEnd() && depth > 0 {
		token := p.tokenCache.peek()
		if token.Type == TokenLess {
			depth++
		} else if token.Type == TokenGreater {
			depth--
		}

		templateParams.WriteString(token.Value)
		p.tokenCache.advance()
	}

	p.tokenCache.skipWhitespaceAndNewlines()

	// Now parse the templated entity
	return p.parseTemplatedEntity(start, templateParams.String())
}

// parseTemplatedEntity parses the entity that follows a template declaration
func (p *Parser) parseTemplatedEntity(templateStart int, templateParams string) error {
	if p.tokenCache.isAtEnd() {
		return fmt.Errorf("expected entity after template")
	}

	token := p.tokenCache.peek()

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

// parseTemplatedClass handles template class declarations
func (p *Parser) parseTemplatedClass(templateStart int, templateParams string) error {
	p.tokenCache.advance() // consume 'class'

	p.tokenCache.skipWhitespace()

	if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected class name")
	}

	nameToken := p.tokenCache.advance()

	// Parse inheritance if present
	inheritance := ""
	p.tokenCache.skipWhitespace()
	if p.tokenCache.match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.tokenCache.skipWhitespace()

	// Build signature with template
	signature := fmt.Sprintf("template %s class %s", templateParams, nameToken.Value)
	if inheritance != "" {
		signature += " : " + inheritance
	}
	if p.tokenCache.match(TokenLeftBrace) {
		signature += " {"
	} else if p.tokenCache.match(TokenSemicolon) {
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
		SourceRange: p.getRangeFromTokens(templateStart, p.tokenCache.getCurrentPosition()-1),
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
	p.tokenCache.advance() // consume 'struct'

	p.tokenCache.skipWhitespace()

	if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected struct name")
	}

	nameToken := p.tokenCache.advance()

	// Parse inheritance if present
	inheritance := ""
	p.tokenCache.skipWhitespace()
	if p.tokenCache.match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.tokenCache.skipWhitespace()

	// Build signature with template
	signature := fmt.Sprintf("template %s struct %s", templateParams, nameToken.Value)
	if inheritance != "" {
		signature += " : " + inheritance
	}
	if p.tokenCache.match(TokenLeftBrace) {
		signature += " {"
	} else if p.tokenCache.match(TokenSemicolon) {
		// Forward declaration - consume the semicolon but don't add it to signature
	}

	entity := &ast.Entity{
		Type:        ast.EntityStruct,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		IsTemplate:  true,
		SourceRange: p.getRangeFromTokens(templateStart, p.tokenCache.getCurrentPosition()-1),
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
	p.tokenCache.advance() // consume 'using'

	p.tokenCache.skipWhitespace()

	if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected identifier after using")
	}

	nameToken := p.tokenCache.advance()

	p.tokenCache.skipWhitespace()

	if !p.tokenCache.match(TokenEquals) {
		return fmt.Errorf("expected '=' in using declaration")
	}

	// Parse the rest until semicolon
	var typeValue strings.Builder
	for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type != TokenSemicolon {
		typeValue.WriteString(p.tokenCache.peek().Value)
		p.tokenCache.advance()
	}

	if p.tokenCache.match(TokenSemicolon) {
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
		SourceRange: p.getRangeFromTokens(templateStart, p.tokenCache.getCurrentPosition()-1),
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
	p.tokenCache.skipWhitespace()
	if !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenLeftBrace {
		// This function has a body - track its range and content for the formatter
		bodyStart := p.tokenCache.getCurrentPosition()
		braceDepth := 1
		var bodyTokens []Token
		bodyTokens = append(bodyTokens, p.tokenCache.advance()) // consume opening brace

		for !p.tokenCache.isAtEnd() && braceDepth > 0 {
			token := p.tokenCache.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			bodyTokens = append(bodyTokens, p.tokenCache.advance())
		}

		if braceDepth == 0 {
			rangeValue := p.getRangeFromTokens(bodyStart, p.tokenCache.getCurrentPosition()-1)
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
		SourceRange:  p.getRangeFromTokens(templateStart, p.tokenCache.getCurrentPosition()-1),
		BodyRange:    bodyRange,
		OriginalText: bodyText,
	}

	p.addEntity(entity)
	return nil
}
