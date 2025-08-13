package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseEnum handles enum declarations
func (p *Parser) parseEnum() (*ast.Entity, error) {

	signature := strings.Builder{}
	p.tokenizer.NextToken()
	p.tokenizer.SkipWhitespace()

	// Build signature and handle body
	signature.WriteString("enum")

	// enums always end with a semicolon
	lastIdentifier := ""
	inInheritance := false
	numTokens := 0
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

	if p.tokenizer.IsAtEnd() {
		return nil, p.formatErrorAtCurrentPosition("expected semicolon after enum declaration")
	}

	// consume semicolon
	p.tokenizer.NextToken()

	entity := &ast.Entity{
		Type:      ast.EntityEnum,
		Name:      lastIdentifier,
		FullName:  p.buildFullName(lastIdentifier),
		Signature: signature.String(),
		Children:  make([]*ast.Entity, 0),
	}

	return entity, nil
}

// parseType parses a type specification
func (p *Parser) parseType() string {
	var typeStr strings.Builder

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenLeftBrace && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		if p.tokenizer.PeekToken(0).Type == TokenLeftBrace {
			break
		}
		typeStr.WriteString(p.tokenizer.PeekToken(0).Value)
		p.tokenizer.NextToken()
	}

	return strings.TrimSpace(typeStr.String())
}
