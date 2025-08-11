package formatter_new

import (
	"strings"
	"testing"

	ast "doxyllm-it/pkg/ast_new"
	parser "doxyllm-it/pkg/parser_new"
)

func TestDebugEntityProperties(t *testing.T) {
	content := `namespace test {
    class TestClass {
    public:
        static const int field;
        virtual void method();
    private:
        int privateField;
    };
}`

	// Parse the content to get an AST
	tree, err := parser.ParseContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse content: %v", err)
	}

	// Walk through the tree and show entity properties
	var walkEntity func(*ast.Entity, int)
	walkEntity = func(entity *ast.Entity, depth int) {
		indent := strings.Repeat("  ", depth)
		t.Logf("%sEntity: %s (%s)", indent, entity.Type, entity.Name)
		t.Logf("%s  FullName: %s", indent, entity.FullName)
		t.Logf("%s  Signature: %s", indent, entity.Signature)
		t.Logf("%s  AccessLevel: %s", indent, entity.AccessLevel)
		t.Logf("%s  IsStatic: %t, IsConst: %t, IsVirtual: %t", indent, entity.IsStatic, entity.IsConst, entity.IsVirtual)
		t.Logf("%s  Template: %t, Params: %v", indent, entity.IsTemplate, entity.TemplateParams)

		for _, child := range entity.Children {
			walkEntity(child, depth+1)
		}
	}

	walkEntity(tree.Root, 0)
}
