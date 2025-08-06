package main

import (
	"fmt"
	"os"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug_parser <file>")
		os.Exit(1)
	}

	filename := os.Args[1]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	p := parser.New()
	tree, err := p.Parse(filename, string(content))
	if err != nil {
		fmt.Printf("Error parsing: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== PARSER DEBUG OUTPUT ===\n")
	fmt.Printf("Total entities found: %d\n\n", len(getAllEntities(tree.Root)))

	printEntities(tree.Root, 0)
}

func getAllEntities(entity *ast.Entity) []*ast.Entity {
	entities := []*ast.Entity{entity}
	for _, child := range entity.Children {
		entities = append(entities, getAllEntities(child)...)
	}
	return entities
}

func printEntities(entity *ast.Entity, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	if entity.Type.String() != "unknown" {
		fmt.Printf("%s%s: %s (line %d)\n", indent, entity.Type, entity.Name, entity.SourceRange.Start.Line)
		if entity.Signature != "" {
			fmt.Printf("%s  Signature: %s\n", indent, entity.Signature)
		}
	}

	for _, child := range entity.Children {
		printEntities(child, depth+1)
	}
}
