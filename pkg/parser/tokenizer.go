// Package parser - improved streaming tokenizer implementation for C++ header files
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

// Streaming Tokenizer - provides tokens on-demand following single responsibility principle
type Tokenizer struct {
	input     string
	pos       int // current position in input
	line      int // current line number
	column    int // current column number
	width     int // width of last rune read
	
	// Streaming tokenizer improvements
	lookahead [3]Token // Small lookahead buffer for peek operations
	lookPos   int      // Number of tokens in lookahead buffer
	
	// Error handling
	errors []Token
}

// NewTokenizer creates a new streaming tokenizer
func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{
		input:  input,
		line:   1,
		column: 1,
		errors: make([]Token, 0),
	}
}

// NextToken returns the next token and advances the position
func (t *Tokenizer) NextToken() Token {
	// If we have tokens in lookahead buffer, return the first one
	if t.lookPos > 0 {
		token := t.lookahead[0]
		// Shift lookahead buffer left
		for i := 0; i < t.lookPos-1; i++ {
			t.lookahead[i] = t.lookahead[i+1]
		}
		t.lookPos--
		return token
	}
	
	// Generate next token
	return t.scanToken()
}

// PeekToken returns the token at the specified offset without advancing
// offset 0 = current token, 1 = next token, etc.
func (t *Tokenizer) PeekToken(offset int) Token {
	if offset < 0 {
		return Token{Type: TokenError, Value: "negative peek offset"}
	}
	
	// Fill lookahead buffer as needed
	for t.lookPos <= offset {
		if t.lookPos >= len(t.lookahead) {
			return Token{Type: TokenError, Value: "peek offset too large"}
		}
		t.lookahead[t.lookPos] = t.scanToken()
		t.lookPos++
	}
	
	return t.lookahead[offset]
}

// Tokenize processes the input and returns all tokens (for backward compatibility)
func (t *Tokenizer) Tokenize() []Token {
	var tokens []Token
	for {
		token := t.NextToken()
		tokens = append(tokens, token)
		if token.Type == TokenEOF {
			break
		}
	}
	return tokens
}

// HasErrors returns true if there were tokenization errors
func (t *Tokenizer) HasErrors() bool {
	return len(t.errors) > 0
}

// GetErrors returns all tokenization errors
func (t *Tokenizer) GetErrors() []Token {
	return t.errors
}

