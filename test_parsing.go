package main

import (
	"doxyllm-it/pkg/parser"
	"fmt"
)

func main() {
	// Test parsing a structured comment with @tparam
	structuredComment := `/**
 * @brief A StringMap is a type alias for the standard library's std::map template class.
 * @tparam T Template parameter
 */`

	fmt.Println("Input structured comment:")
	fmt.Println(structuredComment)

	// Parse it
	comment := parser.ParseDoxygenComment(structuredComment)

	fmt.Println("\nParsed comment:")
	fmt.Printf("Brief: %s\n", comment.Brief)
	fmt.Printf("TParams: %v\n", comment.TParams)
	fmt.Printf("Raw: %s\n", comment.Raw)
}
