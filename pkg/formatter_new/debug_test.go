package formatter_new

import (
	"testing"

	parser "doxyllm-it/pkg/parser_new"
)

func TestDebugASTStructure(t *testing.T) {
	content := `class TestClass {
public:
    int field;
};`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	// Debug the AST structure
	t.Logf("Root type: %s", tree.Root.Type)
	t.Logf("Root OriginalText length: %d", len(tree.Root.OriginalText))
	t.Logf("Root children count: %d", len(tree.Root.Children))

	for i, child := range tree.Root.Children {
		t.Logf("  Child %d: %s (%s) - OriginalText: %q", i, child.Type, child.Name, child.OriginalText)
		if len(child.Children) > 0 {
			for j, grandchild := range child.Children {
				t.Logf("    Grandchild %d: %s (%s) - OriginalText: %q", j, grandchild.Type, grandchild.Name, grandchild.OriginalText)
			}
		}
	}
}
