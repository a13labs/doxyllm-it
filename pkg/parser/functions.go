package parser

import (
	"strings"

	"doxyllm-it/pkg/ast"
)

// parseFunctionOrVariable attempts to parse function or variable declarations
func (p *Parser) parseFunctionOrVariable() error {
	// Collect tokens until we can determine what this is
	checkpoint := p.tokenCache.getCurrentPosition()

	// Skip storage specifiers, cv-qualifiers, etc.
	p.skipSpecifiers()

	// Look for function pattern: type name(args) or name(args)
	if p.isFunction() {
		p.tokenCache.setPosition(checkpoint)
		return p.parseFunction()
	}

	// Otherwise, try as variable
	p.tokenCache.setPosition(checkpoint)
	return p.parseVariable()
}

// parseFunction handles function declarations
func (p *Parser) parseFunction() error {
	start := p.tokenCache.getCurrentPosition()

	// Parse specifiers and attributes first
	var isStatic, isInline, isVirtual, isConst bool

	// Parse function specifiers
	for !p.tokenCache.isAtEnd() {
		token := p.tokenCache.peek()
		if token.Type == TokenStatic {
			isStatic = true
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else if token.Type == TokenInline {
			isInline = true
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else if token.Type == TokenVirtual {
			isVirtual = true
			p.tokenCache.advance()
			p.tokenCache.skipWhitespace()
		} else {
			break
		}
	}

	signature, name, isMethod, err := p.parseFunctionSignature()
	if err != nil {
		return err
	}

	// Check if there's a function body after the signature
	var bodyRange *ast.Range
	var bodyText string
	p.tokenCache.skipWhitespace()
	if !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenLeftBrace {
		// This function has a body - track its range and content for the formatter
		bodyStart := p.tokenCache.getCurrentPosition()
		braceDepth := 1
		var bodyTokens []Token
		bodyTokens = append(bodyTokens, p.tokenCache.advance()) // consume opening brace

		for !p.tokenCache.isAtEnd() && braceDepth > 0 {
			token := p.tokenCache.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			bodyTokens = append(bodyTokens, p.tokenCache.advance())
		}

		if braceDepth == 0 {
			rangeValue := p.getRangeFromTokens(bodyStart, p.tokenCache.getCurrentPosition()-1)
			bodyRange = &rangeValue

			// Reconstruct body text from tokens
			var bodyBuilder strings.Builder
			for _, token := range bodyTokens {
				bodyBuilder.WriteString(token.Value)
			}
			bodyText = bodyBuilder.String()
		}
	}

	// Check for const after function signature
	if strings.Contains(signature, ") const") {
		isConst = true
	}

	entityType := ast.EntityFunction
	if isMethod {
		entityType = ast.EntityMethod

		// Check for special method types
		if strings.Contains(signature, "~") {
			// Detect destructor by checking for ~ in signature
			entityType = ast.EntityDestructor
		} else if name == p.getCurrentScope().Name {
			entityType = ast.EntityConstructor
		}
	}

	entity := &ast.Entity{
		Type:         entityType,
		Name:         name,
		FullName:     p.buildFullName(name),
		Signature:    signature,
		AccessLevel:  p.getCurrentAccessLevel(),
		IsStatic:     isStatic,
		IsInline:     isInline,
		IsVirtual:    isVirtual,
		IsConst:      isConst,
		SourceRange:  p.getRangeFromTokens(start, p.tokenCache.getCurrentPosition()-1),
		BodyRange:    bodyRange,
		OriginalText: bodyText,
	}

	p.addEntity(entity)
	return nil
}

// isFunction tries to determine if current position is a function
func (p *Parser) isFunction() bool {
	saved := p.tokenCache.getCurrentPosition()
	defer func() { p.tokenCache.setPosition(saved) }()

	// Skip specifiers first
	for !p.tokenCache.isAtEnd() {
		token := p.tokenCache.peek()
		if token.Type == TokenStatic || token.Type == TokenInline || token.Type == TokenVirtual ||
			token.Type == TokenExtern || token.Type == TokenConst || token.Type == TokenConstexpr {
			p.tokenCache.advance()
			p.tokenCache.skipWhitespaceAndNewlines()
		} else {
			break
		}
	}

	// Look for pattern: [return_type] function_name(
	identifiersSeen := 0

	for !p.tokenCache.isAtEnd() && identifiersSeen < 6 { // increased limit to handle more complex types
		token := p.tokenCache.peek()

		switch token.Type {
		case TokenVoid, TokenInt, TokenDouble, TokenChar, TokenFloat, TokenBool, TokenIdentifier:
			identifiersSeen++
			p.tokenCache.advance()
			p.tokenCache.skipWhitespaceAndNewlines() // Handle newlines between return type and function name
		case TokenLeftParen:
			// Found opening parenthesis, this looks like a function
			return identifiersSeen >= 1
		case TokenDoubleColon:
			p.tokenCache.advance()
			p.tokenCache.skipWhitespaceAndNewlines()
		case TokenStar, TokenAmpersand:
			p.tokenCache.advance() // Skip pointer/reference indicators
			p.tokenCache.skipWhitespaceAndNewlines()
		case TokenLess:
			// Skip template parameters
			depth := 1
			p.tokenCache.advance()
			for !p.tokenCache.isAtEnd() && depth > 0 {
				t := p.tokenCache.peek()
				switch t.Type {
				case TokenLess:
					depth++
				case TokenGreater:
					depth--
				}
				p.tokenCache.advance()
			}
			p.tokenCache.skipWhitespaceAndNewlines()
		case TokenTilde:
			// Destructor
			p.tokenCache.advance()
			p.tokenCache.skipWhitespaceAndNewlines()
		default:
			break
		}
	}

	return false
}

// parseFunctionSignature parses a complete function signature
func (p *Parser) parseFunctionSignature() (signature string, name string, isMethod bool, err error) {
	var sig strings.Builder
	var funcName string

	// Parse until we find the function name and parameters
	depth := 0
	beforeParen := true
	isDestructor := false

	for !p.tokenCache.isAtEnd() {
		token := p.tokenCache.peek()

		if token.Type == TokenLeftParen && depth == 0 && beforeParen {
			beforeParen = false
		}

		if token.Type == TokenLeftParen {
			depth++
		} else if token.Type == TokenRightParen {
			depth--
		}

		// Stop at semicolon (declaration) or opening brace (definition)
		if token.Type == TokenSemicolon && depth == 0 {
			break
		} else if token.Type == TokenLeftBrace && depth == 0 {
			// This is a function definition with a body
			// Don't include the body in the signature - the formatter handles it separately
			break
		}

		// Resolve defines in token values
		tokenValue := token.Value
		if token.Type == TokenIdentifier {
			tokenValue = p.resolveDefine(token.Value)
		}

		sig.WriteString(tokenValue)

		// Track the function name (last identifier before '(')
		if token.Type == TokenIdentifier && beforeParen {
			funcName = token.Value
		} else if token.Type == TokenTilde && beforeParen {
			// Handle destructor names
			isDestructor = true
			nextToken := p.tokenCache.getTokenAtOffset(1)
			if nextToken.Type == TokenIdentifier {
				funcName = nextToken.Value // Just the class name, not including ~
			}
		}

		p.tokenCache.advance()
	}

	// Consume semicolon if present
	if !p.tokenCache.isAtEnd() && p.tokenCache.peek().Type == TokenSemicolon {
		sig.WriteString(";")
		p.tokenCache.advance()
	}

	isMethod = p.isInsideClass()

	// If we didn't find a function name, try to extract it from the signature
	if funcName == "" && sig.Len() > 0 {
		// Look for the pattern before the opening parenthesis
		sigStr := sig.String()
		parenIndex := strings.Index(sigStr, "(")
		if parenIndex > 0 {
			beforeParen := strings.TrimSpace(sigStr[:parenIndex])
			parts := strings.Fields(beforeParen)
			if len(parts) > 0 {
				funcName = parts[len(parts)-1]
			}
		}
	}

	// For destructors, we need to return additional information
	if isDestructor {
		// We could return this as part of the signature or handle it differently
		// For now, the calling code will detect destructors by checking if the signature contains ~
	}

	return sig.String(), funcName, isMethod, nil
}
