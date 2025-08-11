package formatter_new

import (
	"testing"

	parser "doxyllm-it/pkg/parser_new"
)

func TestDebugFormatOutput(t *testing.T) {
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

	t.Logf("Original:\n%s", content)
	t.Logf("Formatted:\n%s", result)

	// Debug each step
	class := tree.Root.Children[0]
	t.Logf("Class signature: %q", class.Signature)

	for i, child := range class.Children {
		t.Logf("Child %d: %s - signature: %q", i, child.Type, child.Signature)
	}
}
