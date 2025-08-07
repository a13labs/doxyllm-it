package formatter

import (
	"strings"
	"testing"

	"doxyllm-it/pkg/ast"
)

// createTestScopeTree creates a sample AST for testing
func createTestScopeTree() *ast.ScopeTree {
	// Create root entity
	root := &ast.Entity{
		Type:     ast.EntityUnknown,
		Name:     "",
		FullName: "",
		Children: []*ast.Entity{},
	}

	// Create namespace entity
	namespace := &ast.Entity{
		Type:      ast.EntityNamespace,
		Name:      "TestNS",
		FullName:  "TestNS",
		Signature: "namespace TestNS",
		Parent:    root,
		Children:  []*ast.Entity{},
		Comment: &ast.DoxygenComment{
			Brief:      "Test namespace for formatter testing",
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
			Raw:        "/** @brief Test namespace for formatter testing */",
		},
	}

	// Create class entity
	class := &ast.Entity{
		Type:      ast.EntityClass,
		Name:      "Calculator",
		FullName:  "TestNS::Calculator",
		Signature: "class Calculator",
		Parent:    namespace,
		Children:  []*ast.Entity{},
		Comment: &ast.DoxygenComment{
			Brief:      "A simple calculator class",
			Detailed:   "This class provides basic arithmetic operations",
			Params:     make(map[string]string),
			CustomTags: make(map[string]string),
			Ingroup:    []string{"math_utilities"},
			Raw:        "/** @brief A simple calculator class\n * This class provides basic arithmetic operations\n * @ingroup math_utilities */",
		},
	}

	// Create method entity
	method := &ast.Entity{
		Type:      ast.EntityMethod,
		Name:      "add",
		FullName:  "TestNS::Calculator::add",
		Signature: "int add(int a, int b)",
		Parent:    class,
		Children:  []*ast.Entity{},
		Comment: &ast.DoxygenComment{
			Brief: "Adds two numbers",
			Params: map[string]string{
				"a": "First number",
				"b": "Second number",
			},
			Returns:    "Sum of a and b",
			CustomTags: make(map[string]string),
			Raw:        "/** @brief Adds two numbers\n * @param a First number\n * @param b Second number\n * @return Sum of a and b */",
		},
	}

	// Create function without documentation
	undocumentedMethod := &ast.Entity{
		Type:      ast.EntityMethod,
		Name:      "subtract",
		FullName:  "TestNS::Calculator::subtract",
		Signature: "int subtract(int a, int b)",
		Parent:    class,
		Children:  []*ast.Entity{},
	}

	// Build hierarchy
	root.Children = append(root.Children, namespace)
	namespace.Children = append(namespace.Children, class)
	class.Children = append(class.Children, method, undocumentedMethod)

	return &ast.ScopeTree{Root: root}
}

func TestNew(t *testing.T) {
	formatter := New()
	if formatter == nil {
		t.Fatal("New() should not return nil")
	}
}

func TestReconstructCode(t *testing.T) {
	formatter := New()
	tree := createTestScopeTree()

	result := formatter.ReconstructCode(tree)

	// Check that the result contains expected elements
	expectedElements := []string{
		"namespace TestNS",
		"class Calculator",
		"int add(int a, int b)",
		"int subtract(int a, int b)",
		"@brief Test namespace",
		"@brief A simple calculator class",
		"@brief Adds two numbers",
		"@param a First number",
		"@param b Second number",
		"@return Sum of a and b",
		"@ingroup math_utilities",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected reconstructed code to contain '%s', but it didn't.\nFull result:\n%s", element, result)
		}
	}

	// Check structure
	if !strings.Contains(result, "namespace TestNS {") {
		t.Error("Expected namespace to have opening brace")
	}

	if !strings.Contains(result, "class Calculator {") {
		t.Error("Expected class to have opening brace")
	}
}

func TestReconstructScope(t *testing.T) {
	formatter := New()
	tree := createTestScopeTree()

	// Find the Calculator class
	var calculator *ast.Entity
	for _, child := range tree.Root.Children {
		if child.Name == "TestNS" {
			for _, classChild := range child.Children {
				if classChild.Name == "Calculator" {
					calculator = classChild
					break
				}
			}
		}
	}

	if calculator == nil {
		t.Fatal("Could not find Calculator class in test tree")
	}

	result := formatter.ReconstructScope(calculator)

	// Check that it contains the class and its methods
	expectedElements := []string{
		"@brief A simple calculator class",
		"class Calculator",
		"int add(int a, int b)",
		"int subtract(int a, int b)",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected scope reconstruction to contain '%s', but it didn't.\nFull result:\n%s", element, result)
		}
	}
}

