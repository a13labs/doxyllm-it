package main

import (
	"doxyllm-it/pkg/llm"
	"fmt"
)

func main() {
	builder := llm.NewCommentBuilder()

	// Test the exact context from our template alias
	context := "template <typename Key, typename Value> using Dictionary =  std::map<Key, Value>;"
	entityType := "using"
	entityName := "Dictionary"

	// Check detection methods
	fmt.Printf("isTemplateType(%s): %v\n", entityType, builder.IsTemplateType(entityType))
	fmt.Printf("hasTemplateParameters: %v\n", builder.HasTemplateParameters(context))

	// Extract template parameters
	tparams := builder.ExtractTemplateParametersFromContext(context)
	fmt.Printf("Template parameters: %v\n", tparams)

	// Build a comment
	response := &llm.CommentResponse{
		Description: "Template alias for std::map.",
	}

	comment := builder.BuildStructuredComment(response, entityName, entityType, nil, context)
	fmt.Printf("\nGenerated comment:\n%s\n", comment)
}
