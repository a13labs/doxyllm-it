package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseTypedef handles typedef declarations
func (p *Parser) parseTypedef() (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume 'typedef'

	var tokens []string
	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		token := p.tokenizer.NextToken()
		tokens = append(tokens, token.Value)
	}

	// Consume the semicolon if present
	if !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type == TokenSemicolon {
		p.tokenizer.NextToken()
	}

	entity := &ast.Entity{
		Type:      ast.EntityTypedef,
		Signature: "typedef " + strings.Join(tokens, " "),
	}

	return entity, nil
}
