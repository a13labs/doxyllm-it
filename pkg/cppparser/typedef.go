package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseTypedef handles typedef declarations
func (p *Parser) parseTypedef() (*ast.Entity, error) {

	var signature strings.Builder
	p.tokenizer.NextToken() // consume 'typedef
	signature.WriteString("typedef ")
	p.tokenizer.SkipWhitespace()

	lastIdentifier := ""
	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		token := p.tokenizer.PeekToken(0)
		if token.Type == TokenIdentifier {
			lastIdentifier = token.Value
		}

		signature.WriteString(token.Value)
		p.tokenizer.NextToken() // consume the token
	}

	// Consume the semicolon if present
	if !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type == TokenSemicolon {
		p.tokenizer.NextToken()
	}

	entity := &ast.Entity{
		Type:      ast.EntityTypedef,
		Name:      lastIdentifier,
		FullName:  p.buildFullName(lastIdentifier),
		Signature: signature.String(),
	}

	return entity, nil
}
