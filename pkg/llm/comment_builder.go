package llm

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// CommentBuilder handles the construction of structured Doxygen comments
type CommentBuilder struct{}

// NewCommentBuilder creates a new comment builder instance
func NewCommentBuilder() *CommentBuilder {
	return &CommentBuilder{}
}

// BuildStructuredComment creates a properly structured Doxygen comment
func (cb *CommentBuilder) BuildStructuredComment(response *CommentResponse, entityName, entityType string, groupInfo *GroupInfo, context string) string {
	return cb.BuildStructuredCommentWithStyle(response, entityName, entityType, groupInfo, context, false)
}

// BuildStructuredCommentWithStyle creates a structured Doxygen comment with specified style
func (cb *CommentBuilder) BuildStructuredCommentWithStyle(response *CommentResponse, entityName, entityType string, groupInfo *GroupInfo, context string, isInlineStyle bool) string {
	description := response.Comment

	// For inline comments, create simple inline format
	if isInlineStyle {
		// For simple field documentation, use inline style
		lines := strings.Split(description, "\n")
		brief := ""
		if len(lines) > 0 {
			brief = strings.TrimSpace(lines[0])
		}

		if brief != "" && len(brief) < 100 && !strings.Contains(brief, "\n") {
			// Simple single-line description - use inline format
			return fmt.Sprintf("/**< %s */", brief)
		}
	}

	// Fall back to block comment format
	var comment strings.Builder
	comment.WriteString("/**\n")

	// Continue with existing logic...

	// Add brief description (first sentence or line)
	lines := strings.Split(description, "\n")
	brief := ""
	if len(lines) > 0 {
		brief = strings.TrimSpace(lines[0])
		// If first line is too short, combine with second line
		if len(brief) < 50 && len(lines) > 1 {
			secondLine := strings.TrimSpace(lines[1])
			if secondLine != "" {
				brief += " " + secondLine
			}
		}
	}

	if brief != "" {
		comment.WriteString(fmt.Sprintf(" * @brief %s\n", brief))
	}

	// Add detailed description if there's more content
	detailed := ""
	if len(lines) > 1 {
		detailedLines := lines[1:]
		if brief != "" && len(lines) > 2 && strings.Contains(brief, strings.TrimSpace(lines[1])) {
			detailedLines = lines[2:] // Skip second line if it was included in brief
		}

		var detailedParts []string
		for _, line := range detailedLines {
			line = strings.TrimSpace(line)
			if line != "" {
				detailedParts = append(detailedParts, line)
			}
		}
		detailed = strings.Join(detailedParts, " ")
	}

	if detailed != "" {
		comment.WriteString(" *\n")
		// Wrap detailed description
		cb.writeWrappedText(&comment, detailed, 80)
	}

	// Add template parameters for template entities
	if cb.isTemplateType(entityType) || cb.hasTemplateParameters(context) {
		tparams := cb.extractTemplateParametersFromContext(context)
		for _, tparam := range tparams {
			comment.WriteString(fmt.Sprintf(" * @tparam %s Template parameter\n", tparam))
		}
	}

	// Add function-specific tags only for actual functions/methods
	if cb.isFunctionType(entityType) {
		params := cb.extractParametersFromContext(context)
		for _, param := range params {
			comment.WriteString(fmt.Sprintf(" * @param %s \n", param))
		}

		if cb.hasReturnValue(context, entityType) {
			comment.WriteString(" * @return \n")
		}
	}

	comment.WriteString(" */")
	return comment.String()
}

// writeWrappedText writes text with proper line wrapping
func (cb *CommentBuilder) writeWrappedText(comment *strings.Builder, text string, maxWidth int) {
	words := strings.Fields(text)
	currentLine := " * "

	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxWidth {
			comment.WriteString(currentLine + "\n")
			currentLine = " * " + word
		} else {
			if currentLine != " * " {
				currentLine += " "
			}
			currentLine += word
		}
	}
	if currentLine != " * " {
		comment.WriteString(currentLine + "\n")
	}
}

