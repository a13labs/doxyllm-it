package parser_new

import (
	"fmt"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseEnum handles enum declarations
func (p *Parser) parseEnum() (*ast.Entity, error) {

	signature := ""

	p.tokenizer.NextToken() // consume 'enum'
	p.tokenizer.SkipWhitespace()

	// Build signature and handle body
	signature += "enum"

	// Check for 'class' or 'struct' after enum
	if !p.tokenizer.IsAtEnd() && (p.tokenizer.PeekToken(0).Type == TokenClass || p.tokenizer.PeekToken(0).Type == TokenStruct) {
		token := p.tokenizer.NextToken()
		signature += fmt.Sprintf(" %s", token.Value)
		p.tokenizer.SkipWhitespace()
	}

	if p.tokenizer.IsAtEnd() || p.tokenizer.PeekToken(0).Type != TokenIdentifier {
		return nil, p.formatErrorAtCurrentPosition("expected enum name")
	}

	nameToken := p.tokenizer.NextToken()

	p.tokenizer.SkipWhitespace()

	// Parse underlying type if present
	underlyingType := ""
	if p.tokenizer.Match(TokenColon) {
		underlyingType = p.parseType()
	}

	p.tokenizer.SkipWhitespace()

	signature += " " + nameToken.Value
	if underlyingType != "" {
		signature += " : " + underlyingType
	}

	entity := &ast.Entity{
		Type:      ast.EntityEnum,
		Name:      nameToken.Value,
		Signature: signature,
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
