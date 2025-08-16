package cppparser

import (
	"strings"

	ast "doxyllm-it/pkg/ast"
)

// parseCallable handles function declarations
func (p *Parser) parseCallable() (*ast.Entity, error) {
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
		depth            int
	)

	// Read tokens until we find either ; or {
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenWhitespace:
			if numIdentifiers == 0 && numKeywords == 0 {
				p.tokenizer.NextToken()
				continue
			}
		case TokenNewline:
			p.tokenizer.NextToken()
			continue
		case TokenSemicolon, TokenLeftBrace:
			hasBody = token.Type == TokenLeftBrace
			sigReady = true
			continue
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

	var body []string
	if hasBody {
		p.tokenizer.SkipWhitespace()
		body = p.parseCallableBody()
	} else {
		p.tokenizer.NextToken()
	}

	entityType := ast.EntityCallable
	entitySignature := strings.TrimSpace(signature.String())

	return &ast.Entity{
		Type:        entityType,
		Name:        lastIdentifier,
		FullName:    p.buildFullName(lastIdentifier),
		Signature:   entitySignature,
		AccessLevel: p.getCurrentAccessLevel(),
		Body:        body,
	}, nil
}

func (p *Parser) parseCallableBody() []string {
	var bodyLines []string
	var currentLine strings.Builder
	braceDepth := 0
	bodyReady := false
	for !p.tokenizer.IsAtEnd() && !bodyReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenNewline:
			p.tokenizer.NextToken()
			line := cleanSpaces(strings.TrimSpace(currentLine.String()))
			if line != "" {
				bodyLines = append(bodyLines, line)
			}
			currentLine.Reset()
			continue
		case TokenLeftBrace:
			braceDepth++
			p.tokenizer.NextToken()
			if braceDepth > 1 {
				currentLine.WriteString(token.Value)
			}
			continue
		case TokenRightBrace:
			braceDepth--
			p.tokenizer.NextToken()
			if braceDepth > 0 {
				currentLine.WriteString(token.Value)
			}
			if braceDepth == 0 {
				bodyReady = true
			}
			continue
		}
		currentLine.WriteString(token.Value)
		p.tokenizer.NextToken()
	}

	return bodyLines
}
