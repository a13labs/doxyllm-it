package cppparser

import (
	"strings"

	ast "doxyllm-it/pkg/ast"
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
	var signature strings.Builder

	keyword := p.tokenizer.NextToken()
	signature.WriteString(keyword.Value)
	signature.WriteByte(' ')
	p.tokenizer.SkipWhitespace()

	var (
		sigReady             bool
		isForwardDeclaration bool
		inInheritance        bool
		lastIdentifier       string
	)

	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenNewline:
			p.tokenizer.NextToken()
			continue
		case TokenLeftBrace:
			p.tokenizer.NextToken()
			sigReady = true
			continue
		case TokenSemicolon:
			p.tokenizer.NextToken()
			sigReady = true
			isForwardDeclaration = true
			continue
		case TokenColon:
			inInheritance = true
		case TokenIdentifier:
			if !inInheritance {
				lastIdentifier = token.Value
			}
		}
		signature.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	if !sigReady {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "class/struct signature incomplete")
	}

	return &ast.Entity{
		Type:                 entityType,
		Name:                 lastIdentifier,
		Signature:            strings.TrimSpace(signature.String()),
		IsForwardDeclaration: isForwardDeclaration,
		Children:             []*ast.Entity{},
	}, nil
}
