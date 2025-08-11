package parser_new

import (
	"fmt"
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

// parseUsing handles using declarations
func (p *Parser) parseUsing() (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume 'using'

	p.tokenizer.SkipWhitespace()

	if p.tokenizer.IsAtEnd() {
		return nil, p.formatErrorAtCurrentPosition("expected identifier after using")
	}

	// Check for 'namespace' keyword
	if p.tokenizer.PeekToken(0).Type == TokenNamespace {
		return p.parseUsingNamespace()
	}

	// Parse the full qualified name (could include ::)
	var fullNameBuilder strings.Builder

	// First identifier is required
	if p.tokenizer.PeekToken(0).Type != TokenIdentifier {
		return nil, p.formatErrorAtCurrentPosition("expected identifier after using")
	}

	nameToken := p.tokenizer.NextToken()
	fullNameBuilder.WriteString(nameToken.Value)

	// Handle qualified names like std::data
	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type == TokenDoubleColon {
		fullNameBuilder.WriteString("::")
		p.tokenizer.NextToken()

		if p.tokenizer.IsAtEnd() || p.tokenizer.PeekToken(0).Type != TokenIdentifier {
			break
		}

		nextToken := p.tokenizer.NextToken()
		fullNameBuilder.WriteString(nextToken.Value)
		nameToken = nextToken // Update nameToken to be the last identifier
	}

	p.tokenizer.SkipWhitespace()

	var signature string
	var entityType ast.EntityType

	// Check if this is a type alias (using name = type) or a using declaration (using std::name)
	if p.tokenizer.PeekToken(0).Type == TokenEquals {
		// Type alias: using name = type
		p.tokenizer.NextToken() // consume '='

		// Parse the rest until semicolon
		var typeValue strings.Builder
		for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type != TokenSemicolon {
			typeValue.WriteString(p.tokenizer.PeekToken(0).Value)
			p.tokenizer.NextToken()
		}

		signature = fmt.Sprintf("using %s = %s", nameToken.Value, typeValue.String())
		entityType = ast.EntityUsing
	} else {
		// Using declaration: using std::name
		signature = fmt.Sprintf("using %s", fullNameBuilder.String())
		entityType = ast.EntityUsing
	}

	if p.tokenizer.Match(TokenSemicolon) {
		// consumed semicolon
	}

	entity := &ast.Entity{
		Type:      entityType,
		Name:      nameToken.Value,
		Signature: signature,
	}

	return entity, nil
}

// parseUsingNamespace handles using namespace declarations
func (p *Parser) parseUsingNamespace() (*ast.Entity, error) {
	p.tokenizer.NextToken() // consume 'namespace'

	p.tokenizer.SkipWhitespace()

	if p.tokenizer.IsAtEnd() || p.tokenizer.PeekToken(0).Type != TokenIdentifier {
		return nil, p.formatErrorAtCurrentPosition("expected namespace name after using namespace")
	}

	nameToken := p.tokenizer.NextToken()

	// Parse qualified namespace name
	var namespaceName strings.Builder
	namespaceName.WriteString(nameToken.Value)

	for !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type == TokenDoubleColon {
		namespaceName.WriteString("::")
		p.tokenizer.NextToken()

		if p.tokenizer.IsAtEnd() || p.tokenizer.PeekToken(0).Type != TokenIdentifier {
			break
		}

		namespaceName.WriteString(p.tokenizer.NextToken().Value)
	}

	if p.tokenizer.Match(TokenSemicolon) {
		// consumed semicolon
	}

	signature := fmt.Sprintf("using namespace %s", namespaceName.String())

	entity := &ast.Entity{
		Type:      ast.EntityUsing,
		Name:      namespaceName.String(),
		FullName:  namespaceName.String(),
		Signature: signature,
	}

	return entity, nil
}
