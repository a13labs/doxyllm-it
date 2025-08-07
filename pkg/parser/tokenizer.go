// Package parser - tokenizer implementation for C++ header files
package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenWhitespace
	TokenNewline
	TokenLineComment    // //
	TokenBlockComment   // /* */
	TokenDoxygenComment // /** */ or ///

	// Literals
	TokenIdentifier
	TokenNumber
	TokenString
	TokenCharLiteral

	// Operators and punctuation
	TokenLeftParen    // (
	TokenRightParen   // )
	TokenLeftBrace    // {
	TokenRightBrace   // }
	TokenLeftBracket  // [
	TokenRightBracket // ]
	TokenSemicolon    // ;
	TokenColon        // :
	TokenDoubleColon  // ::
	TokenComma        // ,
	TokenDot          // .
	TokenArrow        // ->
	TokenEquals       // =
	TokenDoubleEquals // ==
	TokenNotEquals    // !=
	TokenLess         // <
	TokenGreater      // >
	TokenLessEqual    // <=
	TokenGreaterEqual // >=
	TokenAmpersand    // &
	TokenDoubleAmp    // &&
	TokenPipe         // |
	TokenDoublePipe   // ||
	TokenCaret        // ^
	TokenTilde        // ~
	TokenExclamation  // !
	TokenQuestion     // ?
	TokenPlus         // +
	TokenMinus        // -
	TokenStar         // *
	TokenSlash        // /
	TokenPercent      // %
	TokenPlusPlus     // ++
	TokenMinusMinus   // --
	TokenPlusEquals   // +=
	TokenMinusEquals  // -=
	TokenStarEquals   // *=
	TokenSlashEquals  // /=
	TokenLeftShift    // <<
	TokenRightShift   // >>

	// Preprocessor
	TokenHash      // #
	TokenHashHash  // ##
	TokenBackslash // \

	// Keywords
	TokenKeywordStart // Marker for start of keywords
	TokenNamespace
	TokenClass
	TokenStruct
	TokenEnum
	TokenUnion
	TokenTypedef
	TokenUsing
	TokenTemplate
	TokenTypename
	TokenPublic
	TokenPrivate
	TokenProtected
	TokenStatic
	TokenVirtual
	TokenInline
	TokenConst
	TokenConstexpr
	TokenMutable
	TokenExtern
	TokenVolatile
	TokenFriend
	TokenOperator
	TokenExplicit
	TokenOverride
	TokenFinal
	TokenNoexcept
	TokenThrow
	TokenTry
	TokenCatch
	TokenIf
	TokenElse
	TokenSwitch
	TokenCase
	TokenDefault
	TokenFor
	TokenWhile
	TokenDo
	TokenBreak
	TokenContinue
	TokenReturn
	TokenGoto
	TokenSizeof
	TokenAlignof
	TokenDecltype
	TokenAuto
	TokenVoid
	TokenBool
	TokenChar
	TokenShort
	TokenInt
	TokenLong
	TokenFloat
	TokenDouble
	TokenSigned
	TokenUnsigned
	TokenTrue
	TokenFalse
	TokenNullptr
	TokenThis
	TokenNew
	TokenDelete
	TokenKeywordEnd // Marker for end of keywords
)

// Token represents a single token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
	Offset int
}

// String returns a string representation of the token
func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return fmt.Sprintf("ERROR:%s", t.Value)
	case TokenWhitespace:
		return "WHITESPACE"
	case TokenNewline:
		return "NEWLINE"
	case TokenLineComment:
		return fmt.Sprintf("LINE_COMMENT:%s", t.Value)
	case TokenBlockComment:
		return fmt.Sprintf("BLOCK_COMMENT:%s", t.Value)
	case TokenDoxygenComment:
		return fmt.Sprintf("DOXYGEN_COMMENT:%s", t.Value)
	case TokenIdentifier:
		return fmt.Sprintf("IDENTIFIER:%s", t.Value)
	case TokenNumber:
		return fmt.Sprintf("NUMBER:%s", t.Value)
	case TokenString:
		return fmt.Sprintf("STRING:%s", t.Value)
	case TokenCharLiteral:
		return fmt.Sprintf("CHAR:%s", t.Value)
	default:
		if t.Type >= TokenKeywordStart && t.Type <= TokenKeywordEnd {
			return fmt.Sprintf("KEYWORD:%s", t.Value)
		}
		return fmt.Sprintf("%s:%s", tokenTypeNames[t.Type], t.Value)
	}
}

