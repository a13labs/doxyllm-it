package parser_new

import (
	ast "doxyllm-it/pkg/ast_new"
	"strings"
)

// parseName handles variable declarations
func (p *Parser) parseName() (*ast.Entity, error) {
	var signature strings.Builder
	var lastIdentifier string
	sigReady := false

	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenWhitespace, TokenNewline:
			if signature.Len() == 0 {
				p.tokenizer.NextToken() // skip leading whitespace
				continue
			}
			signature.WriteString(token.Value) // preserve whitespace inside signature
		case TokenSemicolon:
			p.tokenizer.NextToken()
			sigReady = true
			continue
		case TokenIdentifier:
			lastIdentifier = token.Value
			signature.WriteString(token.Value)
		default:
			signature.WriteString(token.Value)
		}
		p.tokenizer.NextToken()
	}

	if !sigReady {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "variable signature is incomplete")
	}

	entityType := ast.EntityName

	return &ast.Entity{
		Type:        entityType,
		Name:        lastIdentifier,
		FullName:    p.buildFullName(lastIdentifier),
		Signature:   signature.String(),
		AccessLevel: p.getCurrentAccessLevel(),
	}, nil
}
