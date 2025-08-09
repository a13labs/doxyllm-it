package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
)

// parsePreprocessor handles preprocessor directives
func (p *Parser) parsePreprocessor() error {
	start := p.current
	p.advance() // consume '#'

	p.skipWhitespace()

	if p.isAtEnd() {
		return nil
	}

	directive := p.peek()

	if directive.Type == TokenIdentifier && directive.Value == "define" {
		return p.parseDefine(start)
	}

	// Other preprocessor directives
	return p.parseOtherPreprocessor(start)
}

// parseDefine handles #define directives
func (p *Parser) parseDefine(start int) error {
	p.advance() // consume 'define'
	p.skipWhitespace()

	if p.isAtEnd() {
		return fmt.Errorf("expected identifier after #define")
	}

	nameToken := p.peek()
	if nameToken.Type != TokenIdentifier {
		return fmt.Errorf("expected identifier after #define, got %v", nameToken.Type)
	}
	p.advance()

	// Collect the definition value until end of line or end of file
	var value strings.Builder
	depth := 0
	lastWasSpace := false

	for !p.isAtEnd() {
		token := p.peek()

		// Handle multiline macros with backslash continuation
		if token.Type == TokenBackslash {
			p.advance()
			// Skip the backslash and any following whitespace/newline
			p.skipWhitespace()
			if !p.isAtEnd() && p.peek().Type == TokenNewline {
				p.advance()
			}
			value.WriteString(" ") // Replace backslash-newline with space
			lastWasSpace = true
			continue
		}

		if token.Type == TokenNewline && depth == 0 {
			break
		}

		// Track brace depth for complex macros
		if token.Type == TokenLeftBrace || token.Type == TokenLeftParen {
			depth++
		} else if token.Type == TokenRightBrace || token.Type == TokenRightParen {
			depth--
		}

		// Normalize whitespace - collapse multiple spaces into one
		if token.Type == TokenWhitespace {
			if !lastWasSpace {
				value.WriteString(" ")
				lastWasSpace = true
			}
		} else {
			value.WriteString(token.Value)
			lastWasSpace = false
		}
		p.advance()
	}

	// Store the define
	defineName := nameToken.Value
	defineValue := strings.TrimSpace(value.String())
	p.defines[defineName] = defineValue

	// Create preprocessor entity
	entity := &ast.Entity{
		Type:        ast.EntityPreprocessor,
		Name:        defineName,
		FullName:    defineName,
		Signature:   fmt.Sprintf("#define %s %s", defineName, defineValue),
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// parseOtherPreprocessor handles other preprocessor directives
func (p *Parser) parseOtherPreprocessor(start int) error {
	// Consume until end of line
	var content strings.Builder
	content.WriteString("#")

	for !p.isAtEnd() && p.peek().Type != TokenNewline {
		content.WriteString(p.peek().Value)
		p.advance()
	}

	entity := &ast.Entity{
		Type:        ast.EntityPreprocessor,
		Name:        strings.TrimSpace(content.String()),
		FullName:    strings.TrimSpace(content.String()),
		Signature:   content.String(),
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
	}

	p.addEntity(entity)
	return nil
}

// resolveDefine recursively resolves defines to their final values
func (p *Parser) resolveDefine(name string) string {
	visited := make(map[string]bool)
	return p.resolveDefineRecursive(name, visited)
}

// resolveDefineRecursive does the recursive resolution with cycle detection
func (p *Parser) resolveDefineRecursive(name string, visited map[string]bool) string {
	// Prevent infinite loops in circular definitions
	if visited[name] {
		return name
	}

	defineValue, exists := p.defines[name]
	if !exists {
		return name
	}

	visited[name] = true

	// Check if the define value is also a define
	if _, isDefine := p.defines[defineValue]; isDefine {
		return p.resolveDefineRecursive(defineValue, visited)
	}

	return defineValue
}
