package parser_new

import (
	"fmt"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parsePreprocessor handles preprocessor directives
func (p *Parser) parsePreprocessor() (*ast.Entity, error) {
	start := p.tokenizer.GetCurrentOffset()
	p.tokenizer.NextToken() // consume '#'

	p.tokenizer.SkipWhitespace()

	if p.tokenizer.IsAtEnd() {
		return nil, fmt.Errorf("unexpected end of input")
	}

	directive := p.tokenizer.PeekToken(0)

	if directive.Type == TokenIdentifier && directive.Value == "define" {
		return p.parseDefine(start)
	}

	// Other preprocessor directives
	return p.parseOtherPreprocessor(start)
}

// parseDefine handles #define directives
func (p *Parser) parseDefine(start int) (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume 'define'
	p.tokenizer.SkipWhitespace()

	if p.tokenizer.IsAtEnd() {
		return nil, p.formatErrorAtCurrentPosition("expected identifier after #define")
	}

	nameToken := p.tokenizer.PeekToken(0)
	if nameToken.Type != TokenIdentifier {
		return nil, p.formatError("expected identifier after #define", nameToken)
	}
	p.tokenizer.NextToken()

	// Collect the definition value until end of line or end of file
	var value strings.Builder
	depth := 0
	lastWasSpace := false

	for !p.tokenizer.IsAtEnd() {
		token := p.tokenizer.PeekToken(0)

		// Handle multiline macros with backslash continuation
		if token.Type == TokenBackslash {
			p.tokenizer.NextToken()
			// Skip the backslash and any following whitespace/newline
			p.tokenizer.SkipWhitespace()
			if !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type == TokenNewline {
				p.tokenizer.NextToken()
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
		p.tokenizer.NextToken()
	}

	// Store the define
	defineName := nameToken.Value
	defineValue := strings.TrimSpace(value.String())
	p.defines[defineName] = defineValue

	// Create preprocessor entity
	var signature string
	if strings.HasPrefix(defineValue, "(") {
		// Function-like macro: no space between name and parameters
		signature = fmt.Sprintf("#define %s%s", defineName, defineValue)
	} else {
		// Object-like macro: add space between name and value
		signature = fmt.Sprintf("#define %s %s", defineName, defineValue)
	}

	entity := &ast.Entity{
		Type:      ast.EntityPreprocessor,
		Name:      defineName,
		FullName:  defineName,
		Signature: signature,
	}

	return entity, nil
}

// parseOtherPreprocessor handles other preprocessor directives
func (p *Parser) parseOtherPreprocessor(start int) (*ast.Entity, error) {
	// Consume until end of line
	var content strings.Builder
	content.WriteString("#")

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenNewline {
		content.WriteString(p.tokenizer.PeekToken(0).Value)
		p.tokenizer.NextToken()
	}

	entity := &ast.Entity{
		Type:      ast.EntityPreprocessor,
		Name:      strings.TrimSpace(content.String()),
		FullName:  strings.TrimSpace(content.String()),
		Signature: content.String(),
	}

	return entity, nil
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
