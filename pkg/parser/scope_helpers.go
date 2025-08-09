package parser

import (
	"doxyllm-it/pkg/ast"
)

// getCurrentScope returns the current scope entity
func (p *Parser) getCurrentScope() *ast.Entity {
	if len(p.scopeStack) == 0 {
		return p.tree.Root
	}
	return p.scopeStack[len(p.scopeStack)-1]
}

// getCurrentAccessLevel returns the current access level
func (p *Parser) getCurrentAccessLevel() ast.AccessLevel {
	if len(p.accessStack) == 0 {
		return ast.AccessPublic
	}
	return p.accessStack[len(p.accessStack)-1]
}

// isInsideClass returns true if currently inside a class or struct
func (p *Parser) isInsideClass() bool {
	scope := p.getCurrentScope()
	return scope.Type == ast.EntityClass || scope.Type == ast.EntityStruct
}

// buildFullName builds the fully qualified name for an entity
func (p *Parser) buildFullName(name string) string {
	scope := p.getCurrentScope()
	if scope == p.tree.Root || scope.Name == "" {
		return name
	}
	return scope.GetFullPath() + "::" + name
}

// addEntity adds an entity to the current scope
func (p *Parser) addEntity(entity *ast.Entity) {
	// Associate pending comment with this entity
	if p.pendingComment != nil {
		entity.Comment = p.pendingComment
		p.pendingComment = nil // Clear the pending comment
	}

	scope := p.getCurrentScope()
	entity.Parent = scope
	scope.AddChild(entity)
	p.tree.AddEntity(entity)
}

// enterScope enters a new scope
func (p *Parser) enterScope(entity *ast.Entity) {
	p.scopeStack = append(p.scopeStack, entity)
}

// exitScope exits the current scope
func (p *Parser) exitScope() {
	if len(p.scopeStack) > 1 {
		p.scopeStack = p.scopeStack[:len(p.scopeStack)-1]
	}
	if len(p.accessStack) > 1 {
		p.accessStack = p.accessStack[:len(p.accessStack)-1]
	}
}

// getRangeFromTokens creates a range from token indices
func (p *Parser) getRangeFromTokens(start, end int) ast.Range {
	return p.tokenCache.getRangeFromPositions(start, end)
}
