package doxygen

import (
	"doxyllm-it/pkg/ast"
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
		e := &ast.Entity{
			Type:      ast.EntityComment,
			Signature: tt.comment,
		}
		t.Run(tt.name, func(t *testing.T) {
			result, _ := createDocumentationEntry(e)
			if result == nil {
				t.Fatalf("NewDoxygenComment returned nil")
			}

			briefTag := result.CustomTags.Get("brief")
			if len(briefTag) == 0 {
				t.Errorf("NewDoxygenComment Brief = nil, want %q", tt.expected)
			} else if briefTag[0].Value != tt.expected {
				t.Errorf("NewDoxygenComment Brief = %q, want %q", briefTag[0].Value, tt.expected)
			}

			// Verify Raw field is preserved
			if result.Raw != e {
				t.Errorf("NewDoxygenComment Raw = %#v, want %q", result.Raw, tt.comment)
			}
		})
	}
}

// TestParseDoxygenCommentParamVsTParam tests that @param and @tparam are parsed into separate maps
func TestParseDoxygenCommentParamVsTParam(t *testing.T) {
	tests := []struct {
		name            string
		comment         string
		expectedParams  map[string]string
		expectedTParams map[string]string
		expectedBrief   string
	}{
		{
			name: "Template alias with tparam",
			comment: `/**
 * @brief Alias for std::vector<T>.
 * @tparam T The type of elements in the list.
 */`,
			expectedParams:  map[string]string{},
			expectedTParams: map[string]string{"T": "The type of elements in the list."},
			expectedBrief:   "Alias for std::vector<T>.",
		},
		{
			name: "Function with params",
			comment: `/**
 * @brief Calculates the distance between two points.
 * @param x1 The x-coordinate of the first point.
 * @param y1 The y-coordinate of the first point.
 * @param x2 The x-coordinate of the second point.
 * @param y2 The y-coordinate of the second point.
 */`,
			expectedParams: map[string]string{
				"x1": "The x-coordinate of the first point.",
				"y1": "The y-coordinate of the first point.",
				"x2": "The x-coordinate of the second point.",
				"y2": "The y-coordinate of the second point.",
			},
			expectedTParams: map[string]string{},
			expectedBrief:   "Calculates the distance between two points.",
		},
		{
			name: "Template function with both params and tparams",
			comment: `/**
 * @brief Generic container insert function.
 * @tparam T The type of elements in the container.
 * @tparam Container The container type.
 * @param container The container to insert into.
 * @param value The value to insert.
 * @param index The index where to insert.
 */`,
			expectedParams: map[string]string{
				"container": "The container to insert into.",
				"value":     "The value to insert.",
				"index":     "The index where to insert.",
			},
			expectedTParams: map[string]string{
				"T":         "The type of elements in the container.",
				"Container": "The container type.",
			},
			expectedBrief: "Generic container insert function.",
		},
		{
			name: "Class template with tparams only",
			comment: `/**
 * @brief A generic smart pointer class.
 * @tparam T The type of object being managed.
 * @tparam Deleter The deleter type for cleanup.
 */`,
			expectedParams: map[string]string{},
			expectedTParams: map[string]string{
				"T":       "The type of object being managed.",
				"Deleter": "The deleter type for cleanup.",
			},
			expectedBrief: "A generic smart pointer class.",
		},
		{
			name: "Mixed with backslash syntax",
			comment: `/**
 * \brief Template function with backslash syntax.
 * \tparam T The template parameter.
 * \param value The function parameter.
 */`,
			expectedParams:  map[string]string{"value": "The function parameter."},
			expectedTParams: map[string]string{"T": "The template parameter."},
			expectedBrief:   "Template function with backslash syntax.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			e := &ast.Entity{
				Type:      ast.EntityComment,
				Signature: tt.comment,
			}
			result, _ := createDocumentationEntry(e)

			if result == nil {
				t.Fatal("ParseDoxygenComment returned nil")
			}

			// Check brief
			briefTag := result.CustomTags.Get("brief")
			if len(briefTag) > 0 {
				if briefTag[0].Value != tt.expectedBrief {
					t.Errorf("Brief = %q, want %q", briefTag[0].Value, tt.expectedBrief)
				}
			}

			// Check Params
			params := result.CustomTags.Get("param")
			if len(params) != len(tt.expectedParams) {
				t.Errorf("Params length = %d, want %d", len(params), len(tt.expectedParams))
			}
			for key, expectedValue := range tt.expectedParams {
				if actualValue, exists := result.CustomTags.GetParam(key, ""); !exists {
					t.Errorf("Params missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("Params[%q] = %q, want %q", key, actualValue, expectedValue)
				}
			}

			// Check TParams
			tparams := result.CustomTags.Get("tparam")
			if len(tparams) != len(tt.expectedTParams) {
				t.Errorf("TParams length = %d, want %d", len(tparams), len(tt.expectedTParams))
			}
			for key, expectedValue := range tt.expectedTParams {
				if actualValue, exists := result.CustomTags.GetTParam(key); !exists {
					t.Errorf("TParams missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("TParams[%q] = %q, want %q", key, actualValue, expectedValue)
				}
			}

			// Verify that tparams are not in params and vice versa
			for key := range tt.expectedTParams {
				if _, exists := result.CustomTags.GetParam(key, ""); exists {
					t.Errorf("TParam %q incorrectly found in Params", key)
				}
			}
			for key := range tt.expectedParams {
				if _, exists := result.CustomTags.GetTParam(key); exists {
					t.Errorf("Param %q incorrectly found in TParams", key)
				}
			}
		})
	}
}

// TestParseDoxygenCommentRegressionPreventParamTParamMixup tests that the original
// bug where @tparam was stored in Params map doesn't happen anymore
func TestParseDoxygenCommentRegressionPreventParamTParamMixup(t *testing.T) {
	comment := `/**
 * @brief Alias for std::vector<T>.
 * @tparam T The type of elements in the list.
 */`

	e := &ast.Entity{
		Type:      ast.EntityComment,
		Signature: comment,
	}
	result, _ := createDocumentationEntry(e)
	if result == nil {
		t.Fatal("ParseDoxygenComment returned nil")
	}

	// The bug: @tparam should NOT be in Params

	if _, exists := result.CustomTags.GetParam("T", ""); exists {
		t.Error("Regression: @tparam 'T' found in Params map, should be in TParams only")
	}

	// Correct behavior: @tparam should be in TParams
	if value, exists := result.CustomTags.GetTParam("T"); !exists {
		t.Error("@tparam 'T' not found in TParams map")
	} else if value != "The type of elements in the list." {
		t.Errorf("TParams['T'] = %q, want %q", value, "The type of elements in the list.")
	}

	// Verify Params is empty for this case
	if len(result.CustomTags.Get("param")) != 0 {
		t.Errorf("Params should be empty, got %d entries: %v", len(result.CustomTags.Get("param")), result.CustomTags.Get("param"))
	}
}