// tokenTypeNames maps token types to their names for debugging
var tokenTypeNames = map[TokenType]string{
	TokenLeftParen:    "LEFT_PAREN",
	TokenRightParen:   "RIGHT_PAREN",
	TokenLeftBrace:    "LEFT_BRACE",
	TokenRightBrace:   "RIGHT_BRACE",
	TokenLeftBracket:  "LEFT_BRACKET",
	TokenRightBracket: "RIGHT_BRACKET",
	TokenSemicolon:    "SEMICOLON",
	TokenColon:        "COLON",
	TokenDoubleColon:  "DOUBLE_COLON",
	TokenComma:        "COMMA",
	TokenDot:          "DOT",
	TokenArrow:        "ARROW",
	TokenEquals:       "EQUALS",
	TokenDoubleEquals: "DOUBLE_EQUALS",
	TokenNotEquals:    "NOT_EQUALS",
	TokenLess:         "LESS",
	TokenGreater:      "GREATER",
	TokenLessEqual:    "LESS_EQUAL",
	TokenGreaterEqual: "GREATER_EQUAL",
	TokenAmpersand:    "AMPERSAND",
	TokenDoubleAmp:    "DOUBLE_AMP",
	TokenPipe:         "PIPE",
	TokenDoublePipe:   "DOUBLE_PIPE",
	TokenCaret:        "CARET",
	TokenTilde:        "TILDE",
	TokenExclamation:  "EXCLAMATION",
	TokenQuestion:     "QUESTION",
	TokenPlus:         "PLUS",
	TokenMinus:        "MINUS",
	TokenStar:         "STAR",
	TokenSlash:        "SLASH",
	TokenPercent:      "PERCENT",
	TokenPlusPlus:     "PLUS_PLUS",
	TokenMinusMinus:   "MINUS_MINUS",
	TokenPlusEquals:   "PLUS_EQUALS",
	TokenMinusEquals:  "MINUS_EQUALS",
	TokenStarEquals:   "STAR_EQUALS",
	TokenSlashEquals:  "SLASH_EQUALS",
	TokenLeftShift:    "LEFT_SHIFT",
	TokenRightShift:   "RIGHT_SHIFT",
	TokenHash:         "HASH",
	TokenHashHash:     "HASH_HASH",
	TokenBackslash:    "BACKSLASH",
}

// Keywords map for quick lookup
var keywords = map[string]TokenType{
	"namespace": TokenNamespace,
	"class":     TokenClass,
	"struct":    TokenStruct,
	"enum":      TokenEnum,
	"union":     TokenUnion,
	"typedef":   TokenTypedef,
	"using":     TokenUsing,
	"template":  TokenTemplate,
	"typename":  TokenTypename,
	"public":    TokenPublic,
	"private":   TokenPrivate,
	"protected": TokenProtected,
	"static":    TokenStatic,
	"virtual":   TokenVirtual,
	"inline":    TokenInline,
	"const":     TokenConst,
	"constexpr": TokenConstexpr,
	"mutable":   TokenMutable,
	"extern":    TokenExtern,
	"volatile":  TokenVolatile,
	"friend":    TokenFriend,
	"operator":  TokenOperator,
	"explicit":  TokenExplicit,
	"override":  TokenOverride,
	"final":     TokenFinal,
	"noexcept":  TokenNoexcept,
	"throw":     TokenThrow,
	"try":       TokenTry,
	"catch":     TokenCatch,
	"if":        TokenIf,
	"else":      TokenElse,
	"switch":    TokenSwitch,
	"case":      TokenCase,
	"default":   TokenDefault,
	"for":       TokenFor,
	"while":     TokenWhile,
	"do":        TokenDo,
	"break":     TokenBreak,
	"continue":  TokenContinue,
	"return":    TokenReturn,
	"goto":      TokenGoto,
	"sizeof":    TokenSizeof,
	"alignof":   TokenAlignof,
	"decltype":  TokenDecltype,
	"auto":      TokenAuto,
	"void":      TokenVoid,
	"bool":      TokenBool,
	"char":      TokenChar,
	"short":     TokenShort,
	"int":       TokenInt,
	"long":      TokenLong,
	"float":     TokenFloat,
	"double":    TokenDouble,
	"signed":    TokenSigned,
	"unsigned":  TokenUnsigned,
	"true":      TokenTrue,
	"false":     TokenFalse,
	"nullptr":   TokenNullptr,
	"this":      TokenThis,
	"new":       TokenNew,
	"delete":    TokenDelete,
}

