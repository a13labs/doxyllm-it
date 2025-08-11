package parser_new

import (
	"fmt"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseNamespace handles namespace declarations
func (p *Parser) parseNamespace() (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume 'namespace'

	p.tokenizer.SkipWhitespace()

	if p.tokenizer.IsAtEnd() || !p.isValidIdentifierToken(p.tokenizer.PeekToken(0)) {
		return nil, p.formatErrorAtCurrentPosition("expected namespace name")
	}

	// Parse namespace name (could be nested like mgl::io)
	var nameBuilder strings.Builder
	nameBuilder.WriteString(p.tokenizer.NextToken().Value) // first identifier

	// Check for :: followed by more identifiers (nested namespace)
	for !p.tokenizer.IsAtEnd() {
		p.tokenizer.SkipWhitespace()
		if p.tokenizer.PeekToken(0).Type == TokenDoubleColon {
			nameBuilder.WriteString(p.tokenizer.NextToken().Value) // add ::
			p.tokenizer.SkipWhitespace()
			if p.isValidIdentifierToken(p.tokenizer.PeekToken(0)) {
				nameBuilder.WriteString(p.tokenizer.NextToken().Value) // add next identifier
			} else {
				break
			}
		} else {
			break
		}
	}

	p.tokenizer.SkipWhitespace()

	// The next token must be {
	if !p.tokenizer.Match(TokenLeftBrace) {
		return nil, p.formatErrorAtCurrentPosition("expected '{' after namespace")
	}

	// Build signature
	namespaceName := nameBuilder.String()
	signature := fmt.Sprintf("namespace %s", namespaceName)

	// Note: Opening braces are now handled by the main parser dispatch
	// We don't look for braces here anymore

	entity := &ast.Entity{
		Type:      ast.EntityNamespace,
		Name:      namespaceName,
		Signature: signature, // Clean signature without braces
		Children:  make([]*ast.Entity, 0),
	}

	return entity, nil
}
