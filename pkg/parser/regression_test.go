package parser

import (
	"strings"
	"testing"

	"doxyllm-it/pkg/ast"
)

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
			name: "documented_forward_declaration",
			content: `/** @brief A forward declared class */
class DocumentedClass;`,
			checks: func(t *testing.T, tree *ast.ScopeTree) {
				if len(tree.Root.Children) != 1 {
					t.Fatalf("Expected 1 entity, got %d", len(tree.Root.Children))
				}
				entity := tree.Root.Children[0]
				if entity.Type != ast.EntityClass {
					t.Errorf("Expected class, got %s", entity.Type)
				}
				if entity.Name != "DocumentedClass" {
					t.Errorf("Expected name DocumentedClass, got %s", entity.Name)
				}
				if entity.Comment == nil || !strings.Contains(entity.Comment.Raw, "@brief A forward declared class") {
					t.Errorf("Expected doxygen comment, got %v", entity.Comment)
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
			content:   "typedef int MyInt;",
			expected:  "MyInt",
			signature: "typedef int MyInt;",
		},
		{
			name:      "struct_typedef",
			content:   "typedef struct zip zip_t;",
			expected:  "zip_t",
			signature: "typedef struct zip zip_t;",
		},
		{
			name:      "pointer_typedef",
			content:   "typedef void* VoidPtr;",
			expected:  "VoidPtr",
			signature: "typedef void * VoidPtr;",
		},
		{
			name:      "function_pointer_typedef",
			content:   "typedef int (*FuncPtr)(int, int);",
			expected:  "FuncPtr",
			signature: "typedef int ( * FuncPtr ) ( int , int );",
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
			if entity.Name != tt.expected {
				t.Errorf("Expected name %s, got %s", tt.expected, entity.Name)
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
					if child.Type == ast.EntityFunction || child.Type == ast.EntityMethod {
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
					if child.Type == ast.EntityFunction || child.Type == ast.EntityMethod {
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
		if child.Type == ast.EntityFunction || child.Type == ast.EntityMethod {
			operators = append(operators, child)
		}
	}
	
	if len(operators) != 4 {
		t.Fatalf("Expected 4 operators, got %d", len(operators))
	}

	expectedOperators := []string{"operator+", "operator=", "operator==", "operator*"}
	for i, expectedName := range expectedOperators {
		op := operators[i]
		if op.Type != ast.EntityFunction && op.Type != ast.EntityMethod {
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
		if child.Type == ast.EntityFunction || child.Type == ast.EntityMethod || child.Type == ast.EntityConstructor || child.Type == ast.EntityDestructor {
			zipFileMethods = append(zipFileMethods, child)
		}
	}

	// This is the key test - methods should be children of zip_file, not namespace
	expectedMethods := []string{"zip_file", "zip_file", "exists", "read", "write", "is_zip"}  // constructor, destructor, methods
	expectedTypes := []ast.EntityType{ast.EntityConstructor, ast.EntityDestructor, ast.EntityMethod, ast.EntityMethod, ast.EntityMethod, ast.EntityMethod}
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
		if child.Type == ast.EntityFunction || child.Type == ast.EntityMethod {
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
					if child.Type == ast.EntityFunction || child.Type == ast.EntityMethod {
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
					if child.Type == ast.EntityFunction || child.Type == ast.EntityMethod {
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
					if child.Type == ast.EntityFunction {
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
					if child.Type == ast.EntityFunction || child.Type == ast.EntityMethod {
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
