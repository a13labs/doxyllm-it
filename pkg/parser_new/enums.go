package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseEnum handles enum declarations
func (p *Parser) parseEnum() (*ast.Entity, error) {

	p.tokenizer.NextToken()
	p.tokenizer.SkipWhitespace()

	// Build signature and handle body
	signature := strings.Builder{}
	signature.WriteString("enum")

	var (
		lastIdentifier string
		inInheritance  bool
		numTokens      int
	)

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenWhitespace, TokenNewline:
			if numTokens == 0 {
				p.tokenizer.NextToken() // skip whitespace at start of signature
				continue
			}
		case TokenDoubleColon, TokenLeftBrace:
			inInheritance = true
		case TokenIdentifier:
			if !inInheritance {
				lastIdentifier = token.Value
			}
		}
		signature.WriteString(token.Value)
		p.tokenizer.NextToken()
		numTokens++
	}

	// consume semicolon
	if p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "expected ';' after enum declaration")
	}
	p.tokenizer.NextToken()

	return &ast.Entity{
		Type:      ast.EntityEnum,
		Name:      lastIdentifier,
		FullName:  p.buildFullName(lastIdentifier),
		Signature: signature.String(),
		Children:  make([]*ast.Entity, 0),
	}, nil
}
