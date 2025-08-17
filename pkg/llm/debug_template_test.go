package llm

import (
	"testing"
)

func TestDebugDictionaryAlias(t *testing.T) {
	builder := NewCommentBuilder()

	// Test the exact scenario from our file
	context := "template <typename Key, typename Value> using Dictionary =  std::map<Key, Value>;"
	entityType := "using"
	entityName := "Dictionary"

	// Check all the detection methods
	isTemplate := builder.isTemplateType(entityType)
	hasTemplate := builder.hasTemplateParameters(context)
	tparams := builder.extractTemplateParametersFromContext(context)

	t.Logf("Entity: %s (%s)", entityName, entityType)
	t.Logf("Context: %s", context)
	t.Logf("isTemplateType: %v", isTemplate)
	t.Logf("hasTemplateParameters: %v", hasTemplate)
	t.Logf("Template parameters: %v", tparams)

	// Create the response and build comment
	response := &CommentResponse{
		Comment: "Template alias for std::map.",
	}

	comment := builder.BuildStructuredComment(response, entityName, entityType, nil, context)
	t.Logf("Generated comment:\n%s", comment)

	// Verify template parameters are included
	if len(tparams) > 0 && (isTemplate || hasTemplate) {
		if !contains(comment, "@tparam Key") {
			t.Errorf("Expected @tparam Key in comment")
		}
		if !contains(comment, "@tparam Value") {
			t.Errorf("Expected @tparam Value in comment")
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		((len(s) == len(substr) && s == substr) ||
			(len(s) > len(substr) && (s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findInString(s, substr))))
}

func findInString(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
