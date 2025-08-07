// Package formatter handles code reconstruction and formatting
package formatter

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"doxyllm-it/pkg/ast"
)

// Formatter handles code reconstruction and formatting
type Formatter struct {
	indentSize int
	useSpaces  bool
}

// New creates a new formatter
func New() *Formatter {
	return &Formatter{
		indentSize: 4,
		useSpaces:  true,
	}
}

// ReconstructCode reconstructs the original code from the scope tree
func (f *Formatter) ReconstructCode(tree *ast.ScopeTree) string {
	return f.reconstructEntity(tree.Root, 0)
}

// ReconstructScope reconstructs code for a specific scope/entity
func (f *Formatter) ReconstructScope(entity *ast.Entity) string {
	return f.reconstructEntity(entity, 0)
}

// reconstructEntity recursively reconstructs code for an entity
func (f *Formatter) reconstructEntity(entity *ast.Entity, depth int) string {
	var result strings.Builder

	// If this is the root entity, reconstruct all children
	if entity.Type == ast.EntityUnknown && entity.Name == "" {
		for _, child := range entity.Children {
			result.WriteString(f.reconstructEntity(child, depth))
		}
		return result.String()
	}

	// Add leading whitespace/comments
	if entity.LeadingWS != "" {
		result.WriteString(entity.LeadingWS)
	}

	// Add doxygen comment if present
	if entity.Comment != nil {
		result.WriteString(f.formatDoxygenComment(entity.Comment, depth))
		result.WriteString("\n")
	}

	// For comment entities, only output the comment, skip signature and other handling
	if entity.Type == ast.EntityComment {
		return result.String()
	}

	// Add the entity declaration with proper multi-line indentation
	indent := f.getIndent(depth)
	signature := entity.Signature

	// Handle multi-line signatures by indenting each line properly
	if strings.Contains(signature, "\n") {
		lines := strings.Split(signature, "\n")
		for i, line := range lines {
			if i == 0 {
				result.WriteString(indent + line)
			} else {
				result.WriteString("\n" + indent + line)
			}
		}
	} else {
		result.WriteString(indent + signature)
	}

	// Handle different entity types
	switch entity.Type {
	case ast.EntityNamespace:
		// Handle namespace specifically to add proper closing
		if len(entity.Children) > 0 {
			// Check if signature already contains opening brace
			if !strings.HasSuffix(strings.TrimSpace(entity.Signature), "{") {
				// If signature contains newline (multi-line), put brace on new line
				if strings.Contains(entity.Signature, "\n") {
					result.WriteString("\n" + indent + "{")
				} else {
					// Single line signature, add brace on same line
					result.WriteString(" {")
				}
			}
			result.WriteString("\n")

			// Add children
			for _, child := range entity.Children {
				result.WriteString(f.reconstructEntity(child, depth+1))
			}

			// Add closing brace with namespace comment
			result.WriteString(indent + "} // namespace " + entity.Name)
		} else {
			// Empty namespace, just add opening and closing braces
			if !strings.HasSuffix(strings.TrimSpace(entity.Signature), "{") {
				if strings.Contains(entity.Signature, "\n") {
					result.WriteString("\n" + indent + "{\n" + indent + "} // namespace " + entity.Name)
				} else {
					result.WriteString(" {\n" + indent + "} // namespace " + entity.Name)
				}
			}
		}

	case ast.EntityClass, ast.EntityStruct, ast.EntityEnum:
		// These entities have bodies with children
		if len(entity.Children) > 0 {
			// Check if signature already contains opening brace
			if !strings.HasSuffix(strings.TrimSpace(entity.Signature), "{") {
				result.WriteString(" {")
			}
			result.WriteString("\n")

			// Add children
			for _, child := range entity.Children {
				result.WriteString(f.reconstructEntity(child, depth+1))
			}

			result.WriteString(indent + "}")
		}

		// Add semicolon for class/struct (always required in C++)
		if entity.Type == ast.EntityClass || entity.Type == ast.EntityStruct {
			result.WriteString(";")
		}

	case ast.EntityFunction, ast.EntityMethod, ast.EntityConstructor, ast.EntityDestructor:
		// Functions end with semicolon or have a body
		if !strings.HasSuffix(entity.Signature, ";") && entity.BodyRange == nil {
			result.WriteString(";")
		}

	case ast.EntityPreprocessor:
		// Preprocessor directives don't need semicolons or additional formatting
		// They are output as-is

	case ast.EntityComment:
		// File-level comments are output as their comment content only
		// Remove the newline that would be added by default since comment formatting adds its own

	case ast.EntityAccessSpecifier:
		// Access specifiers don't need semicolons
		// They are output as-is with their colon

	default:
		// Other entities typically end with semicolon
		if !strings.HasSuffix(entity.Signature, ";") {
			result.WriteString(";")
		}
	}

	result.WriteString("\n")

	// Add trailing whitespace/comments
	if entity.TrailingWS != "" {
		result.WriteString(entity.TrailingWS)
	}

	return result.String()
}

