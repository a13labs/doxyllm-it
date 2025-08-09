package parser

import (
	"fmt"
	"strings"

	"doxyllm-it/pkg/ast"
)

// parseComment handles comment parsing
func (p *Parser) parseComment() error {
	start := p.tokenCache.getCurrentPosition()
	token := p.tokenCache.advance()

	// Check if this is a Doxygen comment
	content := strings.TrimSpace(token.Value)
	if p.isDoxygenComment(content) {
		doxygenComment := p.parseDoxygenComment(content)
		
		// Check if this is a trailing comment (/**< or ///<)
		if p.isTrailingComment(content) {
			// Associate with the last entity instead of storing as pending
			if err := p.attachCommentToLastEntity(doxygenComment); err != nil {
				// If we can't attach to last entity, store as pending
				p.pendingComment = doxygenComment
			}
		} else {
			// Store as pending for next entity
			p.pendingComment = doxygenComment
		}
		return nil
	}

	// For regular comments, create comment entities
	name := "comment"

	// Try to extract first few words as name
	if len(content) > 2 {
		// Remove comment markers
		if strings.HasPrefix(content, "//") {
			content = strings.TrimSpace(content[2:])
		} else if strings.HasPrefix(content, "/*") && strings.HasSuffix(content, "*/") {
			content = strings.TrimSpace(content[2 : len(content)-2])
		}

		words := strings.Fields(content)
		if len(words) > 0 {
			if len(words) == 1 {
				name = words[0]
			} else {
				name = strings.Join(words[:min(3, len(words))], " ")
			}
		}
	}

	entity := &ast.Entity{
		Type:        ast.EntityComment,
		Name:        name,
		FullName:    name,
		Signature:   token.Value,
		AccessLevel: p.getCurrentAccessLevel(),
		SourceRange: p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
	}

	p.addEntity(entity)
	return nil
}

// isDoxygenComment checks if a comment is a Doxygen comment
func (p *Parser) isDoxygenComment(content string) bool {
	// Check for Doxygen comment patterns
	return strings.HasPrefix(content, "/**") ||
		strings.HasPrefix(content, "///") ||
		strings.HasPrefix(content, "//!")
}

// isTrailingComment checks if a comment is a trailing comment (should be associated with previous entity)
func (p *Parser) isTrailingComment(content string) bool {
	// Check for trailing comment patterns (/**< or ///<)
	return strings.HasPrefix(content, "/**<") ||
		strings.HasPrefix(content, "///<") ||
		strings.HasPrefix(content, "//!<")
}

// attachCommentToLastEntity attaches a comment to the last parsed entity
func (p *Parser) attachCommentToLastEntity(comment *ast.DoxygenComment) error {
	// Find the last entity in the current scope
	currentScope := p.getCurrentScope()
	if currentScope != nil && len(currentScope.Children) > 0 {
		lastEntity := currentScope.Children[len(currentScope.Children)-1]
		if lastEntity.Comment == nil {
			lastEntity.Comment = comment
			return nil
		}
	}
	
	// If we're at global scope, check the root entities
	if len(p.tree.Root.Children) > 0 {
		lastEntity := p.tree.Root.Children[len(p.tree.Root.Children)-1]
		if lastEntity.Comment == nil {
			lastEntity.Comment = comment
			return nil
		}
	}
	
	return fmt.Errorf("no suitable entity found to attach trailing comment")
}

// parseDoxygenComment parses a Doxygen comment string
func (p *Parser) parseDoxygenComment(content string) *ast.DoxygenComment {
	// Reuse the existing ParseDoxygenComment function from parser.go
	return ParseDoxygenComment(content)
}

// min returns the minimum of two integers (helper function)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
