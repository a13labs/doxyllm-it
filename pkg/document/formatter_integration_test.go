package document

import (
	"strings"
	"testing"
)

// Test content for formatter integration testing
const formatterIntegrationContent = `
#pragma once

namespace Graphics {

class Renderer {
public:
    void render();
    
    void setBackend(const std::string& backend);
    
private:
    bool initialized_;
};

} // namespace Graphics
`

func TestFormatterIntegration(t *testing.T) {
	doc, err := NewFromContent("test_formatter.hpp", formatterIntegrationContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add documentation to entities
	err = doc.SetEntityBrief("Graphics", "Graphics rendering namespace")
	if err != nil {
		t.Fatalf("Failed to set namespace brief: %v", err)
	}

	err = doc.SetEntityBrief("Graphics::Renderer", "Main rendering class")
	if err != nil {
		t.Fatalf("Failed to set class brief: %v", err)
	}

	err = doc.SetEntityDetailed("Graphics::Renderer", "This class handles all rendering operations and backend management.")
	if err != nil {
		t.Fatalf("Failed to set class detailed: %v", err)
	}

	err = doc.AddEntityGroup("Graphics::Renderer", "rendering")
	if err != nil {
		t.Fatalf("Failed to add class to group: %v", err)
	}

	err = doc.SetEntityBrief("Graphics::Renderer::render", "Renders the current scene")
	if err != nil {
		t.Fatalf("Failed to set method brief: %v", err)
	}

	err = doc.AddEntityParam("Graphics::Renderer::setBackend", "backend", "The rendering backend to use")
	if err != nil {
		t.Fatalf("Failed to add parameter documentation: %v", err)
	}

	err = doc.SetEntityBrief("Graphics::Renderer::setBackend", "Sets the rendering backend")
	if err != nil {
		t.Fatalf("Failed to set method brief: %v", err)
	}

	// Test SaveToString functionality
	result, err := doc.SaveToString()
	if err != nil {
		t.Fatalf("Failed to save to string: %v", err)
	}

	// Verify the reconstructed code contains our documentation
	expectedElements := []string{
		"@brief Graphics rendering namespace",
		"@brief Main rendering class",
		"This class handles all rendering operations",
		"@ingroup rendering",
		"@brief Renders the current scene",
		"@brief Sets the rendering backend",
		"@param backend The rendering backend to use",
		"namespace Graphics",
		"class Renderer",
		"void render()",
		"void setBackend(const std::string& backend)",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected reconstructed code to contain '%s', but it didn't.\nFull result:\n%s", element, result)
		}
	}

	// Verify the document is marked as modified
	if !doc.IsModified() {
		t.Error("Document should be marked as modified after adding documentation")
	}
}