func TestFormatDoxygenComment(t *testing.T) {
	formatter := New()

	tests := []struct {
		name    string
		comment *ast.DoxygenComment
		depth   int
		want    []string // Elements that should be present
	}{
		{
			name: "simple brief comment",
			comment: &ast.DoxygenComment{
				Brief:      "Simple brief description",
				Params:     make(map[string]string),
				CustomTags: make(map[string]string),
				Raw:        "/** @brief Simple brief description */",
			},
			depth: 0,
			want:  []string{"/**", " * @brief Simple brief description", " */"},
		},
		{
			name: "full comment with params and return",
			comment: &ast.DoxygenComment{
				Brief: "Function with parameters",
				Params: map[string]string{
					"input":  "Input parameter",
					"output": "Output parameter",
				},
				Returns:    "Return value description",
				CustomTags: make(map[string]string),
				Raw:        "/** Full comment */",
			},
			depth: 1,
			want: []string{
				"    /**",
				"    * @brief Function with parameters",
				"    * @param input Input parameter",
				"    * @param output Output parameter",
				"    * @return Return value description",
				"    */",
			},
		},
		{
			name: "deprecated function",
			comment: &ast.DoxygenComment{
				Brief:      "Deprecated function",
				Deprecated: "Use newFunction() instead",
				Params:     make(map[string]string),
				CustomTags: make(map[string]string),
				Raw:        "/** Deprecated */",
			},
			depth: 0,
			want:  []string{"@brief Deprecated function", "@deprecated Use newFunction() instead"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatDoxygenComment(tt.comment, tt.depth)

			for _, wanted := range tt.want {
				if !strings.Contains(result, wanted) {
					t.Errorf("Expected comment to contain '%s', but it didn't.\nFull result:\n%s", wanted, result)
				}
			}
		})
	}
}

func TestGetIndent(t *testing.T) {
	formatter := New()

	tests := []struct {
		depth    int
		expected string
	}{
		{0, ""},
		{1, "    "},
		{2, "        "},
		{3, "            "},
	}

	for _, test := range tests {
		result := formatter.getIndent(test.depth)
		if result != test.expected {
			t.Errorf("Expected indent for depth %d to be '%s', got '%s'", test.depth, test.expected, result)
		}
	}
}

func TestExtractEntityContext(t *testing.T) {
	formatter := New()
	tree := createTestScopeTree()

	// Find the add method
	var addMethod *ast.Entity
	for _, child := range tree.Root.Children {
		if child.Name == "TestNS" {
			for _, classChild := range child.Children {
				if classChild.Name == "Calculator" {
					for _, methodChild := range classChild.Children {
						if methodChild.Name == "add" {
							addMethod = methodChild
							break
						}
					}
				}
			}
		}
	}

	if addMethod == nil {
		t.Fatal("Could not find add method in test tree")
	}

	// Test without parent/siblings
	result := formatter.ExtractEntityContext(addMethod, false, false)
	if !strings.Contains(result, "int add(int a, int b)") {
		t.Error("Expected context to contain the method signature")
	}
	if !strings.Contains(result, "Target entity:") {
		t.Error("Expected context to have target entity section")
	}

	// Test with parent
	resultWithParent := formatter.ExtractEntityContext(addMethod, true, false)
	if !strings.Contains(resultWithParent, "Parent context:") {
		t.Error("Expected context with parent to have parent section")
	}
	if !strings.Contains(resultWithParent, "Calculator") {
		t.Error("Expected parent context to contain class name")
	}

	// Test with siblings
	resultWithSiblings := formatter.ExtractEntityContext(addMethod, false, true)
	if !strings.Contains(resultWithSiblings, "Sibling context:") {
		t.Error("Expected context with siblings to have sibling section")
	}
	if !strings.Contains(resultWithSiblings, "subtract") {
		t.Error("Expected sibling context to contain sibling method")
	}
}

func TestGetEntitySummary(t *testing.T) {
	formatter := New()
	tree := createTestScopeTree()

	// Find the Calculator class
	var calculator *ast.Entity
	for _, child := range tree.Root.Children {
		if child.Name == "TestNS" {
			for _, classChild := range child.Children {
				if classChild.Name == "Calculator" {
					calculator = classChild
					break
				}
			}
		}
	}

	if calculator == nil {
		t.Fatal("Could not find Calculator class in test tree")
	}

	summary := formatter.GetEntitySummary(calculator)

	expectedElements := []string{
		"Type: class",
		"Name: Calculator",
		"Full Name: TestNS::Calculator",
		"Signature: class Calculator",
		"Has Documentation: true",
		"Children: 2",
	}

	for _, element := range expectedElements {
		if !strings.Contains(summary, element) {
			t.Errorf("Expected summary to contain '%s', but it didn't.\nFull summary:\n%s", element, summary)
		}
	}
}

