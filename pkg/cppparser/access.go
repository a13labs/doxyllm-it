package cppparser

import (
	ast "doxyllm-it/pkg/ast"
)

// parseAccessSpecifier handles access specifier declarations
func (p *Parser) parseAccessSpecifier() (*ast.Entity, error) {
	token := p.tokenizer.NextToken()

	// Update current access level
	var accessLevel ast.AccessLevel
	switch token.Value {
	case "public":
		accessLevel = ast.AccessPublic
	case "private":
		accessLevel = ast.AccessPrivate
	case "protected":
		accessLevel = ast.AccessProtected
	}

	if p.tokenizer.PeekToken(0).Type != TokenColon {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "expected ':' after access specifier")
	}
	p.tokenizer.NextToken() // consume ':'

	// Update the access stack for current scope
	if len(p.accessStack) > 0 {
		p.accessStack[len(p.accessStack)-1] = accessLevel
	}

	// Create access specifier entity
	entity := &ast.Entity{
		Type:        ast.EntityAccessSpecifier,
		Name:        token.Value,
		FullName:    token.Value,
		Signature:   token.Value + ":",
		AccessLevel: accessLevel,
	}

	return entity, nil
}
