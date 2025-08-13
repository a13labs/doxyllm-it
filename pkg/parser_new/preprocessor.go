package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parsePreprocessor handles preprocessor directives
func (p *Parser) parsePreprocessor() (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume '#'
	p.tokenizer.SkipWhitespace()

	directive := p.tokenizer.PeekToken(0)

	if directive.Type == TokenIdentifier && directive.Value == "define" {
		return p.parseDefine()
	}

	// Other preprocessor directives
	return p.parseOtherPreprocessor()
}

// parseDefine handles #define directives
func (p *Parser) parseDefine() (*ast.Entity, error) {
	var signature strings.Builder

	p.tokenizer.NextToken()
	signature.WriteString("#define ")
	p.tokenizer.SkipWhitespace()

	var (
		processNextLine bool
		entityName      string
		sigReady        bool
	)

	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)

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

	if !sigReady {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "expected newline after preprocessor directive")
	}

	// Store the define
	p.defines[entityName] = signature.String()

	return &ast.Entity{
		Type:      ast.EntityPreprocessor,
		Name:      entityName,
		FullName:  entityName,
		Signature: signature.String(),
	}, nil
}

// parseOtherPreprocessor handles other preprocessor directives
func (p *Parser) parseOtherPreprocessor() (*ast.Entity, error) {
	var signature strings.Builder
	signature.WriteString("#")

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenNewline {
		signature.WriteString(p.tokenizer.PeekToken(0).Value)
		p.tokenizer.NextToken()
	}

	// Consume \n
	if p.tokenizer.PeekToken(0).Type != TokenNewline && !p.tokenizer.IsAtEnd() {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "expected newline after preprocessor directive")
	}
	p.tokenizer.NextToken() // consume newline

	return &ast.Entity{
		Type:      ast.EntityPreprocessor,
		Name:      strings.TrimSpace(signature.String()),
		FullName:  strings.TrimSpace(signature.String()),
		Signature: signature.String(),
	}, nil
}
