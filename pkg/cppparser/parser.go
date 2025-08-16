// Package parser implements a streaming token-driven C++ header file parser with O(1) memory complexity
package cppparser

import (
	ast "doxyllm-it/pkg/ast"
	"fmt"
	"os"
	"strings"
)

// Parser implements a token-driven parser for C++ headers with streaming tokenizer backend
type Parser struct {
	tokenizer   *Tokenizer // The tokenizer used to generate tokens
	tree        *ast.ScopeTree
	scopeStack  []*ast.Entity
	accessStack []ast.AccessLevel
	defines     map[string]string
}

// formatError creates an error message with line and column information
func (p *Parser) formatError(token Token, message string) error {
	return fmt.Errorf("%s at line %d, column %d (token: '%s')", message, token.Line, token.Column, token.Value)
}

func (p *Parser) formatErrorf(token Token, format string, args ...interface{}) error {
	return fmt.Errorf(format+" at line %d, column %d (token: '%s')", append(args, token.Line, token.Column, token.Value)...)
}

// New creates a new token-driven parser
func New() *Parser {
	return &Parser{
		defines: make(map[string]string),
	}
}

// ParseFile is a convenience function to parse content from a file
func ParseFile(filename string) (*ast.ScopeTree, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	parser := New()
	return parser.Parse(filename, string(content))
}

// Parse parses tokens into an AST using streaming tokenizer with compatibility layer
func (p *Parser) Parse(filename, content string) (*ast.ScopeTree, error) {
	// Initialize the tree
	p.tree = ast.NewScopeTree(filename, content)
	p.scopeStack = []*ast.Entity{p.tree.Root}
	p.accessStack = []ast.AccessLevel{ast.AccessPublic} // Global scope is public

	// Initialize streaming tokenizer
	p.tokenizer = NewTokenizer(content)
	if p.tokenizer.HasErrors() {
		errors := p.tokenizer.GetErrors()
		if len(errors) > 0 {
			return nil, fmt.Errorf("tokenizer error: %s", errors[0].Value)
		}
	}

	for !p.tokenizer.IsAtEnd() {
		e, err := p.parseNext()
		if err != nil {
			return nil, err
		}
		if e == nil {
			continue // skip empty entities
		}
		e.Signature = cleanSpaces(e.Signature)
		p.addEntity(e)
	}

	return p.tree, nil
}

// parseNext parses top-level declarations
func (p *Parser) parseNext() (*ast.Entity, error) {

	p.tokenizer.SkipWhitespace()

	// Handle different token types
	token := p.tokenizer.PeekToken(0)

	switch token.Type {
	case TokenHash:
		return p.parsePreprocessor()
	case TokenLineComment, TokenBlockComment:
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
		return nil, p.parseCloseBrace()
	default:
		return p.parseDefault()
	}
}

func (p *Parser) parseDefault() (*ast.Entity, error) {
	defines := make([]string, 0)

	offset := 0 // Start looking after the identifier
	numIdentifiers := 0
	for !p.tokenizer.IsAtEnd() {
		token := p.tokenizer.PeekToken(offset)
		if IsKeyword(token) || IsLiteral(token) || IsWhitespace(token) {
			offset++
			continue
		}
		if IsSymbol(token) {
			switch token.Type {
			case TokenLeftParen:
				// Found a (, it must be a callable
				e, err := p.parseCallable()
				if err != nil {
					return nil, err
				}
				e.Defines = defines
				return e, nil
			case TokenSemicolon:
				// Found a ';' , it must be name(s)
				e, err := p.parseName()
				if err != nil {
					return nil, err
				}
				e.Defines = defines
				return e, nil
			}
			offset++
			continue
		}

		switch token.Type {
		case TokenIdentifier:
			if _, exists := p.defines[token.Value]; exists {
				defines = append(defines, token.Value)
			}
			numIdentifiers++
		case TokenClass, TokenStruct, TokenEnum:
			// Found a class keyword after identifier, parse as class
			if offset != len(defines) {
				return nil, p.formatError(token, "unexpected token")
			}
			// consume all tokens until now since we found a class
			for i := 0; i < offset; i++ {
				p.tokenizer.NextToken()
			}

			var e *ast.Entity
			var err error
			switch token.Type {
			case TokenClass:
				e, err = p.parseClass()
			case TokenEnum:
				e, err = p.parseEnum()
			case TokenStruct:
				e, err = p.parseStruct()
			default:
				return nil, p.formatError(token, "unexpected token")
			}

			if err != nil {
				return nil, err
			}

			e.Defines = defines
			return e, nil
		default:
			return nil, p.formatError(token, "unexpected token")
		}
		offset++
	}
	return nil, nil
}

// cleanSpaces returns a string with consecutive spaces replaced by a single space
func cleanSpaces(s string) string {
	out := make([]rune, 0, len(s))
	space := false
	for _, r := range s {
		if r == ' ' {
			if !space {
				out = append(out, r)
				space = true
			}
		} else {
			out = append(out, r)
			space = false
		}
	}
	// clean all leading and trailing spaces
	return strings.TrimSpace(string(out))
}
