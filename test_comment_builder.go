package main

import (
	"doxyllm-it/pkg/llm"
	"fmt"
)

func main() {
	builder := llm.NewCommentBuilder()

	// Simulate an LLM response
	response := &llm.CommentResponse{
		Description: "A StringMap is a type alias for the standard library's std::map template class.",
	}

	// Test structured comment generation
	comment := builder.BuildStructuredComment(
		response,
		"StringMap",
		"using",
		nil,
		"template <typename T> using StringMap = std::map<std::string, T>",
	)

	fmt.Println("Generated comment:")
	fmt.Println(comment)
}
