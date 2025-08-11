package parser_new

import (
	"fmt"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseClass handles class declarations
func (p *Parser) parseClass() (*ast.Entity, error) {
	return p.parseClassOrStruct(ast.EntityClass)
}

// parseStruct handles struct declarations
func (p *Parser) parseStruct() (*ast.Entity, error) {
	return p.parseClassOrStruct(ast.EntityStruct)
}

func (p *Parser) parseClassOrStruct(entityType ast.EntityType) (*ast.Entity, error) {
	signature := ""

	keyword := p.tokenizer.NextToken()

	p.tokenizer.SkipWhitespace()

	if p.tokenizer.IsAtEnd() || p.tokenizer.PeekToken(0).Type != TokenIdentifier {
		return nil, p.formatErrorAtCurrentPosition(fmt.Sprintf("expected %s name", keyword.Value))
	}

	nameToken := p.tokenizer.NextToken()

	// Build signature
	signature = fmt.Sprintf("%s %s", keyword.Value, nameToken.Value)

	// Parse inheritance if present
	inheritance := ""
	p.tokenizer.SkipWhitespace()
	if p.tokenizer.Match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.tokenizer.SkipWhitespace()

	if inheritance != "" {
		signature += fmt.Sprintf(" : %s", inheritance)
	}

	p.tokenizer.SkipWhitespace()

	// Check for semicolon (forward declaration)
	isForwardDeclaration := p.tokenizer.Match(TokenSemicolon)
	if !isForwardDeclaration && !p.tokenizer.Match(TokenLeftBrace) {
		return nil, p.formatErrorAtCurrentPosition("expected semicolon for forward declaration or '{' to start class body")
	}

	entity := &ast.Entity{
		Type:                 entityType,
		Name:                 nameToken.Value,
		Signature:            signature, // Clean signature without braces
		IsForwardDeclaration: isForwardDeclaration,
		Children:             make([]*ast.Entity, 0),
	}

	return entity, nil
}

// parseInheritance parses class inheritance specification
func (p *Parser) parseInheritance() string {
	var inheritance strings.Builder

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenLeftBrace && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		inheritance.WriteString(p.tokenizer.NextToken().Value)
	}

	return strings.TrimSpace(inheritance.String())
}