// formatDoxygenComment formats a doxygen comment with proper indentation
func (f *Formatter) formatDoxygenComment(comment *ast.DoxygenComment, depth int) string {
	if comment == nil || comment.Raw == "" {
		return ""
	}

	var result strings.Builder
	indent := f.getIndent(depth)

	result.WriteString(indent + "/**\n")

	// Brief description
	if comment.Brief != "" {
		result.WriteString(indent + " * @brief " + comment.Brief + "\n")
	}

	// Detailed description
	if comment.Detailed != "" {
		lines := strings.Split(comment.Detailed, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.WriteString(indent + " * " + strings.TrimSpace(line) + "\n")
			} else {
				result.WriteString(indent + " *\n")
			}
		}
	}

	// Parameters
	if len(comment.Params) > 0 {
		result.WriteString(indent + " *\n")
		for param, desc := range comment.Params {
			result.WriteString(indent + " * @param " + param + " " + desc + "\n")
		}
	}

	// Return value
	if comment.Returns != "" {
		result.WriteString(indent + " * @return " + comment.Returns + "\n")
	}

	// Exceptions
	for _, exception := range comment.Throws {
		result.WriteString(indent + " * @throws " + exception + "\n")
	}

	// Other tags
	if comment.Since != "" {
		result.WriteString(indent + " * @since " + comment.Since + "\n")
	}

	if comment.Deprecated != "" {
		result.WriteString(indent + " * @deprecated " + comment.Deprecated + "\n")
	}

	if comment.Author != "" {
		result.WriteString(indent + " * @author " + comment.Author + "\n")
	}

	if comment.Version != "" {
		result.WriteString(indent + " * @version " + comment.Version + "\n")
	}

	// See also
	for _, see := range comment.See {
		result.WriteString(indent + " * @see " + see + "\n")
	}

	// Ingroup tags
	for _, group := range comment.Ingroup {
		result.WriteString(indent + " * @ingroup " + group + "\n")
	}

	// Group definition tags
	if comment.Defgroup != "" {
		result.WriteString(indent + " * @defgroup " + comment.Defgroup + "\n")
	}

	if comment.Addtogroup != "" {
		result.WriteString(indent + " * @addtogroup " + comment.Addtogroup + "\n")
	}

	// Structural tags
	if comment.File != "" {
		result.WriteString(indent + " * @file " + comment.File + "\n")
	}

	if comment.Namespace != "" {
		result.WriteString(indent + " * @namespace " + comment.Namespace + "\n")
	}

	if comment.Class != "" {
		result.WriteString(indent + " * @class " + comment.Class + "\n")
	}

	// Custom tags
	for tag, value := range comment.CustomTags {
		result.WriteString(indent + " * @" + tag + " " + value + "\n")
	}

	result.WriteString(indent + " */")

	return result.String()
}

// getIndent returns the indentation string for the given depth
func (f *Formatter) getIndent(depth int) string {
	if f.useSpaces {
		return strings.Repeat(" ", depth*f.indentSize)
	}
	return strings.Repeat("\t", depth)
}

// FormatWithClang formats the code using clang-format
func (f *Formatter) FormatWithClang(code string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "doxyllm-*.cpp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write code to temp file
	if _, err := tmpFile.WriteString(code); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Run clang-format
	cmd := exec.Command("clang-format", tmpFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("clang-format failed: %w", err)
	}

	return string(output), nil
}

// UpdateEntityComment updates an entity's doxygen comment
func (f *Formatter) UpdateEntityComment(entity *ast.Entity, newComment *ast.DoxygenComment) {
	entity.Comment = newComment
}

// ExtractEntityContext extracts code context for an entity suitable for LLM input
func (f *Formatter) ExtractEntityContext(entity *ast.Entity, includeParent bool, includeSiblings bool) string {
	var result strings.Builder

	// Include parent context if requested
	if includeParent && entity.Parent != nil {
		parent := entity.Parent
		if parent.Type != ast.EntityUnknown {
			result.WriteString("// Parent context:\n")
			result.WriteString(f.formatEntitySignature(parent))
			result.WriteString("\n\n")
		}
	}

	// Include sibling context if requested
	if includeSiblings && entity.Parent != nil {
		result.WriteString("// Sibling context:\n")
		for _, sibling := range entity.Parent.Children {
			if sibling != entity {
				result.WriteString(f.formatEntitySignature(sibling))
				result.WriteString("\n")
			}
		}
		result.WriteString("\n")
	}

	// Include the entity itself
	result.WriteString("// Target entity:\n")
	result.WriteString(f.ReconstructScope(entity))

	return result.String()
}

// formatEntitySignature formats just the signature of an entity
func (f *Formatter) formatEntitySignature(entity *ast.Entity) string {
	signature := entity.Signature

	// Add type prefix for clarity
	switch entity.Type {
	case ast.EntityNamespace:
		signature = "namespace " + entity.Name + " { /* ... */ }"
	case ast.EntityClass:
		signature = "class " + entity.Name + " { /* ... */ };"
	case ast.EntityStruct:
		signature = "struct " + entity.Name + " { /* ... */ };"
	case ast.EntityEnum:
		signature = "enum " + entity.Name + " { /* ... */ };"
	}

	return signature
}

// GetEntitySummary returns a summary of an entity for LLM context
func (f *Formatter) GetEntitySummary(entity *ast.Entity) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Type: %s\n", entity.Type))
	result.WriteString(fmt.Sprintf("Name: %s\n", entity.Name))
	result.WriteString(fmt.Sprintf("Full Name: %s\n", entity.FullName))
	result.WriteString(fmt.Sprintf("Signature: %s\n", entity.Signature))

	if entity.AccessLevel != ast.AccessUnknown {
		result.WriteString(fmt.Sprintf("Access: %s\n", entity.AccessLevel))
	}

	if entity.IsStatic {
		result.WriteString("Static: true\n")
	}
	if entity.IsVirtual {
		result.WriteString("Virtual: true\n")
	}
	if entity.IsConst {
		result.WriteString("Const: true\n")
	}

	if entity.HasDoxygenComment() {
		result.WriteString("Has Documentation: true\n")
	} else {
		result.WriteString("Has Documentation: false\n")
	}

	if len(entity.Children) > 0 {
		result.WriteString(fmt.Sprintf("Children: %d\n", len(entity.Children)))
	}

	return result.String()
}
