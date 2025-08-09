package llm

import (
	"testing"
)

func TestOllamaProvider_CleanDoxygenResponse(t *testing.T) {
	provider := &OllamaProvider{}

	tests := []struct {
		name     string
		response string
		expected string
	}{
		{
			name: "Complete Doxygen comment with @brief",
			response: `/**
 * @brief "This is a template alias named 'Dictionary' that represents an associative container in C++."
 */`,
			expected: "This is a template alias named 'Dictionary' that represents an associative container in C++.",
		},
		{
			name: "Complete comment with @brief and @tparam",
			response: `/**
 * @brief Template alias for std::map.
 * @tparam Key The key type
 * @tparam Value The value type
 */`,
			expected: "Template alias for std::map.",
		},
		{
			name:     "Simple description without Doxygen tags",
			response: "Simple description text.",
			expected: "Simple description text.",
		},
		{
			name:     "Description with quotes",
			response: `@brief "Template alias for std::map."`,
			expected: "Template alias for std::map.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.cleanResponse(tt.response)
			if result != tt.expected {
				t.Errorf("cleanResponse() = %q, want %q", result, tt.expected)
			}
		})
	}
}
