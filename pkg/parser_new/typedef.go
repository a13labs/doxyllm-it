package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseTypedef handles typedef declarations
func (p *Parser) parseTypedef() (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume 'typedef'

	// Parse until we find the identifier and semicolon
	var signature strings.Builder
	signature.WriteString("typedef")

	childs := make([]*ast.Entity, 0)
	var structEntity *ast.Entity
	var err error
	var lastIdentifier string

	entity := &ast.Entity{
		Type:      ast.EntityTypedef,
		Signature: signature.String(),
	}

	p.tokenizer.SkipWhitespace()

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenSemicolon {

		p.tokenizer.SkipWhitespace()

		if p.tokenizer.IsAtEnd() || p.tokenizer.PeekToken(0).Type == TokenSemicolon {
			break
		}

		token := p.tokenizer.PeekToken(0)

		switch token.Type {
		case TokenIdentifier:
			childs = append(childs, &ast.Entity{
				Type:      ast.EntityIdentifier,
				Name:      token.Value,
				Signature: token.Value,
			})
			lastIdentifier = token.Value
		case TokenStruct:
			if structEntity != nil {
				return nil, p.formatErrorAtCurrentPosition("multiple structs in typedef")
			}
			structEntity, err = p.parseStruct()
			if err != nil {
				return nil, err
			}
			signature.WriteString(" " + structEntity.Name)
		}

		p.tokenizer.NextToken()
	}

	entity.Name = lastIdentifier // The last identifier is typically the typedef name

	if p.tokenizer.Match(TokenSemicolon) {
		signature.WriteString(";")
	}

	if structEntity != nil {
		entity.AddChild(structEntity)
	}

	for _, child := range childs {
		entity.AddChild(child)
	}

	return entity, nil
}
