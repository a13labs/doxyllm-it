package document_new

import (
	"testing"
)

func TestDebugSetEntityBrief(t *testing.T) {
	content := `class TestClass {
public:
    int field;
};`

	doc, err := NewFromContent("test.hpp", content)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	classEntity := doc.FindEntity("TestClass")
	if classEntity == nil {
		t.Fatal("Could not find TestClass")
	}

	t.Logf("Before: Parent children count: %d", len(classEntity.Parent.Children))

	// Manually test the logic step by step
	comment := NewDoxygenComment()
	comment.Brief = "A test class"

	t.Logf("Created comment: %+v", comment)
	t.Logf("HasDoxygenContent: %t", comment.HasDoxygenContent())

	// Test comment text generation
	commentText := doc.generateCommentText(comment)
	t.Logf("Generated comment text: %s", commentText)

	// Test comment entity creation
	commentEntity := doc.createCommentEntity(comment)
	t.Logf("Created comment entity: %+v", commentEntity)

	// Test insertion
	err = doc.insertCommentEntityBeforeTarget(commentEntity, classEntity)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	t.Logf("After: Parent children count: %d", len(classEntity.Parent.Children))
	for i, child := range classEntity.Parent.Children {
		t.Logf("  Child %d: %s (%s)", i, child.Type, child.Name)
	}
}
