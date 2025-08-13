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
	lastIdentifier := ""
	inDefinition := false

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		if p.tokenizer.PeekToken(0).Type == TokenEquals {
			inDefinition = true
		}
		if p.tokenizer.PeekToken(0).Type == TokenIdentifier && !inDefinition {
			lastIdentifier = p.tokenizer.PeekToken(0).Value
		}
		signature.WriteString(p.tokenizer.NextToken().Value)
	}

	// Consume the semicolon
	if p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "expected semicolon after using declaration")
	}
	p.tokenizer.NextToken() // consume semicolon

	entity := &ast.Entity{
		Type:      ast.EntityUsing,
		Name:      lastIdentifier,
		FullName:  p.buildFullName(lastIdentifier),
		Signature: signature.String(),
	}

	return entity, nil
}
