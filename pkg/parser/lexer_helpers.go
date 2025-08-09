package parser

import (
	"strings"
)

// advance returns the current token and moves to the next
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// isAtEnd checks if we're at the end of tokens
func (p *Parser) isAtEnd() bool {
	return p.current >= len(p.tokens) || p.peek().Type == TokenEOF
}

// peek returns the current token without advancing
func (p *Parser) peek() Token {
	if p.current >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.current]
}

// previous returns the previous token
func (p *Parser) previous() Token {
	if p.current <= 0 {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.current-1]
}

// peekAhead looks ahead by offset tokens (for compatibility)
func (p *Parser) peekAhead(offset int) Token {
	targetIndex := p.current + offset
	if targetIndex >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[targetIndex]
}

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

// match checks if current token matches any of the given types
func (p *Parser) match(types ...TokenType) bool {
	for _, tokenType := range types {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}
	return false
}

// check returns true if current token is of given type
func (p *Parser) check(tokenType TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
}

// skipWhitespace skips whitespace tokens
func (p *Parser) skipWhitespace() {
	for !p.isAtEnd() && p.peek().Type == TokenWhitespace {
		p.advance()
	}
}

// skipWhitespaceAndNewlines skips whitespace and newline tokens
func (p *Parser) skipWhitespaceAndNewlines() {
	for !p.isAtEnd() && (p.peek().Type == TokenWhitespace || p.peek().Type == TokenNewline) {
		p.advance()
	}
}

// skipSpecifiers skips storage and cv specifiers
func (p *Parser) skipSpecifiers() {
	specifiers := []TokenType{
		TokenStatic, TokenExtern, TokenInline, TokenVirtual,
		TokenConst, TokenConstexpr, TokenMutable, TokenVolatile,
		TokenExplicit, TokenFriend,
	}

	for !p.isAtEnd() {
		found := false
		for _, spec := range specifiers {
			if p.check(spec) {
				p.advance()
				p.skipWhitespace()
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

	for !p.isAtEnd() && p.peek().Type != TokenLeftBrace && p.peek().Type != TokenSemicolon {
		if p.peek().Type == TokenLeftBrace {
			break
		}
		typeStr.WriteString(p.peek().Value)
		p.advance()
	}

	return strings.TrimSpace(typeStr.String())
}
