package parser

import (
	"fmt"

	"doxyllm-it/pkg/ast"
)

// parseEnum handles enum declarations
func (p *Parser) parseEnum() error {
	start := p.tokenCache.getCurrentPosition()
	p.tokenCache.advance() // consume 'enum'

	p.tokenCache.skipWhitespace()

	// Check for 'class' or 'struct' after enum
	isScoped := false
	if !p.tokenCache.isAtEnd() && (p.tokenCache.peek().Type == TokenClass || p.tokenCache.peek().Type == TokenStruct) {
		isScoped = true
		p.tokenCache.advance()
		p.tokenCache.skipWhitespace()
	}

	if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
		return p.formatErrorAtCurrentPosition("expected enum name")
	}

	nameToken := p.tokenCache.advance()

	p.tokenCache.skipWhitespaceAndNewlines()

	// Parse underlying type if present
	underlyingType := ""
	if p.tokenCache.match(TokenColon) {
		underlyingType = p.parseType()
	}

	p.tokenCache.skipWhitespaceAndNewlines()

	// Build signature and handle body
	signature := "enum"
	if isScoped {
		signature += " class"
	}
	signature += " " + nameToken.Value
	if underlyingType != "" {
		signature += " : " + underlyingType
	}

	// Check if this enum has a body (opening brace)
	if p.tokenCache.match(TokenLeftBrace) {
		signature += " {"

		// Consume the entire enum body
		braceDepth := 1
		for !p.tokenCache.isAtEnd() && braceDepth > 0 {
			token := p.tokenCache.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			signature += token.Value
			p.tokenCache.advance()
		}

		// Consume optional semicolon after enum body
		if !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenSemicolon {
			signature += ";"
			p.tokenCache.advance()
		}
	}

	entity := &ast.Entity{
		Type:        ast.EntityEnum,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	return nil
}

// parseEnumWithMacro handles enum declarations preceded by macros
func (p *Parser) parseEnumWithMacro() error {
	start := p.tokenCache.getCurrentPosition()

	// Resolve the macro
	macroToken := p.tokenCache.advance()
	macroValue := p.resolveDefine(macroToken.Value)

	p.tokenCache.skipWhitespace()

	// Now parse the enum normally but include the macro in the signature
	enumToken := p.tokenCache.advance() // consume 'enum'

	p.tokenCache.skipWhitespace()

	if p.tokenCache.isAtEnd() || p.tokenCache.peek().Type != TokenIdentifier {
		return p.formatErrorAtCurrentPosition("expected enum name")
	}

	nameToken := p.tokenCache.advance()

	// Build signature with macro
	signature := fmt.Sprintf("%s %s %s", macroValue, enumToken.Value, nameToken.Value)

	p.tokenCache.skipWhitespace()

	// Check for enum class or enum struct
	if p.tokenCache.match(TokenClass) || p.tokenCache.match(TokenStruct) {
		signature += fmt.Sprintf(" %s", p.tokenCache.previous().Value)
	}

	// Check for base type
	if p.tokenCache.match(TokenColon) {
		signature += " :"
		for !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type != TokenLeftBrace && p.tokenCache.peek().Type != TokenSemicolon {
			signature += p.tokenCache.advance().Value
		}
	}

	// Handle body if present
	if p.tokenCache.match(TokenLeftBrace) {
		signature += " {"

		// Consume the entire enum body
		braceDepth := 1
		for !p.tokenCache.isAtEnd() && braceDepth > 0 {
			token := p.tokenCache.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			signature += token.Value
			p.tokenCache.advance()
		}
	} else if p.tokenCache.match(TokenSemicolon) {
		signature += ";"
	}

	entity := &ast.Entity{
		Type:        ast.EntityEnum,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)
	return nil
}
