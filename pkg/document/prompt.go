// Package document provides a high-level service for processing documentation
// requests using the document abstraction
package document

const defaultOllamaPromptTemplate = `You are a C++ documentation expert. Generate a Doxygen comment for the specific entity requested.

CRITICAL INSTRUCTIONS:
- Generate ONLY the Doxygen comment, no code is allowed
- You must include Doxygen tags (@brief, @param, @return, etc.)
- You must include comment markers (/** */) - the system requires this format
- Document ONLY the target entity.
- Focus on describing the purpose, behavior, and usage
- For functions: Describe what it does, not parameters/return (those will be handled separately)
- For classes: Describe the class responsibility and main purpose
- For namespaces: Describe the purpose and scope
- For templates: Describe the template parameters and their usage

%s

%s

ONLY DOXYGEN TAGS:
 - @brief
 - @param
 - @tparam
 - @return

Generate focused doxygen content for this entity (NO SOURCE CODE).`

const fieldPromptTemplate = `You are a C++ documentation expert. Generate a very concise description for a struct/class field.

CRITICAL REQUIREMENTS:
- Generate ONLY a brief description (ONE sentence, under 60 characters)
- Do NOT include Doxygen tags, comment markers, or < symbols
- Start with "The" (e.g., "The x-coordinate of the point")
- Be specific and concise
- Examples of good responses:
  * "The width of the rectangle"
  * "The player's health points"  
  * "The file handle for reading"

%s

%s

Generate ONE concise sentence (under 60 characters, no < symbols).`

// getPromptTemplate returns the appropriate prompt template based on entity type
func (p *DocumentationService) getPromptTemplate(entityType string) string {
	switch entityType {
	case "name":
		return fieldPromptTemplate
	default:
		return defaultOllamaPromptTemplate
	}
}
