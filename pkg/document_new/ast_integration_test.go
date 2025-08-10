package document_new

import (
	"testing"

	ast "doxyllm-it/pkg/ast_new"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestCommentASTreeIntegration(t *testing.T) {
	content := `class TestClass {
public:
    int field;
    void method();
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Initially, there should be no comments
	classEntity := doc.FindEntity("TestClass")
	if classEntity == nil {
		t.Fatal("Could not find TestClass")
	}

	if doc.HasDoxygenComment(classEntity) {
		t.Error("TestClass should not initially have Doxygen comment")
	}

	// Get initial children count of root
	initialChildrenCount := len(doc.tree.Root.Children)

	// Add a brief comment to the class
	t.Logf("Setting brief for TestClass...")
	err = doc.SetEntityBrief("TestClass", "A test class for demonstration")
	if err != nil {
		t.Fatalf("Failed to set brief: %v", err)
	}
	t.Logf("Brief set successfully")

	// Debug: Check parent AFTER setting brief
	t.Logf("After setting brief - Parent children count: %d", len(classEntity.Parent.Children))

	// Debug: List all parent children
	for i, child := range classEntity.Parent.Children {
		t.Logf("  Parent child %d: %s (%s)", i, child.Type, child.Name)
	}

	// Debug: Check comment cache
	cachedComment := doc.GetDoxygenComment(classEntity)
	if cachedComment == nil {
		t.Fatal("Comment should be in cache")
	}
	t.Logf("Comment in cache - Brief: '%s'", cachedComment.Brief)

	// Debug: Print all root children
	t.Logf("Root children after adding comment:")
	for i, child := range doc.tree.Root.Children {
		t.Logf("  %d: %s (%s) - %s", i, child.Type, child.Name, child.OriginalText[:min(50, len(child.OriginalText))])
	}

	// Now the class should have a comment
	if !doc.HasDoxygenComment(classEntity) {
		t.Error("TestClass should now have Doxygen comment")
	}

	// Check that a comment entity was added to the AST
	// The comment should be inserted before the class
	afterChildrenCount := len(doc.tree.Root.Children)
	if afterChildrenCount != initialChildrenCount+1 {
		t.Errorf("Expected %d children after adding comment, got %d",
			initialChildrenCount+1, afterChildrenCount)
	}

	// Find the comment entity
	var commentEntity *ast.Entity
	for i, child := range doc.tree.Root.Children {
		if child.Type == ast.EntityComment {
			commentEntity = child
			// Check that the comment comes before the class
			if i+1 < len(doc.tree.Root.Children) {
				nextChild := doc.tree.Root.Children[i+1]
				if nextChild.Type == ast.EntityClass && nextChild.Name == "TestClass" {
					t.Logf("Comment correctly placed before TestClass")
				}
			}
			break
		}
	}

	if commentEntity == nil {
		t.Error("Comment entity not found in AST")
	} else {
		// Verify comment content
		if commentEntity.OriginalText == "" {
			t.Error("Comment entity has empty OriginalText")
		}
		t.Logf("Comment text: %s", commentEntity.OriginalText)
	}

	// Test modifying the comment
	err = doc.SetEntityDetailed("TestClass", "This class provides detailed functionality for testing purposes")
	if err != nil {
		t.Fatalf("Failed to set detailed description: %v", err)
	}

	// Should still have same number of children (comment updated, not added)
	finalChildrenCount := len(doc.tree.Root.Children)
	if finalChildrenCount != afterChildrenCount {
		t.Errorf("Expected %d children after updating comment, got %d",
			afterChildrenCount, finalChildrenCount)
	}

	// Verify the updated comment
	comment := doc.GetDoxygenComment(classEntity)
	if comment == nil {
		t.Fatal("Failed to get updated comment")
	}

	if comment.Brief != "A test class for demonstration" {
		t.Errorf("Expected brief 'A test class for demonstration', got '%s'", comment.Brief)
	}

	if comment.Detailed != "This class provides detailed functionality for testing purposes" {
		t.Errorf("Expected detailed description, got '%s'", comment.Detailed)
	}
}

func TestCommentForMethod(t *testing.T) {
	content := `class TestClass {
public:
    void method();
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Get the method entity
	methodEntity := doc.FindEntity("TestClass::method")
	if methodEntity == nil {
		t.Fatal("Could not find TestClass::method")
	}

	// Add comment to method
	err = doc.AddEntityParam("TestClass::method", "x", "Input parameter")
	if err != nil {
		t.Fatalf("Failed to add param: %v", err)
	}

	// Check that comment was added
	if !doc.HasDoxygenComment(methodEntity) {
		t.Error("Method should have Doxygen comment")
	}

	// Find the TestClass entity to check its children
	classEntity := doc.FindEntity("TestClass")
	if classEntity == nil {
		t.Fatal("Could not find TestClass")
	}

	// Look for comment entity in class children (should be before method)
	var foundComment bool
	for i, child := range classEntity.Children {
		if child.Type == ast.EntityComment {
			foundComment = true
			// Check that next child is the method (or access specifier before method)
			if i+1 < len(classEntity.Children) {
				nextChild := classEntity.Children[i+1]
				t.Logf("Comment followed by: %s (%s)", nextChild.Type, nextChild.Name)
			}
			break
		}
	}

	if !foundComment {
		t.Error("Comment entity not found in class children")
	}
}