// Tokenizer represents the tokenizer state
type Tokenizer struct {
	input     string
	pos       int // current position in input
	line      int // current line number
	column    int // current column number
	width     int // width of last rune read
	start     int // start position of current token
	tokens    []Token
	maxTokens int // Maximum number of tokens to prevent OOM
	maxPos    int // Maximum position to prevent infinite loops
}

// NewTokenizer creates a new tokenizer
func NewTokenizer(input string) *Tokenizer {
	const maxTokensLimit = 100000 // Prevent OOM from too many tokens
	return &Tokenizer{
		input:     input,
		line:      1,
		column:    1,
		tokens:    make([]Token, 0, 1024), // Pre-allocate reasonable capacity
		maxTokens: maxTokensLimit,
		maxPos:    len(input) + 1000, // Allow some buffer but prevent runaway
	}
}

// next reads the next rune and advances position
func (t *Tokenizer) next() rune {
	if t.pos >= len(t.input) {
		t.width = 0
		return 0
	}

	r, w := utf8.DecodeRuneInString(t.input[t.pos:])
	t.width = w
	t.pos += w

	if r == '\n' {
		t.line++
		t.column = 1
	} else {
		t.column++
	}

	return r
}

// backup steps back one rune
func (t *Tokenizer) backup() {
	t.pos -= t.width
	if t.pos < len(t.input) && t.input[t.pos] == '\n' {
		t.line--
		// Recalculate column by scanning back to start of line
		col := 1
		for i := t.pos - 1; i >= 0 && t.input[i] != '\n'; i-- {
			col++
		}
		t.column = col
	} else {
		t.column--
	}
}

// peek returns the next rune without advancing position
func (t *Tokenizer) peek() rune {
	r := t.next()
	t.backup()
	return r
}

// peekN returns the nth rune ahead without advancing position
// Currently unused but kept for future extensions
func (t *Tokenizer) peekN(n int) rune {
	pos := t.pos
	line := t.line
	column := t.column

	var r rune
	for i := 0; i < n; i++ {
		r = t.next()
		if r == 0 {
			break
		}
	}

	// Restore position
	t.pos = pos
	t.line = line
	t.column = column

	return r
}

// emit creates a token and adds it to the tokens slice
func (t *Tokenizer) emit(tokenType TokenType) {
	// Safeguard: Check if we've exceeded maximum tokens
	if len(t.tokens) >= t.maxTokens {
		if tokenType != TokenError { // Avoid infinite recursion
			// Create error token manually to ensure it gets added
			errorToken := Token{
				Type:   TokenError,
				Value:  "too many tokens - possible infinite loop or memory exhaustion",
				Line:   t.line,
				Column: t.column,
				Offset: t.start,
			}
			t.tokens = append(t.tokens, errorToken)
		}
		return
	}

	value := t.input[t.start:t.pos]
	token := Token{
		Type:   tokenType,
		Value:  value,
		Line:   t.line,
		Column: t.column - len(value),
		Offset: t.start,
	}
	t.tokens = append(t.tokens, token)
	t.start = t.pos
}

// emitError creates an error token
func (t *Tokenizer) emitError(message string) {
	token := Token{
		Type:   TokenError,
		Value:  message,
		Line:   t.line,
		Column: t.column,
		Offset: t.start,
	}
	t.tokens = append(t.tokens, token)
	t.start = t.pos
}

