package parser

import (
	"strings"

	"doxyllm-it/pkg/ast"
)

// parseFunctionOrVariable attempts to parse function or variable declarations
func (p *Parser) parseFunctionOrVariable() error {
	// Collect tokens until we can determine what this is
	checkpoint := p.current

	// Skip storage specifiers, cv-qualifiers, etc.
	p.skipSpecifiers()

	// Look for function pattern: type name(args) or name(args)
	if p.isFunction() {
		p.current = checkpoint
		return p.parseFunction()
	}

	// Otherwise, try as variable
	p.current = checkpoint
	return p.parseVariable()
}

// parseFunction handles function declarations
func (p *Parser) parseFunction() error {
	start := p.current

	// Parse specifiers and attributes first
	var isStatic, isInline, isVirtual, isConst bool

	// Parse function specifiers
	for !p.isAtEnd() {
		token := p.peek()
		if token.Type == TokenStatic {
			isStatic = true
			p.advance()
			p.skipWhitespace()
		} else if token.Type == TokenInline {
			isInline = true
			p.advance()
			p.skipWhitespace()
		} else if token.Type == TokenVirtual {
			isVirtual = true
			p.advance()
			p.skipWhitespace()
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
	p.skipWhitespace()
	if !p.isAtEnd() && p.peek().Type == TokenLeftBrace {
		// This function has a body - track its range and content for the formatter
		bodyStart := p.current
		braceDepth := 1
		var bodyTokens []Token
		bodyTokens = append(bodyTokens, p.advance()) // consume opening brace

		for !p.isAtEnd() && braceDepth > 0 {
			token := p.peek()
			if token.Type == TokenLeftBrace {
				braceDepth++
			} else if token.Type == TokenRightBrace {
				braceDepth--
			}
			bodyTokens = append(bodyTokens, p.advance())
		}

		if braceDepth == 0 {
			rangeValue := p.getRangeFromTokens(bodyStart, p.current-1)
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
		SourceRange:  p.getRangeFromTokens(start, p.current-1),
		BodyRange:    bodyRange,
		OriginalText: bodyText,
	}

	p.addEntity(entity)
	return nil
}

// isFunction tries to determine if current position is a function
func (p *Parser) isFunction() bool {
	saved := p.current
	defer func() { p.current = saved }()

	// Skip specifiers first
	for !p.isAtEnd() {
		token := p.peek()
		if token.Type == TokenStatic || token.Type == TokenInline || token.Type == TokenVirtual ||
			token.Type == TokenExtern || token.Type == TokenConst || token.Type == TokenConstexpr {
			p.advance()
			p.skipWhitespaceAndNewlines()
		} else {
			break
		}
	}

	// Look for pattern: [return_type] function_name(
	identifiersSeen := 0

	for !p.isAtEnd() && identifiersSeen < 6 { // increased limit to handle more complex types
		token := p.peek()

		if token.Type == TokenVoid || token.Type == TokenInt || token.Type == TokenDouble ||
			token.Type == TokenChar || token.Type == TokenFloat || token.Type == TokenBool ||
			token.Type == TokenIdentifier {
			identifiersSeen++
			p.advance()
			p.skipWhitespaceAndNewlines() // Handle newlines between return type and function name
		} else if token.Type == TokenLeftParen {
			// Found opening parenthesis, this looks like a function
			return identifiersSeen >= 1
		} else if token.Type == TokenDoubleColon {
			p.advance()
			p.skipWhitespaceAndNewlines()
		} else if token.Type == TokenStar || token.Type == TokenAmpersand {
			p.advance() // Skip pointer/reference indicators
			p.skipWhitespaceAndNewlines()
		} else if token.Type == TokenLess {
			// Skip template parameters
			depth := 1
			p.advance()
			for !p.isAtEnd() && depth > 0 {
				t := p.peek()
				if t.Type == TokenLess {
					depth++
				} else if t.Type == TokenGreater {
					depth--
				}
				p.advance()
			}
			p.skipWhitespaceAndNewlines()
		} else if token.Type == TokenTilde {
			// Destructor
			p.advance()
			p.skipWhitespaceAndNewlines()
		} else {
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

	for !p.isAtEnd() {
		token := p.peek()

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
			if !p.isAtEnd() && p.tokens[p.current+1].Type == TokenIdentifier {
				funcName = p.tokens[p.current+1].Value // Just the class name, not including ~
			}
		}

		p.advance()
	}

	// Consume semicolon if present
	if !p.isAtEnd() && p.peek().Type == TokenSemicolon {
		sig.WriteString(";")
		p.advance()
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
