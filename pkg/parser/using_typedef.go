package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
)

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