// ignore discards the current token
// Currently unused but kept for future extensions
func (t *Tokenizer) ignore() {
	t.start = t.pos
}

// acceptRun consumes a run of runes from the valid set
// Currently unused but kept for future extensions
func (t *Tokenizer) acceptRun(valid string) {
	for strings.ContainsRune(valid, t.next()) {
	}
	t.backup()
}

// Tokenize processes the input and returns all tokens
func (t *Tokenizer) Tokenize() []Token {
	iterations := 0
	const maxIterations = 1000000 // Prevent infinite loops

	for t.pos < len(t.input) {
		// Safeguard: Check for infinite loops
		iterations++
		if iterations > maxIterations {
			t.emitError("tokenizer exceeded maximum iterations - possible infinite loop")
			break
		}

		// Safeguard: Check position bounds
		if t.pos > t.maxPos {
			t.emitError("tokenizer position exceeded maximum bounds")
			break
		}

		// Safeguard: Check if position is advancing
		oldPos := t.pos

		r := t.next()

		switch {
		case r == 0:
			// EOF
			return append(t.tokens, Token{Type: TokenEOF, Line: t.line, Column: t.column, Offset: t.pos})

		case unicode.IsSpace(r):
			if r == '\n' {
				t.emit(TokenNewline)
			} else {
				t.scanWhitespace()
			}

		case r == '/':
			if !t.scanComment() {
				// If it's not a comment, we need to handle operators like /=
				t.scanSlashOperator()
			}

		case r == '#':
			t.scanHash()

		case r == '"':
			t.scanString()

		case r == '\'':
			t.scanChar()

		case unicode.IsLetter(r) || r == '_':
			t.scanIdentifier()

		case unicode.IsDigit(r):
			t.scanNumber()

		default:
			t.scanOperator()
		}

		// Safeguard: Ensure position advanced
		if t.pos == oldPos {
			t.emitError(fmt.Sprintf("tokenizer stuck at position %d", t.pos))
			t.pos++ // Force advance to prevent infinite loop
		}

		// Safeguard: Check if we have too many tokens
		if len(t.tokens) >= t.maxTokens {
			break
		}
	}

	// Only add EOF if we haven't exceeded the token limit
	if len(t.tokens) < t.maxTokens {
		return append(t.tokens, Token{Type: TokenEOF, Line: t.line, Column: t.column, Offset: t.pos})
	}

	return t.tokens
}

// HasErrors returns true if the tokenizer encountered any errors
func (t *Tokenizer) HasErrors() bool {
	for _, token := range t.tokens {
		if token.Type == TokenError {
			return true
		}
	}
	return false
}

// GetErrors returns all error tokens
func (t *Tokenizer) GetErrors() []Token {
	var errors []Token
	for _, token := range t.tokens {
		if token.Type == TokenError {
			errors = append(errors, token)
		}
	}
	return errors
}

// SetMaxTokens sets the maximum number of tokens (for testing purposes)
func (t *Tokenizer) SetMaxTokens(max int) {
	t.maxTokens = max
}

// scanSlashOperator handles the / character that wasn't part of a comment
func (t *Tokenizer) scanSlashOperator() {
	// We've already consumed the first '/', check for operators
	if t.peek() == '=' {
		t.next()
		t.emit(TokenSlashEquals)
	} else {
		t.emit(TokenSlash)
	}
}

// scanWhitespace scans whitespace characters
func (t *Tokenizer) scanWhitespace() {
	count := 0
	const maxWhitespace = 10000 // Prevent infinite loops on whitespace

	for {
		r := t.peek()
		if !unicode.IsSpace(r) || r == '\n' {
			break
		}
		count++
		if count > maxWhitespace {
			t.emitError("excessive whitespace - possible infinite loop")
			break
		}
		t.next()
	}
	t.emit(TokenWhitespace)
}

