package parser_new

import (
	"fmt"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseTemplate handles template declarations
func (p *Parser) parseTemplate() (*ast.Entity, error) {
	p.tokenizer.NextToken()
	p.tokenizer.SkipWhitespace()

	var signature strings.Builder
	signature.WriteString("template ")

	var (
		depth    int
		sigReady bool
	)

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
		case TokenRightShift:
			depth--
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

	p.tokenizer.SkipWhitespace()
	token := p.tokenizer.PeekToken(0)

	if token.Type == TokenLineComment || token.Type == TokenBlockComment {
		for !p.tokenizer.IsAtEnd() && (token.Type == TokenLineComment || token.Type == TokenBlockComment) {
			// Consume all comments
			p.tokenizer.NextToken() // consume comment
			p.tokenizer.SkipWhitespace()
			token = p.tokenizer.PeekToken(0)
		}
	}

	var (
		entity *ast.Entity
		err    error
	)
	switch token.Type {
	case TokenClass:
		entity, err = p.parseClass()
	case TokenStruct:
		entity, err = p.parseStruct()
	case TokenUsing:
		entity, err = p.parseUsing()
	default:
		entity, err = p.parseDefault()
	}

	if err != nil {
		return nil, err
	}

	entity.Signature = fmt.Sprintf("%s %s", signature.String(), entity.Signature)
	entity.IsTemplate = true

	return entity, nil
}
