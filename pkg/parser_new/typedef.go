package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseTypedef handles typedef declarations
func (p *Parser) parseTypedef() (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume 'typedef'

	// Parse until we find the identifier and semicolon
	var signature strings.Builder
	signature.WriteString("typedef")

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		token := p.tokenizer.PeekToken(0)
		signature.WriteString(" " + token.Value)
		p.tokenizer.NextToken()
	}

	// Consume the semicolon
	p.tokenizer.NextToken()

	entity := &ast.Entity{
		Type:      ast.EntityTypedef,
		Signature: signature.String(),
	}

	return entity, nil
}