func TestFormatEntitySignature(t *testing.T) {
	formatter := New()

	tests := []struct {
		name     string
		entity   *ast.Entity
		expected string
	}{
		{
			name: "namespace",
			entity: &ast.Entity{
				Type:      ast.EntityNamespace,
				Name:      "TestNS",
				Signature: "namespace TestNS",
			},
			expected: "namespace TestNS { /* ... */ }",
		},
		{
			name: "class",
			entity: &ast.Entity{
				Type:      ast.EntityClass,
				Name:      "MyClass",
				Signature: "class MyClass",
			},
			expected: "class MyClass { /* ... */ };",
		},
		{
			name: "function",
			entity: &ast.Entity{
				Type:      ast.EntityFunction,
				Name:      "func",
				Signature: "int func()",
			},
			expected: "int func()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatEntitySignature(tt.entity)
			if result != tt.expected {
				t.Errorf("Expected signature '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestUpdateEntityComment(t *testing.T) {
	formatter := New()

	entity := &ast.Entity{
		Type:      ast.EntityFunction,
		Name:      "test",
		Signature: "void test()",
	}

	newComment := &ast.DoxygenComment{
		Brief:      "Test function",
		Params:     make(map[string]string),
		CustomTags: make(map[string]string),
		Raw:        "/** @brief Test function */",
	}

	// Before update
	if entity.Comment != nil {
		t.Error("Entity should not have a comment initially")
	}

	// Update comment
	formatter.UpdateEntityComment(entity, newComment)

	// After update
	if entity.Comment == nil {
		t.Fatal("Entity should have a comment after update")
	}

	if entity.Comment.Brief != "Test function" {
		t.Errorf("Expected brief 'Test function', got '%s'", entity.Comment.Brief)
	}
}

// Integration test with nil/empty cases
func TestReconstructCodeWithEmptyTree(t *testing.T) {
	formatter := New()

	// Test with nil tree
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ReconstructCode should handle nil tree gracefully, but panicked: %v", r)
		}
	}()

	// Test with empty root
	emptyTree := &ast.ScopeTree{
		Root: &ast.Entity{
			Type:     ast.EntityUnknown,
			Name:     "",
			Children: []*ast.Entity{},
		},
	}

	result := formatter.ReconstructCode(emptyTree)
	if result != "" {
		t.Errorf("Expected empty result for empty tree, got: '%s'", result)
	}
}

func TestFormatDoxygenCommentWithNilComment(t *testing.T) {
	formatter := New()

	result := formatter.formatDoxygenComment(nil, 0)
	if result != "" {
		t.Errorf("Expected empty result for nil comment, got: '%s'", result)
	}

	// Test with comment that has empty Raw field
	emptyComment := &ast.DoxygenComment{
		Params:     make(map[string]string),
		CustomTags: make(map[string]string),
		Raw:        "",
	}

	result = formatter.formatDoxygenComment(emptyComment, 0)
	if result != "" {
		t.Errorf("Expected empty result for comment with empty Raw field, got: '%s'", result)
	}
}

// Benchmark tests
func BenchmarkReconstructCode(b *testing.B) {
	formatter := New()
	tree := createTestScopeTree()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.ReconstructCode(tree)
	}
}

func BenchmarkFormatDoxygenComment(b *testing.B) {
	formatter := New()
	comment := &ast.DoxygenComment{
		Brief: "Test function with multiple parameters",
		Params: map[string]string{
			"param1": "First parameter description",
			"param2": "Second parameter description",
			"param3": "Third parameter description",
		},
		Returns:    "Return value description",
		CustomTags: make(map[string]string),
		Raw:        "/** Complex comment */",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.formatDoxygenComment(comment, 1)
	}
}

func TestFormatWithClang(t *testing.T) {
	formatter := New()

	testCode := `namespace Test{class MyClass{public:int func(int a,int b);};}`

	result, err := formatter.FormatWithClang(testCode)

	// If clang-format is not available, that's okay - just log it
	if err != nil {
		t.Logf("clang-format not available (this is okay): %v", err)

		// Verify that the error mentions clang-format
		if !strings.Contains(err.Error(), "clang-format") {
			t.Errorf("Expected error to mention clang-format, got: %v", err)
		}
		return
	}

	// If clang-format is available, verify it formatted the code
	if result == testCode {
		t.Error("Expected clang-format to change the formatting")
	}

	// Should still contain the basic structure
	if !strings.Contains(result, "namespace Test") {
		t.Error("Expected formatted result to contain namespace")
	}
	if !strings.Contains(result, "class MyClass") {
		t.Error("Expected formatted result to contain class")
	}
}
