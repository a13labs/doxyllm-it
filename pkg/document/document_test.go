package document

import (
	"testing"

	"doxyllm-it/pkg/ast"
)

// Test content for document operations
const testHeaderContent = `
/**
 * @file test.hpp
 * @brief Test header file for document API testing
 */

#pragma once

/**
 * @namespace TestNS
 * @brief Test namespace for document testing
 */
namespace TestNS {

/**
 * @class Calculator
 * @brief A simple calculator class
 * @ingroup math_utilities
 */
class Calculator {
public:
    /**
     * @brief Default constructor
     */
    Calculator();

    /**
     * @brief Adds two numbers
     * @param a First number
     * @param b Second number
     * @return Sum of a and b
     */
    int add(int a, int b);

    /**
     * @brief Multiplies two numbers
     * @param a First number
     * @param b Second number
     * @return Product of a and b
     */
    int multiply(int a, int b);

    // This function lacks documentation
    int subtract(int a, int b);

private:
    int result_;
};

// Global function without documentation
void globalFunction();

/**
 * @brief Global documented function
 * @param value Input value
 * @return Processed value
 */
int processValue(int value);

} // namespace TestNS
`

func TestNewFromContent(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	if doc.GetFilename() != "test.hpp" {
		t.Errorf("Expected filename 'test.hpp', got '%s'", doc.GetFilename())
	}

	if doc.IsModified() {
		t.Error("Newly created document should not be marked as modified")
	}
}

func TestFindEntity(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Test finding entities by path
	tests := []struct {
		path     string
		expected bool
		name     string
	}{
		{"TestNS", true, "TestNS"},
		{"TestNS::Calculator", true, "Calculator"},
		{"TestNS::Calculator::add", true, "add"},
		{"TestNS::Calculator::multiply", true, "multiply"},
		{"TestNS::Calculator::subtract", true, "subtract"},
		{"TestNS::processValue", true, "processValue"},
		{"NonExistent", false, ""},
		{"TestNS::NonExistent", false, ""},
	}

	for _, test := range tests {
		entity := doc.FindEntity(test.path)
		if test.expected {
			if entity == nil {
				t.Errorf("Expected to find entity at path '%s'", test.path)
			} else if entity.Name != test.name {
				t.Errorf("Expected entity name '%s', got '%s'", test.name, entity.Name)
			}
		} else {
			if entity != nil {
				t.Errorf("Expected not to find entity at path '%s', but found '%s'", test.path, entity.Name)
			}
		}
	}
}

func TestFindEntitiesByType(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Test finding by type
	classes := doc.FindEntitiesByType(ast.EntityClass)
	if len(classes) != 1 {
		t.Errorf("Expected 1 class, found %d", len(classes))
	}
	if len(classes) > 0 && classes[0].Name != "Calculator" {
		t.Errorf("Expected class 'Calculator', got '%s'", classes[0].Name)
	}

	methods := doc.FindEntitiesByType(ast.EntityMethod)
	expectedMethods := []string{"add", "multiply", "subtract"}
	if len(methods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, found %d", len(expectedMethods), len(methods))
	}

	constructors := doc.FindEntitiesByType(ast.EntityConstructor)
	if len(constructors) != 1 {
		t.Errorf("Expected 1 constructor, found %d", len(constructors))
	}
}

func TestGetUndocumentedEntities(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	undocumented := doc.GetUndocumentedEntities()

	// Should find subtract method, globalFunction, and result_ field as undocumented
	expectedUndocumented := []string{
		"TestNS::Calculator::subtract",
		"TestNS::globalFunction",
		"TestNS::Calculator::result_",
	}

	if len(undocumented) != len(expectedUndocumented) {
		t.Errorf("Expected %d undocumented entities, found %d", len(expectedUndocumented), len(undocumented))
		// Print what we actually found for debugging
		for _, entity := range undocumented {
			t.Logf("Found undocumented: %s", entity.GetFullPath())
		}
	}

	// Check that the right entities are undocumented
	for _, entity := range undocumented {
		path := entity.GetFullPath()
		found := false
		for _, expected := range expectedUndocumented {
			if path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected undocumented entity: %s", path)
		}
	}
}

func TestSetEntityBrief(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Test setting brief for undocumented entity
	err = doc.SetEntityBrief("TestNS::Calculator::subtract", "Subtracts two numbers")
	if err != nil {
		t.Errorf("Failed to set brief: %v", err)
	}

	if !doc.IsModified() {
		t.Error("Document should be marked as modified after setting brief")
	}

	entity := doc.FindEntity("TestNS::Calculator::subtract")
	if entity == nil {
		t.Fatal("Entity not found after setting brief")
	}

	if entity.Comment == nil || entity.Comment.Brief != "Subtracts two numbers" {
		t.Errorf("Expected brief 'Subtracts two numbers', got '%s'", entity.Comment.Brief)
	}
}

func TestAddEntityParam(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add parameter documentation to subtract method
	err = doc.AddEntityParam("TestNS::Calculator::subtract", "a", "First number")
	if err != nil {
		t.Errorf("Failed to add param: %v", err)
	}

	err = doc.AddEntityParam("TestNS::Calculator::subtract", "b", "Second number")
	if err != nil {
		t.Errorf("Failed to add param: %v", err)
	}

	entity := doc.FindEntity("TestNS::Calculator::subtract")
	if entity == nil || entity.Comment == nil {
		t.Fatal("Entity or comment not found")
	}

	if len(entity.Comment.Params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(entity.Comment.Params))
	}

	if entity.Comment.Params["a"] != "First number" {
		t.Errorf("Expected param 'a' = 'First number', got '%s'", entity.Comment.Params["a"])
	}

	if entity.Comment.Params["b"] != "Second number" {
		t.Errorf("Expected param 'b' = 'Second number', got '%s'", entity.Comment.Params["b"])
	}
}

