package parser_new

import (
	"fmt"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseTemplate handles template declarations
func (p *Parser) parseTemplate() (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume 'template'

	// Skip whitespace and newlines
	p.tokenizer.SkipWhitespace()

	// Parse template parameters
	if !p.tokenizer.Match(TokenLess) {
		return nil, p.formatErrorAtCurrentPosition("expected '<' after template")
	}

	depth := 1
	var templateParams strings.Builder
	templateParams.WriteString("<")

	for !p.tokenizer.IsAtEnd() && depth > 0 {
		token := p.tokenizer.PeekToken(0)
		if token.Type == TokenLess {
			depth++
		} else if token.Type == TokenGreater {
			depth--
		}

		templateParams.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	p.tokenizer.SkipWhitespace()
	var newSig string
	var e *ast.Entity
	var err error
	token := p.tokenizer.PeekToken(0)
	switch token.Type {
	case TokenClass:
		e, err = p.parseClass()
	case TokenStruct:
		e, err = p.parseStruct()
	case TokenUsing:
		e, err = p.parseUsing()
	default:
		// Template function?
		e, err = p.parseDefault()
	}

	if err != nil {
		return nil, err
	}

	newSig = fmt.Sprintf("template %s %s", templateParams.String(), e.Signature)
	e.Signature = newSig
	e.IsTemplate = true

	return e, nil
}
