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

	documentable := doc.GetDocumentableEntities()
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

	if !doc.HasDoxygenComment(classEntity) {
		t.Error("TestClass should have Doxygen comment")
	}

	classComment := doc.GetDoxygenComment(classEntity)
	if classComment == nil {
		t.Error("TestClass should have parsed Doxygen comment")
	} else {
		if classComment.Brief != "A test class" {
			t.Errorf("Expected brief 'A test class', got '%s'", classComment.Brief)
		}
	}

	// Test method has Doxygen comment
	methodEntity := doc.FindEntity("TestClass::testMethod")
	if methodEntity == nil {
		t.Fatal("Could not find TestClass::testMethod")
	}

	if !doc.HasDoxygenComment(methodEntity) {
		t.Error("testMethod should have Doxygen comment")
	}

	methodComment := doc.GetDoxygenComment(methodEntity)
	if methodComment == nil {
		t.Error("testMethod should have parsed Doxygen comment")
	} else {
		if methodComment.Brief != "A test method" {
			t.Errorf("Expected brief 'A test method', got '%s'", methodComment.Brief)
		}
		if len(methodComment.Params) != 1 {
			t.Errorf("Expected 1 param, got %d", len(methodComment.Params))
		}
		if methodComment.Params["x"] != "The parameter" {
			t.Errorf("Expected param x='The parameter', got '%s'", methodComment.Params["x"])
		}
		if methodComment.Returns != "The result" {
			t.Errorf("Expected return 'The result', got '%s'", methodComment.Returns)
		}
	}

	// Test field does not have Doxygen comment
	fieldEntity := doc.FindEntity("TestClass::regularField")
	if fieldEntity == nil {
		t.Fatal("Could not find TestClass::regularField")
	}

	if doc.HasDoxygenComment(fieldEntity) {
		t.Error("regularField should not have Doxygen comment")
	}
}

func TestDocumentationStats(t *testing.T) {
	content := `/// @brief Documented class
class DocumentedClass {
public:
    int undocumentedField;
    
    /// @brief Documented method
    void documentedMethod();
    
    void undocumentedMethod();
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	stats := doc.GetDocumentationStats()
	if stats.TotalEntities == 0 {
		t.Error("Expected some total entities")
	}

	if stats.DocumentedEntities == 0 {
		t.Error("Expected some documented entities")
	}

	if stats.UndocumentedEntities == 0 {
		t.Error("Expected some undocumented entities")
	}

	if stats.DocumentationCoverage <= 0 || stats.DocumentationCoverage > 100 {
		t.Errorf("Expected coverage between 0-100, got %.1f", stats.DocumentationCoverage)
	}

	// Verify the sum
	if stats.DocumentedEntities+stats.UndocumentedEntities != stats.TotalEntities {
		t.Error("Documented + undocumented should equal total")
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

	// Set brief for method
	err = doc.SetEntityBrief("TestClass::testMethod", "This is a test method")
	if err != nil {
		t.Fatalf("Failed to set brief: %v", err)
	}

	if !doc.IsModified() {
		t.Error("Document should be marked as modified")
	}

	// Verify the brief was set
	methodEntity := doc.FindEntity("TestClass::testMethod")
	if methodEntity == nil {
		t.Fatal("Could not find method entity")
	}

	comment := doc.GetDoxygenComment(methodEntity)
	if comment == nil {
		t.Fatal("Method should have comment after setting brief")
	}

	if comment.Brief != "This is a test method" {
		t.Errorf("Expected brief 'This is a test method', got '%s'", comment.Brief)
	}
}
