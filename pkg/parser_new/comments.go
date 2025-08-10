package parser_new

import (
	ast "doxyllm-it/pkg/ast_new"
)

// parseComment handles comment parsing - NEW ARCHITECTURE
// Only identifies C++ comment syntax, does NOT interpret content
func (p *Parser) parseComment() error {
	startToken := p.tokenCache.advance()

	var fullCommentText string
	var endPosition ast.Position

	// Determine comment type and parse accordingly
	commentStart := startToken.Value

	if len(commentStart) >= 2 && commentStart[:2] == "//" {
		// Line comment - read until end of line
		fullCommentText = commentStart
		endPosition = ast.Position{
			Line:   startToken.Line,
			Column: startToken.Column + len(commentStart),
			Offset: startToken.Offset + len(commentStart),
		}

		// For line comments, the tokenizer should have captured the full line already

	} else if len(commentStart) >= 2 && commentStart[:2] == "/*" {
		// Block comment - should be fully contained in the token already
		fullCommentText = commentStart
		endPosition = ast.Position{
			Line:   startToken.Line,
			Column: startToken.Column + len(commentStart),
			Offset: startToken.Offset + len(commentStart),
		}
	} else {
		// Unknown comment format - use as is
		fullCommentText = commentStart
		endPosition = ast.Position{
			Line:   startToken.Line,
			Column: startToken.Column + len(commentStart),
			Offset: startToken.Offset + len(commentStart),
		}
	}

	// Create comment entity
	comment := &ast.Entity{
		Type:         ast.EntityComment,
		Name:         "comment",
		FullName:     "comment",
		Signature:    fullCommentText,
		AccessLevel:  ast.AccessUnknown,
		OriginalText: fullCommentText,
		LeadingWS:    "",
		TrailingWS:   "",
		SourceRange: ast.Range{
			Start: ast.Position{
				Line:   startToken.Line,
				Column: startToken.Column,
				Offset: startToken.Offset,
			},
			End: endPosition,
		},
	}

	// Store comment directly in the current scope
	// The Document layer will handle proper association with entities
	p.getCurrentScope().AddChild(comment)

	return nil
}
