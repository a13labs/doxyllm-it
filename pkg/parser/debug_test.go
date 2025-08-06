package parser

import (
	"fmt"
	"testing"

	"doxyllm-it/pkg/ast"
)

func TestDebugFunctionParsing(t *testing.T) {
	content := `void globalFunction(int param);
static void staticFunction();
inline void inlineFunction();
virtual void virtualFunction();

class TestClass {
public:
    TestClass();
    ~TestClass();
    void publicMethod() const;
    static void staticMethod();
    virtual void virtualMethod() override;
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	fmt.Printf("Root has %d children:\n", len(tree.Root.Children))
	for i, child := range tree.Root.Children {
		fmt.Printf("  [%d] %s: %s (line %d)\n", i, child.Type, child.Name, child.SourceRange.Start.Line)
	}

	// Find the class
	var class *ast.Entity
	for _, child := range tree.Root.Children {
		if child.Type == ast.EntityClass && child.Name == "TestClass" {
			class = child
			break
		}
	}

	if class != nil {
		fmt.Printf("\nTestClass has %d children:\n", len(class.Children))
		for i, child := range class.Children {
			fmt.Printf("  [%d] %s: %s (access: %s, line %d)\n", i, child.Type, child.Name, child.AccessLevel, child.SourceRange.Start.Line)
			fmt.Printf("      Signature: %s\n", child.Signature)
			fmt.Printf("      IsConst: %v, IsStatic: %v, IsVirtual: %v\n", child.IsConst, child.IsStatic, child.IsVirtual)
		}
	} else {
		fmt.Println("TestClass not found!")
	}
}
