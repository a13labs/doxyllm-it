package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseFunction handles function declarations
func (p *Parser) parseFunction() (*ast.Entity, error) {
	var (
		lastIdentifier   string
		signature        strings.Builder
		hasBody          bool
		numIdentifiers   int
		numKeywords      int
		inParameters     bool
		sigReady         bool
		inInheritance    bool
		inSpecialization bool
		isOperator       bool
		isDeduction      bool
		depth            int
	)

	// Read tokens until we find either ; or {
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenWhitespace, TokenNewline:
			if numIdentifiers == 0 && numKeywords == 0 {
				p.tokenizer.NextToken()
				continue
			}
		case TokenSemicolon, TokenLeftBrace:
			hasBody = token.Type == TokenLeftBrace
			sigReady = true
			continue
		case TokenArrow:
			if !isOperator {
				isDeduction = true
			}
		case TokenLess:
			depth++
			inSpecialization = true
		case TokenGreater:
			depth--
			if depth == 0 {
				inSpecialization = false
			}
		case TokenColon:
			inInheritance = true
		case TokenLeftParen:
			inParameters = true
		case TokenIdentifier:
			if !(inParameters || inInheritance || inSpecialization || isOperator) {
				numIdentifiers++
				lastIdentifier = token.Value
			}
		default:
			if IsSymbol(token) && isOperator && !(inParameters || inInheritance || inSpecialization) {
				lastIdentifier += token.Value
			}
			if IsKeyword(token) && !(inParameters || inInheritance || inSpecialization) {
				numKeywords++
				if token.Type == TokenOperator {
					isOperator = true
					lastIdentifier = token.Value
				}
			}
		}
		signature.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	if !sigReady {
		return nil, p.formatError(p.tokenizer.PeekToken(0), "function signature is incomplete")
	}

	var bodyText string
	if hasBody {
		p.tokenizer.SkipWhitespace()
		if !p.tokenizer.IsAtEnd() && p.tokenizer.PeekToken(0).Type == TokenLeftBrace {
			braceDepth := 1
			var bodyTokens []Token
			bodyTokens = append(bodyTokens, p.tokenizer.NextToken())
			for !p.tokenizer.IsAtEnd() && braceDepth > 0 {
				token := p.tokenizer.PeekToken(0)
				switch token.Type {
				case TokenLeftBrace:
					braceDepth++
				case TokenRightBrace:
					braceDepth--
				}
				bodyTokens = append(bodyTokens, p.tokenizer.NextToken())
			}
			if braceDepth == 0 {
				var bodyBuilder strings.Builder
				for _, t := range bodyTokens {
					bodyBuilder.WriteString(t.Value)
				}
				bodyText = bodyBuilder.String()
			}
		}
	} else {
		p.tokenizer.NextToken()
	}

	entityType := ast.EntityFunction
	entitySignature := strings.TrimSpace(signature.String())
	switch {
	case strings.HasPrefix(entitySignature, "~"):
		entityType = ast.EntityDestructor
	case numIdentifiers == 1 && numKeywords == 0:
		entityType = ast.EntityConstructor
	}

	return &ast.Entity{
		Type:        entityType,
		Name:        lastIdentifier,
		FullName:    p.buildFullName(lastIdentifier),
		Signature:   entitySignature,
		IsDeduction: isDeduction,
		AccessLevel: p.getCurrentAccessLevel(),
		Body:        bodyText,
	}, nil
}