// scanToken scans and returns the next token from input
func (t *Tokenizer) scanToken() Token {
	start := t.pos
	startLine := t.line
	startColumn := t.column
	
	if t.pos >= len(t.input) {
		return Token{Type: TokenEOF, Line: t.line, Column: t.column, Offset: t.pos}
	}
	
	r := t.next()
	
	switch {
	case r == 0:
		return Token{Type: TokenEOF, Line: startLine, Column: startColumn, Offset: start}
		
	case unicode.IsSpace(r):
		if r == '\n' {
			return Token{
				Type:   TokenNewline,
				Value:  "\n",
				Line:   startLine,
				Column: startColumn,
				Offset: start,
			}
		}
		return t.scanWhitespaceToken(start, startLine, startColumn)
		
	case r == '/':
		return t.scanCommentToken(start, startLine, startColumn)
		
	case r == '#':
		if t.peek() == '#' {
			t.next()
			return Token{Type: TokenHashHash, Value: "##", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenHash, Value: "#", Line: startLine, Column: startColumn, Offset: start}
		
	case unicode.IsLetter(r) || r == '_':
		return t.scanIdentifierToken(start, startLine, startColumn)
		
	case unicode.IsDigit(r):
		return t.scanNumberToken(start, startLine, startColumn)
		
	case r == '"':
		return t.scanStringToken(start, startLine, startColumn)
		
	case r == '\'':
		return t.scanCharToken(start, startLine, startColumn)
		
	// Operators and punctuation
	case r == '(':
		return Token{Type: TokenLeftParen, Value: "(", Line: startLine, Column: startColumn, Offset: start}
	case r == ')':
		return Token{Type: TokenRightParen, Value: ")", Line: startLine, Column: startColumn, Offset: start}
	case r == '{':
		return Token{Type: TokenLeftBrace, Value: "{", Line: startLine, Column: startColumn, Offset: start}
	case r == '}':
		return Token{Type: TokenRightBrace, Value: "}", Line: startLine, Column: startColumn, Offset: start}
	case r == '[':
		return Token{Type: TokenLeftBracket, Value: "[", Line: startLine, Column: startColumn, Offset: start}
	case r == ']':
		return Token{Type: TokenRightBracket, Value: "]", Line: startLine, Column: startColumn, Offset: start}
	case r == ';':
		return Token{Type: TokenSemicolon, Value: ";", Line: startLine, Column: startColumn, Offset: start}
	case r == ',':
		return Token{Type: TokenComma, Value: ",", Line: startLine, Column: startColumn, Offset: start}
	case r == '\\':
		return Token{Type: TokenBackslash, Value: "\\", Line: startLine, Column: startColumn, Offset: start}
	case r == '.':
		return Token{Type: TokenDot, Value: ".", Line: startLine, Column: startColumn, Offset: start}
	case r == ':':
		if t.peek() == ':' {
			t.next()
			return Token{Type: TokenDoubleColon, Value: "::", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenColon, Value: ":", Line: startLine, Column: startColumn, Offset: start}
	case r == '=':
		if t.peek() == '=' {
			t.next()
			return Token{Type: TokenDoubleEquals, Value: "==", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenEquals, Value: "=", Line: startLine, Column: startColumn, Offset: start}
	case r == '!':
		if t.peek() == '=' {
			t.next()
			return Token{Type: TokenNotEquals, Value: "!=", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenExclamation, Value: "!", Line: startLine, Column: startColumn, Offset: start}
	case r == '<':
		if t.peek() == '=' {
			t.next()
			return Token{Type: TokenLessEqual, Value: "<=", Line: startLine, Column: startColumn, Offset: start}
		} else if t.peek() == '<' {
			t.next()
			return Token{Type: TokenLeftShift, Value: "<<", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenLess, Value: "<", Line: startLine, Column: startColumn, Offset: start}
	case r == '>':
		if t.peek() == '=' {
			t.next()
			return Token{Type: TokenGreaterEqual, Value: ">=", Line: startLine, Column: startColumn, Offset: start}
		} else if t.peek() == '>' {
			t.next()
			return Token{Type: TokenRightShift, Value: ">>", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenGreater, Value: ">", Line: startLine, Column: startColumn, Offset: start}
	case r == '&':
		if t.peek() == '&' {
			t.next()
			return Token{Type: TokenDoubleAmp, Value: "&&", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenAmpersand, Value: "&", Line: startLine, Column: startColumn, Offset: start}
	case r == '|':
		if t.peek() == '|' {
			t.next()
			return Token{Type: TokenDoublePipe, Value: "||", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenPipe, Value: "|", Line: startLine, Column: startColumn, Offset: start}
	case r == '+':
		if t.peek() == '+' {
			t.next()
			return Token{Type: TokenPlusPlus, Value: "++", Line: startLine, Column: startColumn, Offset: start}
		} else if t.peek() == '=' {
			t.next()
			return Token{Type: TokenPlusEquals, Value: "+=", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenPlus, Value: "+", Line: startLine, Column: startColumn, Offset: start}
	case r == '-':
		if t.peek() == '-' {
			t.next()
			return Token{Type: TokenMinusMinus, Value: "--", Line: startLine, Column: startColumn, Offset: start}
		} else if t.peek() == '=' {
			t.next()
			return Token{Type: TokenMinusEquals, Value: "-=", Line: startLine, Column: startColumn, Offset: start}
		} else if t.peek() == '>' {
			t.next()
			return Token{Type: TokenArrow, Value: "->", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenMinus, Value: "-", Line: startLine, Column: startColumn, Offset: start}
	case r == '*':
		if t.peek() == '=' {
			t.next()
			return Token{Type: TokenStarEquals, Value: "*=", Line: startLine, Column: startColumn, Offset: start}
		}
		return Token{Type: TokenStar, Value: "*", Line: startLine, Column: startColumn, Offset: start}
	case r == '%':
		return Token{Type: TokenPercent, Value: "%", Line: startLine, Column: startColumn, Offset: start}
	case r == '^':
		return Token{Type: TokenCaret, Value: "^", Line: startLine, Column: startColumn, Offset: start}
	case r == '~':
		return Token{Type: TokenTilde, Value: "~", Line: startLine, Column: startColumn, Offset: start}
	case r == '?':
		return Token{Type: TokenQuestion, Value: "?", Line: startLine, Column: startColumn, Offset: start}
	default:
		t.errors = append(t.errors, Token{
			Type:   TokenError,
			Value:  fmt.Sprintf("unexpected character: %c", r),
			Line:   startLine,
			Column: startColumn,
			Offset: start,
		})
		return Token{Type: TokenError, Value: string(r), Line: startLine, Column: startColumn, Offset: start}
	}
}

// Helper methods for character reading and lookahead
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

func (t *Tokenizer) peek() rune {
	if t.pos >= len(t.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(t.input[t.pos:])
	return r
}

func (t *Tokenizer) backup() {
	t.pos -= t.width
	if t.width == 1 && t.pos >= 0 && t.pos < len(t.input) && t.input[t.pos] == '\n' {
		t.line--
		// Column calculation would be complex here, so we keep it simple
	} else {
		t.column--
	}
}

// Helper methods for scanning specific token types

func (t *Tokenizer) scanWhitespaceToken(start, startLine, startColumn int) Token {
	for t.pos < len(t.input) && unicode.IsSpace(rune(t.input[t.pos])) && t.input[t.pos] != '\n' {
		t.next()
	}
	return Token{
		Type:   TokenWhitespace,
		Value:  t.input[start:t.pos],
		Line:   startLine,
		Column: startColumn,
		Offset: start,
	}
}

func (t *Tokenizer) scanCommentToken(start, startLine, startColumn int) Token {
	if t.peek() == '/' {
		// Line comment
		t.next() // consume second /
		for t.pos < len(t.input) && t.input[t.pos] != '\n' {
			t.next()
		}
		
		value := t.input[start:t.pos]
		tokenType := TokenLineComment
		
		// Check for Doxygen comment
		if strings.HasPrefix(value, "///") || strings.HasPrefix(value, "//!") {
			tokenType = TokenDoxygenComment
		}
		
		return Token{
			Type:   tokenType,
			Value:  value,
			Line:   startLine,
			Column: startColumn,
			Offset: start,
		}
	} else if t.peek() == '*' {
		// Block comment
		t.next() // consume *
		isDoxygen := false
		
		// Check for /** at start
		if t.peek() == '*' {
			isDoxygen = true
		}
		
		for t.pos < len(t.input)-1 {
			if t.input[t.pos] == '*' && t.input[t.pos+1] == '/' {
				t.next() // consume *
				t.next() // consume /
				break
			}
			t.next()
		}
		
		value := t.input[start:t.pos]
		tokenType := TokenBlockComment
		if isDoxygen {
			tokenType = TokenDoxygenComment
		}
		
		return Token{
			Type:   tokenType,
			Value:  value,
			Line:   startLine,
			Column: startColumn,
			Offset: start,
		}
	} else if t.peek() == '=' {
		// /= operator
		t.next()
		return Token{Type: TokenSlashEquals, Value: "/=", Line: startLine, Column: startColumn, Offset: start}
	}
	
	// Just a single /
	return Token{Type: TokenSlash, Value: "/", Line: startLine, Column: startColumn, Offset: start}
}

func (t *Tokenizer) scanIdentifierToken(start, startLine, startColumn int) Token {
	for t.pos < len(t.input) {
		r := rune(t.input[t.pos])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		t.next()
	}
	
	value := t.input[start:t.pos]
	tokenType := TokenIdentifier
	
	// Check if it's a keyword
	if keywordType, isKeyword := keywords[value]; isKeyword {
		tokenType = keywordType
	}
	
	return Token{
		Type:   tokenType,
		Value:  value,
		Line:   startLine,
		Column: startColumn,
		Offset: start,
	}
}

func (t *Tokenizer) scanNumberToken(start, startLine, startColumn int) Token {
	// Handle integers and floats
	for t.pos < len(t.input) && unicode.IsDigit(rune(t.input[t.pos])) {
		t.next()
	}
	
	// Handle decimal point
	if t.pos < len(t.input) && t.input[t.pos] == '.' {
		t.next()
		for t.pos < len(t.input) && unicode.IsDigit(rune(t.input[t.pos])) {
			t.next()
		}
	}
	
	// Handle scientific notation (e/E)
	if t.pos < len(t.input) && (t.input[t.pos] == 'e' || t.input[t.pos] == 'E') {
		t.next()
		if t.pos < len(t.input) && (t.input[t.pos] == '+' || t.input[t.pos] == '-') {
			t.next()
		}
		for t.pos < len(t.input) && unicode.IsDigit(rune(t.input[t.pos])) {
			t.next()
		}
	}
	
	// Handle suffixes (f, l, u, etc.)
	for t.pos < len(t.input) {
		r := rune(t.input[t.pos])
		if r == 'f' || r == 'F' || r == 'l' || r == 'L' || r == 'u' || r == 'U' {
			t.next()
		} else {
			break
		}
	}
	
	return Token{
		Type:   TokenNumber,
		Value:  t.input[start:t.pos],
		Line:   startLine,
		Column: startColumn,
		Offset: start,
	}
}

func (t *Tokenizer) scanStringToken(start, startLine, startColumn int) Token {
	for t.pos < len(t.input) {
		r := t.next()
		if r == '"' {
			break
		}
		if r == '\\' && t.pos < len(t.input) {
			t.next() // Skip escaped character
		}
	}
	
	return Token{
		Type:   TokenString,
		Value:  t.input[start:t.pos],
		Line:   startLine,
		Column: startColumn,
		Offset: start,
	}
}

func (t *Tokenizer) scanCharToken(start, startLine, startColumn int) Token {
	for t.pos < len(t.input) {
		r := t.next()
		if r == '\'' {
			break
		}
		if r == '\\' && t.pos < len(t.input) {
			t.next() // Skip escaped character
		}
	}
	
	return Token{
		Type:   TokenCharLiteral,
		Value:  t.input[start:t.pos],
		Line:   startLine,
		Column: startColumn,
		Offset: start,
	}
}
