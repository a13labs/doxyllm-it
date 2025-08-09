package main

import (
	"doxyllm-it/pkg/llm"
	"fmt"
)

func main() {
	// Test the cleaning method directly
	sampleResponse := `@brief A StringMap is a type alias for the standard library's std::map template class.
	It maps keys of type std::string to values of any other type T.`

	// Use reflection to call the private method for testing
	// Since we can't access private methods directly, let's test the logic
	// by simulating what should happen

	description := sampleResponse

	// Simulate the cleaning logic
	if description[0:6] == "@brief" {
		// Extract content after @brief
		description = description[7:] // Remove "@brief "
	}

	fmt.Println("Original response:")
	fmt.Println(sampleResponse)
	fmt.Println("\nCleaned description:")
	fmt.Println(description)

	// Now test comment builder
	builder := llm.NewCommentBuilder()
	response := &llm.CommentResponse{
		Description: description,
	}

	comment := builder.BuildStructuredComment(
		response,
		"StringMap",
		"using",
		nil,
		"template <typename T> using StringMap = std::map<std::string, T>",
	)

	fmt.Println("\nGenerated structured comment:")
	fmt.Println(comment)
}
