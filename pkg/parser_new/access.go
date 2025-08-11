package parser_new

import (
	ast "doxyllm-it/pkg/ast_new"
)

// parseAccessSpecifier handles access specifier declarations
func (p *Parser) parseAccessSpecifier() error {
	start := p.tokenCache.getCurrentPosition()
	accessToken := p.tokenCache.advance()

	if !p.tokenCache.match(TokenColon) {
		return p.formatErrorAtCurrentPosition("expected ':' after access specifier")
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

// parseOpenBrace handles opening braces - creates scope entity and enters it
func (p *Parser) parseOpenBrace() error {
	start := p.tokenCache.getCurrentPosition()
	p.tokenCache.advance() // consume '{'

	// Create explicit scope open entity
	openBrace := &ast.Entity{
		Type:        ast.EntityScopeOpen,
		Name:        "{",
		Signature:   "{",
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
		Children:    make([]*ast.Entity, 0),
	}

	// Add to current scope
	currentScope := p.getCurrentScope()
	if currentScope != nil {
		currentScope.AddChild(openBrace)
	}

	// Enter the new scope - the opening brace entity becomes the new scope
	p.enterScope(openBrace)

	return nil
}

// parseCloseBrace handles closing braces - creates scope close entity and exits scope
func (p *Parser) parseCloseBrace() error {
	start := p.tokenCache.getCurrentPosition()
	p.tokenCache.advance() // consume '}'

	// Create explicit scope close entity
	closeBrace := &ast.Entity{
		Type:        ast.EntityScopeClose,
		Name:        "}",
		Signature:   "}",
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
		Children:    make([]*ast.Entity, 0),
	}

	// Add to current scope before exiting
	currentScope := p.getCurrentScope()
	if currentScope != nil {
		currentScope.AddChild(closeBrace)
	}

	// Check for optional semicolon after brace (for class/struct)
	if !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenSemicolon {
		p.tokenCache.advance()
	}

	// Exit current scope
	p.exitScope()

	return nil
}
