package formatter

import (
	"strings"
	"testing"

	"doxyllm-it/pkg/ast"
)

func TestFormatDoxygenCommentWithTParams(t *testing.T) {
	tests := []struct {
		name     string
		comment  *ast.DoxygenComment
		expected []string // Lines that should be present
	}{
		{
			name: "Template function with both params and tparams",
			comment: &ast.DoxygenComment{
				Raw:   "/**\n * @brief A generic container function.\n */",
				Brief: "A generic container function.",
				TParams: map[string]string{
					"T":         "The type of elements",
					"Container": "The container type",
				},
				Params: map[string]string{
					"container": "The container instance",
					"value":     "The value to process",
				},
				Returns: "true if successful",
			},
			expected: []string{
				"* @brief A generic container function.",
				"* @tparam T The type of elements",
				"* @tparam Container The container type",
				"* @param container The container instance", 
				"* @param value The value to process",
				"* @return true if successful",
			},
		},
		{
			name: "Template alias with only tparams",
			comment: &ast.DoxygenComment{
				Raw:   "/**\n * @brief Alias for std::vector<T>.\n */",
				Brief: "Alias for std::vector<T>.",
				TParams: map[string]string{
					"T": "The type of elements in the list.",
				},
			},
			expected: []string{
				"* @brief Alias for std::vector<T>.",
				"* @tparam T The type of elements in the list.",
			},
		},
		{
			name: "Regular function with only params",
			comment: &ast.DoxygenComment{
				Raw:   "/**\n * @brief Calculates distance.\n */",
				Brief: "Calculates distance.",
				Params: map[string]string{
					"x": "X coordinate",
					"y": "Y coordinate",
				},
			},
			expected: []string{
				"* @brief Calculates distance.",
				"* @param x X coordinate",
				"* @param y Y coordinate",
			},
		},
	}

	formatter := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatDoxygenComment(tt.comment, 0)
			
			// Check that all expected lines are present
			for _, expectedLine := range tt.expected {
				if !strings.Contains(result, expectedLine) {
					t.Errorf("Expected line not found: %q\nActual result:\n%s", expectedLine, result)
				}
			}

			// Verify tparams come before params
			if len(tt.comment.TParams) > 0 && len(tt.comment.Params) > 0 {
				tparamIndex := strings.Index(result, "@tparam")
				paramIndex := strings.Index(result, "@param")
				if tparamIndex == -1 {
					t.Error("@tparam not found in result")
				}
				if paramIndex == -1 {
					t.Error("@param not found in result")
				}
				if tparamIndex > paramIndex {
					t.Error("@tparam should come before @param")
				}
			}
		})
	}
}

func TestFormatDoxygenCommentIsEmpty(t *testing.T) {
	formatter := New()

	// Comment with TParams should not be considered empty
	commentWithTParams := &ast.DoxygenComment{
		Raw:   "/**\n * @brief Test brief\n */",
		Brief: "Test brief",
		TParams: map[string]string{
			"T": "Template param",
		},
	}

	if formatter.shouldUseInlineComment(commentWithTParams) {
		t.Error("Comment with TParams should not be considered inline-suitable")
	}

	// Empty comment should be considered inline-suitable if it has the inline marker
	emptyComment := &ast.DoxygenComment{
		Raw:   "/**< Short brief */",
		Brief: "< Short brief",
	}

	if !formatter.shouldUseInlineComment(emptyComment) {
		t.Error("Simple brief-only comment should be considered inline-suitable")
	}
}
