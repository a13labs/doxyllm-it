// Package main demonstrates the usage of the document abstraction layer
package main

import (
	"fmt"
	"log"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/document"
)

func main() {
	// Example: Working with a C++ header file using the document abstraction

	headerContent := `
#pragma once

namespace Graphics {

class Renderer {
public:
    Renderer();
    ~Renderer();
    
    void initialize();
    void render();
    void cleanup();
    
    void setWidth(int width);
    void setHeight(int height);
    
    int getWidth() const;
    int getHeight() const;

private:
    int width_;
    int height_;
    bool initialized_;
};

void initializeGraphics();
void shutdownGraphics();

} // namespace Graphics
`

	// Create a document from content
	fmt.Println("=== Creating Document ===")
	doc, err := document.NewFromContent("renderer.hpp", headerContent)
	if err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}

	fmt.Printf("Document created: %s\n", doc)

	// Get initial documentation statistics
	fmt.Println("\n=== Initial Documentation Statistics ===")
	stats := doc.GetDocumentationStats()
	fmt.Printf("Total entities: %d\n", stats.TotalEntities)
	fmt.Printf("Documented: %d\n", stats.DocumentedEntities)
	fmt.Printf("Undocumented: %d\n", stats.UndocumentedEntities)
	fmt.Printf("Coverage: %.1f%%\n", stats.DocumentationCoverage)

	// Find all undocumented entities
	fmt.Println("\n=== Undocumented Entities ===")
	undocumented := doc.GetUndocumentedEntities()
	for _, entity := range undocumented {
		fmt.Printf("- %s (%s)\n", entity.GetFullPath(), entity.Type)
	}

	// Add documentation to some entities
	fmt.Println("\n=== Adding Documentation ===")

	// Document the namespace
	err = doc.SetEntityBrief("Graphics", "Graphics rendering subsystem")
	if err != nil {
		log.Printf("Error setting namespace brief: %v", err)
	}

	// Document the class
	err = doc.SetEntityBrief("Graphics::Renderer", "A renderer for 2D graphics")
	if err != nil {
		log.Printf("Error setting class brief: %v", err)
	}

	err = doc.SetEntityDetailed("Graphics::Renderer",
		"The Renderer class provides functionality for rendering 2D graphics. "+
			"It manages the rendering context and provides methods for drawing operations.")
	if err != nil {
		log.Printf("Error setting class detailed: %v", err)
	}

	// Document constructor and destructor
	err = doc.SetEntityBrief("Graphics::Renderer::Renderer", "Constructs a new renderer instance")
	if err != nil {
		log.Printf("Error setting constructor brief: %v", err)
	}

	// Find the destructor with the correct name format
	destructors := doc.FindEntitiesByType(ast.EntityDestructor)
	for _, destructor := range destructors {
		err = doc.SetEntityBrief(destructor.GetFullPath(), "Destroys the renderer and cleans up resources")
		if err != nil {
			log.Printf("Error setting destructor brief: %v", err)
		}
	}

	// Document methods with parameters and return values
	methodDocs := []struct {
		path   string
		brief  string
		params map[string]string
		ret    string
	}{
		{
			path:  "Graphics::Renderer::initialize",
			brief: "Initializes the renderer",
		},
		{
			path:  "Graphics::Renderer::render",
			brief: "Renders the current frame",
		},
		{
			path:  "Graphics::Renderer::cleanup",
			brief: "Cleans up renderer resources",
		},
		{
			path:  "Graphics::Renderer::setWidth",
			brief: "Sets the rendering width",
			params: map[string]string{
				"width": "The new width in pixels",
			},
		},
		{
			path:  "Graphics::Renderer::setHeight",
			brief: "Sets the rendering height",
			params: map[string]string{
				"height": "The new height in pixels",
			},
		},
		{
			path:  "Graphics::Renderer::getWidth",
			brief: "Gets the current rendering width",
			ret:   "The current width in pixels",
		},
		{
			path:  "Graphics::Renderer::getHeight",
			brief: "Gets the current rendering height",
			ret:   "The current height in pixels",
		},
	}

	for _, methodDoc := range methodDocs {
		err = doc.SetEntityBrief(methodDoc.path, methodDoc.brief)
		if err != nil {
			log.Printf("Error setting brief for %s: %v", methodDoc.path, err)
			continue
		}

		for paramName, paramDesc := range methodDoc.params {
			err = doc.AddEntityParam(methodDoc.path, paramName, paramDesc)
			if err != nil {
				log.Printf("Error adding param for %s: %v", methodDoc.path, err)
			}
		}

		if methodDoc.ret != "" {
			err = doc.SetEntityReturn(methodDoc.path, methodDoc.ret)
			if err != nil {
				log.Printf("Error setting return for %s: %v", methodDoc.path, err)
			}
		}
	}

	// Document global functions
	err = doc.SetEntityBrief("Graphics::initializeGraphics", "Initializes the graphics subsystem")
	if err != nil {
		log.Printf("Error setting function brief: %v", err)
	}

	err = doc.SetEntityBrief("Graphics::shutdownGraphics", "Shuts down the graphics subsystem")
	if err != nil {
		log.Printf("Error setting function brief: %v", err)
	}

	// Add some entities to groups
	fmt.Println("\n=== Adding to Groups ===")
	rendererMethods := []string{
		"Graphics::Renderer::initialize",
		"Graphics::Renderer::render",
		"Graphics::Renderer::cleanup",
	}

	for _, method := range rendererMethods {
		err = doc.AddEntityGroup(method, "renderer_core")
		if err != nil {
			log.Printf("Error adding %s to group: %v", method, err)
		}
	}

	getterMethods := []string{
		"Graphics::Renderer::getWidth",
		"Graphics::Renderer::getHeight",
	}

	for _, method := range getterMethods {
		err = doc.AddEntityGroup(method, "getters")
		if err != nil {
			log.Printf("Error adding %s to group: %v", method, err)
		}
	}

	// Use batch updates for remaining entities
	fmt.Println("\n=== Using Batch Updates ===")
	brief1 := "Width of the rendering area"
	brief2 := "Height of the rendering area"
	brief3 := "Whether the renderer has been initialized"

	batchUpdates := []document.BatchUpdate{
		{
			EntityPath: "Graphics::Renderer::width_",
			Brief:      &brief1,
			Groups:     []string{"private_members"},
		},
		{
			EntityPath: "Graphics::Renderer::height_",
			Brief:      &brief2,
			Groups:     []string{"private_members"},
		},
		{
			EntityPath: "Graphics::Renderer::initialized_",
			Brief:      &brief3,
			Groups:     []string{"private_members"},
		},
	}

	err = doc.ApplyBatchUpdates(batchUpdates)
	if err != nil {
		log.Printf("Error applying batch updates: %v", err)
	}

	// Get updated statistics
	fmt.Println("\n=== Updated Documentation Statistics ===")
	stats = doc.GetDocumentationStats()
	fmt.Printf("Total entities: %d\n", stats.TotalEntities)
	fmt.Printf("Documented: %d\n", stats.DocumentedEntities)
	fmt.Printf("Undocumented: %d\n", stats.UndocumentedEntities)
	fmt.Printf("Coverage: %.1f%%\n", stats.DocumentationCoverage)

	// Show remaining undocumented entities
	fmt.Println("\n=== Remaining Undocumented Entities ===")
	undocumented = doc.GetUndocumentedEntities()
	if len(undocumented) == 0 {
		fmt.Println("All entities are now documented!")
	} else {
		for _, entity := range undocumented {
			fmt.Printf("- %s (%s)\n", entity.GetFullPath(), entity.Type)
		}
	}

	// Validate the document
	fmt.Println("\n=== Validation Results ===")
	issues := doc.Validate()
	if len(issues) == 0 {
		fmt.Println("No validation issues found!")
	} else {
		for _, issue := range issues {
			fmt.Printf("- %s [%s]: %s (%s)\n",
				issue.EntityPath, issue.Severity, issue.Message, issue.IssueType)
		}
	}

	// Show some entity summaries
	fmt.Println("\n=== Entity Summaries ===")
	summaryPaths := []string{
		"Graphics::Renderer",
		"Graphics::Renderer::render",
		"Graphics::Renderer::setWidth",
		"Graphics::Renderer::getWidth",
	}

	for _, path := range summaryPaths {
		summary, err := doc.GetEntitySummary(path)
		if err != nil {
			log.Printf("Error getting summary for %s: %v", path, err)
			continue
		}

		fmt.Printf("- %s (%s): doc=%t, brief=%t, detailed=%t, params=%d, return=%t\n",
			summary.Path, summary.Type, summary.HasDoc, summary.HasBrief,
			summary.HasDetailed, summary.ParamCount, summary.HasReturn)
	}

	// Demonstrate finding entities by type
	fmt.Println("\n=== Entities by Type ===")
	classes := doc.FindEntitiesByType(ast.EntityClass)
	fmt.Printf("Classes (%d):\n", len(classes))
	for _, class := range classes {
		fmt.Printf("  - %s\n", class.GetFullPath())
	}

	methods := doc.FindEntitiesByType(ast.EntityMethod)
	fmt.Printf("Methods (%d):\n", len(methods))
	for _, method := range methods {
		fmt.Printf("  - %s\n", method.GetFullPath())
	}

	functions := doc.FindEntitiesByType(ast.EntityFunction)
	fmt.Printf("Functions (%d):\n", len(functions))
	for _, function := range functions {
		fmt.Printf("  - %s\n", function.GetFullPath())
	}

	fmt.Printf("\nDocument modification status: %t\n", doc.IsModified())
	fmt.Println("\n=== Documentation Process Complete ===")
}
