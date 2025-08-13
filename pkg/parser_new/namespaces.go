package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseNamespace handles namespace declarations
func (p *Parser) parseNamespace() (*ast.Entity, error) {

	signature := strings.Builder{}
	nameBuilder := strings.Builder{}
	signature.WriteString("namespace ")
	p.tokenizer.NextToken() // consume 'namespace'
	p.tokenizer.SkipWhitespace()

	sigReady := false
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		if token.Type == TokenLeftBrace {
			sigReady = true
			break
		}

		nameBuilder.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	if !sigReady {
		return nil, p.formatErrorAtCurrentPosition("expected '{' after namespace name")
	}

	// Consume the opening brace (entering a scope, this will be taken care addEntity)
	p.tokenizer.NextToken()

	namespaceName := strings.Trim(nameBuilder.String(), " ")
	signature.WriteString(namespaceName)

	entity := &ast.Entity{
		Type:      ast.EntityNamespace,
		Name:      namespaceName,
		FullName:  namespaceName,
		Signature: signature.String(), // Clean signature without braces
		Children:  make([]*ast.Entity, 0),
	}

	return entity, nil
}
