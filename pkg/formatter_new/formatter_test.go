package formatter_new

import (
	"strings"
	"testing"

	ast "doxyllm-it/pkg/ast_new"
	parser "doxyllm-it/pkg/parser_new"
)

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()
	if formatter == nil {
		t.Fatal("NewFormatter should not return nil")
	}

	if !formatter.preserveWhitespace {
		t.Error("Default formatter should preserve whitespace")
	}

	if formatter.indentSize != 4 {
		t.Errorf("Expected default indent size 4, got %d", formatter.indentSize)
	}

	if !formatter.useSpaces {
		t.Error("Default formatter should use spaces")
	}
}

func TestNewFormatterWithOptions(t *testing.T) {
	opts := FormatOptions{
		PreserveWhitespace: false,
		IndentSize:         2,
		UseSpaces:          false,
	}

	formatter := NewFormatterWithOptions(opts)
	if formatter == nil {
		t.Fatal("NewFormatterWithOptions should not return nil")
	}

	if formatter.preserveWhitespace {
		t.Error("Formatter should not preserve whitespace with custom options")
	}

	if formatter.indentSize != 2 {
		t.Errorf("Expected indent size 2, got %d", formatter.indentSize)
	}

	if formatter.useSpaces {
		t.Error("Formatter should not use spaces with custom options")
	}
}

func TestFormatToStringEmpty(t *testing.T) {
	formatter := NewFormatter()

	// Test with nil tree
	_, err := formatter.FormatToString(nil)
	if err == nil {
		t.Error("FormatToString should return error for nil tree")
	}

	// Test with empty tree
	tree := ast.NewScopeTree("test.hpp", "")
	result, err := formatter.FormatToString(tree)
	if err != nil {
		t.Fatalf("FormatToString failed with empty tree: %v", err)
	}

	// Should return empty content from root's OriginalText
	if result != "" {
		t.Errorf("Expected empty result for empty tree, got: %q", result)
	}
}

func TestFormatToStringSimple(t *testing.T) {
	content := `class TestClass {
public:
    int field;
};`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	formatter := NewFormatter()
	result, err := formatter.FormatToString(tree)
	if err != nil {
		t.Fatalf("FormatToString failed: %v", err)
	}

	// The result should be the original content (since we preserve original text)
	if result != content {
		t.Errorf("Expected original content, got:\n%q\nExpected:\n%q", result, content)
	}
}

func TestFormatToStringWithComments(t *testing.T) {
	content := `/**
 * @brief A test class
 */
class TestClass {
public:
    int field;
};`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	formatter := NewFormatter()
	result, err := formatter.FormatToString(tree)
	if err != nil {
		t.Fatalf("FormatToString failed: %v", err)
	}

	// The result should preserve the comments and original structure
	if !strings.Contains(result, "* @brief A test class") {
		t.Errorf("Result should contain the comment, got:\n%s", result)
	}

	if !strings.Contains(result, "class TestClass") {
		t.Errorf("Result should contain the class, got:\n%s", result)
	}
}

func TestFormatToStringNamespace(t *testing.T) {
	content := `namespace test {
    class MyClass {
    public:
        void method();
    };
}`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	formatter := NewFormatter()
	result, err := formatter.FormatToString(tree)
	if err != nil {
		t.Fatalf("FormatToString failed: %v", err)
	}

	// Should preserve namespace structure
	if !strings.Contains(result, "namespace test") {
		t.Errorf("Result should contain namespace, got:\n%s", result)
	}

	if !strings.Contains(result, "class MyClass") {
		t.Errorf("Result should contain class, got:\n%s", result)
	}

	if !strings.Contains(result, "void method()") {
		t.Errorf("Result should contain method, got:\n%s", result)
	}
}

func TestFormatToFile(t *testing.T) {
	content := `class TestClass {
public:
    int field;
};`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	formatter := NewFormatter()

	// Create temporary file path
	tempFile := "/tmp/test_formatter_output.hpp"

	err = formatter.FormatToFile(tree, tempFile)
	if err != nil {
		t.Fatalf("FormatToFile failed: %v", err)
	}

	// Read the file back and verify content
	// Note: In a real test, we would read the file and compare
	// For now, just ensure no error occurred
}

func TestFormatEntitySorting(t *testing.T) {
	// Test that entities are formatted in the correct source order
	content := `// First comment
class FirstClass {
};

// Second comment  
class SecondClass {
};`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	formatter := NewFormatter()
	result, err := formatter.FormatToString(tree)
	if err != nil {
		t.Fatalf("FormatToString failed: %v", err)
	}

	// Check that FirstClass appears before SecondClass
	firstPos := strings.Index(result, "FirstClass")
	secondPos := strings.Index(result, "SecondClass")

	if firstPos == -1 || secondPos == -1 {
		t.Errorf("Both classes should be present in result")
	}

	if firstPos > secondPos {
		t.Errorf("FirstClass should appear before SecondClass in the formatted output")
	}
}

func TestFormatterPreserveWhitespace(t *testing.T) {
	content := `   class TestClass {
	public:
		int field;
   };`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	// Test with whitespace preservation
	formatter := NewFormatterWithOptions(FormatOptions{
		PreserveWhitespace: true,
		IndentSize:         4,
		UseSpaces:          true,
	})

	result, err := formatter.FormatToString(tree)
	if err != nil {
		t.Fatalf("FormatToString failed: %v", err)
	}

	// The result should preserve original formatting
	if result != content {
		t.Errorf("Expected preserved whitespace, got:\n%q\nExpected:\n%q", result, content)
	}
}
