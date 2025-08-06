package main

import (
	"fmt"
	"log"
	"os"
	
	"doxyllm-it/pkg/parser"
	"doxyllm-it/pkg/formatter"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test.go <header-file>")
		os.Exit(1)
	}
	
	filename := os.Args[1]
	
	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	
	// Parse the file
	p := parser.New()
	tree, err := p.Parse(filename, string(content))
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}
	
	fmt.Printf("=== Parsing Results for %s ===\n\n", filename)
	
	// Show documentable entities
	entities := tree.GetDocumentableEntities()
	fmt.Printf("Found %d documentable entities:\n\n", len(entities))
	
	for i, entity := range entities {
		if i >= 10 { // Limit output for testing
			fmt.Printf("... and %d more entities\n", len(entities)-i)
			break
		}
		
		fmt.Printf("%d. %s: %s\n", i+1, entity.Type.String(), entity.FullName)
		fmt.Printf("   Signature: %s\n", entity.Signature)
		fmt.Printf("   Location: Line %d\n", entity.SourceRange.Start.Line)
		
		if entity.HasDoxygenComment() {
			fmt.Printf("   [DOCUMENTED]\n")
		} else {
			fmt.Printf("   [UNDOCUMENTED]\n")
		}
		fmt.Println()
	}
	
	// Test context extraction for first function
	for _, entity := range entities {
		if entity.Type.String() == "function" || entity.Type.String() == "method" {
			fmt.Printf("=== Context for %s ===\n", entity.FullName)
			
			f := formatter.New()
			context := f.ExtractEntityContext(entity, true, true)
			fmt.Println(context)
			
			fmt.Printf("=== Summary for %s ===\n", entity.FullName)
			summary := f.GetEntitySummary(entity)
			fmt.Println(summary)
			break
		}
	}
}