// scanComment scans comments and returns true if a comment was found
func (t *Tokenizer) scanComment() bool {
	// We've already consumed one '/'
	r := t.peek()

	if r == '/' {
		// Line comment
		t.next() // consume second '/'

		// Check if it's a doxygen comment (///)
		if t.peek() == '/' {
			t.next() // consume third '/'
			t.scanLineComment()
			t.emit(TokenDoxygenComment)
		} else if t.peek() == '!' {
			t.next() // consume '!'
			t.scanLineComment()
			t.emit(TokenDoxygenComment)
		} else {
			t.scanLineComment()
			t.emit(TokenLineComment)
		}
		return true

	} else if r == '*' {
		// Block comment
		t.next() // consume '*'

		// Check if it's a doxygen comment (/**)
		if t.peek() == '*' {
			t.next() // consume second '*'
			t.scanBlockComment()
			t.emit(TokenDoxygenComment)
		} else if t.peek() == '!' {
			t.next() // consume '!'
			t.scanBlockComment()
			t.emit(TokenDoxygenComment)
		} else {
			t.scanBlockComment()
			t.emit(TokenBlockComment)
		}
		return true
	}

	// Not a comment, don't backup since we haven't consumed anything extra
	return false
}

// scanLineComment scans until end of line
func (t *Tokenizer) scanLineComment() {
	count := 0
	const maxCommentLength = 100000 // Prevent infinite loops in comments

	for {
		r := t.next()
		count++
		if count > maxCommentLength {
			t.emitError("comment too long - possible infinite loop")
			break
		}
		if r == '\n' || r == 0 {
			t.backup()
			break
		}
	}
}

// scanBlockComment scans until */
func (t *Tokenizer) scanBlockComment() {
	count := 0
	const maxCommentLength = 100000 // Prevent infinite loops in comments

	for {
		r := t.next()
		count++
		if count > maxCommentLength {
			t.emitError("block comment too long - possible infinite loop")
			return
		}
		if r == 0 {
			t.emitError("unterminated block comment")
			return
		}
		if r == '*' && t.peek() == '/' {
			t.next() // consume '/'
			break
		}
	}
}

// scanHash scans hash and hash-hash operators
func (t *Tokenizer) scanHash() {
	if t.peek() == '#' {
		t.next()
		t.emit(TokenHashHash)
	} else {
		t.emit(TokenHash)
	}
}

// scanString scans a string literal
func (t *Tokenizer) scanString() {
	count := 0
	const maxStringLength = 100000 // Prevent infinite loops in strings

	for {
		r := t.next()
		count++
		if count > maxStringLength {
			t.emitError("string literal too long - possible infinite loop")
			return
		}
		if r == 0 || r == '\n' {
			t.emitError("unterminated string literal")
			return
		}
		if r == '"' {
			break
		}
		if r == '\\' {
			// Skip escaped character
			nextR := t.next()
			if nextR == 0 {
				t.emitError("unterminated string literal - EOF after escape")
				return
			}
			count++
		}
	}
	t.emit(TokenString)
}

// scanChar scans a character literal
func (t *Tokenizer) scanChar() {
	count := 0
	const maxCharLength = 10 // Character literals should be very short

	for {
		r := t.next()
		count++
		if count > maxCharLength {
			t.emitError("character literal too long - possible infinite loop")
			return
		}
		if r == 0 || r == '\n' {
			t.emitError("unterminated character literal")
			return
		}
		if r == '\'' {
			break
		}
		if r == '\\' {
			// Skip escaped character
			nextR := t.next()
			if nextR == 0 {
				t.emitError("unterminated character literal - EOF after escape")
				return
			}
			count++
		}
	}
	t.emit(TokenCharLiteral)
}

// scanIdentifier scans an identifier or keyword
func (t *Tokenizer) scanIdentifier() {
	count := 0
	const maxIdentifierLength = 1000 // Reasonable limit for identifiers

	for {
		r := t.peek()
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		count++
		if count > maxIdentifierLength {
			t.emitError("identifier too long - possible infinite loop")
			break
		}
		t.next()
	}

	value := t.input[t.start:t.pos]
	if tokenType, isKeyword := keywords[value]; isKeyword {
		t.emit(tokenType)
	} else {
		t.emit(TokenIdentifier)
	}
}

