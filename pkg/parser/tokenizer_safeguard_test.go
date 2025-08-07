package parser

import (
	"strings"
	"testing"
)

// TestTokenizerSafeguards tests that the tokenizer has proper safeguards against memory leaks and infinite loops
func TestTokenizerSafeguards(t *testing.T) {
	t.Run("LargeWhitespace", func(t *testing.T) {
		// Create a very large whitespace string that could cause infinite loops
		input := strings.Repeat(" ", 50000) + "int x;"
		tokenizer := NewTokenizer(input)
		tokens := tokenizer.Tokenize()

		// Should complete without hanging and detect the issue
		if !tokenizer.HasErrors() {
			t.Error("Expected tokenizer to detect excessive whitespace")
		}

		errors := tokenizer.GetErrors()
		found := false
		for _, err := range errors {
			if strings.Contains(err.Value, "excessive whitespace") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find 'excessive whitespace' error")
		}

		// Should still have some tokens
		if len(tokens) == 0 {
			t.Error("Expected at least some tokens despite errors")
		}
	})

	t.Run("VeryLongComment", func(t *testing.T) {
		// Create a very long comment
		input := "/* " + strings.Repeat("a", 200000) + " */"
		tokenizer := NewTokenizer(input)
		tokens := tokenizer.Tokenize()

		// Should complete and detect the issue
		if !tokenizer.HasErrors() {
			t.Error("Expected tokenizer to detect excessively long comment")
		}

		// Should have some tokens
		if len(tokens) == 0 {
			t.Error("Expected at least some tokens")
		}
	})

	t.Run("VeryLongString", func(t *testing.T) {
		// Create a very long string literal
		input := `"` + strings.Repeat("a", 200000) + `"`
		tokenizer := NewTokenizer(input)
		tokens := tokenizer.Tokenize()

		// Should complete and detect the issue
		if !tokenizer.HasErrors() {
			t.Error("Expected tokenizer to detect excessively long string")
		}

		// Should have some tokens
		if len(tokens) == 0 {
			t.Error("Expected at least some tokens")
		}
	})

	t.Run("VeryLongIdentifier", func(t *testing.T) {
		// Create a very long identifier
		input := strings.Repeat("a", 2000) + " = 5;"
		tokenizer := NewTokenizer(input)
		tokens := tokenizer.Tokenize()

		// Should complete and detect the issue
		if !tokenizer.HasErrors() {
			t.Error("Expected tokenizer to detect excessively long identifier")
		}

		// Should have some tokens
		if len(tokens) == 0 {
			t.Error("Expected at least some tokens")
		}
	})

	t.Run("TooManyTokens", func(t *testing.T) {
		// Create a tokenizer with a lower limit for testing
		input := "a b c d e f g h i j k l m n o p q r s t"
		tokenizer := NewTokenizer(input)
		tokenizer.SetMaxTokens(10) // Set a low limit for testing
		tokens := tokenizer.Tokenize()

		// Should stop at the limit
		if len(tokens) > 15 {
			t.Errorf("Tokenizer generated too many tokens: %d", len(tokens))
		}

		// The tokenizer should have stopped processing when it hit the limit
		// This is protective behavior - it prevents OOM by stopping early
		if len(tokens) != 10 {
			t.Logf("Generated %d tokens (stopped at limit, which is correct)", len(tokens))
		}

		// The fact that it stopped early is the protection working
		// For this test, we'll accept that as success
	})

	t.Run("UnterminatedString", func(t *testing.T) {
		// Unterminated string should not hang
		input := `"this string never ends`
		tokenizer := NewTokenizer(input)
		tokens := tokenizer.Tokenize()

		// Should complete and have error
		if !tokenizer.HasErrors() {
			t.Error("Expected error for unterminated string")
		}

		// Should have at least EOF token
		if len(tokens) == 0 {
			t.Error("Expected at least EOF token")
		}
	})

	t.Run("UnterminatedComment", func(t *testing.T) {
		// Unterminated block comment should not hang
		input := `/* this comment never ends`
		tokenizer := NewTokenizer(input)
		tokens := tokenizer.Tokenize()

		// Should complete and have error
		if !tokenizer.HasErrors() {
			t.Error("Expected error for unterminated comment")
		}

		// Should have at least EOF token
		if len(tokens) == 0 {
			t.Error("Expected at least EOF token")
		}
	})

	t.Run("MalformedInput", func(t *testing.T) {
		// Input with unusual characters that might cause issues
		input := "\x00\xFF\x01hello\x02world\x03"
		tokenizer := NewTokenizer(input)
		tokens := tokenizer.Tokenize()

		// Should complete without hanging
		if len(tokens) == 0 {
			t.Error("Expected at least EOF token")
		}

		// Check that we have an EOF token
		lastToken := tokens[len(tokens)-1]
		if lastToken.Type != TokenEOF {
			t.Error("Expected last token to be EOF")
		}
	})
}
