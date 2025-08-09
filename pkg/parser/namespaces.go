package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
)

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
