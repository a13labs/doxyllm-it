package parser

import (
	"strings"

	ast "doxyllm-it/pkg/ast"
)

// parseNamespace handles namespace declarations
func (p *Parser) parseNamespace() (*ast.Entity, error) {

	signature := strings.Builder{}
	nameBuilder := strings.Builder{}
	signature.WriteString("namespace ")
	p.tokenizer.NextToken() // consume 'namespace'
	p.tokenizer.SkipWhitespace()

	var sigReady bool

	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		if token.Type == TokenLeftBrace {
			sigReady = true
			break
		}
		if token.Type == TokenNewline {
			p.tokenizer.NextToken() // consume newline
			continue
		}

		nameBuilder.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	if !sigReady {
		token := p.tokenizer.PeekToken(0)
		return nil, p.formatError(token, "unexpected end of namespace declaration")
	}

	// Consume the opening brace (entering a scope, this will be taken care addEntity)
	p.tokenizer.NextToken()

	namespaceName := strings.TrimRight(nameBuilder.String(), " ")
	signature.WriteString(namespaceName)

	return &ast.Entity{
		Type:      ast.EntityNamespace,
		Name:      namespaceName,
		FullName:  p.buildFullName(namespaceName),
		Signature: signature.String(), // Clean signature without braces
		Children:  make([]*ast.Entity, 0),
	}, nil
}
