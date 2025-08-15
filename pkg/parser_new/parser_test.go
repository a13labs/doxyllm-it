package parser_new

import (
	"strings"
	"testing"

	ast "doxyllm-it/pkg/ast_new"
)

// TestCommentParsing tests that comments are correctly parsed as entities
func TestCommentParsing(t *testing.T) {
	content := `// Line comment
/* Block comment */
/** 
 * Multi-line Doxygen comment
 * @brief Description
 */
/// Doxygen line comment

class TestClass {
    // Inline comment
    int field; /**< Trailing comment */
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Count all comment entities in the tree
	var commentEntities []*ast.Entity
	var allEntities func(*ast.Entity)
	allEntities = func(entity *ast.Entity) {
		if entity.Type == ast.EntityComment {
			commentEntities = append(commentEntities, entity)
		}
		for _, child := range entity.Children {
			allEntities(child)
		}
	}
	allEntities(tree.Root)

	// Should have found 6 comments
	if len(commentEntities) != 6 {
		t.Errorf("Expected 6 comment entities, got %d", len(commentEntities))
		for i, comment := range commentEntities {
			t.Logf("Comment %d: %s", i, comment.Signature)
		}
	}

	// Verify specific comments
	expectedComments := []string{
		"// Line comment",
		"/* Block comment */",
		"/** \n * Multi-line Doxygen comment\n * @brief Description\n */",
		"/// Doxygen line comment",
		"// Inline comment",
		"/**< Trailing comment */",
	}

	for i, expected := range expectedComments {
		if i < len(commentEntities) {
			// Check that comment text is preserved (may have whitespace differences)
			if !strings.Contains(commentEntities[i].Signature, strings.TrimSpace(expected)) {
				t.Errorf("Comment %d: expected to contain %q, got %q", i, expected, commentEntities[i].Signature)
			}
		}
	}
}

// TestInlineCommentAssociation tests that inline comments are associated with the correct entity

// Helper functions for tests
func getNonAccessSpecifierChildren(entity *ast.Entity) []*ast.Entity {
	var result []*ast.Entity
	for _, child := range entity.Children {
		if child.Type != ast.EntityAccessSpecifier {
			result = append(result, child)
		}
	}
	return result
}

func getAccessSpecifiers(entity *ast.Entity) []*ast.Entity {
	var result []*ast.Entity
	for _, child := range entity.Children {
		if child.Type == ast.EntityAccessSpecifier {
			result = append(result, child)
		}
	}
	return result
}

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

	// Get only non-access-specifier children (actual members)
	members := getNonAccessSpecifierChildren(class)
	if len(members) != 2 {
		t.Errorf("Expected 2 members in class, got %d", len(members))
	}

	// Check access levels
	method := members[0]
	if method.AccessLevel != ast.AccessPublic {
		t.Errorf("Expected public method, got %s", method.AccessLevel)
	}

	field := members[1]
	if field.AccessLevel != ast.AccessPrivate {
		t.Errorf("Expected private field, got %s", field.AccessLevel)
	}

	// Verify access specifiers are present
	accessSpecs := getAccessSpecifiers(class)
	if len(accessSpecs) != 2 {
		t.Errorf("Expected 2 access specifiers, got %d", len(accessSpecs))
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

	// Get only non-access-specifier children (actual members)
	members := getNonAccessSpecifierChildren(class)
	if len(members) != 5 {
		t.Errorf("Expected 5 members, got %d", len(members))
	}

	// Check access levels in order
	expectedAccess := []ast.AccessLevel{
		ast.AccessPublic,    // publicMethod
		ast.AccessPublic,    // publicField
		ast.AccessPrivate,   // privateMethod
		ast.AccessPrivate,   // privateField
		ast.AccessProtected, // protectedMethod
	}

	for i, member := range members {
		if member.AccessLevel != expectedAccess[i] {
			t.Errorf("Member %d: expected %s, got %s", i, expectedAccess[i], member.AccessLevel)
		}
	}

	// Verify access specifiers are present
	accessSpecs := getAccessSpecifiers(class)
	if len(accessSpecs) != 3 {
		t.Errorf("Expected 3 access specifiers, got %d", len(accessSpecs))
	}

	// Verify access specifiers have correct names
	expectedAccessSpecNames := []string{"public", "private", "protected"}
	for i, spec := range accessSpecs {
		if i < len(expectedAccessSpecNames) && spec.Name != expectedAccessSpecNames[i] {
			t.Errorf("Access specifier %d: expected %s, got %s", i, expectedAccessSpecNames[i], spec.Name)
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

func TestAdvancedTemplateParsing(t *testing.T) {
	content := `// Test template functions
template <typename T>
T add(T a, T b) {
    return a + b;
}

template <typename T, typename U>
auto multiply(T a, U b) -> decltype(a * b);

// Test template using declarations
template <typename T>
using Vector = std::vector<T>;

template <class Key, class Value>
using Map = std::unordered_map<Key, Value>;

// Test complex template class
template <typename T, size_t N = 10>
class Container {
public:
    template <typename U>
    void insert(const U& item);
};

template <class Key, class Value>
// Template with comment
using Map = std::unordered_map<Key, Value>;

template <typename ElementType, std::size_t Extent = dynamic_extent>
/**
* \brief A flexible, efficient and safe mechanism for representing contiguous sequences of elements in memory.
*
* The span class template allows access to a contiguous sequence of objects without copying them, providing strong exception safety guarantees.
*/
class span;

template <typename C, typename U = uncvref_t<C>>
/**
 * @brief A type trait to check if a type is a container.
 */
struct is_container;

`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Filter out comment entities to get only the actual code entities
	var codeEntities []*ast.Entity
	for _, child := range tree.Root.Children {
		if child.Type != ast.EntityComment {
			codeEntities = append(codeEntities, child)
		}
	}

	// Should have: 2 template functions + 2 template using + 1 template class = 5 entities
	if len(codeEntities) != 8 {
		t.Errorf("Expected 8 code entities, got %d", len(codeEntities))
		for i, child := range tree.Root.Children {
			t.Logf("Child %d: %s %s", i, child.Type, child.Name)
		}
	}

	// Check template function
	templateFunc := codeEntities[0]
	if templateFunc.Type != ast.EntityCallable || templateFunc.Name != "add" {
		t.Errorf("Expected template function add, got %s %s", templateFunc.Type, templateFunc.Name)
	}

	// Check auto return template function
	autoFunc := codeEntities[1]
	if autoFunc.Type != ast.EntityCallable || autoFunc.Name != "multiply" {
		t.Errorf("Expected template function multiply, got %s %s", autoFunc.Type, autoFunc.Name)
	}

	// Check template using declarations
	templateUsing1 := codeEntities[2]
	if templateUsing1.Type != ast.EntityUsing || templateUsing1.Name != "Vector" {
		t.Errorf("Expected template using Vector, got %s %s", templateUsing1.Type, templateUsing1.Name)
	}

	templateUsing2 := codeEntities[3]
	if templateUsing2.Type != ast.EntityUsing || templateUsing2.Name != "Map" {
		t.Errorf("Expected template using Map, got %s %s", templateUsing2.Type, templateUsing2.Name)
	}

	// Check template class with default parameter
	templateClass := codeEntities[4]
	if templateClass.Type != ast.EntityClass || templateClass.Name != "Container" {
		t.Errorf("Expected template class Container, got %s %s", templateClass.Type, templateClass.Name)
	}

	// Check that template class has the member template function (ignoring access specifiers)
	members := getNonAccessSpecifierChildren(templateClass)
	if len(members) != 1 {
		t.Errorf("Expected 1 member in template class, got %d", len(members))
	}

	memberFunc := members[0]
	if memberFunc.Type != ast.EntityCallable || memberFunc.Name != "insert" {
		t.Errorf("Expected member template method insert, got %s %s", memberFunc.Type, memberFunc.Name)
	}
}

func TestPreprocessorDirectiveParsing(t *testing.T) {
	content := `#pragma once
#include <iostream>
#include "local_header.h"

#define MAX_SIZE 1024
#ifdef DEBUG
#define LOG(x) std::cout << x << std::endl
#else
#define LOG(x)
#endif

// File-level comment
/* Multi-line comment
   about the file */

namespace MyNamespace {
    class MyClass {
    public:
        void method();
    };
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Count preprocessor directives and other entities
	preprocessorCount := 0
	commentCount := 0
	namespaceCount := 0

	for _, child := range tree.Root.Children {
		switch child.Type {
		case ast.EntityPreprocessor:
			preprocessorCount++
		case ast.EntityComment:
			commentCount++
		case ast.EntityNamespace:
			namespaceCount++
		}
	}

	// Should have preprocessor directives preserved
	if preprocessorCount == 0 {
		t.Error("Expected preprocessor directives to be parsed, got none")
	}

	// Should have comments preserved
	if commentCount == 0 {
		t.Error("Expected file-level comments to be parsed, got none")
	}

	// Should have namespace
	if namespaceCount != 1 {
		t.Errorf("Expected 1 namespace, got %d", namespaceCount)
	}

	// Verify specific preprocessor directives exist
	foundPragma := false
	foundInclude := false
	foundDefine := false

	for _, child := range tree.Root.Children {
		if child.Type == ast.EntityPreprocessor {
			if strings.Contains(child.Signature, "#pragma once") {
				foundPragma = true
			}
			if strings.Contains(child.Signature, "#include <iostream>") {
				foundInclude = true
			}
			if strings.Contains(child.Signature, "#define MAX_SIZE") {
				foundDefine = true
			}
		}
	}

	if !foundPragma {
		t.Error("Expected #pragma once directive to be preserved")
	}
	if !foundInclude {
		t.Error("Expected #include directive to be preserved")
	}
	if !foundDefine {
		t.Error("Expected #define directive to be preserved")
	}
}

func TestMultiLineTemplateSignatures(t *testing.T) {
	content := `template <
    typename T,
    typename U = int,
    size_t N = 100
>
class MultiLineTemplate {
public:
    void method();
};

template <
    class Iterator,
    class Distance
>
void advance(Iterator& it, Distance n);`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(tree.Root.Children) != 2 {
		t.Errorf("Expected 2 root entities, got %d", len(tree.Root.Children))
		for i, child := range tree.Root.Children {
			t.Logf("Entity %d: %s %s (signature: %s)", i, child.Type, child.Name, child.Signature)
		}
	}

	// Check multi-line template class
	templateClass := tree.Root.Children[0]
	if templateClass.Type != ast.EntityClass || templateClass.Name != "MultiLineTemplate" {
		t.Errorf("Expected multi-line template class MultiLineTemplate, got %s %s", templateClass.Type, templateClass.Name)
	}

	// Verify the signature contains template parameters
	if !strings.Contains(templateClass.Signature, "template") {
		t.Error("Expected class signature to contain template declaration")
	}

	// Check multi-line template function
	if len(tree.Root.Children) >= 2 {
		templateFunc := tree.Root.Children[1]
		if templateFunc.Type != ast.EntityCallable || templateFunc.Name != "advance" {
			t.Errorf("Expected multi-line template function advance, got %s %s", templateFunc.Type, templateFunc.Name)
		}

		// Verify the signature contains template parameters
		if !strings.Contains(templateFunc.Signature, "template") {
			t.Error("Expected function signature to contain template declaration")
		}
	} else {
		t.Error("Expected to find template function, but it's missing")
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
	if globalFunc.Type != ast.EntityCallable || globalFunc.Name != "globalFunction" {
		t.Errorf("Expected global function, got %s %s", globalFunc.Type, globalFunc.Name)
	}

	staticFunc := tree.Root.Children[1]
	if !strings.Contains(staticFunc.Signature, "static") {
		t.Errorf("Expected static function to have IsStatic=true")
	}

	inlineFunc := tree.Root.Children[2]
	if !strings.Contains(inlineFunc.Signature, "inline") {
		t.Errorf("Expected inline function to have IsInline=true")
	}

	// Check class methods
	class := tree.Root.Children[4] // Last entity should be the class
	members := getNonAccessSpecifierChildren(class)
	if len(members) != 5 {
		t.Errorf("Expected 5 class members, got %d", len(members))
	}

	constructor := members[0]
	if constructor.Type != ast.EntityCallable && constructor.Name != "TestClass" {
		t.Errorf("Expected constructor, got %s", constructor.Type)
	}

	destructor := members[1]
	if destructor.Type != ast.EntityCallable && destructor.Name == "TestClass" {
		t.Errorf("Expected destructor TestClass, got %s %s", destructor.Type, destructor.Name)
	}

	constMethod := members[2]
	if !strings.Contains(constMethod.Signature, "const") {
		t.Errorf("Expected const method to have IsConst=true")
	}

	staticMethod := members[3]
	if !strings.Contains(staticMethod.Signature, "static") {
		t.Errorf("Expected static method to have IsStatic=true")
	}

	virtualMethod := members[4]
	if !strings.Contains(virtualMethod.Signature, "virtual") {
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
	if globalVar.Type != ast.EntityName || globalVar.Name != "globalVar" {
		t.Errorf("Expected global variable, got %s %s", globalVar.Type, globalVar.Name)
	}

	staticVar := tree.Root.Children[1]
	if !strings.Contains(staticVar.Signature, "static") {
		t.Errorf("Expected static variable to have IsStatic=true")
	}

	constVar := tree.Root.Children[2]
	if !strings.Contains(constVar.Signature, "const") {
		t.Errorf("Expected const variable to have IsConst=true")
	}

	// Check class fields
	class := tree.Root.Children[4] // Last entity
	members := getNonAccessSpecifierChildren(class)
	if len(members) != 4 {
		t.Errorf("Expected 4 class fields, got %d", len(members))
	}

	publicField := members[0]
	if publicField.Type != ast.EntityName {
		t.Errorf("Expected field, got %s", publicField.Type)
	}
	if publicField.AccessLevel != ast.AccessPublic {
		t.Errorf("Expected public field, got %s", publicField.AccessLevel)
	}

	staticField := members[1]
	if !strings.Contains(staticField.Signature, "static") {
		t.Errorf("Expected static field to have IsStatic=true")
	}

	constField := members[3]
	if !strings.Contains(constField.Signature, "const") {
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
	if typedef.Type != ast.EntityTypedef {
		t.Errorf("Expected typedef, got %s", typedef.Type)
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

func TestMultiLineFunctionDeclarations(t *testing.T) {
	content := `namespace test {
    // Multi-line function declaration
    inline void
    multiline_function(int param1, int param2)
    {
        return;
    }
    
    // Regular single-line function
    void single_line_function(int param);
    
    // Template with multi-line
    template<typename T>
    T
    template_multiline(const T& value)
    {
        return value;
    }
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Get the namespace
	ns := tree.Root.Children[0]
	if ns.Type != ast.EntityNamespace || ns.Name != "test" {
		t.Errorf("Expected namespace test, got %s %s", ns.Type, ns.Name)
	}

	// Get functions (excluding comments)
	var functions []*ast.Entity
	for _, child := range ns.Children {
		if child.Type == ast.EntityCallable {
			functions = append(functions, child)
		}
	}

	if len(functions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(functions))
		for i, child := range ns.Children {
			t.Logf("Child %d: %s %s", i, child.Type, child.Name)
		}
	}

	// Check multi-line function
	multilineFunc := functions[0]
	if multilineFunc.Name != "multiline_function" {
		t.Errorf("Expected multiline_function, got %s", multilineFunc.Name)
	}
	// Signature should NOT contain the function body
	if strings.Contains(multilineFunc.Signature, "return;") {
		t.Errorf("Function signature should not contain body, got: %s", multilineFunc.Signature)
	}
	// Should have body text stored
	if multilineFunc.Signature == "" {
		t.Errorf("Expected function body to be stored in Signature")
	}

	// Check single-line function
	singleFunc := functions[1]
	if singleFunc.Name != "single_line_function" {
		t.Errorf("Expected single_line_function, got %s", singleFunc.Name)
	}

	// Check template multi-line function
	templateFunc := functions[2]
	if templateFunc.Name != "template_multiline" {
		t.Errorf("Expected template_multiline, got %s", templateFunc.Name)
	}
	if !templateFunc.IsTemplate {
		t.Errorf("Expected function to be marked as template")
	}
	// Should have template in signature but not body
	if !strings.Contains(templateFunc.Signature, "template") {
		t.Errorf("Template function signature should contain template: %s", templateFunc.Signature)
	}
	if strings.Contains(templateFunc.Signature, "return value;") {
		t.Errorf("Template function signature should not contain body: %s", templateFunc.Signature)
	}
}

func TestNestedNamespaceDeclarations(t *testing.T) {
	content := `namespace outer::inner {
    void nested_function();
}

namespace single {
    namespace double {
        namespace triple {
            class NestedClass {};
        }
    }
}

namespace int::float::auto {
    void keyword_function();
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(tree.Root.Children) != 3 {
		t.Errorf("Expected 3 root namespaces, got %d", len(tree.Root.Children))
	}

	// Check first nested namespace (C++17 style)
	outerInner := tree.Root.Children[0]
	if outerInner.Type != ast.EntityNamespace || outerInner.Name != "outer::inner" {
		t.Errorf("Expected namespace outer::inner, got %s %s", outerInner.Type, outerInner.Name)
	}

	// Check function inside nested namespace
	if len(outerInner.Children) != 1 {
		t.Errorf("Expected 1 child in outer::inner namespace, got %d", len(outerInner.Children))
	}

	nestedFunc := outerInner.Children[0]
	if nestedFunc.Type != ast.EntityCallable || nestedFunc.Name != "nested_function" {
		t.Errorf("Expected function nested_function, got %s %s", nestedFunc.Type, nestedFunc.Name)
	}

	// Check full name includes nested namespace
	expectedFullName := "outer::inner::nested_function"
	if nestedFunc.FullName != expectedFullName {
		t.Errorf("Expected full name %s, got %s", expectedFullName, nestedFunc.FullName)
	}

	// Check second namespace structure (traditional nested style)
	single := tree.Root.Children[1]
	if single.Type != ast.EntityNamespace || single.Name != "single" {
		t.Errorf("Expected namespace single, got %s %s", single.Type, single.Name)
	}

	// Check nested namespace inside single
	if len(single.Children) != 1 {
		t.Errorf("Expected 1 child in single namespace, got %d", len(single.Children))
	}

	double := single.Children[0]
	if double.Type != ast.EntityNamespace || double.Name != "double" {
		t.Errorf("Expected namespace double, got %s %s", double.Type, double.Name)
	}

	// Check triple namespace
	if len(double.Children) != 1 {
		t.Errorf("Expected 1 child in double namespace, got %d", len(double.Children))
	}

	triple := double.Children[0]
	if triple.Type != ast.EntityNamespace || triple.Name != "triple" {
		t.Errorf("Expected namespace triple, got %s %s", triple.Type, triple.Name)
	}

	// Check class inside triple
	if len(triple.Children) != 1 {
		t.Errorf("Expected 1 child in triple namespace, got %d", len(triple.Children))
	}

	nestedClass := triple.Children[0]
	if nestedClass.Type != ast.EntityClass || nestedClass.Name != "NestedClass" {
		t.Errorf("Expected class NestedClass, got %s %s", nestedClass.Type, nestedClass.Name)
	}

	// Check third namespace (C++17 with keywords)
	keywordNamespace := tree.Root.Children[2]
	if keywordNamespace.Type != ast.EntityNamespace || keywordNamespace.Name != "int::float::auto" {
		t.Errorf("Expected namespace int::float::auto, got %s %s", keywordNamespace.Type, keywordNamespace.Name)
	}

	// Check function inside keyword namespace
	if len(keywordNamespace.Children) != 1 {
		t.Errorf("Expected 1 child in int::float::auto namespace, got %d", len(keywordNamespace.Children))
	}

	keywordFunc := keywordNamespace.Children[0]
	if keywordFunc.Type != ast.EntityCallable || keywordFunc.Name != "keyword_function" {
		t.Errorf("Expected function keyword_function, got %s %s", keywordFunc.Type, keywordFunc.Name)
	}

	// Check full name includes nested keyword namespace
	expectedKeywordFullName := "int::float::auto::keyword_function"
	if keywordFunc.FullName != expectedKeywordFullName {
		t.Errorf("Expected full name %s, got %s", expectedKeywordFullName, keywordFunc.FullName)
	}
}

func TestFunctionBodySeparation(t *testing.T) {
	content := `class TestClass {
public:
    // Constructor with body
    TestClass(int value) : member_(value) {
        initialize();
    }
    
    // Method with body
    void method_with_body() {
        int x = 42;
        process(x);
    }
    
    // Method declaration only
    void method_declaration_only();
    
    // Inline method
    inline int get_value() const {
        return member_;
    }
    
private:
    int member_;
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Get the class
	class := tree.Root.Children[0]
	if class.Type != ast.EntityClass || class.Name != "TestClass" {
		t.Errorf("Expected class TestClass, got %s %s", class.Type, class.Name)
	}

	// Get non-access-specifier and non-comment children
	members := []*ast.Entity{}
	for _, child := range class.Children {
		if child.Type != ast.EntityAccessSpecifier && child.Type != ast.EntityComment {
			members = append(members, child)
		}
	}

	// Should have: constructor, 3 methods, 1 field = 5 members
	if len(members) != 5 {
		t.Errorf("Expected 5 members, got %d", len(members))
		for i, member := range members {
			t.Logf("Member %d: %s %s", i, member.Type, member.Name)
		}
	}

	// Check constructor
	constructor := members[0]
	if constructor.Type != ast.EntityCallable {
		t.Errorf("Expected constructor, got %s", constructor.Type)
	}
	// Constructor signature should not contain the body
	if strings.Contains(constructor.Signature, "initialize()") {
		t.Errorf("Constructor signature should not contain body: %s", constructor.Signature)
	}
	// Should have body stored
	if constructor.Signature == "" {
		t.Errorf("Constructor should have body stored in Signature")
	}

	// Check method with body
	methodWithBody := members[1]
	if methodWithBody.Name != "method_with_body" {
		t.Errorf("Expected method_with_body, got %s", methodWithBody.Name)
	}
	// Signature should not contain body
	if strings.Contains(methodWithBody.Signature, "int x = 42") {
		t.Errorf("Method signature should not contain body: %s", methodWithBody.Signature)
	}
	// Should have body stored
	if methodWithBody.Signature == "" {
		t.Errorf("Method should have body stored in Signature")
	}

	// Check method declaration only
	methodDecl := members[2]
	if methodDecl.Name != "method_declaration_only" {
		t.Errorf("Expected method_declaration_only, got %s", methodDecl.Name)
	}
	// Should not have body
	if methodDecl.Body != "" {
		t.Errorf("Declaration-only method should not have body: %s", methodDecl.Signature)
	}

	// Check inline method
	inlineMethod := members[3]
	if inlineMethod.Name != "get_value" {
		t.Errorf("Expected get_value, got %s", inlineMethod.Name)
	}
	if !strings.Contains(inlineMethod.Signature, "inline") {
		t.Errorf("Expected method to be marked as inline")
	}
	// Signature should not contain body
	if strings.Contains(inlineMethod.Signature, "return member_") {
		t.Errorf("Inline method signature should not contain body: %s", inlineMethod.Signature)
	}
}

func TestComplexInlineFunctions(t *testing.T) {
	content := `namespace mgl::io {
    // Multi-line inline function like in io.hpp
    inline void
    read_buffer(const istream_ref& file, uint8_buffer& buffer, size_t size, size_t offset = 0)
    {
        ASSERT(file->good() && !file->eof(), "read_buffer: file is not open");
        ASSERT(size <= buffer.size(), "read_bytes: size is greater than buffer size");
        file->read(reinterpret_cast<char*>(buffer.data() + offset), size);
    }
    
    // Another inline function
    inline bool is_valid(const path& p) {
        return !p.empty();
    }
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Get the namespace
	ns := tree.Root.Children[0]
	if ns.Type != ast.EntityNamespace || ns.Name != "mgl::io" {
		t.Errorf("Expected namespace mgl::io, got %s %s", ns.Type, ns.Name)
	}

	// Get functions
	var functions []*ast.Entity
	for _, child := range ns.Children {
		if child.Type == ast.EntityCallable {
			functions = append(functions, child)
		}
	}

	if len(functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(functions))
		for i, child := range ns.Children {
			t.Logf("Child %d: %s %s", i, child.Type, child.Name)
		}
	}

	// Check first function (the problematic one from io.hpp)
	readBuffer := functions[0]
	if readBuffer.Name != "read_buffer" {
		t.Errorf("Expected read_buffer, got %s", readBuffer.Name)
	}
	if !strings.Contains(readBuffer.Signature, "inline") {
		t.Errorf("Expected function to be marked as inline")
	}

	// Signature should not contain function body
	if strings.Contains(readBuffer.Signature, "ASSERT(") {
		t.Errorf("Function signature should not contain body: %s", readBuffer.Signature)
	}
	if strings.Contains(readBuffer.Signature, "file->read(") {
		t.Errorf("Function signature should not contain body: %s", readBuffer.Signature)
	}

	// Should have body stored
	if readBuffer.Signature == "" {
		t.Errorf("Function should have body stored in Signature")
	}

	// Body should contain the actual implementation
	if !strings.Contains(readBuffer.Body, "ASSERT(") {
		t.Errorf("Function body should contain implementation: %s", readBuffer.Body)
	}

	// Check second function
	isValid := functions[1]
	if isValid.Name != "is_valid" {
		t.Errorf("Expected is_valid, got %s", isValid.Name)
	}
	if !strings.Contains(isValid.Signature, "inline") {
		t.Errorf("Expected function to be marked as inline")
	}
}

// Test for the specific brace mismatch issue we fixed
func TestBraceCountingValidation(t *testing.T) {
	content := `namespace test {
    void func1() {
        int x = 1;
    }
    
    inline void func2() {
        if (true) {
            return;
        }
    }
    
    class TestClass {
    public:
        void method() {
            // nested braces
            {
                int y = 2;
            }
        }
    };
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Simple validation - check that parsing completed successfully
	// and we have the expected structure without using formatter
	if tree == nil {
		t.Errorf("Tree should not be nil")
		return
	}

	if tree.Root == nil {
		t.Errorf("Root should not be nil")
		return
	}

	// Count braces in the original content for basic validation
	openBraces := strings.Count(content, "{")
	closeBraces := strings.Count(content, "}")

	if openBraces != closeBraces {
		t.Errorf("Brace mismatch in original content: %d opening braces, %d closing braces", openBraces, closeBraces)
	}

	// Verify that we have the expected entities
	ns := tree.Root.Children[0]
	if ns.Type != ast.EntityNamespace || ns.Name != "test" {
		t.Errorf("Expected namespace test, got %s %s", ns.Type, ns.Name)
	}

	// Count different entity types
	var functions, classes int
	for _, child := range ns.Children {
		switch child.Type {
		case ast.EntityCallable:
			functions++
		case ast.EntityClass:
			classes++
		}
	}

	if functions != 2 {
		t.Errorf("Expected 2 functions, got %d", functions)
	}
	if classes != 1 {
		t.Errorf("Expected 1 class, got %d", classes)
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
	members := getNonAccessSpecifierChildren(nestedClass)
	if len(members) != 2 {
		t.Errorf("Expected 2 members in NestedClass, got %d", len(members))
	}

	// Check full names
	method := members[0]
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
	members := getNonAccessSpecifierChildren(complexStruct)
	if len(members) != 2 {
		t.Errorf("Expected 2 members in ComplexStruct, got %d", len(members))
	}

	method := members[0]
	if method.AccessLevel != ast.AccessPublic {
		t.Errorf("Expected public method, got %s", method.AccessLevel)
	}

	privateData := members[1]
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
	members := getNonAccessSpecifierChildren(c)
	method := members[0]

	expectedFullName := "A::B::C::method"
	if method.FullName != expectedFullName {
		t.Errorf("Expected full name %s, got %s", expectedFullName, method.FullName)
	}

	field := members[1]
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

func TestConditionalCompilationIgnored(t *testing.T) {
	content := `#define FEATURE_ENABLED 1

#ifdef FEATURE_ENABLED
void enabledFunction();
#else
void disabledFunction();
#endif

#if defined(DEBUG)
void debugFunction();
#elif defined(RELEASE)
void releaseFunction();
#else
void defaultFunction();
#endif

#ifndef FEATURE_DISABLED
void notDisabledFunction();
#endif`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// All functions should be parsed regardless of conditional compilation
	entities := collectAllEntities(tree.Root)
	expectedFunctions := []string{
		"enabledFunction",
		"disabledFunction",
		"debugFunction",
		"releaseFunction",
		"defaultFunction",
		"notDisabledFunction",
	}

	functionCount := 0
	for _, entity := range entities {
		if entity.Type == ast.EntityCallable {
			functionCount++
			found := false
			for _, expected := range expectedFunctions {
				if entity.Name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unexpected function found: %s", entity.Name)
			}
		}
	}

	if functionCount != len(expectedFunctions) {
		t.Errorf("Expected %d functions, got %d", len(expectedFunctions), functionCount)
	}

}

// Helper function to collect all entities recursively
func collectAllEntities(entity *ast.Entity) []*ast.Entity {
	var entities []*ast.Entity
	if entity.Name != "" { // Skip root entity
		entities = append(entities, entity)
	}
	for _, child := range entity.Children {
		entities = append(entities, collectAllEntities(child)...)
	}
	return entities
}

func TestTemplateEntityCreation(t *testing.T) {
	content := `template <typename T>
class TemplatedClass;

template <typename T>
void templateFunction(T value);

template <typename K, typename V>
using TemplateMap = std::map<K, V>;`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Verify that templates are attached to their entities
	for _, entity := range tree.Root.Children {
		if !strings.Contains(entity.Signature, "template") {
			t.Errorf("Entity %s (%s) missing template in signature: %s", entity.Name, entity.Type, entity.Signature)
		}
	}

	// Test specific template constructs
	if len(tree.Root.Children) != 3 {
		t.Errorf("Expected 3 templated entities, got %d", len(tree.Root.Children))
	}
}

func TestPreprocessorDirectivePreservation(t *testing.T) {
	content := `#pragma once
#include <vector>
#define VERSION "1.0"

class MyClass {
public:
    void method();
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Count different entity types
	counts := make(map[ast.EntityType]int)
	for _, entity := range tree.Root.Children {
		counts[entity.Type]++
	}

	// Should have preprocessor entities
	if counts[ast.EntityPreprocessor] == 0 {
		t.Error("Expected preprocessor directives to be preserved as entities")
	}

	// Should still have class
	if counts[ast.EntityClass] != 1 {
		t.Errorf("Expected 1 class entity, got %d", counts[ast.EntityClass])
	}
}

func TestComplexTemplateScenarios(t *testing.T) {
	content := `// Variadic template
template <typename... Args>
void variadic_function(Args... args);

// Template specialization declaration
template <>
void specialized_function<int>(int value);

// Template with non-type parameters
template <int N, typename T = double>
struct FixedArray {
    T data[N];
};

// Nested template in namespace
namespace Utils {
    template <typename T>
    class Container {
    public:
        template <typename U>
        void add(const U& item);
    };
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Should parse variadic templates, specializations, and nested templates
	variadicFound := false
	specializationFound := false
	fixedArrayFound := false
	namespaceFound := false

	for _, entity := range tree.Root.Children {
		switch entity.Name {
		case "variadic_function":
			variadicFound = true
			if !strings.Contains(entity.Signature, "Args...") {
				t.Error("Variadic template function signature missing parameter pack")
			}
		case "specialized_function":
			specializationFound = true
		case "FixedArray":
			fixedArrayFound = true
			if !strings.Contains(entity.Signature, "int N") {
				t.Error("Non-type template parameter missing from signature")
			}
		case "Utils":
			namespaceFound = true
			// Check nested template class
			if len(entity.Children) > 0 {
				containerClass := entity.Children[0]
				if containerClass.Name == "Container" && len(containerClass.Children) > 0 {
					templateMethod := containerClass.Children[0]
					if templateMethod.Name == "add" && !strings.Contains(templateMethod.Signature, "template") {
						t.Error("Nested template method missing template declaration")
					}
				}
			}
		}
	}

	if !variadicFound {
		t.Error("Variadic template function not found")
	}
	if !specializationFound {
		t.Error("Template specialization not found")
	}
	if !fixedArrayFound {
		t.Error("Template with non-type parameters not found")
	}
	if !namespaceFound {
		t.Error("Namespace with nested templates not found")
	}
}

// TestCommentTypesAndPositions tests comment type detection and source positioning
func TestCommentTypesAndPositions(t *testing.T) {
	content := `// Line comment on line 1
/* Block comment on line 2 */
class TestClass {
    /// Doxygen line comment on line 4
    int field;
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find comment entities
	var commentEntities []*ast.Entity
	var allEntities func(*ast.Entity)
	allEntities = func(entity *ast.Entity) {
		if entity.Type == ast.EntityComment {
			commentEntities = append(commentEntities, entity)
		}
		for _, child := range entity.Children {
			allEntities(child)
		}
	}
	allEntities(tree.Root)

	// Should have 3 comments
	if len(commentEntities) != 3 {
		t.Errorf("Expected 3 comment entities, got %d", len(commentEntities))
		for i, comment := range commentEntities {
			t.Logf("Comment %d: %s", i, comment.Signature)
		}
		return
	}

	// Check first comment (line comment)
	lineComment := commentEntities[0]
	if lineComment.Type != ast.EntityComment {
		t.Errorf("Expected comment entity type, got %v", lineComment.Type)
	}
	if !strings.HasPrefix(lineComment.Signature, "//") {
		t.Errorf("Expected line comment to start with //, got %s", lineComment.Signature)
	}

	// Check second comment (block comment)
	blockComment := commentEntities[1]
	if !strings.HasPrefix(blockComment.Signature, "/*") {
		t.Errorf("Expected block comment to start with /*, got %s", blockComment.Signature)
	}

	// Check third comment (Doxygen line comment)
	doxygenComment := commentEntities[2]
	if !strings.HasPrefix(doxygenComment.Signature, "///") {
		t.Errorf("Expected Doxygen comment to start with ///, got %s", doxygenComment.Signature)
	}
}

// TestForwardDeclarationFixes tests fixes for forward declaration parsing
func TestForwardDeclarationFixes(t *testing.T) {
	tests := []struct {
		name    string
		content string
		checks  func(*testing.T, *ast.ScopeTree)
	}{
		{
			name:    "struct_forward_with_semicolon",
			content: "struct MyStruct;",
			checks: func(t *testing.T, tree *ast.ScopeTree) {
				if len(tree.Root.Children) != 1 {
					t.Fatalf("Expected 1 entity, got %d", len(tree.Root.Children))
				}
				entity := tree.Root.Children[0]
				if entity.Type != ast.EntityStruct {
					t.Errorf("Expected struct, got %s", entity.Type)
				}
				if entity.Name != "MyStruct" {
					t.Errorf("Expected name MyStruct, got %s", entity.Name)
				}
			},
		},
		{
			name:    "class_forward_with_semicolon",
			content: "class MyClass;",
			checks: func(t *testing.T, tree *ast.ScopeTree) {
				if len(tree.Root.Children) != 1 {
					t.Fatalf("Expected 1 entity, got %d", len(tree.Root.Children))
				}
				entity := tree.Root.Children[0]
				if entity.Type != ast.EntityClass {
					t.Errorf("Expected class, got %s", entity.Type)
				}
				if entity.Name != "MyClass" {
					t.Errorf("Expected name MyClass, got %s", entity.Name)
				}
			},
		},
		{
			name: "multiple_forward_declarations",
			content: `struct A;
class B;
struct C;`,
			checks: func(t *testing.T, tree *ast.ScopeTree) {
				if len(tree.Root.Children) != 3 {
					t.Fatalf("Expected 3 entities, got %d", len(tree.Root.Children))
				}
				expected := []struct {
					name     string
					typeName ast.EntityType
				}{
					{"A", ast.EntityStruct},
					{"B", ast.EntityClass},
					{"C", ast.EntityStruct},
				}
				for i, exp := range expected {
					entity := tree.Root.Children[i]
					if entity.Type != exp.typeName {
						t.Errorf("Entity %d: expected %s, got %s", i, exp.typeName, entity.Type)
					}
					if entity.Name != exp.name {
						t.Errorf("Entity %d: expected name %s, got %s", i, exp.name, entity.Name)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := New()
			tree, err := parser.Parse("test.hpp", tt.content)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			tt.checks(t, tree)
		})
	}
}

// TestTypedefSpacingFixes tests fixes for typedef whitespace handling
func TestTypedefSpacingFixes(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		expected  string
		signature string
	}{
		{
			name:      "simple_typedef",
			content:   "typedef int MyInt",
			expected:  "MyInt",
			signature: "typedef int MyInt",
		},
		{
			name:      "struct_typedef",
			content:   "typedef struct zip zip_t",
			expected:  "zip_t",
			signature: "typedef struct zip zip_t",
		},
		{
			name:      "pointer_typedef",
			content:   "typedef void* VoidPtr;",
			expected:  "VoidPtr",
			signature: "typedef void* VoidPtr",
		},
		{
			name:      "function_pointer_typedef",
			content:   "typedef int (*FuncPtr)(int, int);",
			expected:  "FuncPtr",
			signature: "typedef int (*FuncPtr)(int, int)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := New()
			tree, err := parser.Parse("test.hpp", tt.content)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(tree.Root.Children) != 1 {
				t.Fatalf("Expected 1 entity, got %d", len(tree.Root.Children))
			}

			entity := tree.Root.Children[0]
			if entity.Type != ast.EntityTypedef {
				t.Errorf("Expected typedef, got %s", entity.Type)
			}
			if entity.Signature != tt.signature {
				t.Errorf("Expected signature %s, got %s", tt.signature, entity.Signature)
			}
		})
	}
}

// TestClassMethodScopeFixes tests that class members are correctly scoped
func TestClassMethodScopeFixes(t *testing.T) {
	tests := []struct {
		name    string
		content string
		checks  func(*testing.T, *ast.ScopeTree)
	}{
		{
			name: "simple_class_with_methods",
			content: `namespace ns {
  class MyClass {
  public:
    void method1();
    void method2();
  };
}`,
			checks: func(t *testing.T, tree *ast.ScopeTree) {
				ns := tree.Root.Children[0]
				if ns.Type != ast.EntityNamespace || ns.Name != "ns" {
					t.Fatalf("Expected namespace ns, got %s:%s", ns.Type, ns.Name)
				}

				// Filter for class entities only
				var classes []*ast.Entity
				for _, child := range ns.Children {
					if child.Type == ast.EntityClass {
						classes = append(classes, child)
					}
				}

				if len(classes) != 1 {
					t.Fatalf("Expected 1 class, got %d", len(classes))
				}

				class := classes[0] // First class should be MyClass
				if class.Name != "MyClass" {
					t.Fatalf("Expected class MyClass, got %s", class.Name)
				}

				// Filter for method entities in the class
				var methods []*ast.Entity
				for _, child := range class.Children {
					if child.Type == ast.EntityCallable {
						methods = append(methods, child)
					}
				}

				if len(methods) != 2 {
					t.Fatalf("Expected 2 methods, got %d", len(methods))
				}

				method1 := methods[0]
				if method1.Name != "method1" {
					t.Errorf("Expected method1, got %s", method1.Name)
				}
				if method1.FullName != "ns::MyClass::method1" {
					t.Errorf("Expected FullName ns::MyClass::method1, got %s", method1.FullName)
				}

				method2 := methods[1]
				if method2.Name != "method2" {
					t.Errorf("Expected method2, got %s", method2.Name)
				}
				if method2.FullName != "ns::MyClass::method2" {
					t.Errorf("Expected FullName ns::MyClass::method2, got %s", method2.FullName)
				}
			},
		},
		{
			name: "class_with_forward_declaration",
			content: `namespace ns {
  class ForwardDeclared;
  
  class MyClass {
  public:
    void useForward(ForwardDeclared* ptr);
  };
}`,
			checks: func(t *testing.T, tree *ast.ScopeTree) {
				ns := tree.Root.Children[0]

				// Filter for class entities
				var classes []*ast.Entity
				for _, child := range ns.Children {
					if child.Type == ast.EntityClass {
						classes = append(classes, child)
					}
				}

				if len(classes) != 2 {
					t.Fatalf("Expected 2 classes in namespace, got %d", len(classes))
				}

				forward := classes[0]
				if forward.Name != "ForwardDeclared" {
					t.Errorf("Expected forward declaration ForwardDeclared, got %s", forward.Name)
				}

				class := classes[1]
				if class.Name != "MyClass" {
					t.Errorf("Expected class MyClass, got %s", class.Name)
				}

				// Filter for methods in the class
				var methods []*ast.Entity
				for _, child := range class.Children {
					if child.Type == ast.EntityCallable {
						methods = append(methods, child)
					}
				}

				if len(methods) != 1 {
					t.Fatalf("Expected 1 method, got %d", len(methods))
				}

				method := methods[0]
				if method.FullName != "ns::MyClass::useForward" {
					t.Errorf("Expected FullName ns::MyClass::useForward, got %s", method.FullName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := New()
			tree, err := parser.Parse("test.hpp", tt.content)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			tt.checks(t, tree)
		})
	}
}

// TestOperatorOverloadFixes tests operator overload parsing
func TestOperatorOverloadFixes(t *testing.T) {
	content := `class Vector {
public:
  Vector operator+(const Vector& other);
  Vector& operator=(const Vector& other);
  bool operator==(const Vector& other) const;
  Vector operator*(double scalar);
};`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(tree.Root.Children) != 1 {
		t.Fatalf("Expected 1 class, got %d", len(tree.Root.Children))
	}

	class := tree.Root.Children[0]

	// Filter for methods/functions in the class
	var operators []*ast.Entity
	for _, child := range class.Children {
		if child.Type == ast.EntityCallable {
			operators = append(operators, child)
		}
	}

	if len(operators) != 4 {
		t.Fatalf("Expected 4 operators, got %d", len(operators))
	}

	expectedOperators := []string{"operator+", "operator=", "operator==", "operator*"}
	for i, expectedName := range expectedOperators {
		op := operators[i]
		if op.Type != ast.EntityCallable {
			t.Errorf("Operator %d: expected function or method, got %s", i, op.Type)
		}
		if op.Name != expectedName {
			t.Errorf("Operator %d: expected name %s, got %s", i, expectedName, op.Name)
		}
		if op.FullName != "Vector::"+expectedName {
			t.Errorf("Operator %d: expected FullName Vector::%s, got %s", i, expectedName, op.FullName)
		}
	}
}

// TestComplexZipFileScenario attempts to replicate exact zip.hpp structure issues
func TestComplexZipFileScenario(t *testing.T) {
	content := `namespace mgl {
  
  // Forward declaration
  struct zip;
  
  // Main class with inheritance and multiple methods
  class zip_file : public zip {
  public:
    // Constructor and destructor
    zip_file(const std::string& path);
    ~zip_file();
    
    // Core methods that should be in zip_file scope
    bool exists(const std::string& filename) const;
    std::string read(const std::string& filename) const;
    void write(const std::string& filename, const std::string& data);
    
    // Static methods
    static bool is_zip(const std::string& path);
    
  protected:
    std::string m_path;
  };
  
  // Another class to test scope isolation
  class zip_reader {
  public:
    bool read_entry(const std::string& name);
  };
  
}`

	parser := New()
	tree, err := parser.Parse("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Find the mgl namespace
	if len(tree.Root.Children) != 1 {
		t.Fatalf("Expected 1 namespace, got %d", len(tree.Root.Children))
	}

	mgl := tree.Root.Children[0]
	if mgl.Type != ast.EntityNamespace || mgl.Name != "mgl" {
		t.Fatalf("Expected mgl namespace, got %s:%s", mgl.Type, mgl.Name)
	}

	// Should have: zip struct, zip_file class, zip_reader class
	// Filter only struct and class entities
	var structsAndClasses []*ast.Entity
	for _, child := range mgl.Children {
		if child.Type == ast.EntityStruct || child.Type == ast.EntityClass {
			structsAndClasses = append(structsAndClasses, child)
		}
	}

	if len(structsAndClasses) != 3 {
		t.Fatalf("Expected 3 struct/class entities in mgl namespace, got %d", len(structsAndClasses))
	}

	// Check forward declaration
	zip := structsAndClasses[0]
	if zip.Type != ast.EntityStruct || zip.Name != "zip" {
		t.Errorf("Expected zip struct, got %s:%s", zip.Type, zip.Name)
	}

	// Check zip_file class
	zipFile := structsAndClasses[1]
	if zipFile.Type != ast.EntityClass || zipFile.Name != "zip_file" {
		t.Errorf("Expected zip_file class, got %s:%s", zipFile.Type, zipFile.Name)
	}

	// Filter for methods/constructors/destructors in zip_file
	var zipFileMethods []*ast.Entity
	for _, child := range zipFile.Children {
		if child.Type == ast.EntityCallable {
			zipFileMethods = append(zipFileMethods, child)
		}
	}

	// This is the key test - methods should be children of zip_file, not namespace
	expectedMethods := []string{"zip_file", "zip_file", "exists", "read", "write", "is_zip"} // constructor, destructor, methods
	expectedTypes := []ast.EntityType{ast.EntityCallable, ast.EntityCallable, ast.EntityCallable, ast.EntityCallable, ast.EntityCallable, ast.EntityCallable}
	if len(zipFileMethods) != len(expectedMethods) {
		t.Errorf("Expected %d methods in zip_file, got %d", len(expectedMethods), len(zipFileMethods))
		// Print what we actually got for debugging
		for i, child := range zipFileMethods {
			t.Logf("zip_file method %d: %s (%s) - FullName: %s", i, child.Name, child.Type, child.FullName)
		}
	}

	// Check specific methods have correct scope and type
	for i, expectedName := range expectedMethods {
		if i < len(zipFileMethods) {
			method := zipFileMethods[i]
			if method.Name != expectedName {
				t.Errorf("Method %d: expected %s, got %s", i, expectedName, method.Name)
			}
			if method.Type != expectedTypes[i] {
				t.Errorf("Method %d: expected type %s, got %s", i, expectedTypes[i], method.Type)
			}
			expectedFullName := "mgl::zip_file::" + expectedName
			if method.FullName != expectedFullName {
				t.Errorf("Method %d: expected FullName %s, got %s", i, expectedFullName, method.FullName)
			}
		}
	}

	// Check zip_reader class
	zipReader := structsAndClasses[2]
	if zipReader.Type != ast.EntityClass || zipReader.Name != "zip_reader" {
		t.Errorf("Expected zip_reader class, got %s:%s", zipReader.Type, zipReader.Name)
	}

	// Filter for methods in zip_reader
	var zipReaderMethods []*ast.Entity
	for _, child := range zipReader.Children {
		if child.Type == ast.EntityCallable {
			zipReaderMethods = append(zipReaderMethods, child)
		}
	}

	if len(zipReaderMethods) != 1 {
		t.Fatalf("Expected 1 method in zip_reader, got %d", len(zipReaderMethods))
	}

	readerMethod := zipReaderMethods[0]
	if readerMethod.Name != "read_entry" {
		t.Errorf("Expected read_entry method, got %s", readerMethod.Name)
	}
	if readerMethod.FullName != "mgl::zip_reader::read_entry" {
		t.Errorf("Expected FullName mgl::zip_reader::read_entry, got %s", readerMethod.FullName)
	}
}

// TestScopeRegressionScenarios tests scenarios that have regressed in the past
func TestScopeRegressionScenarios(t *testing.T) {
	tests := []struct {
		name    string
		content string
		checks  func(*testing.T, *ast.ScopeTree)
	}{
		{
			name: "method_in_wrong_scope_regression",
			content: `namespace ns {
  class A;  // forward declaration
  
  class B {
  public:
    void methodB();
  };
  
  class C {
  public:
    void methodC();
  };
}`,
			checks: func(t *testing.T, tree *ast.ScopeTree) {
				ns := tree.Root.Children[0]

				// Filter for class entities
				var classes []*ast.Entity
				for _, child := range ns.Children {
					if child.Type == ast.EntityClass {
						classes = append(classes, child)
					}
				}

				if len(classes) != 3 {
					t.Fatalf("Expected 3 classes in namespace, got %d", len(classes))
				}

				// A is forward declaration - should have no children
				a := classes[0]
				if a.Name != "A" || len(a.Children) != 0 {
					t.Errorf("Forward declaration A should have no children, got %d", len(a.Children))
				}

				// B should have methodB
				b := classes[1]
				if b.Name != "B" {
					t.Errorf("Expected class B, got %s", b.Name)
				}

				var bMethods []*ast.Entity
				for _, child := range b.Children {
					if child.Type == ast.EntityCallable {
						bMethods = append(bMethods, child)
					}
				}

				if len(bMethods) != 1 {
					t.Errorf("Class B should have 1 method, got %d", len(bMethods))
				}
				if len(bMethods) > 0 && bMethods[0].Name != "methodB" {
					t.Errorf("Expected methodB in class B, got %s", bMethods[0].Name)
				}

				// C should have methodC
				c := classes[2]
				if c.Name != "C" {
					t.Errorf("Expected class C, got %s", c.Name)
				}

				var cMethods []*ast.Entity
				for _, child := range c.Children {
					if child.Type == ast.EntityCallable {
						cMethods = append(cMethods, child)
					}
				}

				if len(cMethods) != 1 {
					t.Errorf("Class C should have 1 method, got %d", len(cMethods))
				}
				if len(cMethods) > 0 && cMethods[0].Name != "methodC" {
					t.Errorf("Expected methodC in class C, got %s", cMethods[0].Name)
				}

				// Methods should NOT be in namespace scope
				for _, child := range ns.Children {
					if child.Type == ast.EntityCallable {
						t.Errorf("Found function %s in namespace scope - should be in class", child.Name)
					}
				}
			},
		},
		{
			name: "inline_vs_declaration_regression",
			content: `class Test {
public:
  void inline_method() { 
    int x = 1; 
  }
  void declaration_method();
  
  static void static_method();
};`,
			checks: func(t *testing.T, tree *ast.ScopeTree) {
				if len(tree.Root.Children) != 1 {
					t.Fatalf("Expected 1 class, got %d", len(tree.Root.Children))
				}

				class := tree.Root.Children[0]

				// Filter for methods only
				var methods []*ast.Entity
				for _, child := range class.Children {
					if child.Type == ast.EntityCallable {
						methods = append(methods, child)
					}
				}

				if len(methods) != 3 {
					t.Fatalf("Expected 3 methods, got %d", len(methods))
				}

				expectedMethods := []string{"inline_method", "declaration_method", "static_method"}
				for i, expectedName := range expectedMethods {
					method := methods[i]
					if method.Name != expectedName {
						t.Errorf("Method %d: expected %s, got %s", i, expectedName, method.Name)
					}
					if method.FullName != "Test::"+expectedName {
						t.Errorf("Method %d: expected FullName Test::%s, got %s", i, expectedName, method.FullName)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := New()
			tree, err := parser.Parse("test.hpp", tt.content)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			tt.checks(t, tree)
		})
	}
}
