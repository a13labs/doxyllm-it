// Package parser implements a streaming token-driven C++ header file parser with O(1) memory complexity
package parser

import (
	"doxyllm-it/pkg/ast"
)

// Parser implements a token-driven parser for C++ headers with streaming tokenizer backend
type Parser struct {
	tokenCache     *TokenCache // Token cache abstraction
	tree           *ast.ScopeTree
	scopeStack     []*ast.Entity
	accessStack    []ast.AccessLevel
	defines        map[string]string
	pendingComment *ast.DoxygenComment
}

// New creates a new token-driven parser
func New() *Parser {
	return &Parser{
		defines: make(map[string]string),
	}
}

// Parse parses tokens into an AST using streaming tokenizer with compatibility layer
func (p *Parser) Parse(filename, content string) (*ast.ScopeTree, error) {
	// Initialize the tree
	p.tree = ast.NewScopeTree(filename, content)
	p.scopeStack = []*ast.Entity{p.tree.Root}
	p.accessStack = []ast.AccessLevel{ast.AccessPublic} // Global scope is public

	// Initialize streaming tokenizer
	var err error
	p.tokenCache, err = NewTokenCache(content)
	if err != nil {
		return nil, err
	}

	// Parse tokens
	for !p.tokenCache.isAtEnd() {
		if err := p.parseTopLevel(); err != nil {
			return nil, err
		}
	}

	return p.tree, nil
}

// parseTopLevel parses top-level declarations
func (p *Parser) parseTopLevel() error {
	// Skip whitespace and newlines
	p.tokenCache.skipWhitespaceAndNewlines()

	if p.tokenCache.isAtEnd() {
		return nil
	}

	// Handle different token types
	token := p.tokenCache.peek()

	switch token.Type {
	case TokenHash:
		return p.parsePreprocessor()
	case TokenLineComment, TokenBlockComment, TokenDoxygenComment:
		return p.parseComment()
	case TokenTemplate:
		return p.parseTemplate()
	case TokenNamespace:
		return p.parseNamespace()
	case TokenClass:
		return p.parseClass()
	case TokenStruct:
		return p.parseStruct()
	case TokenEnum:
		return p.parseEnum()
	case TokenTypedef:
		return p.parseTypedef()
	case TokenUsing:
		return p.parseUsing()
	case TokenPublic, TokenPrivate, TokenProtected:
		return p.parseAccessSpecifier()
	case TokenRightBrace:
		return p.parseCloseBrace()
	case TokenIdentifier:
		// Check if this identifier is a macro that should be resolved
		if _, exists := p.defines[token.Value]; exists {
			// Look ahead to see if after macro resolution we have a keyword
			offset := 1
			nextToken := p.tokenCache.peekAhead(offset)
			// Skip whitespace in lookahead
			for nextToken.Type == TokenWhitespace {
				offset++
				nextToken = p.tokenCache.peekAhead(offset)
			}

			switch nextToken.Type {
			case TokenClass:
				return p.parseClassWithMacro()
			case TokenStruct:
				return p.parseStructWithMacro()
			case TokenEnum:
				return p.parseEnumWithMacro()
			}
		}
		// Fall through to default if not a macro or not followed by keyword
		return p.parseFunctionOrVariable()
	default:
		// Try to parse as function or variable
		return p.parseFunctionOrVariable()
	}
}