func TestGetEntityContext(t *testing.T) {
	doc, err := NewFromContent("test_context.hpp", formatterIntegrationContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add some documentation
	err = doc.SetEntityBrief("Graphics::Renderer::render", "Renders the current scene")
	if err != nil {
		t.Fatalf("Failed to set method brief: %v", err)
	}

	// Test GetEntityContext
	context, err := doc.GetEntityContext("Graphics::Renderer::render", true, true)
	if err != nil {
		t.Fatalf("Failed to get entity context: %v", err)
	}

	expectedElements := []string{
		"Parent context:",
		"Renderer",
		"Sibling context:",
		"setBackend",
		"Target entity:",
		"@brief Renders the current scene",
		"void render()",
	}

	for _, element := range expectedElements {
		if !strings.Contains(context, element) {
			t.Errorf("Expected context to contain '%s', but it didn't.\nFull context:\n%s", element, context)
		}
	}
}

func TestGetEntitySummaryFormatted(t *testing.T) {
	doc, err := NewFromContent("test_summary.hpp", formatterIntegrationContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add documentation
	err = doc.SetEntityBrief("Graphics::Renderer", "Main rendering class")
	if err != nil {
		t.Fatalf("Failed to set class brief: %v", err)
	}

	// Test GetEntitySummaryFormatted
	summary, err := doc.GetEntitySummaryFormatted("Graphics::Renderer")
	if err != nil {
		t.Fatalf("Failed to get entity summary: %v", err)
	}

	expectedElements := []string{
		"Type: class",
		"Name: Renderer",
		"Full Name: Graphics::Renderer",
		"Has Documentation: true",
		"Children:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(summary, element) {
			t.Errorf("Expected summary to contain '%s', but it didn't.\nFull summary:\n%s", element, summary)
		}
	}
}

func TestReconstructScope(t *testing.T) {
	doc, err := NewFromContent("test_scope.hpp", formatterIntegrationContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add documentation
	err = doc.SetEntityBrief("Graphics::Renderer", "Main rendering class")
	if err != nil {
		t.Fatalf("Failed to set class brief: %v", err)
	}

	// Test ReconstructScope
	scope, err := doc.ReconstructScope("Graphics::Renderer")
	if err != nil {
		t.Fatalf("Failed to reconstruct scope: %v", err)
	}

	expectedElements := []string{
		"@brief Main rendering class",
		"class Renderer",
		"void render()",
		"void setBackend(const std::string& backend)",
	}

	for _, element := range expectedElements {
		if !strings.Contains(scope, element) {
			t.Errorf("Expected scope to contain '%s', but it didn't.\nFull scope:\n%s", element, scope)
		}
	}

	// Should not contain the namespace since we're only reconstructing the class scope
	if strings.Contains(scope, "namespace Graphics") {
		t.Error("Scope reconstruction should not contain parent namespace when reconstructing class scope")
	}
}

func TestSaveToStringFormatted(t *testing.T) {
	doc, err := NewFromContent("test_formatted.hpp", formatterIntegrationContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add documentation
	err = doc.SetEntityBrief("Graphics", "Graphics namespace")
	if err != nil {
		t.Fatalf("Failed to set namespace brief: %v", err)
	}

	// Test SaveToStringFormatted - this may fail if clang-format is not available
	result, err := doc.SaveToStringFormatted()

	// If clang-format is not available, we should still get the unformatted result
	if err != nil && !strings.Contains(err.Error(), "clang-format not available") {
		t.Fatalf("Unexpected error from SaveToStringFormatted: %v", err)
	}

	// The result should contain our documentation regardless of formatting
	if !strings.Contains(result, "@brief Graphics namespace") {
		t.Error("Expected formatted result to contain documentation")
	}

	if !strings.Contains(result, "namespace Graphics") {
		t.Error("Expected formatted result to contain namespace declaration")
	}
}

func TestBatchUpdatesWithFormatterIntegration(t *testing.T) {
	doc, err := NewFromContent("test_batch.hpp", formatterIntegrationContent)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Create batch updates
	updates := []BatchUpdate{
		{
			EntityPath: "Graphics",
			Brief:      stringPtr("Graphics rendering namespace"),
			Groups:     []string{"graphics"},
		},
		{
			EntityPath: "Graphics::Renderer",
			Brief:      stringPtr("Main rendering class"),
			Detailed:   stringPtr("Handles all rendering operations"),
			Groups:     []string{"rendering", "graphics"},
		},
		{
			EntityPath: "Graphics::Renderer::render",
			Brief:      stringPtr("Renders the current scene"),
		},
	}

	err = doc.ApplyBatchUpdates(updates)
	if err != nil {
		t.Fatalf("Failed to apply batch updates: %v", err)
	}

	// Test that formatter integration works with batch updates
	result, err := doc.SaveToString()
	if err != nil {
		t.Fatalf("Failed to save to string after batch updates: %v", err)
	}

	expectedElements := []string{
		"@brief Graphics rendering namespace",
		"@ingroup graphics",
		"@brief Main rendering class",
		"Handles all rendering operations",
		"@ingroup rendering",
		"@brief Renders the current scene",
	}

	for _, element := range expectedElements {
		if !strings.Contains(result, element) {
			t.Errorf("Expected batch update result to contain '%s', but it didn't.\nFull result:\n%s", element, result)
		}
	}
}

// Helper function to create string pointers for batch updates
func stringPtr(s string) *string {
	return &s
}
