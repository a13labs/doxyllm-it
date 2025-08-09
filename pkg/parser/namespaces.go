package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
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

	// Look for opening brace, which might be on the same line or next line
	if p.tokenCache.match(TokenLeftBrace) {
		signature += " {"
	} else {
		// The brace might be on the next line, so we need to look ahead
		// Save current position to check for brace
		checkpoint := p.tokenCache.getCurrentPosition()

		// Skip any whitespace/newlines to find the brace
		for !p.tokenCache.isAtEnd() && (p.tokenCache.peek().Type == TokenWhitespace || p.tokenCache.peek().Type == TokenNewline) {
			p.tokenCache.advance()
		}

		if !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenLeftBrace {
			p.tokenCache.advance() // consume the brace
			signature += " {"
		} else {
			// No brace found, restore position
			p.tokenCache.setPosition(checkpoint)
		}
	}

	entity := &ast.Entity{
		Type:        ast.EntityNamespace,
		Name:        namespaceName,
		FullName:    p.buildFullName(namespaceName),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// Enter scope if we found opening brace
	if strings.Contains(signature, "{") {
		p.enterScope(entity)
	}

	return nil
}
