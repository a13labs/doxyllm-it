package formatter_new

import (
	"testing"

	parser "doxyllm-it/pkg/parser_new"
)

func TestDebugEntityFields(t *testing.T) {
	content := `class TestClass {
public:
    int field;
};`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	// Debug the first child (should be the class)
	if len(tree.Root.Children) > 0 {
		class := tree.Root.Children[0]
		t.Logf("Class entity:")
		t.Logf("  Type: %s", class.Type)
		t.Logf("  Name: %s", class.Name)
		t.Logf("  FullName: %s", class.FullName)
		t.Logf("  Signature: %s", class.Signature)
		t.Logf("  OriginalText: %q", class.OriginalText)
		t.Logf("  SourceRange: %+v", class.SourceRange)
		t.Logf("  HeaderRange: %+v", class.HeaderRange)

		// Check if we can extract text using source range from root
		if class.SourceRange.Start.Offset < len(tree.Root.OriginalText) &&
			class.SourceRange.End.Offset <= len(tree.Root.OriginalText) {
			extractedText := tree.Root.OriginalText[class.SourceRange.Start.Offset:class.SourceRange.End.Offset]
			t.Logf("  Extracted from root: %q", extractedText)
		}

		// Debug children (access specifier and field)
		for i, child := range class.Children {
			t.Logf("  Child %d:", i)
			t.Logf("    Type: %s", child.Type)
			t.Logf("    Name: %s", child.Name)
			t.Logf("    Signature: %s", child.Signature)
			t.Logf("    OriginalText: %q", child.OriginalText)
			t.Logf("    SourceRange: %+v", child.SourceRange)
		}
	}
}
