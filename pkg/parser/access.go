package parser

import (
	"fmt"

	"doxyllm-it/pkg/ast"
)

// parseAccessSpecifier handles access specifier declarations
func (p *Parser) parseAccessSpecifier() error {
	start := p.tokenCache.getCurrentPosition()
	accessToken := p.tokenCache.advance()

	if !p.tokenCache.match(TokenColon) {
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
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
	}

	p.addEntity(entity)
	return nil
}

// parseCloseBrace handles closing braces
func (p *Parser) parseCloseBrace() error {
	p.tokenCache.advance() // consume '}'

	// Check for optional semicolon after brace (for class/struct)
	if !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenSemicolon {
		p.tokenCache.advance()
	}

	// Exit current scope
	p.exitScope()

	return nil
}
