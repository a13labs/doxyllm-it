package document_new

import (
	"testing"
)

func TestParseDoxygenComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected DoxygenComment
	}{
		{
			name:  "Simple brief comment",
			input: "/// @brief A simple function",
			expected: DoxygenComment{
				Brief:      "A simple function",
				Params:     map[string]string{},
				TParams:    map[string]string{},
				CustomTags: map[string]string{},
				Groups:     []string{},
			},
		},
		{
			name: "Block comment with multiple tags",
			input: `/** @brief A complex function
			 *  @details This function does complex things
			 *  @param x The first parameter
			 *  @param y The second parameter
			 *  @return The computed result
			 *  @since version 1.0
			 */`,
			expected: DoxygenComment{
				Brief:    "A complex function",
				Detailed: "This function does complex things",
				Params: map[string]string{
					"x": "The first parameter",
					"y": "The second parameter",
				},
				TParams:    map[string]string{},
				Returns:    "The computed result",
				Since:      "version 1.0",
				CustomTags: map[string]string{},
				Groups:     []string{},
			},
		},
		{
			name: "Template function comment",
			input: `/** @brief Template function
			 *  @tparam T The template type
			 *  @tparam N The size parameter
			 *  @param data The input data
			 */`,
			expected: DoxygenComment{
				Brief: "Template function",
				TParams: map[string]string{
					"T": "The template type",
					"N": "The size parameter",
				},
				Params: map[string]string{
					"data": "The input data",
				},
				CustomTags: map[string]string{},
				Groups:     []string{},
			},
		},
		{
			name: "Comment with groups and custom tags",
			input: `/// @brief Group member function
			/// @ingroup math_functions
			/// @ingroup utility
			/// @author John Doe
			/// @deprecated Use newFunction() instead`,
			expected: DoxygenComment{
				Brief: "Group member function",
				Groups: []string{
					"math_functions",
					"utility",
				},
				Deprecated: "Use newFunction() instead",
				CustomTags: map[string]string{
					"author": "John Doe",
				},
				Params:  map[string]string{},
				TParams: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseDoxygenComment(tt.input)

			if result.Brief != tt.expected.Brief {
				t.Errorf("Brief: expected '%s', got '%s'", tt.expected.Brief, result.Brief)
			}

			if result.Detailed != tt.expected.Detailed {
				t.Errorf("Detailed: expected '%s', got '%s'", tt.expected.Detailed, result.Detailed)
			}

			if result.Returns != tt.expected.Returns {
				t.Errorf("Returns: expected '%s', got '%s'", tt.expected.Returns, result.Returns)
			}

			if result.Since != tt.expected.Since {
				t.Errorf("Since: expected '%s', got '%s'", tt.expected.Since, result.Since)
			}

			if result.Deprecated != tt.expected.Deprecated {
				t.Errorf("Deprecated: expected '%s', got '%s'", tt.expected.Deprecated, result.Deprecated)
			}

			// Check params
			if len(result.Params) != len(tt.expected.Params) {
				t.Errorf("Params count: expected %d, got %d", len(tt.expected.Params), len(result.Params))
			}
			for key, expectedValue := range tt.expected.Params {
				if actualValue, exists := result.Params[key]; !exists {
					t.Errorf("Param '%s' not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("Param '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
				}
			}

			// Check tparams
			if len(result.TParams) != len(tt.expected.TParams) {
				t.Errorf("TParams count: expected %d, got %d", len(tt.expected.TParams), len(result.TParams))
			}
			for key, expectedValue := range tt.expected.TParams {
				if actualValue, exists := result.TParams[key]; !exists {
					t.Errorf("TParam '%s' not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("TParam '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
				}
			}

			// Check groups
			if len(result.Groups) != len(tt.expected.Groups) {
				t.Errorf("Groups count: expected %d, got %d", len(tt.expected.Groups), len(result.Groups))
			}
			for i, expectedGroup := range tt.expected.Groups {
				if i >= len(result.Groups) || result.Groups[i] != expectedGroup {
					t.Errorf("Group %d: expected '%s', got '%s'", i, expectedGroup,
						func() string {
							if i < len(result.Groups) {
								return result.Groups[i]
							}
							return "missing"
						}())
				}
			}

			// Check custom tags
			if len(result.CustomTags) != len(tt.expected.CustomTags) {
				t.Errorf("CustomTags count: expected %d, got %d", len(tt.expected.CustomTags), len(result.CustomTags))
			}
			for key, expectedValue := range tt.expected.CustomTags {
				if actualValue, exists := result.CustomTags[key]; !exists {
					t.Errorf("CustomTag '%s' not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("CustomTag '%s': expected '%s', got '%s'", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestCleanCommentContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Single line comment",
			input:    "/// This is a comment",
			expected: "This is a comment",
		},
		{
			name:     "Block comment",
			input:    "/** This is a block comment */",
			expected: "This is a block comment",
		},
		{
			name: "Multi-line block comment",
			input: `/** This is a
			 * multi-line
			 * comment */`,
			expected: "This is a\nmulti-line\ncomment",
		},
		{
			name: "Multiple single line comments",
			input: `/// First line
			/// Second line
			/// Third line`,
			expected: "First line\n Second line\n Third line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanCommentContent(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestDoxygenCommentMethods(t *testing.T) {
	// Test IsEmpty
	emptyComment := NewDoxygenComment()
	if !emptyComment.IsEmpty() {
		t.Error("New comment should be empty")
	}

	// Test HasDoxygenContent
	doxyComment := ParseDoxygenComment("/** @brief A function */")
	if !doxyComment.HasDoxygenContent() {
		t.Error("Parsed comment should have Doxygen content")
	}

	regularComment := ParseDoxygenComment("// Just a regular comment")
	// Debug: check what was parsed
	if regularComment.Brief != "" {
		t.Logf("Regular comment unexpectedly got brief: '%s'", regularComment.Brief)
	}
	if regularComment.HasDoxygenContent() {
		t.Error("Regular comment should not have Doxygen content")
	}

	// Test with populated comment
	comment := NewDoxygenComment()
	comment.Brief = "Test brief"
	if comment.IsEmpty() {
		t.Error("Comment with brief should not be empty")
	}
}
