package parser_new

import (
	"fmt"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseNamespace handles namespace declarations
func (p *Parser) parseNamespace() error {
	start := p.tokenCache.getCurrentPosition()
	p.tokenCache.advance() // consume 'namespace'

	p.tokenCache.skipWhitespace()

	if p.tokenCache.isAtEnd() || !p.isValidIdentifierToken(p.tokenCache.peek()) {
		return p.formatErrorAtCurrentPosition("expected namespace name")
	}

	// Parse namespace name (could be nested like mgl::io)
	var nameBuilder strings.Builder
	nameBuilder.WriteString(p.tokenCache.advance().Value) // first identifier

	// Check for :: followed by more identifiers (nested namespace)
	for !p.tokenCache.isAtEnd() {
		p.tokenCache.skipWhitespace()
		if p.tokenCache.peek().Type == TokenDoubleColon {
			nameBuilder.WriteString(p.tokenCache.advance().Value) // add ::
			p.tokenCache.skipWhitespace()
			if p.isValidIdentifierToken(p.tokenCache.peek()) {
				nameBuilder.WriteString(p.tokenCache.advance().Value) // add next identifier
			} else {
				break
			}
		} else {
			break
		}
	}

	namespaceName := nameBuilder.String()

	p.tokenCache.skipWhitespace()

	// Build signature
	signature := fmt.Sprintf("namespace %s", namespaceName)

	// Note: Opening braces are now handled by the main parser dispatch
	// We don't look for braces here anymore

	entity := &ast.Entity{
		Type:        ast.EntityNamespace,
		Name:        namespaceName,
		FullName:    p.buildFullName(namespaceName),
		Signature:   signature, // Clean signature without braces
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// The opening brace (if present) will be handled by the main parser
	// and will create its own scope

	return nil
}
