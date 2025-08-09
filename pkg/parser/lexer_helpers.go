package parser

import (
	"strings"
)

// isValidIdentifierToken checks if a token can be used as an identifier
func (p *Parser) isValidIdentifierToken(token Token) bool {
	// Regular identifiers are always valid
	if token.Type == TokenIdentifier {
		return true
	}

	// Some keywords can also be used as identifiers in certain contexts
	switch token.Type {
	case TokenVoid, TokenBool, TokenChar, TokenShort, TokenInt, TokenLong,
		TokenFloat, TokenDouble, TokenSigned, TokenUnsigned, TokenAuto:
		return true
	default:
		return false
	}
}

// skipSpecifiers skips storage and cv specifiers
func (p *Parser) skipSpecifiers() {
	specifiers := []TokenType{
		TokenStatic, TokenExtern, TokenInline, TokenVirtual,
		TokenConst, TokenConstexpr, TokenMutable, TokenVolatile,
		TokenExplicit, TokenFriend,
	}

	for !p.tokenCache.isAtEnd() {
		found := false
		for _, spec := range specifiers {
			if p.tokenCache.check(spec) {
				p.tokenCache.advance()
				p.tokenCache.skipWhitespace()
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
}

// parseType parses a type specification
func (p *Parser) parseType() string {
	var typeStr strings.Builder

	for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type != TokenLeftBrace && p.tokenCache.peek().Type != TokenSemicolon {
		if p.tokenCache.peek().Type == TokenLeftBrace {
			break
		}
		typeStr.WriteString(p.tokenCache.peek().Value)
		p.tokenCache.advance()
	}

	return strings.TrimSpace(typeStr.String())
}
