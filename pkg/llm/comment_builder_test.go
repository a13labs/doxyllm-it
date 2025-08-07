package llm

import (
	"strings"
	"testing"
)

func TestCommentBuilder_BuildStructuredComment(t *testing.T) {
	builder := NewCommentBuilder()

	tests := []struct {
		name       string
		response   *CommentResponse
		entityName string
		entityType string
		groupInfo  *GroupInfo
		context    string
		checkFunc  func(t *testing.T, comment string)
	}{
		{
			name: "simple function comment",
			response: &CommentResponse{
				Description: "A simple test function that performs basic operations.",
			},
			entityName: "testFunction",
			entityType: "function",
			context:    "void testFunction(int param);",
			checkFunc: func(t *testing.T, comment string) {
				if !strings.Contains(comment, "@brief A simple test function") {
					t.Errorf("comment should contain brief description")
				}
				if !strings.Contains(comment, "@param param") {
					t.Errorf("comment should contain parameter")
				}
				if strings.Contains(comment, "@return") {
					t.Errorf("void function should not have @return tag")
				}
			},
		},
		{
			name: "class comment with group",
			response: &CommentResponse{
				Description: "A test class.\nThis class provides test functionality for the application.",
			},
			entityName: "TestClass",
			entityType: "class",
			groupInfo:  nil, // GroupInfo is no longer used by CommentBuilder
			context:    "class TestClass {};",
			checkFunc: func(t *testing.T, comment string) {
				if !strings.Contains(comment, "@brief A test class") {
					t.Errorf("comment should contain brief description")
				}
				if !strings.Contains(comment, "This class provides test functionality") {
					t.Errorf("comment should contain detailed description")
				}
				// Note: @ingroup is now handled by post-processor in cmd layer, not here
				if strings.Contains(comment, "@param") {
					t.Errorf("class should not have @param tags")
				}
			},
		},
		{
			name: "function with return value",
			response: &CommentResponse{
				Description: "Returns the result of a calculation.",
			},
			entityName: "calculate",
			entityType: "function",
			context:    "int calculate(int a, int b);",
			checkFunc: func(t *testing.T, comment string) {
				if !strings.Contains(comment, "@param a") {
					t.Errorf("comment should contain parameter a")
				}
				if !strings.Contains(comment, "@param b") {
					t.Errorf("comment should contain parameter b")
				}
				if !strings.Contains(comment, "@return") {
					t.Errorf("non-void function should have @return tag")
				}
			},
		},
		{
			name: "namespace comment",
			response: &CommentResponse{
				Description: "Test namespace for organizing test utilities.",
			},
			entityName: "test",
			entityType: "namespace",
			context:    "namespace test {",
			checkFunc: func(t *testing.T, comment string) {
				if !strings.Contains(comment, "@brief Test namespace") {
					t.Errorf("comment should contain brief description")
				}
				if strings.Contains(comment, "@param") {
					t.Errorf("namespace should not have @param tags")
				}
				if strings.Contains(comment, "@return") {
					t.Errorf("namespace should not have @return tags")
				}
			},
		},
		{
			name: "constructor comment",
			response: &CommentResponse{
				Description: "Creates a new instance of the class.",
			},
			entityName: "TestClass",
			entityType: "constructor",
			context:    "TestClass(int value);",
			checkFunc: func(t *testing.T, comment string) {
				if !strings.Contains(comment, "@param value") {
					t.Errorf("constructor should have parameter")
				}
				if strings.Contains(comment, "@return") {
					t.Errorf("constructor should not have @return tag")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment := builder.BuildStructuredComment(
				tt.response,
				tt.entityName,
				tt.entityType,
				tt.groupInfo,
				tt.context,
			)

			// Basic structure checks
			if !strings.HasPrefix(comment, "/**") {
				t.Errorf("comment should start with /**")
			}
			if !strings.HasSuffix(comment, "*/") {
				t.Errorf("comment should end with */")
			}

			// Run specific test checks
			if tt.checkFunc != nil {
				tt.checkFunc(t, comment)
			}
		})
	}
}

func TestCommentBuilder_ExtractParametersFromContext(t *testing.T) {
	builder := NewCommentBuilder()

	tests := []struct {
		name     string
		context  string
		expected []string
	}{
		{
			name:     "simple function",
			context:  "void testFunction(int param);",
			expected: []string{"param"},
		},
		{
			name:     "multiple parameters",
			context:  "int calculate(int a, int b, float c);",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "parameters with default values",
			context:  "void function(int a, int b = 10, const char* name = nullptr);",
			expected: []string{"a", "b", "name"},
		},
		{
			name:     "no parameters",
			context:  "void function();",
			expected: []string{},
		},
		{
			name:     "void parameter",
			context:  "void function(void);",
			expected: []string{},
		},
		{
			name:     "complex parameter types",
			context:  "void function(const std::string& text, std::vector<int>* data);",
			expected: []string{"text", "data"},
		},
		{
			name:     "reference and pointer parameters",
			context:  "void function(int& ref, int* ptr, const int& constRef);",
			expected: []string{"ref", "ptr", "constRef"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := builder.extractParametersFromContext(tt.context)

			if len(params) != len(tt.expected) {
				t.Errorf("expected %d parameters, got %d: %v", len(tt.expected), len(params), params)
				return
			}

			for i, expected := range tt.expected {
				if params[i] != expected {
					t.Errorf("parameter %d: expected %q, got %q", i, expected, params[i])
				}
			}
		})
	}
}

func TestCommentBuilder_HasReturnValue(t *testing.T) {
	builder := NewCommentBuilder()

	tests := []struct {
		name       string
		context    string
		entityType string
		expected   bool
	}{
		{
			name:       "void function",
			context:    "void function();",
			entityType: "function",
			expected:   false,
		},
		{
			name:       "int function",
			context:    "int calculate();",
			entityType: "function",
			expected:   true,
		},
		{
			name:       "constructor",
			context:    "TestClass();",
			entityType: "constructor",
			expected:   false,
		},
		{
			name:       "destructor",
			context:    "~TestClass();",
			entityType: "destructor",
			expected:   false,
		},
		{
			name:       "string function",
			context:    "std::string getName();",
			entityType: "method",
			expected:   true,
		},
		{
			name:       "void set function",
			context:    "void setName(const std::string& name);",
			entityType: "method",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.hasReturnValue(tt.context, tt.entityType)
			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestCommentBuilder_IsFunctionType(t *testing.T) {
	builder := NewCommentBuilder()

	tests := []struct {
		name       string
		entityType string
		expected   bool
	}{
		{
			name:       "function",
			entityType: "function",
			expected:   true,
		},
		{
			name:       "method",
			entityType: "method",
			expected:   true,
		},
		{
			name:       "constructor",
			entityType: "constructor",
			expected:   true,
		},
		{
			name:       "destructor",
			entityType: "destructor",
			expected:   true,
		},
		{
			name:       "class",
			entityType: "class",
			expected:   false,
		},
		{
			name:       "namespace",
			entityType: "namespace",
			expected:   false,
		},
		{
			name:       "variable",
			entityType: "variable",
			expected:   false,
		},
		{
			name:       "mixed case function",
			entityType: "Function/Method",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.isFunctionType(tt.entityType)
			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestCommentBuilder_IsValidIdentifier(t *testing.T) {
	builder := NewCommentBuilder()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid identifier",
			input:    "validName",
			expected: true,
		},
		{
			name:     "identifier with underscore",
			input:    "valid_name",
			expected: true,
		},
		{
			name:     "identifier with numbers",
			input:    "name123",
			expected: true,
		},
		{
			name:     "starts with underscore",
			input:    "_private",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "starts with number",
			input:    "123name",
			expected: false,
		},
		{
			name:     "contains special characters",
			input:    "name@test",
			expected: false,
		},
		{
			name:     "contains space",
			input:    "name test",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.isValidIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}
