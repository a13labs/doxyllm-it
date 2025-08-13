package parser_new

import (
	"fmt"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseTemplate handles template declarations
func (p *Parser) parseTemplate() (*ast.Entity, error) {
	var signature strings.Builder
	p.tokenizer.NextToken() // consume 'template'
	p.tokenizer.SkipWhitespace()
	signature.WriteString("template ")

	depth := 0
	sigReady := false
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenLess:
			depth++
		case TokenGreater:
			depth--
			if depth == 0 {
				sigReady = true
			}
		}
		signature.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	if !sigReady {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "template signature is incomplete")
	}

	var entity *ast.Entity
	var err error

	// templates are always followed by a class, struct, or other type of entities
	p.tokenizer.SkipWhitespace()
	token := p.tokenizer.PeekToken(0)
	switch token.Type {
	case TokenClass:
		entity, err = p.parseClass()
	case TokenStruct:
		entity, err = p.parseStruct()
	case TokenUsing:
		entity, err = p.parseUsing()
	default:
		// Template function?
		entity, err = p.parseDefault()
	}

	if err != nil {
		return nil, err
	}

	entity.Signature = fmt.Sprintf("%s %s", signature.String(), entity.Signature)
	entity.IsTemplate = true

	return entity, nil
}
