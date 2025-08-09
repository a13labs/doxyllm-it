package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
)

// parseClass handles class declarations
func (p *Parser) parseClass() error {
	return p.parseClassOrStruct(ast.EntityClass)
}

// parseStruct handles struct declarations
func (p *Parser) parseStruct() error {
	return p.parseClassOrStruct(ast.EntityStruct)
}

// parseClassOrStruct handles both class and struct declarations
func (p *Parser) parseClassOrStruct(entityType ast.EntityType) error {
	start := p.current
	keyword := p.advance() // consume 'class' or 'struct'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected %s name", keyword.Value)
	}

	nameToken := p.advance()

	// Parse inheritance if present
	inheritance := ""
	p.skipWhitespace()
	if p.match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.skipWhitespace()

	// Build signature
	signature := fmt.Sprintf("%s %s", keyword.Value, nameToken.Value)
	if inheritance != "" {
		signature += " : " + inheritance
	}
	if p.match(TokenLeftBrace) {
		signature += " {"
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// Enter scope if we found opening brace
	if strings.Contains(signature, "{") {
		p.enterScope(entity)
		// Set default access level for the new scope
		if entityType == ast.EntityClass {
			p.accessStack = append(p.accessStack, ast.AccessPrivate) // class default
		} else {
			p.accessStack = append(p.accessStack, ast.AccessPublic) // struct default
		}
	}

	return nil
}

// parseInheritance parses class inheritance specification
func (p *Parser) parseInheritance() string {
	var inheritance strings.Builder

	for !p.isAtEnd() && p.peek().Type != TokenLeftBrace && p.peek().Type != TokenSemicolon {
		inheritance.WriteString(p.peek().Value)
		p.advance()
	}

	return strings.TrimSpace(inheritance.String())
}

// parseClassWithMacro handles class declarations preceded by macros
func (p *Parser) parseClassWithMacro() error {
	return p.parseClassOrStructWithMacro(ast.EntityClass)
}

// parseStructWithMacro handles struct declarations preceded by macros
func (p *Parser) parseStructWithMacro() error {
	return p.parseClassOrStructWithMacro(ast.EntityStruct)
}

// parseClassOrStructWithMacro handles class/struct declarations preceded by macros
func (p *Parser) parseClassOrStructWithMacro(entityType ast.EntityType) error {
	start := p.current

	// Resolve the macro
	macroToken := p.advance()
	macroValue := p.resolveDefine(macroToken.Value)

	p.skipWhitespace()

	// Now parse the class/struct normally but include the macro in the signature
	keyword := p.advance() // consume 'class' or 'struct'

	p.skipWhitespace()

	if p.isAtEnd() || p.peek().Type != TokenIdentifier {
		return fmt.Errorf("expected %s name", keyword.Value)
	}

	nameToken := p.advance()

	// Parse inheritance if present
	inheritance := ""
	p.skipWhitespace()
	if p.match(TokenColon) {
		inheritance = p.parseInheritance()
	}

	p.skipWhitespace()

	// Build signature with macro
	signature := fmt.Sprintf("%s %s %s", macroValue, keyword.Value, nameToken.Value)
	if inheritance != "" {
		signature += " : " + inheritance
	}
	if p.match(TokenLeftBrace) {
		signature += " {"
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        nameToken.Value,
		FullName:    p.buildFullName(nameToken.Value),
		Signature:   signature,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.current-1),
		Children:    make([]*ast.Entity, 0),
	}

	p.addEntity(entity)

	// Enter scope if we found opening brace
	if strings.Contains(signature, "{") {
		p.enterScope(entity)
		// Set default access level for the new scope
		if entityType == ast.EntityClass {
			p.accessStack = append(p.accessStack, ast.AccessPrivate) // class default
		} else {
			p.accessStack = append(p.accessStack, ast.AccessPublic) // struct default
		}
	}

	return nil
}
