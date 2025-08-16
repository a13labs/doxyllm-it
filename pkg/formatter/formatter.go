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
	case ast.EntityRoot:
		for _, child := range entity.Children {
			result.WriteString(f.reconstructEntity(child, depth))
		}
		return result.String()
	case ast.EntityNamespace:
		// Handle namespace specifically to add proper closing
		result.WriteString(fmt.Sprintf("\n%s{\n", indent))
		for _, child := range entity.Children {
			result.WriteString(f.reconstructEntity(child, depth+1))
		}
		result.WriteString(fmt.Sprintf("%s}", indent))

	case ast.EntityClass, ast.EntityStruct:
		if !entity.IsForwardDeclaration {
			result.WriteString("\n")
			result.WriteString(fmt.Sprintf("%s{\n", indent))
			for _, child := range entity.Children {
				result.WriteString(f.reconstructEntity(child, depth+1))
			}
			result.WriteString(fmt.Sprintf("%s};", indent))
		} else {
			result.WriteString(";")
		}
	case ast.EntityCallable:
		// Functions with bodies include the body content, others end with semicolon
		if len(entity.Body) > 0 {
			// Function has a body - include it from the stored original text
			result.WriteString("\n")
			result.WriteString(fmt.Sprintf("%s{\n", indent))
			codeDepth := depth + 1
			for _, line := range entity.Body {
				if !strings.Contains(line, "{") && strings.Contains(line, "}") {
					codeDepth--
				}
				codeIndent := f.getIndent(codeDepth)
				result.WriteString(fmt.Sprintf("%s%s\n", codeIndent, line))
				if strings.Contains(line, "{") && !strings.Contains(line, "}") {
					codeDepth++
				}
			}
			result.WriteString(fmt.Sprintf("%s}\n", indent))
		} else {
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
		result.WriteString(";")
	}
	if entity.LineComment != nil && entity.LineComment.Signature != "" {
		result.WriteString(fmt.Sprintf(" %s", entity.LineComment.Signature))
	}
	result.WriteString("\n")
	return result.String()
}

func (f *Formatter) ExtractEntityContext(entity *ast.Entity, includeParent bool, includeSiblings bool) string {
	var result strings.Builder

	// Include parent context if requested
	if includeParent && entity.Parent != nil {
		parent := entity.Parent
		result.WriteString("// Parent context:\n")
		result.WriteString(f.formatEntitySignature(parent))
		result.WriteString("\n\n")
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

// getIndent returns the indentation string for the given depth
func (f *Formatter) getIndent(depth int) string {
	if f.useSpaces {
		return strings.Repeat(" ", depth*f.indentSize)
	}
	return strings.Repeat("\t", depth)
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
