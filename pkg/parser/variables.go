package parser

import (
	"strings"

	"doxyllm-it/pkg/ast"
)

// parseVariable handles variable declarations
func (p *Parser) parseVariable() error {
	start := p.tokenCache.getCurrentPosition()

	// Parse specifiers first
	var isStatic, isConst bool

	// Parse variable specifiers
	for !p.tokenCache.isAtEnd() {
		token := p.tokenCache.peek()
		if token.Type == TokenStatic {
			isStatic = true
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else if token.Type == TokenConst || token.Type == TokenConstexpr {
			isConst = true
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else if token.Type == TokenExtern || token.Type == TokenMutable {
			p.tokenCache.advance() // consume but don't track these for now
			p.tokenCache.skipWhitespace()
		} else {
			break
		}
	}

	// Parse until semicolon
	var signature strings.Builder
	var name string
	lastIdentifier := ""

	for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type != TokenSemicolon {
		token := p.tokenCache.peek()

		// Resolve defines in token values
		tokenValue := token.Value
		if token.Type == TokenIdentifier {
			tokenValue = p.resolveDefine(token.Value)
		}

		signature.WriteString(tokenValue)

		if token.Type == TokenIdentifier {
			lastIdentifier = token.Value
		}

		p.tokenCache.advance()
	}

	if p.tokenCache.match(TokenSemicolon) {
		signature.WriteString(";")
	}

	name = lastIdentifier

	entityType := ast.EntityVariable
	if p.isInsideClass() {
		entityType = ast.EntityField
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        name,
		FullName:    p.buildFullName(name),
		Signature:   signature.String(),
		AccessLevel: p.getCurrentAccessLevel(),
		IsStatic:    isStatic,
		IsConst:     isConst,
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
	}

	p.addEntity(entity)
	return nil
}
