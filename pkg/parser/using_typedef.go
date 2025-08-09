package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
)

// parseTypedef handles typedef declarations
func (p *Parser) parseTypedef() error {
	start := p.tokenCache.getCurrentPosition()
	p.tokenCache.advance() // consume 'typedef'

	// Parse until we find the identifier and semicolon
	var signature strings.Builder
	signature.WriteString("typedef ")

	var name string
	lastIdentifier := ""

	for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type != TokenSemicolon {
		token := p.tokenCache.peek()
		signature.WriteString(token.Value)

		if token.Type == TokenIdentifier {
			lastIdentifier = token.Value
		}

		p.tokenCache.advance()
	}

	if p.tokenCache.match(TokenSemicolon) {
		signature.WriteString(";")
	}

	name = lastIdentifier // The last identifier is typically the typedef name

	entity := &ast.Entity{
		Type:        ast.EntityTypedef,
		Name:        name,
		FullName:    p.buildFullName(name),
		Signature:   signature.String(),
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
	}

	p.addEntity(entity)
	return nil
}

// parseUsing handles using declarations
func (p *Parser) parseUsing() error {
	start := p.tokenCache.getCurrentPosition()
	p.tokenCache.advance() // consume 'using'

	p.tokenCache.skipWhitespace()

	if p.tokenCache.isAtEnd() {
		return fmt.Errorf("expected identifier after using")
	}

	// Check for 'namespace' keyword
	if p.tokenCache.peek().Type == TokenNamespace {
		return p.parseUsingNamespace(start)
	}

	// Regular using declaration
	if p.tokenCache.peek().Type != TokenIdentifier {
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

	signature := fmt.Sprintf("using %s = %s", nameToken.Value, typeValue.String())

	entity := &ast.Entity{
		Type:        ast.EntityUsing,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
	}

	p.addEntity(entity)
	return nil
}

// parseUsingNamespace handles using namespace declarations
func (p *Parser) parseUsingNamespace(start int) error {
	p.tokenCache.advance() // consume 'namespace'

	p.tokenCache.skipWhitespace()

	if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected namespace name after using namespace")
	}

	nameToken := p.tokenCache.advance()

	// Parse qualified namespace name
	var namespaceName strings.Builder
	namespaceName.WriteString(nameToken.Value)

	for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenDoubleColon {
		namespaceName.WriteString("::")
		p.tokenCache.advance()

		if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
			break
		}

		namespaceName.WriteString(p.tokenCache.advance().Value)
	}

	if p.tokenCache.match(TokenSemicolon) {
		// consumed semicolon
	}

	signature := fmt.Sprintf("using namespace %s", namespaceName.String())

	entity := &ast.Entity{
		Type:        ast.EntityUsing,
		Name:        namespaceName.String(),
		FullName:    namespaceName.String(),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
	}

	p.addEntity(entity)
	return nil
}
