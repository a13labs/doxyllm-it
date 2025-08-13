package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseUsing handles using declarations
func (p *Parser) parseUsing() (*ast.Entity, error) {
	var signature strings.Builder
	p.tokenizer.NextToken() // consume 'using'
	p.tokenizer.SkipWhitespace()

	signature.WriteString("using ")

	var (
		lastIdentifier string
		inDefinition   bool
		sigReady       bool
	)

	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenSemicolon:
			sigReady = true
			continue
		case TokenEquals:
			inDefinition = true
		case TokenIdentifier:
			if !inDefinition {
				lastIdentifier = token.Value
			}
		}
		signature.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	// Consume the semicolon
	if p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "expected semicolon after using declaration")
	}
	p.tokenizer.NextToken() // consume semicolon

	return &ast.Entity{
		Type:      ast.EntityUsing,
		Name:      lastIdentifier,
		FullName:  p.buildFullName(lastIdentifier),
		Signature: signature.String(),
	}, nil
}
