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
			result := ParseDoxygenComment(tt.comment)

			if result == nil {
				t.Fatal("ParseDoxygenComment returned nil")
			}

			// Check brief
			if result.Brief != tt.expectedBrief {
				t.Errorf("Brief = %q, want %q", result.Brief, tt.expectedBrief)
			}

			// Check Params
			if len(result.Params) != len(tt.expectedParams) {
				t.Errorf("Params length = %d, want %d", len(result.Params), len(tt.expectedParams))
			}
			for key, expectedValue := range tt.expectedParams {
				if actualValue, exists := result.Params[key]; !exists {
					t.Errorf("Params missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("Params[%q] = %q, want %q", key, actualValue, expectedValue)
				}
			}

			// Check TParams
			if len(result.TParams) != len(tt.expectedTParams) {
				t.Errorf("TParams length = %d, want %d", len(result.TParams), len(tt.expectedTParams))
			}
			for key, expectedValue := range tt.expectedTParams {
				if actualValue, exists := result.TParams[key]; !exists {
					t.Errorf("TParams missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("TParams[%q] = %q, want %q", key, actualValue, expectedValue)
				}
			}

			// Verify that tparams are not in params and vice versa
			for key := range tt.expectedTParams {
				if _, exists := result.Params[key]; exists {
					t.Errorf("TParam %q incorrectly found in Params", key)
				}
			}
			for key := range tt.expectedParams {
				if _, exists := result.TParams[key]; exists {
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

	result := ParseDoxygenComment(comment)
	if result == nil {
		t.Fatal("ParseDoxygenComment returned nil")
	}

	// The bug: @tparam should NOT be in Params
	if _, exists := result.Params["T"]; exists {
		t.Error("Regression: @tparam 'T' found in Params map, should be in TParams only")
	}

	// Correct behavior: @tparam should be in TParams
	if value, exists := result.TParams["T"]; !exists {
		t.Error("@tparam 'T' not found in TParams map")
	} else if value != "The type of elements in the list." {
		t.Errorf("TParams['T'] = %q, want %q", value, "The type of elements in the list.")
	}

	// Verify Params is empty for this case
	if len(result.Params) != 0 {
		t.Errorf("Params should be empty, got %d entries: %v", len(result.Params), result.Params)
	}
}