func TestBatchUpdates(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Prepare batch updates
	brief := "Subtracts second number from first"
	detailed := "This method performs subtraction of two integer values."
	returnDesc := "Difference of a and b"
	deprecated := "Use Calculator::subtract_safe instead"

	updates := []BatchUpdate{
		{
			EntityPath: "TestNS::Calculator::subtract",
			Brief:      &brief,
			Detailed:   &detailed,
			Params: map[string]string{
				"a": "Minuend",
				"b": "Subtrahend",
			},
			Return: &returnDesc,
			Groups: []string{"arithmetic", "basic_ops"},
			CustomTags: map[string]string{
				"complexity":    "O(1)",
				"thread_safety": "safe",
			},
			Deprecated: &deprecated,
		},
	}

	err = doc.ApplyBatchUpdates(updates)
	if err != nil {
		t.Errorf("Failed to apply batch updates: %v", err)
	}

	// Verify updates
	entity := doc.FindEntity("TestNS::Calculator::subtract")
	if entity == nil || entity.Comment == nil {
		t.Fatal("Entity or comment not found after batch update")
	}

	comment := entity.Comment

	if comment.Brief != brief {
		t.Errorf("Expected brief '%s', got '%s'", brief, comment.Brief)
	}

	if comment.Detailed != detailed {
		t.Errorf("Expected detailed '%s', got '%s'", detailed, comment.Detailed)
	}

	if comment.Returns != returnDesc {
		t.Errorf("Expected return '%s', got '%s'", returnDesc, comment.Returns)
	}

	if comment.Deprecated != deprecated {
		t.Errorf("Expected deprecated '%s', got '%s'", deprecated, comment.Deprecated)
	}

	if len(comment.Ingroup) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(comment.Ingroup))
	}

	if comment.CustomTags["complexity"] != "O(1)" {
		t.Errorf("Expected custom tag 'complexity' = 'O(1)', got '%s'", comment.CustomTags["complexity"])
	}
}

func TestGetDocumentationStats(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	stats := doc.GetDocumentationStats()

	// We expect several documentable entities:
	// - TestNS namespace (documented)
	// - Calculator class (documented)
	// - Calculator constructor (documented)
	// - add method (documented)
	// - multiply method (documented)
	// - subtract method (undocumented)
	// - result_ field (undocumented - fields are usually undocumented)
	// - globalFunction (undocumented)
	// - processValue function (documented)

	if stats.TotalEntities == 0 {
		t.Error("Expected some documentable entities")
	}

	if stats.DocumentationCoverage < 0 || stats.DocumentationCoverage > 100 {
		t.Errorf("Documentation coverage should be between 0 and 100, got %.1f", stats.DocumentationCoverage)
	}

	if stats.DocumentedEntities+stats.UndocumentedEntities != stats.TotalEntities {
		t.Error("Sum of documented and undocumented entities should equal total")
	}
}

func TestValidation(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	issues := doc.Validate()

	// Should find validation issues for undocumented entities
	if len(issues) == 0 {
		t.Error("Expected some validation issues for undocumented entities")
	}

	// Check that we find missing documentation issues
	foundMissingDoc := false
	for _, issue := range issues {
		if issue.IssueType == "missing_documentation" {
			foundMissingDoc = true
			break
		}
	}

	if !foundMissingDoc {
		t.Error("Expected to find missing documentation issues")
	}
}

func TestGetEntitySummary(t *testing.T) {
	doc, err := NewFromContent("test.hpp", testHeaderContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Test documented entity
	summary, err := doc.GetEntitySummary("TestNS::Calculator::add")
	if err != nil {
		t.Errorf("Failed to get summary: %v", err)
	}

	if !summary.HasDoc {
		t.Error("add method should be documented")
	}

	if !summary.HasBrief {
		t.Error("add method should have brief description")
	}

	if summary.ParamCount != 2 {
		t.Errorf("add method should have 2 parameters, got %d", summary.ParamCount)
	}

	if !summary.HasReturn {
		t.Error("add method should have return documentation")
	}

	// Test undocumented entity
	summary, err = doc.GetEntitySummary("TestNS::Calculator::subtract")
	if err != nil {
		t.Errorf("Failed to get summary: %v", err)
	}

	if summary.HasDoc {
		t.Error("subtract method should not be documented initially")
	}
}

// Example usage test
func ExampleDocument_usage() {
	// Create document from content
	doc, _ := NewFromContent("example.hpp", testHeaderContent)

	// Find undocumented entities
	undocumented := doc.GetUndocumentedEntities()

	// Add documentation to undocumented entities
	for _, entity := range undocumented {
		path := entity.GetFullPath()
		switch entity.Type {
		case ast.EntityMethod, ast.EntityFunction:
			doc.SetEntityBrief(path, "TODO: Add brief description")
		}
	}

	// Get documentation statistics
	stats := doc.GetDocumentationStats()
	_ = stats.DocumentationCoverage // Now should be higher

	// Validate the document
	issues := doc.Validate()
	_ = len(issues) // Should be fewer issues now
}
