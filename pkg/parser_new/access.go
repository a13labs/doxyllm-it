package parser_new

import (
	ast "doxyllm-it/pkg/ast_new"
)

// parseAccessSpecifier handles access specifier declarations
func (p *Parser) parseAccessSpecifier() (*ast.Entity, error) {
	accessToken := p.tokenizer.NextToken()

	if !p.tokenizer.Match(TokenColon) {
		return nil, p.formatErrorAtCurrentPosition("expected ':' after access specifier")
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
	}

	return entity, nil
}
