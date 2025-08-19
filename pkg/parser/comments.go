package parser

import (
	ast "doxyllm-it/pkg/ast"
)

// parseComment handles comment parsing - NEW ARCHITECTURE
// Only identifies C++ comment syntax, does NOT interpret content
func (p *Parser) parseComment() (*ast.Entity, error) {

	fullCommentText := p.tokenizer.NextToken().Value

	// Create comment entity
	comment := &ast.Entity{
		Type:        ast.EntityComment,
		Name:        "comment",
		AccessLevel: ast.AccessUnknown,
		Signature:   fullCommentText,
	}

	return comment, nil
}
