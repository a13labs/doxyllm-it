package parser

import (
	"doxyllm-it/pkg/ast"
	"fmt"
)

// TokenCache provides an abstraction layer for navigating cached tokens
// It encapsulates token array access and position management
type TokenCache struct {
	tokenizer *Tokenizer // The tokenizer used to generate tokens
	tokens    []Token    // The token array
	current   int        // Current position in the token array
}

// NewTokenCache creates a new token cache with the provided tokens
func NewTokenCache(content string) (*TokenCache, error) {

	tokenizer := NewTokenizer(content)
	tokens := tokenizer.Tokenize() // Pre-tokenize for compatibility
	// Check for tokenizer errors
	if tokenizer.HasErrors() {
		errors := tokenizer.GetErrors()
		if len(errors) > 0 {
			return nil, fmt.Errorf("tokenizer error: %s", errors[0].Value)
		}
	}

	return &TokenCache{
		tokenizer: tokenizer,
		tokens:    tokens,
		current:   0,
	}, nil
}

// advance returns the current token and moves to the next
func (tc *TokenCache) advance() Token {
	if !tc.isAtEnd() {
		tc.current++
	}
	return tc.previous()
}

// isAtEnd checks if we're at the end of tokens
func (tc *TokenCache) isAtEnd() bool {
	return tc.current >= len(tc.tokens) || tc.peek().Type == TokenEOF
}

// peek returns the current token without advancing
func (tc *TokenCache) peek() Token {
	if tc.current >= len(tc.tokens) {
		return Token{Type: TokenEOF}
	}
	return tc.tokens[tc.current]
}

// previous returns the previous token
func (tc *TokenCache) previous() Token {
	if tc.current <= 0 {
		return Token{Type: TokenEOF}
	}
	return tc.tokens[tc.current-1]
}

// peekAhead looks ahead by offset tokens
func (tc *TokenCache) peekAhead(offset int) Token {
	targetIndex := tc.current + offset
	if targetIndex >= len(tc.tokens) {
		return Token{Type: TokenEOF}
	}
	return tc.tokens[targetIndex]
}

// getCurrentPosition returns the current position in the token array
func (tc *TokenCache) getCurrentPosition() int {
	return tc.current
}

// check returns true if current token is of given type
func (tc *TokenCache) check(tokenType TokenType) bool {
	if tc.isAtEnd() {
		return false
	}
	return tc.peek().Type == tokenType
}

// match checks if current token matches any of the given types
func (tc *TokenCache) match(types ...TokenType) bool {
	for _, tokenType := range types {
		if tc.check(tokenType) {
			tc.advance()
			return true
		}
	}
	return false
}

// skipWhitespace skips whitespace tokens
func (tc *TokenCache) skipWhitespace() {
	for !tc.isAtEnd() && tc.peek().Type == TokenWhitespace {
		tc.advance()
	}
}

// skipWhitespaceAndNewlines skips whitespace and newline tokens
func (tc *TokenCache) skipWhitespaceAndNewlines() {
	for !tc.isAtEnd() && (tc.peek().Type == TokenWhitespace || tc.peek().Type == TokenNewline) {
		tc.advance()
	}
}

// getRangeFromPositions creates a range from token positions
func (tc *TokenCache) getRangeFromPositions(start, end int) ast.Range {
	if start >= len(tc.tokens) {
		start = len(tc.tokens) - 1
	}
	if end >= len(tc.tokens) {
		end = len(tc.tokens) - 1
	}
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}

	startToken := tc.tokens[start]
	endToken := tc.tokens[end]

	return ast.Range{
		Start: ast.Position{Line: startToken.Line, Column: startToken.Column, Offset: startToken.Offset},
		End:   ast.Position{Line: endToken.Line, Column: endToken.Column, Offset: endToken.Offset},
	}
}

// setPosition sets the current position (for checkpointing)
func (tc *TokenCache) setPosition(position int) {
	if position < 0 {
		tc.current = 0
	} else if position >= len(tc.tokens) {
		tc.current = len(tc.tokens)
	} else {
		tc.current = position
	}
}

// getTokenAtOffset returns the token at current position + offset
func (tc *TokenCache) getTokenAtOffset(offset int) Token {
	targetIndex := tc.current + offset
	if targetIndex < 0 || targetIndex >= len(tc.tokens) {
		return Token{Type: TokenEOF}
	}
	return tc.tokens[targetIndex]
}
