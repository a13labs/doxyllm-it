package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseVariable handles variable declarations
func (p *Parser) parseVariable() error {
	start := p.tokenCache.getCurrentPosition()

	// Parse specifiers first
	var isStatic, isConst, isConstexpr, isExtern bool

	// Parse variable specifiers
	for !p.tokenCache.isAtEnd() {
		token := p.tokenCache.peek()
		if token.Type == TokenStatic {
			isStatic = true
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else if token.Type == TokenConst {
			isConst = true
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else if token.Type == TokenConstexpr {
			isConstexpr = true
			isConst = true // constexpr implies const
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else if token.Type == TokenExtern {
			isExtern = true
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else if token.Type == TokenMutable {
			p.tokenCache.advance() // consume but don't track mutable for now
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
		IsConstexpr: isConstexpr,
		IsExtern:    isExtern,
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
	}

	p.addEntity(entity)
	return nil
}
