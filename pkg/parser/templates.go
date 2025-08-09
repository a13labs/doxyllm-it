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
		return p.formatErrorAtCurrentPosition("expected '<' after template")
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
		// If we reach end of file, just return successfully instead of failing
		// This can happen with incomplete template declarations at the end of files
		return nil
	}

	token := p.tokenCache.peek()

	switch token.Type {
	case TokenClass:
		err := p.parseTemplatedClass(templateStart, templateParams)
		if err != nil {
			// If parsing fails, try to skip to the next semicolon or closing brace
			return p.skipToNextEntity()
		}
		return nil
	case TokenStruct:
		err := p.parseTemplatedStruct(templateStart, templateParams)
		if err != nil {
			return p.skipToNextEntity()
		}
		return nil
	case TokenUsing:
		err := p.parseTemplatedUsing(templateStart, templateParams)
		if err != nil {
			return p.skipToNextEntity()
		}
		return nil
	default:
		// Template function
		err := p.parseTemplatedFunction(templateStart, templateParams)
		if err != nil {
			return p.skipToNextEntity()
		}
		return nil
	}
}

// skipToNextEntity attempts to recover from template parsing errors by skipping to the next entity
func (p *Parser) skipToNextEntity() error {
	// Skip tokens until we find a semicolon, closing brace, or the start of a new entity
	for !p.tokenCache.isAtEnd() {
		token := p.tokenCache.peek()

		// Stop at semicolon (end of declaration)
		if token.Type == TokenSemicolon {
			p.tokenCache.advance() // consume the semicolon
			return nil
		}

		// Stop at closing brace
		if token.Type == TokenRightBrace {
			return nil
		}

		// Stop at tokens that typically start new entities
		if token.Type == TokenClass || token.Type == TokenStruct || token.Type == TokenEnum ||
			token.Type == TokenNamespace || token.Type == TokenTemplate ||
			token.Type == TokenTypedef || token.Type == TokenUsing {
			return nil
		}

		p.tokenCache.advance()
	}

	return nil // End of file reached
}

// parseTemplatedClass handles template class declarations
func (p *Parser) parseTemplatedClass(templateStart int, templateParams string) error {
	p.tokenCache.advance() // consume 'class'

	p.tokenCache.skipWhitespace()

	if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
		if p.tokenCache.isAtEnd() {
			return p.formatErrorAtCurrentPosition("expected class name, but reached end of file")
		} else {
			return p.formatErrorAtCurrentPosition("expected class name")
		}
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
		return p.formatErrorAtCurrentPosition("expected struct name")
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
		return p.formatErrorAtCurrentPosition("expected identifier after using")
	}

	nameToken := p.tokenCache.advance()

	p.tokenCache.skipWhitespace()

	if !p.tokenCache.match(TokenEquals) {
		return p.formatErrorAtCurrentPosition("expected '=' in using declaration")
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
