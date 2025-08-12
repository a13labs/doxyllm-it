package parser_new

import (
	ast "doxyllm-it/pkg/ast_new"
	"strings"
)

// parseVariable handles variable declarations
func (p *Parser) parseVariable() (*ast.Entity, error) {

	// Parse specifiers first
	var isStatic, isConst, isConstexpr, isExtern bool

	signature := strings.Builder{}
	lastIdentifier := ""
	sigReady := false
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenWhitespace, TokenNewline:
			if signature.Len() == 0 {
				p.tokenizer.NextToken() // skip whitespace at start of signature
				continue
			}
		case TokenSemicolon:
			p.tokenizer.NextToken()
			sigReady = true
			continue
		case TokenExtern:
			isExtern = true
		case TokenStatic:
			isStatic = true
		case TokenConstexpr:
			isConstexpr = true
		case TokenConst:
			isConst = true
		case TokenIdentifier:
			lastIdentifier = token.Value
		default:
		}
		signature.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	if !sigReady {
		return nil, p.formatErrorAtCurrentPosition("variable signature is incomplete")
	}

	entityType := ast.EntityVariable
	if p.isInsideClass() {
		entityType = ast.EntityField
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        lastIdentifier,
		FullName:    p.buildFullName(lastIdentifier),
		Signature:   signature.String(),
		AccessLevel: p.getCurrentAccessLevel(),
		IsStatic:    isStatic,
		IsConst:     isConst,
		IsConstexpr: isConstexpr,
		IsExtern:    isExtern,
	}

	return entity, nil
}
