package parser_new

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

	for !p.tokenizer.IsAtEnd() {
		found := false
		for _, spec := range specifiers {
			if p.tokenizer.Match(spec) {
				p.tokenizer.NextToken()
				p.tokenizer.SkipWhitespace()
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

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenLeftBrace && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
		if p.tokenizer.PeekToken(0).Type == TokenLeftBrace {
			break
		}
		typeStr.WriteString(p.tokenizer.PeekToken(0).Value)
		p.tokenizer.NextToken()
	}

	return strings.TrimSpace(typeStr.String())
}
