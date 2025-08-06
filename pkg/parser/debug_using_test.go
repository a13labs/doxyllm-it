package parser

import (
	"fmt"
	"testing"
)

func TestDebugUsingAndTypedef(t *testing.T) {
	content := `typedef int MyInt;
using MyString = std::string;
using namespace std;

namespace TestNamespace {
    typedef std::vector<int> IntVector;
    using StringVector = std::vector<std::string>;
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	fmt.Printf("Root has %d children:\n", len(tree.Root.Children))
	for i, child := range tree.Root.Children {
		fmt.Printf("  [%d] %s: %s (line %d)\n", i, child.Type, child.Name, child.SourceRange.Start.Line)
	}
}
