package parser

import (
	"testing"
)

// TestParseDoxygenCommentWithInlineMarker tests parsing of inline comments
func TestParseDoxygenCommentWithInlineMarker(t *testing.T) {
	tests := []struct {
		name     string
		comment  string
		expected string
	}{
		{
			name:     "Inline comment with /**< marker",
			comment:  "/**< The x-coordinate of the point. */",
			expected: "< The x-coordinate of the point.",
		},
		{
			name:     "Inline comment with ///< marker",
			comment:  "///< The y-coordinate of the point.",
			expected: "< The y-coordinate of the point.",
		},
		{
			name:     "Inline comment with //!< marker",
			comment:  "//!< The z-coordinate of the point.",
			expected: "< The z-coordinate of the point.",
		},
		{
			name:     "Regular block comment",
			comment:  "/** @brief Regular comment. */",
			expected: "Regular comment.",
		},
		{
			name:     "Multi-line inline comment",
			comment:  "/**< The coordinate value\n * for this point. */",
			expected: "< The coordinate value for this point.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseDoxygenComment(tt.comment)
			if result == nil {
				t.Fatalf("ParseDoxygenComment returned nil")
			}
			
			if result.Brief != tt.expected {
				t.Errorf("ParseDoxygenComment Brief = %q, want %q", result.Brief, tt.expected)
			}
			
			// Verify Raw field is preserved
			if result.Raw != tt.comment {
				t.Errorf("ParseDoxygenComment Raw = %q, want %q", result.Raw, tt.comment)
			}
		})
	}
}

// TestInlineCommentDetection tests detection of inline comment patterns
func TestInlineCommentDetection(t *testing.T) {
	parser := New()
	
	tests := []struct {
		name     string
		comment  string
		expected bool
	}{
		{
			name:     "/**< pattern",
			comment:  "/**< Description */",
			expected: true,
		},
		{
			name:     "///< pattern",
			comment:  "///< Description",
			expected: true,
		},
		{
			name:     "//!< pattern",
			comment:  "//!< Description",
			expected: true,
		},
		{
			name:     "Regular /** pattern",
			comment:  "/** Description */",
			expected: false,
		},
		{
			name:     "Regular /// pattern",
			comment:  "/// Description",
			expected: false,
		},
		{
			name:     "Regular //! pattern",
			comment:  "//! Description",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isTrailingComment(tt.comment)
			if result != tt.expected {
				t.Errorf("isTrailingComment(%q) = %v, want %v", tt.comment, result, tt.expected)
			}
		})
	}
}
