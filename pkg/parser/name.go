package parser

import (
	ast "doxyllm-it/pkg/ast"
	"strings"
)

// parseName handles variable declarations
func (p *Parser) parseName() (*ast.Entity, error) {
	var signature strings.Builder
	var lastIdentifier string
	var sigReady bool
	var lineComment string

	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenWhitespace:
			if signature.Len() == 0 {
				p.tokenizer.NextToken() // skip leading whitespace
				continue
			}
		case TokenNewline:
			p.tokenizer.NextToken()
			continue
		case TokenSemicolon:
			p.tokenizer.NextToken()
			p.tokenizer.SkipSpaces()
			if p.tokenizer.PeekToken(0).Type == TokenLineComment || p.tokenizer.PeekToken(0).Type == TokenBlockComment {
				lineComment = p.tokenizer.NextToken().Value
			}
			sigReady = true
			continue
		case TokenIdentifier:
			lastIdentifier = token.Value
		}
		signature.WriteString(token.Value) // preserve whitespace inside signature
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
		LineComment: &ast.Entity{
			Type:      ast.EntityComment,
			Signature: lineComment,
		},
	}, nil
}
