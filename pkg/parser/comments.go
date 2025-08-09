package parser

import (
	"strings"

	"doxyllm-it/pkg/ast"
)

// parseComment handles comment parsing
func (p *Parser) parseComment() error {
	start := p.current
	token := p.advance()

	// Check if this is a Doxygen comment
	content := strings.TrimSpace(token.Value)
	if p.isDoxygenComment(content) {
		// Parse as Doxygen comment and store as pending
		p.pendingComment = p.parseDoxygenComment(content)
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
		SourceRange: p.getRangeFromTokens(start, p.current-1),
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
