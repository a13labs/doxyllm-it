package main

import (
	"fmt"
	"log"

	"doxyllm-it/pkg/document"
	"doxyllm-it/pkg/llm"
)

func main() {
	// Create document from our test file
	doc, err := document.NewFromFile("test_simple_templates.hpp")
	if err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}

	// Find the template alias
	entity := doc.FindEntity("Map")
	if entity == nil {
		log.Fatal("Map entity not found")
	}

	fmt.Printf("Entity: %s\n", entity.Name)
	fmt.Printf("Type: %s\n", entity.Type)
	fmt.Printf("Signature: %s\n", entity.Signature)
	fmt.Printf("IsTemplate: %t\n", entity.IsTemplate)
	fmt.Printf("TemplateParams: %v\n", entity.TemplateParams)

	// Create LLM provider
	provider, err := llm.NewProvider("ollama", &llm.Config{
		URL:   "http://10.19.4.136:11434/api/generate",
		Model: "deepseek-coder:6.7b",
	})
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	// Create documentation service
	service := llm.NewDocumentationService(provider)

	// Get context for the entity
	context := doc.GetEntityContext(entity.GetFullPath(), true, true)

	// Generate documentation
	request := &llm.DocumentationRequest{
		EntityName:        entity.Name,
		EntityType:        entity.Type.String(),
		Context:           context,
		ForceInline:       false,
		AdditionalContext: "This is a template alias for std::map. Please include @tparam documentation for template parameters.",
	}

	response, err := service.GenerateDocumentation(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}

	fmt.Printf("\nGenerated comment:\n%s\n", response.Comment)

	// Parse the generated comment to see if it contains tparams
	comment := llm.ParseDoxygenComment(response.Comment)
	if comment != nil {
		fmt.Printf("\nParsed comment:\n")
		fmt.Printf("Brief: %s\n", comment.Brief)
		fmt.Printf("TParams: %v\n", comment.TParams)
		fmt.Printf("Params: %v\n", comment.Params)
	}
}