// isFunctionType determines if an entity type represents a function/method
func (cb *CommentBuilder) isFunctionType(entityType string) bool {
	functionTypes := []string{"function", "method", "constructor", "destructor"}
	lowerType := strings.ToLower(entityType)

	for _, ft := range functionTypes {
		if strings.Contains(lowerType, ft) {
			return true
		}
	}
	return false
}

// extractParametersFromContext extracts parameter names from function context
func (cb *CommentBuilder) extractParametersFromContext(context string) []string {
	var params []string

	// Simple regex to find parameters in function signatures
	funcRegex := regexp.MustCompile(`\([^)]*\)`)
	matches := funcRegex.FindAllString(context, -1)

	for _, match := range matches {
		// Remove parentheses
		paramStr := strings.Trim(match, "()")
		if paramStr == "" || paramStr == "void" {
			continue
		}

		// Split by comma and extract parameter names
		parts := strings.Split(paramStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)

			// Handle default values - split by = and take the left side
			if strings.Contains(part, "=") {
				part = strings.Split(part, "=")[0]
				part = strings.TrimSpace(part)
			}

			// Extract the parameter name (last word)
			words := strings.Fields(part)
			if len(words) > 0 {
				paramName := words[len(words)-1]
				// Remove reference/pointer markers
				paramName = strings.TrimPrefix(paramName, "&")
				paramName = strings.TrimPrefix(paramName, "*")
				paramName = strings.TrimSpace(paramName)

				if paramName != "" && cb.isValidIdentifier(paramName) {
					params = append(params, paramName)
				}
			}
		}
	}

	return params
}

// hasReturnValue checks if a function has a return value
func (cb *CommentBuilder) hasReturnValue(context, entityType string) bool {
	// Constructors and destructors don't have return values
	lowerType := strings.ToLower(entityType)
	if strings.Contains(lowerType, "constructor") || strings.Contains(lowerType, "destructor") {
		return false
	}

	// Check if function returns void explicitly
	if strings.Contains(context, "void ") {
		// Look for void at the beginning of function declarations
		lines := strings.Split(context, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "void ") {
				return false
			}
		}
	}

	return true
}

// isValidIdentifier checks if a string is a valid C++ identifier
func (cb *CommentBuilder) isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Must start with letter or underscore
	first := rune(s[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Rest must be letters, digits, or underscores
	for _, r := range s[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}

// isTemplateType determines if an entity type represents a template
func (cb *CommentBuilder) isTemplateType(entityType string) bool {
	templateTypes := []string{"template", "using", "typedef", "class template", "function template"}
	lowerType := strings.ToLower(entityType)

	for _, tt := range templateTypes {
		if strings.Contains(lowerType, tt) {
			return true
		}
	}
	return false
}

// hasTemplateParameters checks if the context contains template parameters
func (cb *CommentBuilder) hasTemplateParameters(context string) bool {
	return strings.Contains(context, "template") && strings.Contains(context, "<") && strings.Contains(context, ">")
}

// extractTemplateParametersFromContext extracts template parameter names from context
func (cb *CommentBuilder) extractTemplateParametersFromContext(context string) []string {
	var tparams []string

	// Look for template parameter declarations
	templateRegex := regexp.MustCompile(`template\s*<([^>]+)>`)
	matches := templateRegex.FindAllStringSubmatch(context, -1)

	for _, match := range matches {
		if len(match) > 1 {
			paramStr := match[1]
			// Split by comma to get individual parameters
			parts := strings.Split(paramStr, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)

				// Extract parameter name from declarations like "typename T" or "class U"
				words := strings.Fields(part)
				if len(words) >= 2 {
					// Handle cases like "typename T", "class U", "int N"
					paramName := words[len(words)-1]
					if cb.isValidIdentifier(paramName) {
						tparams = append(tparams, paramName)
					}
				}
			}
		}
	}

	return tparams
}

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
		Comment: "Alias for std::pair<T, U>.",
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
				Comment: "A simple test function that performs basic operations.",
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
				Comment: "A test class.\nThis class provides test functionality for the application.",
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
				Comment: "Returns the result of a calculation.",
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
				Comment: "Test namespace for organizing test utilities.",
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
				Comment: "Creates a new instance of the class.",
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
