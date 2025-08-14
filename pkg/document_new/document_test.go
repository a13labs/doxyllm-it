package document_new

import (
	"testing"

	ast "doxyllm-it/pkg/ast_new"
)

func TestNewFromContent(t *testing.T) {
	content := `/** @brief A test class */
class TestClass {
public:
    /// @brief A test method
    /// @param x The parameter
    /// @return The result
    int testMethod(int x);
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	if doc.GetFilename() != "test.hpp" {
		t.Errorf("Expected filename test.hpp, got %s", doc.GetFilename())
	}

	if doc.IsModified() {
		t.Errorf("New document should not be modified")
	}
}

func TestFindEntity(t *testing.T) {
	content := `/** @brief A test class */
class TestClass {
public:
    /// @brief A test method
    int testMethod(int x);
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Find the class
	classEntity := doc.FindEntity("TestClass")
	if classEntity == nil {
		t.Error("Could not find TestClass entity")
	} else if classEntity.Type != ast.EntityClass {
		t.Errorf("Expected class entity, got %s", classEntity.Type)
	}

	// Find the method
	methodEntity := doc.FindEntity("TestClass::testMethod")
	if methodEntity == nil {
		t.Error("Could not find TestClass::testMethod entity")
	} else if methodEntity.Type != ast.EntityFunction {
		t.Errorf("Expected method entity, got %s", methodEntity.Type)
	}
}

func TestGetDocumentableEntities(t *testing.T) {
	content := `// Regular comment
class TestClass {
public:
    int field;
    void method();
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	documentable := doc.GetInstructions()
	if len(documentable) == 0 {
		t.Error("Expected some documentable entities")
	}

	// Should include class, field, and method but not comments
	var hasClass, hasField, hasFunction bool
	for _, entity := range documentable {
		switch entity.Type {
		case ast.EntityClass:
			hasClass = true
		case ast.EntityField:
			hasField = true
		case ast.EntityFunction:
			hasFunction = true
		case ast.EntityComment:
			t.Error("Comments should not be in documentable entities")
		}
	}

	if !hasClass {
		t.Error("Expected to find class in documentable entities")
	}
	if !hasField {
		t.Error("Expected to find field in documentable entities")
	}
	if !hasFunction {
		t.Error("Expected to find method in documentable entities")
	}
}

func TestDoxygenCommentDetection(t *testing.T) {
	content := `/// @brief A test class
class TestClass {
public:
    /** @brief A test method
     *  @param x The parameter
     *  @return The result
     */
    int testMethod(int x);
    
    // Regular comment, not Doxygen
    int regularField;
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Test class has Doxygen comment
	classEntity := doc.FindEntity("TestClass")
	if classEntity == nil {
		t.Fatal("Could not find TestClass")
	}

	// Test method has Doxygen comment
	methodEntity := doc.FindEntity("TestClass::testMethod")
	if methodEntity == nil {
		t.Fatal("Could not find TestClass::testMethod")
	}

	// Test field does not have Doxygen comment
	fieldEntity := doc.FindEntity("TestClass::regularField")
	if fieldEntity == nil {
		t.Fatal("Could not find TestClass::regularField")
	}
}

func TestSetEntityBrief(t *testing.T) {
	content := `class TestClass {
public:
    void testMethod();
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Verify the brief was set
	methodEntity := doc.FindEntity("TestClass::testMethod")
	if methodEntity == nil {
		t.Fatal("Could not find method entity")
	}
}