// scanNumber scans a numeric literal
func (t *Tokenizer) scanNumber() {
	count := 0
	const maxNumberLength = 100 // Reasonable limit for numbers

	// Simple number scanning - can be enhanced for hex, octal, float, etc.
	for {
		r := t.peek()
		if !unicode.IsDigit(r) && r != '.' && r != 'e' && r != 'E' && r != '+' && r != '-' &&
			r != 'x' && r != 'X' && r != 'a' && r != 'b' && r != 'c' && r != 'd' && r != 'f' &&
			r != 'A' && r != 'B' && r != 'C' && r != 'D' && r != 'F' && r != 'u' && r != 'U' &&
			r != 'l' && r != 'L' {
			break
		}
		count++
		if count > maxNumberLength {
			t.emitError("number too long - possible infinite loop")
			break
		}
		t.next()
	}
	t.emit(TokenNumber)
}

// scanOperator scans operators and punctuation
func (t *Tokenizer) scanOperator() {
	r := t.input[t.pos-1] // Current character (already consumed)

	switch r {
	case '(':
		t.emit(TokenLeftParen)
	case ')':
		t.emit(TokenRightParen)
	case '{':
		t.emit(TokenLeftBrace)
	case '}':
		t.emit(TokenRightBrace)
	case '[':
		t.emit(TokenLeftBracket)
	case ']':
		t.emit(TokenRightBracket)
	case ';':
		t.emit(TokenSemicolon)
	case ',':
		t.emit(TokenComma)
	case '\\':
		t.emit(TokenBackslash)
	case '?':
		t.emit(TokenQuestion)
	case '~':
		t.emit(TokenTilde)
	case '^':
		t.emit(TokenCaret)
	case '%':
		t.emit(TokenPercent)

	case ':':
		if t.peek() == ':' {
			t.next()
			t.emit(TokenDoubleColon)
		} else {
			t.emit(TokenColon)
		}

	case '.':
		t.emit(TokenDot)

	case '=':
		if t.peek() == '=' {
			t.next()
			t.emit(TokenDoubleEquals)
		} else {
			t.emit(TokenEquals)
		}

	case '!':
		if t.peek() == '=' {
			t.next()
			t.emit(TokenNotEquals)
		} else {
			t.emit(TokenExclamation)
		}

	case '<':
		next := t.peek()
		if next == '=' {
			t.next()
			t.emit(TokenLessEqual)
		} else if next == '<' {
			t.next()
			t.emit(TokenLeftShift)
		} else {
			t.emit(TokenLess)
		}

	case '>':
		next := t.peek()
		if next == '=' {
			t.next()
			t.emit(TokenGreaterEqual)
		} else if next == '>' {
			t.next()
			t.emit(TokenRightShift)
		} else {
			t.emit(TokenGreater)
		}

	case '&':
		if t.peek() == '&' {
			t.next()
			t.emit(TokenDoubleAmp)
		} else {
			t.emit(TokenAmpersand)
		}

	case '|':
		if t.peek() == '|' {
			t.next()
			t.emit(TokenDoublePipe)
		} else {
			t.emit(TokenPipe)
		}

	case '+':
		next := t.peek()
		if next == '+' {
			t.next()
			t.emit(TokenPlusPlus)
		} else if next == '=' {
			t.next()
			t.emit(TokenPlusEquals)
		} else {
			t.emit(TokenPlus)
		}

	case '-':
		next := t.peek()
		if next == '-' {
			t.next()
			t.emit(TokenMinusMinus)
		} else if next == '=' {
			t.next()
			t.emit(TokenMinusEquals)
		} else if next == '>' {
			t.next()
			t.emit(TokenArrow)
		} else {
			t.emit(TokenMinus)
		}

	case '*':
		if t.peek() == '=' {
			t.next()
			t.emit(TokenStarEquals)
		} else {
			t.emit(TokenStar)
		}

	default:
		t.emitError(fmt.Sprintf("unexpected character: %c", r))
	}
}
