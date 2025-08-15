package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parsePreprocessor handles preprocessor directives
func (p *Parser) parsePreprocessor() (*ast.Entity, error) {
	var signature strings.Builder
	p.tokenizer.NextToken() // consume '#'
	signature.WriteString("#")
	p.tokenizer.SkipWhitespace()

	var (
		processNextLine bool
		entityName      string
		sigReady        bool
		isFirstToken    bool
		directive       string
	)

	isFirstToken = true
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		if isFirstToken {
			directive = token.Value
		}
		isFirstToken = false
		switch token.Type {
		case TokenBackslash:
			processNextLine = true
		case TokenNewline:
			if !processNextLine {
				p.tokenizer.NextToken() // consume newline
				sigReady = true
				continue
			}
			processNextLine = false
		case TokenIdentifier:
			if entityName == "" {
				entityName = token.Value
			}
		}

		signature.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	if !sigReady && !p.tokenizer.IsAtEnd() {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "expected newline after preprocessor directive")
	}

	if directive == "define" {
		// Store the define
		p.defines[entityName] = signature.String()
	} else {
		entityName = "preprocessor"
	}

	return &ast.Entity{
		Type:      ast.EntityPreprocessor,
		Name:      entityName,
		FullName:  entityName,
		Signature: signature.String(),
	}, nil
}
