package parser

import (
	"testing"
)

func TestTokenizerBasics(t *testing.T) {
	input := `namespace Test {
    class MyClass {
    public:
        void method();
    };
}`

	tokenizer := NewTokenizer(input)
	tokens := tokenizer.Tokenize()

	// Verify we get the expected tokens
	expectedTokens := []TokenType{
		TokenNamespace, TokenWhitespace, TokenIdentifier, TokenWhitespace, TokenLeftBrace, TokenNewline,
		TokenWhitespace, TokenClass, TokenWhitespace, TokenIdentifier, TokenWhitespace, TokenLeftBrace, TokenNewline,
		TokenWhitespace, TokenPublic, TokenColon, TokenNewline,
		TokenWhitespace, TokenVoid, TokenWhitespace, TokenIdentifier, TokenLeftParen, TokenRightParen, TokenSemicolon, TokenNewline,
		TokenWhitespace, TokenRightBrace, TokenSemicolon, TokenNewline,
		TokenRightBrace, TokenEOF,
	}

	if len(tokens) < len(expectedTokens) {
		t.Errorf("Expected at least %d tokens, got %d", len(expectedTokens), len(tokens))
	}

	// Check some key tokens
	foundNamespace := false
	foundClass := false
	foundPublic := false

	for _, token := range tokens {
		switch token.Type {
		case TokenNamespace:
			foundNamespace = true
		case TokenClass:
			foundClass = true
		case TokenPublic:
			foundPublic = true
		}
	}

	if !foundNamespace {
		t.Error("Expected to find namespace token")
	}
	if !foundClass {
		t.Error("Expected to find class token")
	}
	if !foundPublic {
		t.Error("Expected to find public token")
	}
}

func TestTokenizerComments(t *testing.T) {
	input := `// Line comment
/* Block comment */
/** Doxygen block */
/// Doxygen line`

	tokenizer := NewTokenizer(input)
	tokens := tokenizer.Tokenize()

	commentTypes := []TokenType{}
	for _, token := range tokens {
		if token.Type == TokenLineComment || token.Type == TokenBlockComment || token.Type == TokenDoxygenComment {
			commentTypes = append(commentTypes, token.Type)
		}
	}

	expectedCommentTypes := []TokenType{
		TokenLineComment,
		TokenBlockComment,
		TokenDoxygenComment,
		TokenDoxygenComment,
	}

	if len(commentTypes) != len(expectedCommentTypes) {
		t.Errorf("Expected %d comment tokens, got %d", len(expectedCommentTypes), len(commentTypes))
	}

	for i, expected := range expectedCommentTypes {
		if i < len(commentTypes) && commentTypes[i] != expected {
			t.Errorf("Comment %d: expected %v, got %v", i, expected, commentTypes[i])
		}
	}
}

func TestTokenizerOperators(t *testing.T) {
	input := `:: -> == != <= >= && || ++ -- += -= *= /= << >>`

	tokenizer := NewTokenizer(input)
	tokens := tokenizer.Tokenize()

	expectedOperators := []TokenType{
		TokenDoubleColon, TokenWhitespace,
		TokenArrow, TokenWhitespace,
		TokenDoubleEquals, TokenWhitespace,
		TokenNotEquals, TokenWhitespace,
		TokenLessEqual, TokenWhitespace,
		TokenGreaterEqual, TokenWhitespace,
		TokenDoubleAmp, TokenWhitespace,
		TokenDoublePipe, TokenWhitespace,
		TokenPlusPlus, TokenWhitespace,
		TokenMinusMinus, TokenWhitespace,
		TokenPlusEquals, TokenWhitespace,
		TokenMinusEquals, TokenWhitespace,
		TokenStarEquals, TokenWhitespace,
		TokenSlashEquals, TokenWhitespace,
		TokenLeftShift, TokenWhitespace,
		TokenRightShift,
		TokenEOF,
	}

	if len(tokens) != len(expectedOperators) {
		t.Errorf("Expected %d tokens, got %d", len(expectedOperators), len(tokens))
		for i, token := range tokens {
			t.Logf("Token %d: %s", i, token)
		}
	}

	for i, expected := range expectedOperators {
		if i < len(tokens) && tokens[i].Type != expected {
			t.Errorf("Token %d: expected %v, got %v", i, expected, tokens[i].Type)
		}
	}
}

func TestTokenizerPreprocessor(t *testing.T) {
	input := `#define MAX_SIZE 100
#include <iostream>`

	tokenizer := NewTokenizer(input)
	tokens := tokenizer.Tokenize()

	foundDefine := false
	foundInclude := false

	for _, token := range tokens {
		if token.Type == TokenHash {
			// Look at next non-whitespace token
			continue
		}
		if token.Value == "define" {
			foundDefine = true
		}
		if token.Value == "include" {
			foundInclude = true
		}
	}

	if !foundDefine {
		t.Error("Expected to find 'define' identifier")
	}
	if !foundInclude {
		t.Error("Expected to find 'include' identifier")
	}
}
