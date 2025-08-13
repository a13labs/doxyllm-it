package parser_new

import (
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// parseFunction handles function declarations
func (p *Parser) parseFunction() (*ast.Entity, error) {

	// Parse specifiers and attributes first
	var isStatic, isInline, isVirtual, isConst, isConstexpr bool

	lastIdentifier := ""
	signature := strings.Builder{}
	hasBody := false
	numIdentifiers := 0
	numKeywords := 0
	inParameters := false
	sigReady := false
	inInheritance := false
	inSpecialization := false
	isOperator := false
	// Since this is a function read all tokens until we find either ; or {
	for !p.tokenizer.IsAtEnd() && !sigReady {
		token := p.tokenizer.PeekToken(0)
		switch token.Type {
		case TokenWhitespace, TokenNewline:
			if numIdentifiers == 0 && numKeywords == 0 {
				p.tokenizer.NextToken() // skip whitespace at start of signature
				continue
			}
		case TokenSemicolon, TokenLeftBrace:
			if token.Type == TokenLeftBrace {
				hasBody = true
			}
			sigReady = true
			continue
		case TokenLess:
			inSpecialization = true
		case TokenColon:
			inInheritance = true
		case TokenLeftParen:
			inParameters = true
		case TokenIdentifier:
			if !inParameters && !inInheritance && !inSpecialization && !isOperator {
				numIdentifiers++
				lastIdentifier = token.Value
			}
		default:
			if IsSymbol(token) {
				if isOperator && !inParameters && !inInheritance && !inSpecialization {
					lastIdentifier += token.Value
				}
			}
			if IsKeyword(token) {
				if !inParameters && !inInheritance && !inSpecialization {
					numKeywords++
				}
				switch token.Type {
				case TokenStatic:
					isStatic = true
				case TokenInline:
					isInline = true
				case TokenVirtual:
					isVirtual = true
				case TokenConstexpr:
					isConstexpr = true
				case TokenConst:
					isConst = true
				case TokenOperator:
					isOperator = true
					lastIdentifier = token.Value // operator name is the keyword itself
				}
			}
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
	} else {
		// consume ;
		p.tokenizer.NextToken()
	}

	entityType := ast.EntityFunction
	entitySignature := strings.TrimSpace(signature.String())

	// Check for special method types
	if strings.HasPrefix(entitySignature, "~") {
		// Detect destructor by checking for ~ in signature
		entityType = ast.EntityDestructor
	} else {
		if numIdentifiers == 1 && numKeywords == 0 {
			// If there's exactly one identifier until the parameter list, this is a constructor
			entityType = ast.EntityConstructor
		}
	}

	entity := &ast.Entity{
		Type:        entityType,
		Name:        lastIdentifier,
		FullName:    p.buildFullName(lastIdentifier),
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
