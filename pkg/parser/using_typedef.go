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
	signature.WriteString("typedef")

	var name string
	lastIdentifier := ""
	previousWasSpace := false

	for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type != TokenSemicolon {
		token := p.tokenCache.peek()
		
		if token.Type == TokenWhitespace {
			if !previousWasSpace {
				signature.WriteString(" ")
				previousWasSpace = true
			}
		} else {
			if !previousWasSpace && signature.Len() > 0 {
				signature.WriteString(" ")
			}
			signature.WriteString(token.Value)
			previousWasSpace = false
		}

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
		return p.formatErrorAtCurrentPosition("expected identifier after using")
	}

	// Check for 'namespace' keyword
	if p.tokenCache.peek().Type == TokenNamespace {
		return p.parseUsingNamespace(start)
	}

	// Parse the full qualified name (could include ::)
	var fullNameBuilder strings.Builder

	// First identifier is required
	if p.tokenCache.peek().Type != TokenIdentifier {
		return p.formatErrorAtCurrentPosition("expected identifier after using")
	}

	nameToken := p.tokenCache.advance()
	fullNameBuilder.WriteString(nameToken.Value)

	// Handle qualified names like std::data
	for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenDoubleColon {
		fullNameBuilder.WriteString("::")
		p.tokenCache.advance()

		if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
			break
		}

		nextToken := p.tokenCache.advance()
		fullNameBuilder.WriteString(nextToken.Value)
		nameToken = nextToken // Update nameToken to be the last identifier
	}

	p.tokenCache.skipWhitespace()

	var signature string
	var entityType ast.EntityType

	// Check if this is a type alias (using name = type) or a using declaration (using std::name)
	if p.tokenCache.peek().Type == TokenEquals {
		// Type alias: using name = type
		p.tokenCache.advance() // consume '='

		// Parse the rest until semicolon
		var typeValue strings.Builder
		for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type != TokenSemicolon {
			typeValue.WriteString(p.tokenCache.peek().Value)
			p.tokenCache.advance()
		}

		signature = fmt.Sprintf("using %s = %s", nameToken.Value, typeValue.String())
		entityType = ast.EntityUsing
	} else {
		// Using declaration: using std::name
		signature = fmt.Sprintf("using %s", fullNameBuilder.String())
		entityType = ast.EntityUsing
	}

	if p.tokenCache.match(TokenSemicolon) {
		// consumed semicolon
	}

	entity := &ast.Entity{
		Type:        entityType,
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
		return p.formatErrorAtCurrentPosition("expected namespace name after using namespace")
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
