package parser_new

import (
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
	signature := strings.Builder{}

	keyword := p.tokenizer.NextToken()
	signature.WriteString(keyword.Value + " ")
	p.tokenizer.SkipWhitespace()

	sigReady := false
	isForwardDeclaration := false
	inInheritance := false
	lastIdentifier := ""
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenLeftBrace, TokenSemicolon:
			p.tokenizer.NextToken()
			sigReady = true
			isForwardDeclaration = (token.Type == TokenSemicolon)
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
		return nil, p.formatError(p.tokenizer.PeekToken(0), "class signature is incomplete")
	}

	entity := &ast.Entity{
		Type:                 entityType,
		Name:                 lastIdentifier,
		Signature:            strings.Trim(signature.String(), " "),
		IsForwardDeclaration: isForwardDeclaration,
		Children:             make([]*ast.Entity, 0),
	}

	return entity, nil
}
