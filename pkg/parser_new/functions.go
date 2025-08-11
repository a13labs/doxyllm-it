package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseFunction handles function declarations
func (p *Parser) parseFunction() (*ast.Entity, error) {

	// Parse specifiers and attributes first
	var isStatic, isInline, isVirtual, isConst, isConstexpr bool

	name := ""
	signature := strings.Builder{}
	hasBody := false
	numIdentifiers := 0
	inParameters := false
	sigReady := false
	// Since this is a function read all tokens until we find either ; or {
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenWhitespace, TokenNewline:
			if numIdentifiers == 0 {
				p.tokenizer.NextToken() // skip whitespace at start of signature
				continue
			}
			signature.WriteString(" ")
		case TokenSemicolon, TokenLeftBrace:
			if token.Type == TokenLeftBrace {
				hasBody = true
			}
			signature.WriteString(token.Value)
			p.tokenizer.NextToken()
			sigReady = true
			continue
		case TokenLeftParen:
			inParameters = true
			signature.WriteString(token.Value)
		case TokenStatic:
			isStatic = true
			signature.WriteString(token.Value)
		case TokenInline:
			isInline = true
			signature.WriteString(token.Value)
		case TokenVirtual:
			isVirtual = true
			signature.WriteString(token.Value)
		case TokenConstexpr:
			isConstexpr = true
			signature.WriteString(token.Value)
		case TokenIdentifier:
			if !inParameters {
				// first identifier that is not a macro is the name of the entity
				if _, exists := p.defines[token.Value]; !exists && name == "" {
					name = token.Value
				}
				numIdentifiers++
			}
			signature.WriteString(token.Value)
		default:
			signature.WriteString(token.Value)
		}
		signature.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	if !sigReady {
		return nil, p.formatErrorAtCurrentPosition("function signature is incomplete")
	}

	bodyText := ""
	if hasBody {
		// Check if there's a function body after the signature
		p.tokenizer.SkipWhitespace()
		if !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type == TokenLeftBrace {
			// This function has a body - track its range and content for the formatter
			braceDepth := 1
			var bodyTokens []Token
			bodyTokens = append(bodyTokens, p.tokenizer.NextToken()) // consume opening brace

			for !p.tokenizer.IsAtEnd() && braceDepth > 0 {
				token := p.tokenizer.PeekToken(0)
				if token.Type == TokenLeftBrace {
					braceDepth++
				} else if token.Type == TokenRightBrace {
					braceDepth--
				}
				bodyTokens = append(bodyTokens, p.tokenizer.NextToken())
			}

			if braceDepth == 0 {
				// Reconstruct body text from tokens
				var bodyBuilder strings.Builder
				for _, token := range bodyTokens {
					bodyBuilder.WriteString(token.Value)
				}
				bodyText = bodyBuilder.String()
			}
		}
	}

	entityType := ast.EntityFunction
	entitySignature := signature.String()
	entityName := p.buildFullName(name)

	// Check for special method types
	if strings.HasPrefix(entitySignature, "~") {
		// Detect destructor by checking for ~ in signature
		entityType = ast.EntityDestructor
	} else {
		if numIdentifiers == 1 {
			// If there's exactly one identifier until the parameter list, this is a constructor
			entityType = ast.EntityConstructor
		}
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        entityName,
		FullName:    p.buildFullName(name),
		Signature:   entitySignature,
		AccessLevel: p.getCurrentAccessLevel(),
		IsStatic:    isStatic,
		IsInline:    isInline,
		IsVirtual:   isVirtual,
		IsConst:     isConst,
		IsConstexpr: isConstexpr,
		Body:        bodyText,
	}

	return entity, nil
}
