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
	p.tokenizer.NextToken() // consume 'define'
	p.tokenizer.SkipWhitespace()

	var signature strings.Builder
	nameToken := p.tokenizer.PeekToken(0)
	if nameToken.Type != TokenIdentifier {
		return nil, p.formatError(nameToken, "expected identifier after #define")
	}
	p.tokenizer.NextToken()
	signature.WriteString("#define " + nameToken.Value)

	// Collect the definition value until end of line or end of file

	nextLine := false
	for !p.tokenizer.IsAtEnd() {
		token := p.tokenizer.PeekToken(0)

		// Handle nextLine macros with backslash continuation
		if token.Type == TokenBackslash {
			nextLine = true
			continue
		}

		if token.Type == TokenNewline {
			if !nextLine {
				p.tokenizer.NextToken() // consume newline
				break
			}
			nextLine = false
		}

		signature.WriteString(token.Value)

		p.tokenizer.NextToken()
	}

	// Store the define
	p.defines[nameToken.Value] = signature.String()

	entity := &ast.Entity{
		Type:      ast.EntityPreprocessor,
		Name:      nameToken.Value,
		FullName:  nameToken.Value,
		Signature: signature.String(),
	}

	return entity, nil
}

// parseOtherPreprocessor handles other preprocessor directives
func (p *Parser) parseOtherPreprocessor() (*ast.Entity, error) {
	// Consume until end of line
	var content strings.Builder
	content.WriteString("#")

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenNewline {
		content.WriteString(p.tokenizer.PeekToken(0).Value)
		p.tokenizer.NextToken()
	}

	// Consume \n
	if !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type == TokenNewline {
		p.tokenizer.NextToken()
	}

	entity := &ast.Entity{
		Type:      ast.EntityPreprocessor,
		Name:      strings.TrimSpace(content.String()),
		FullName:  strings.TrimSpace(content.String()),
		Signature: content.String(),
	}

	return entity, nil
}
