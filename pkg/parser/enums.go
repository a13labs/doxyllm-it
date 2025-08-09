package parser

import (
	"fmt"

	"doxyllm-it/pkg/ast"
)

// parseEnum handles enum declarations
func (p *Parser) parseEnum() error {
	start := p.current
	p.advance() // consume 'enum'

	p.skipWhitespace()

	// Check for 'class' or 'struct' after enum
	isScoped := false
	if !p.isAtEnd() && (p.peek().Type == TokenClass || p.peek().Type == TokenStruct) {
		isScoped = true
		p.advance()
		p.skipWhitespace()
	}

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected enum name")
	}

	nameToken := p.advance()

	p.skipWhitespace()

	// Parse underlying type if present
	underlyingType := ""
	if p.match(TokenColon) {
		underlyingType = p.parseType()
	}

	p.skipWhitespace()

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
	if p.match(TokenLeftBrace) {
		signature += " {"

		// Consume the entire enum body
		braceDepth := 1
		for !p.isAtEnd() && braceDepth > 0 {
			token := p.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			signature += token.Value
			p.advance()
		}

		// Consume optional semicolon after enum body
		if !p.isAtEnd() && p.peek().Type == TokenSemicolon {
			signature += ";"
			p.advance()
		}
	}

	entity := &ast.Entity{
		Type:        ast.EntityEnum,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	return nil
}

// parseEnumWithMacro handles enum declarations preceded by macros
func (p *Parser) parseEnumWithMacro() error {
	start := p.current

	// Resolve the macro
	macroToken := p.advance()
	macroValue := p.resolveDefine(macroToken.Value)

	p.skipWhitespace()

	// Now parse the enum normally but include the macro in the signature
	enumToken := p.advance() // consume 'enum'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected enum name")
	}

	nameToken := p.advance()

	// Build signature with macro
	signature := fmt.Sprintf("%s %s %s", macroValue, enumToken.Value, nameToken.Value)

	p.skipWhitespace()

	// Check for enum class or enum struct
	if p.match(TokenClass) || p.match(TokenStruct) {
		signature += fmt.Sprintf(" %s", p.previous().Value)
	}

	// Check for base type
	if p.match(TokenColon) {
		signature += " :"
		for !p.isAtEnd() && p.peek().Type != TokenLeftBrace && p.peek().Type != TokenSemicolon {
			signature += p.advance().Value
		}
	}

	// Handle body if present
	if p.match(TokenLeftBrace) {
		signature += " {"

		// Consume the entire enum body
		braceDepth := 1
		for !p.isAtEnd() && braceDepth > 0 {
			token := p.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			signature += token.Value
			p.advance()
		}
	} else if p.match(TokenSemicolon) {
		signature += ";"
	}

	entity := &ast.Entity{
		Type:        ast.EntityEnum,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)
	return nil
}
