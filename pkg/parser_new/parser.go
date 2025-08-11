// Package parser implements a streaming token-driven C++ header file parser with O(1) memory complexity
package parser_new

import (
	ast "doxyllm-it/pkg/ast_new"
	"fmt"
	"os"
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
func (p *Parser) formatError(message string, token Token) error {
	return fmt.Errorf("%s at line %d, column %d (token: '%s')", message, token.Line, token.Column, token.Value)
}

// formatErrorAtCurrentPosition creates an error message with current position information
func (p *Parser) formatErrorAtCurrentPosition(message string) error {
	if p.tokenizer.IsAtEnd() {
		return fmt.Errorf("%s at end of file", message)
	}
	token := p.tokenizer.PeekToken(0)
	return p.formatError(message, token)
}

// formatErrorAtCurrentPosition creates an error message with current position information
func (p *Parser) formatErrorAtCurrentPositionf(format string, args ...interface{}) error {
	if p.tokenizer.IsAtEnd() {
		return fmt.Errorf(format+" at end of file", args...)
	}
	token := p.tokenizer.PeekToken(0)
	return p.formatError(fmt.Sprintf(format, args...), token)
}

// New creates a new token-driven parser
func New() *Parser {
	return &Parser{
		defines: make(map[string]string),
	}
}

// ParseContent is a convenience function to parse content from a string
func ParseContent(filename, content string) (*ast.ScopeTree, error) {
	parser := New()
	return parser.Parse(filename, content)
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
		p.addEntity(e)
	}

	return p.tree, nil
}

// parseNext parses top-level declarations
func (p *Parser) parseNext() (*ast.Entity, error) {

	p.tokenizer.SkipWhitespaceAndNewlines()

	// Handle different token types
	token := p.tokenizer.PeekToken(0)

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
		return nil, p.parseCloseBrace()
	case TokenIdentifier:
		return p.parseIdentifier()
	default:
		return p.parseIdentifier()
	}
}

func (p *Parser) parseIdentifier() (*ast.Entity, error) {
	defines := make([]string, 0)

	offset := 0 // Start looking after the identifier
	numIdentifiers := 0
	for !p.tokenizer.IsAtEnd() {
		nextToken := p.tokenizer.PeekToken(offset)
		switch nextToken.Type {
		case TokenWhitespace, TokenNewline:
			offset++
		case TokenIdentifier:
			if _, exists := p.defines[nextToken.Value]; exists {
				defines = append(defines, nextToken.Value)
			}
			numIdentifiers++
			offset++ // Keep looking for the next token
		case TokenVoid, TokenInt, TokenDouble, TokenChar, TokenFloat, TokenBool, TokenStar, TokenAmpersand:
			// These are valid types, continue
			offset++ // Keep looking for the next token
		case TokenOperator, TokenDoubleColon, TokenLess:
			if numIdentifiers == 0 {
				return nil, p.formatErrorAtCurrentPositionf("There must be some identifiers before a '%s'", nextToken.Value)
			}
			offset++
		case TokenLeftParen:
			// Found a (, it must be a function
			e, err := p.parseFunction()
			if err != nil {
				return nil, err
			}
			e.Defines = defines
			return e, nil
		case TokenSemicolon:
			// Found a ';' , it must be variable(s)
			e, err := p.parseVariable()
			if err != nil {
				return nil, err
			}
			e.Defines = defines
			return e, nil
		case TokenClass, TokenStruct, TokenEnum:
			// Found a class keyword after identifier, parse as class
			if offset != len(defines) {
				return nil, p.formatErrorAtCurrentPositionf("All identifiers before a class must be a macro")
			}
			// consume all tokens until now since we found a class
			for i := 0; i < offset; i++ {
				p.tokenizer.NextToken()
			}

			var e *ast.Entity
			var err error
			switch nextToken.Type {
			case TokenClass:
				e, err = p.parseClass()
			case TokenEnum:
				e, err = p.parseEnum()
			case TokenStruct:
				e, err = p.parseStruct()
			default:
				return nil, p.formatErrorAtCurrentPositionf("unexpected token '%s' after identifier", nextToken.Value)
			}

			if err != nil {
				return nil, err
			}

			e.Defines = defines
			return e, nil
		default:
			return nil, p.formatErrorAtCurrentPositionf("unexpected token '%s' after identifier", nextToken.Value)
		}
	}
	return nil, p.formatErrorAtCurrentPositionf("unexpected end of input after identifier")
}

// // advance returns the current token and moves to the next
// func (p *Parser) advance() (Token, error) {
// 	if p.tokenizer.isAtEnd() {
// 		return Token{}, fmt.Errorf("unexpected end of input")
// 	}
// 	return p.tokenizer.NextToken(), nil
// }

// // isAtEnd checks if we're at the end of tokens
// func (p *Parser) isAtEnd() bool {
// 	return p.tokenizer.isAtEnd()
// }

// // peek returns the current token without advancing
// func (p *Parser) peek() Token {
// 	return p.tokenizer.PeekToken(1)
// }

// // peekAhead looks ahead by offset tokens
// func (p *Parser) peekAhead(offset int) Token {
// 	return p.tokenizer.PeekToken(offset)
// }

// // check returns true if current token is of given type
// func (p *Parser) check(tokenType TokenType) bool {
// 	if p.tokenizer.isAtEnd() {
// 		return false
// 	}
// 	return p.tokenizer.PeekToken(1).Type == tokenType
// }

// // skipWhitespace skips whitespace tokens
// func (p *Parser) skipWhitespace() {
// 	for !p.tokenizer.isAtEnd() && p.tokenizer.PeekToken(1).Type == TokenWhitespace {
// 		p.tokenizer.NextToken()
// 	}
// }

// // skipWhitespaceAndNewlines skips whitespace and newline tokens
// func (p *Parser) skipWhitespaceAndNewlines() {
// 	for !p.tokenizer.isAtEnd() && (p.tokenizer.PeekToken(1).Type == TokenWhitespace || p.tokenizer.PeekToken(1).Type == TokenNewline) {
// 		p.tokenizer.NextToken()
// 	}
// }
