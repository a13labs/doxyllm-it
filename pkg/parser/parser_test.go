package parser

import (
	"fmt"
	"strings"
	"testing"

	"doxyllm-it/pkg/ast"
)

func TestBasicNamespaceParsing(t *testing.T) {
	content := `namespace TestNamespace {
    class TestClass {
    public:
        void publicMethod();
    private:
        int privateField;
    };
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should have 1 namespace at root level
	if len(tree.Root.Children) != 1 {
		t.Errorf("Expected 1 root entity, got %d", len(tree.Root.Children))
	}

	ns := tree.Root.Children[0]
	if ns.Type != ast.EntityNamespace || ns.Name != "TestNamespace" {
		t.Errorf("Expected namespace TestNamespace, got %s %s", ns.Type, ns.Name)
	}

	// Namespace should have 1 class
	if len(ns.Children) != 1 {
		t.Errorf("Expected 1 child in namespace, got %d", len(ns.Children))
	}

	class := ns.Children[0]
	if class.Type != ast.EntityClass || class.Name != "TestClass" {
		t.Errorf("Expected class TestClass, got %s %s", class.Type, class.Name)
	}

	// Class should have 2 members
	if len(class.Children) != 2 {
		t.Errorf("Expected 2 children in class, got %d", len(class.Children))
	}

	// Check access levels
	method := class.Children[0]
	if method.AccessLevel != ast.AccessPublic {
		t.Errorf("Expected public method, got %s", method.AccessLevel)
	}

	field := class.Children[1]
	if field.AccessLevel != ast.AccessPrivate {
		t.Errorf("Expected private field, got %s", field.AccessLevel)
	}
}

func TestAccessLevelParsing(t *testing.T) {
	content := `class TestClass {
public:
    void publicMethod();
    int publicField;
private:
    void privateMethod();
    int privateField;
protected:
    void protectedMethod();
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	class := tree.Root.Children[0]
	if len(class.Children) != 5 {
		t.Errorf("Expected 5 children, got %d", len(class.Children))
	}

	// Check access levels in order
	expectedAccess := []ast.AccessLevel{
		ast.AccessPublic,    // publicMethod
		ast.AccessPublic,    // publicField
		ast.AccessPrivate,   // privateMethod
		ast.AccessPrivate,   // privateField
		ast.AccessProtected, // protectedMethod
	}

	for i, child := range class.Children {
		if child.AccessLevel != expectedAccess[i] {
			t.Errorf("Child %d: expected %s, got %s", i, expectedAccess[i], child.AccessLevel)
		}
	}
}

func TestTemplateParsing(t *testing.T) {
	content := `template <typename T>
class TemplateClass {
public:
    T getValue();
};

template <typename T, size_t N>
struct TemplateStruct {
    T data[N];
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(tree.Root.Children) != 2 {
		t.Errorf("Expected 2 root entities, got %d", len(tree.Root.Children))
	}

	// Check template class
	templateClass := tree.Root.Children[0]
	if templateClass.Type != ast.EntityClass || templateClass.Name != "TemplateClass" {
		t.Errorf("Expected template class TemplateClass, got %s %s", templateClass.Type, templateClass.Name)
	}

	// Check template struct
	templateStruct := tree.Root.Children[1]
	if templateStruct.Type != ast.EntityStruct || templateStruct.Name != "TemplateStruct" {
		t.Errorf("Expected template struct TemplateStruct, got %s %s", templateStruct.Type, templateStruct.Name)
	}
}

func TestFunctionParsing(t *testing.T) {
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

	// Should have 4 global functions + 1 class
	if len(tree.Root.Children) != 5 {
		t.Errorf("Expected 5 root entities, got %d", len(tree.Root.Children))
	}

	// Check global functions
	globalFunc := tree.Root.Children[0]
	if globalFunc.Type != ast.EntityFunction || globalFunc.Name != "globalFunction" {
		t.Errorf("Expected global function, got %s %s", globalFunc.Type, globalFunc.Name)
	}

	staticFunc := tree.Root.Children[1]
	if !staticFunc.IsStatic {
		t.Errorf("Expected static function to have IsStatic=true")
	}

	inlineFunc := tree.Root.Children[2]
	if !inlineFunc.IsInline {
		t.Errorf("Expected inline function to have IsInline=true")
	}

	// Check class methods
	class := tree.Root.Children[4] // Last entity should be the class
	if len(class.Children) != 5 {
		t.Errorf("Expected 5 class members, got %d", len(class.Children))
	}

	constructor := class.Children[0]
	if constructor.Type != ast.EntityConstructor {
		t.Errorf("Expected constructor, got %s", constructor.Type)
	}

	destructor := class.Children[1]
	if destructor.Type != ast.EntityDestructor || destructor.Name != "TestClass" {
		t.Errorf("Expected destructor TestClass, got %s %s", destructor.Type, destructor.Name)
	}

	constMethod := class.Children[2]
	if !constMethod.IsConst {
		t.Errorf("Expected const method to have IsConst=true")
	}

	staticMethod := class.Children[3]
	if !staticMethod.IsStatic {
		t.Errorf("Expected static method to have IsStatic=true")
	}

	virtualMethod := class.Children[4]
	if !virtualMethod.IsVirtual {
		t.Errorf("Expected virtual method to have IsVirtual=true")
	}
}

func TestVariableParsing(t *testing.T) {
	content := `int globalVar;
static int staticVar;
const int constVar = 42;
extern int externVar;

class TestClass {
public:
    int publicField;
    static int staticField;
private:
    mutable int mutableField;
    const int constField;
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Check global variables
	globalVar := tree.Root.Children[0]
	if globalVar.Type != ast.EntityVariable || globalVar.Name != "globalVar" {
		t.Errorf("Expected global variable, got %s %s", globalVar.Type, globalVar.Name)
	}

	staticVar := tree.Root.Children[1]
	if !staticVar.IsStatic {
		t.Errorf("Expected static variable to have IsStatic=true")
	}

	constVar := tree.Root.Children[2]
	if !constVar.IsConst {
		t.Errorf("Expected const variable to have IsConst=true")
	}

	// Check class fields
	class := tree.Root.Children[4] // Last entity
	if len(class.Children) != 4 {
		t.Errorf("Expected 4 class fields, got %d", len(class.Children))
	}

	publicField := class.Children[0]
	if publicField.Type != ast.EntityField {
		t.Errorf("Expected field, got %s", publicField.Type)
	}
	if publicField.AccessLevel != ast.AccessPublic {
		t.Errorf("Expected public field, got %s", publicField.AccessLevel)
	}

	staticField := class.Children[1]
	if !staticField.IsStatic {
		t.Errorf("Expected static field to have IsStatic=true")
	}

	constField := class.Children[3]
	if !constField.IsConst {
		t.Errorf("Expected const field to have IsConst=true")
	}
}

func TestUsingAndTypedefParsing(t *testing.T) {
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

	// Should have typedef, using, using namespace, and namespace at root
	if len(tree.Root.Children) < 4 {
		t.Errorf("Expected at least 4 root entities, got %d", len(tree.Root.Children))
	}

	// Check typedef
	typedef := tree.Root.Children[0]
	if typedef.Type != ast.EntityTypedef || typedef.Name != "MyInt" {
		t.Errorf("Expected typedef MyInt, got %s %s", typedef.Type, typedef.Name)
	}

	// Check using
	using := tree.Root.Children[1]
	if using.Type != ast.EntityUsing || using.Name != "MyString" {
		t.Errorf("Expected using MyString, got %s %s", using.Type, using.Name)
	}

	// Check using namespace
	usingNamespace := tree.Root.Children[2]
	if usingNamespace.Type != ast.EntityUsing || usingNamespace.Name != "std" {
		t.Errorf("Expected using namespace std, got %s %s", usingNamespace.Type, usingNamespace.Name)
	}

	// Check namespace and its contents
	ns := tree.Root.Children[3]
	if ns.Type != ast.EntityNamespace {
		t.Errorf("Expected namespace, got %s", ns.Type)
	}

	if len(ns.Children) != 2 {
		t.Errorf("Expected 2 children in namespace, got %d", len(ns.Children))
	}
}

func TestEnumParsing(t *testing.T) {
	content := `enum Color {
    RED,
    GREEN,
    BLUE
};

enum class Status : int {
    PENDING,
    COMPLETE
};

namespace TestNamespace {
    enum Priority {
        LOW,
        HIGH
    };
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Check global enums
	colorEnum := tree.Root.Children[0]
	if colorEnum.Type != ast.EntityEnum || colorEnum.Name != "Color" {
		t.Errorf("Expected enum Color, got %s %s", colorEnum.Type, colorEnum.Name)
	}

	statusEnum := tree.Root.Children[1]
	if statusEnum.Type != ast.EntityEnum || statusEnum.Name != "Status" {
		t.Errorf("Expected enum Status, got %s %s", statusEnum.Type, statusEnum.Name)
	}

	// Check namespaced enum
	ns := tree.Root.Children[2]
	if len(ns.Children) != 1 {
		t.Errorf("Expected 1 child in namespace, got %d", len(ns.Children))
	}

	priorityEnum := ns.Children[0]
	if priorityEnum.Type != ast.EntityEnum || priorityEnum.Name != "Priority" {
		t.Errorf("Expected enum Priority, got %s %s", priorityEnum.Type, priorityEnum.Name)
	}
}

func TestComplexNesting(t *testing.T) {
	content := `namespace Outer {
    namespace Inner {
        class NestedClass {
        public:
            void method();
            int field;
        };
        
        void nestedFunction();
    }
    
    class OuterClass {
    private:
        Inner::NestedClass member;
    };
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Check outer namespace
	outer := tree.Root.Children[0]
	if outer.Type != ast.EntityNamespace || outer.Name != "Outer" {
		t.Errorf("Expected namespace Outer, got %s %s", outer.Type, outer.Name)
	}

	if len(outer.Children) != 2 {
		t.Errorf("Expected 2 children in Outer namespace, got %d", len(outer.Children))
	}

	// Check inner namespace
	inner := outer.Children[0]
	if inner.Type != ast.EntityNamespace || inner.Name != "Inner" {
		t.Errorf("Expected namespace Inner, got %s %s", inner.Type, inner.Name)
	}

	if len(inner.Children) != 2 {
		t.Errorf("Expected 2 children in Inner namespace, got %d", len(inner.Children))
	}

	// Check nested class
	nestedClass := inner.Children[0]
	if nestedClass.Type != ast.EntityClass || nestedClass.Name != "NestedClass" {
		t.Errorf("Expected class NestedClass, got %s %s", nestedClass.Type, nestedClass.Name)
	}

	// Check nested class members
	if len(nestedClass.Children) != 2 {
		t.Errorf("Expected 2 members in NestedClass, got %d", len(nestedClass.Children))
	}

	// Check full names
	method := nestedClass.Children[0]
	expectedFullName := "Outer::Inner::NestedClass::method"
	if method.FullName != expectedFullName {
		t.Errorf("Expected full name %s, got %s", expectedFullName, method.FullName)
	}
}

func TestStructParsing(t *testing.T) {
	content := `struct SimpleStruct {
    int x;
    int y;
};

struct ComplexStruct : public SimpleStruct {
public:
    void method();
private:
    int privateData;
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Check simple struct
	simpleStruct := tree.Root.Children[0]
	if simpleStruct.Type != ast.EntityStruct || simpleStruct.Name != "SimpleStruct" {
		t.Errorf("Expected struct SimpleStruct, got %s %s", simpleStruct.Type, simpleStruct.Name)
	}

	// Struct members should default to public
	if len(simpleStruct.Children) != 2 {
		t.Errorf("Expected 2 members in SimpleStruct, got %d", len(simpleStruct.Children))
	}

	x := simpleStruct.Children[0]
	if x.AccessLevel != ast.AccessPublic {
		t.Errorf("Expected public access for struct member, got %s", x.AccessLevel)
	}

	// Check complex struct with explicit access specifiers
	complexStruct := tree.Root.Children[1]
	if len(complexStruct.Children) != 2 {
		t.Errorf("Expected 2 members in ComplexStruct, got %d", len(complexStruct.Children))
	}

	method := complexStruct.Children[0]
	if method.AccessLevel != ast.AccessPublic {
		t.Errorf("Expected public method, got %s", method.AccessLevel)
	}

	privateData := complexStruct.Children[1]
	if privateData.AccessLevel != ast.AccessPrivate {
		t.Errorf("Expected private field, got %s", privateData.AccessLevel)
	}
}

func TestFullNameGeneration(t *testing.T) {
	content := `namespace A {
    namespace B {
        class C {
        public:
            void method();
            static int field;
        };
    }
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Navigate to the deeply nested method
	a := tree.Root.Children[0]
	b := a.Children[0]
	c := b.Children[0]
	method := c.Children[0]

	expectedFullName := "A::B::C::method"
	if method.FullName != expectedFullName {
		t.Errorf("Expected full name %s, got %s", expectedFullName, method.FullName)
	}

	field := c.Children[1]
	expectedFieldFullName := "A::B::C::field"
	if field.FullName != expectedFieldFullName {
		t.Errorf("Expected full name %s, got %s", expectedFieldFullName, field.FullName)
	}
}

func TestConditionalCompilation(t *testing.T) {
	content := `#ifdef FEATURE_A
void featureAFunction();
#endif

#ifndef FEATURE_B
void defaultFunction();
#endif

namespace Test {
#if defined(DEBUG)
    void debugFunction();
#else
    void releaseFunction();
#endif
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// The parser should see all functions regardless of preprocessor directives
	// (since it doesn't evaluate preprocessor conditions)
	if len(tree.Root.Children) < 2 {
		t.Errorf("Expected at least 2 root entities, got %d", len(tree.Root.Children))
	}

	// Check that namespace and its functions are parsed
	var ns *ast.Entity
	for _, child := range tree.Root.Children {
		if child.Type == ast.EntityNamespace && child.Name == "Test" {
			ns = child
			break
		}
	}

	if ns == nil {
		t.Errorf("Expected to find namespace Test")
	}
}

func TestRegexMatching(t *testing.T) {
	testCases := []string{
		"TestClass();",
		"~TestClass();",
		"void publicMethod() const;",
		"static void staticMethod();",
		"virtual void virtualMethod() override;",
		"TCB_SPAN_CONSTEXPR11 span(pointer ptr, size_type count);",
		"constexpr size_type size() const noexcept { return storage_.size; }",
		"TCB_SPAN_NODISCARD constexpr bool empty() const noexcept { return size() == 0; }",
		"TCB_SPAN_ARRAY_CONSTEXPR reverse_iterator rbegin() const noexcept;",
	}

	for _, testCase := range testCases {
		fmt.Printf("Testing: %s\n", testCase)
		isFunction := functionRegex.MatchString(testCase)
		fmt.Printf("  isFunction: %t\n", isFunction)

		if isFunction {
			matches := functionRegex.FindStringSubmatch(testCase)
			fmt.Printf("  Matches: %v\n", matches)
		}
		fmt.Println()
	}
}

// Benchmark tests
func BenchmarkParseSimpleClass(b *testing.B) {
	content := `class TestClass {
public:
    void method1();
    void method2();
    int field1;
    int field2;
private:
    void privateMethod();
    int privateField;
};`

	for i := 0; i < b.N; i++ {
		parser := New()
		_, err := parser.Parse("test.hpp", content)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}

func BenchmarkParseComplexFile(b *testing.B) {
	content := strings.Repeat(`namespace NS%d {
    template<typename T>
    class TemplateClass {
    public:
        T getValue();
        void setValue(const T& value);
    private:
        T data_;
    };
    
    void function%d();
}
`, 10)

	for i := 0; i < b.N; i++ {
		parser := New()
		_, err := parser.Parse("test.hpp", content)
		if err != nil {
			b.Fatalf("Failed to parse: %v", err)
		}
	}
}
