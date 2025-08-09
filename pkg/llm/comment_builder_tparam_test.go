package llm

import (
	"strings"
	"testing"
)

func TestCommentBuilder_ExtractTemplateParametersFromContext(t *testing.T) {
	builder := NewCommentBuilder()

	testCases := []struct {
		name     string
		context  string
		expected []string
	}{
		{
			name:     "Single template parameter",
			context:  "template <typename T> using list = std::vector<T>",
			expected: []string{"T"},
		},
		{
			name:     "Two template parameters",
			context:  "template <typename T, typename U> using pair = std::pair<T, U>",
			expected: []string{"T", "U"},
		},
		{
			name:     "Three template parameters",
			context:  "template <typename K, typename V, typename H> using map = std::unordered_map<K, V, H>",
			expected: []string{"K", "V", "H"},
		},
		{
			name:     "Mixed parameter types",
			context:  "template <class T, int N> using array = std::array<T, N>",
			expected: []string{"T", "N"},
		},
		{
			name:     "No template parameters",
			context:  "using string = std::string",
			expected: []string{},
		},
		{
			name:     "Complex template with spaces",
			context:  "template < typename T , typename U > using dict = std::map<T, U>",
			expected: []string{"T", "U"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := builder.extractTemplateParametersFromContext(tc.context)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d parameters, got %d. Expected: %v, Got: %v",
					len(tc.expected), len(result), tc.expected, result)
				return
			}

			for i, expected := range tc.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Parameter %d: expected %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

func TestCommentBuilder_IsTemplateType(t *testing.T) {
	builder := NewCommentBuilder()

	testCases := []struct {
		entityType string
		expected   bool
	}{
		{"using", true},
		{"template", true},
		{"typedef", true},
		{"class template", true},
		{"function template", true},
		{"function", false},
		{"variable", false},
		{"namespace", false},
	}

	for _, tc := range testCases {
		t.Run(tc.entityType, func(t *testing.T) {
			result := builder.isTemplateType(tc.entityType)
			if result != tc.expected {
				t.Errorf("Expected %v for entity type %s, got %v", tc.expected, tc.entityType, result)
			}
		})
	}
}

func TestCommentBuilder_HasTemplateParameters(t *testing.T) {
	builder := NewCommentBuilder()

	testCases := []struct {
		name     string
		context  string
		expected bool
	}{
		{
			name:     "Has template parameters",
			context:  "template <typename T> using list = std::vector<T>",
			expected: true,
		},
		{
			name:     "No template keyword",
			context:  "using string = std::string",
			expected: false,
		},
		{
			name:     "Template without angle brackets",
			context:  "template void func()",
			expected: false,
		},
		{
			name:     "Angle brackets without template",
			context:  "std::vector<int> vec",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := builder.hasTemplateParameters(tc.context)
			if result != tc.expected {
				t.Errorf("Expected %v for context %s, got %v", tc.expected, tc.context, result)
			}
		})
	}
}

func TestCommentBuilder_BuildStructuredCommentWithTemplateParams(t *testing.T) {
	builder := NewCommentBuilder()

	// Test that template parameters are included in the generated comment
	response := &CommentResponse{
		Description: "Alias for std::pair<T, U>.",
	}

	context := "template <typename T, typename U> using pair = std::pair<T, U>"
	entityName := "pair"
	entityType := "using"

	result := builder.BuildStructuredComment(response, entityName, entityType, nil, context)

	t.Logf("Generated comment:\n%s", result)

	// Check that both template parameters are documented
	if !strings.Contains(result, "@tparam T") {
		t.Errorf("Expected @tparam T in comment, but not found. Comment:\n%s", result)
	}

	if !strings.Contains(result, "@tparam U") {
		t.Errorf("Expected @tparam U in comment, but not found. Comment:\n%s", result)
	}

	// Also test the individual functions
	isTemplate := builder.isTemplateType(entityType)
	hasTemplate := builder.hasTemplateParameters(context)
	tparams := builder.extractTemplateParametersFromContext(context)

	t.Logf("isTemplateType(%s): %v", entityType, isTemplate)
	t.Logf("hasTemplateParameters: %v", hasTemplate)
	t.Logf("Template parameters found: %v", tparams)
}

// Helper function to check if a string contains a specific line
func containsLine(text, line string) bool {
	lines := splitLines(text)
	for _, l := range lines {
		if l == line {
			return true
		}
	}
	return false
}

// Helper function to split text into lines
func splitLines(text string) []string {
	lines := []string{}
	for _, line := range []string{} {
		lines = append(lines, line)
	}
	// Simple implementation for test
	return []string{} // This would need proper implementation
}
