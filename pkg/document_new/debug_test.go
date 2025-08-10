package document_new

import (
	"fmt"
	"testing"
)

func TestDebugDoxygen(t *testing.T) {
	input := `/// @brief Group member function
		/// @ingroup math_functions
		/// @ingroup utility
		/// @author John Doe
		/// @deprecated Use newFunction() instead`

	// Debug the cleaning
	cleaned := cleanCommentContent(input)
	fmt.Printf("Cleaned content:\n%s\n\n", cleaned)

	// Debug individual extractions
	fmt.Printf("Brief: '%s'\n", extractBrief(cleaned))
	fmt.Printf("Groups: %+v\n", extractGroups(cleaned))
	fmt.Printf("Deprecated: '%s'\n", extractDeprecated(cleaned))
	fmt.Printf("Custom tags: %+v\n", extractCustomTags(cleaned))

	// Test parse function
	result := ParseDoxygenComment(input)
	fmt.Printf("\nFull result:\n")
	fmt.Printf("Brief: '%s'\n", result.Brief)
	fmt.Printf("Groups: %+v\n", result.Groups)
	fmt.Printf("Deprecated: '%s'\n", result.Deprecated)
	fmt.Printf("CustomTags: %+v\n", result.CustomTags)
}
