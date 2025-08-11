package formatter_new

import (
	"fmt"
	"os"
	"sort"
	"strings"

	ast "doxyllm-it/pkg/ast_new"
)

// Formatter reconstructs source code from AST trees
type Formatter struct {
	// Configuration options
	preserveWhitespace bool
	indentSize         int
	useSpaces          bool
}

// NewFormatter creates a new formatter with default settings
func NewFormatter() *Formatter {
	return &Formatter{
		preserveWhitespace: true,
		indentSize:         4,
		useSpaces:          true,
	}
}

// FormatOptions configures the formatter behavior
type FormatOptions struct {
	PreserveWhitespace bool
	IndentSize         int
	UseSpaces          bool
}

// NewFormatterWithOptions creates a formatter with custom options
func NewFormatterWithOptions(opts FormatOptions) *Formatter {
	return &Formatter{
		preserveWhitespace: opts.PreserveWhitespace,
		indentSize:         opts.IndentSize,
		useSpaces:          opts.UseSpaces,
	}
}

// FormatToString reconstructs the source code from an AST tree and returns it as a string
func (f *Formatter) FormatToString(tree *ast.ScopeTree) (string, error) {
	if tree == nil || tree.Root == nil {
		return "", fmt.Errorf("invalid AST tree")
	}

	var result strings.Builder
	err := f.formatEntity(tree.Root, &result, 0)
	if err != nil {
		return "", fmt.Errorf("failed to format AST: %w", err)
	}

	return result.String(), nil
}

// FormatToFile reconstructs the source code from an AST tree and saves it to a file
func (f *Formatter) FormatToFile(tree *ast.ScopeTree, filename string) error {
	content, err := f.FormatToString(tree)
	if err != nil {
		return fmt.Errorf("failed to format content: %w", err)
	}

	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

// formatEntity recursively formats an entity and its children
func (f *Formatter) formatEntity(entity *ast.Entity, result *strings.Builder, depth int) error {
	if entity == nil {
		return nil
	}

	// Handle root entity specially - just format its children
	if entity.Type == ast.EntityRoot {
		return f.formatChildren(entity, result, depth)
	}

	// Add leading whitespace if preserved
	if f.preserveWhitespace && entity.LeadingWS != "" {
		result.WriteString(entity.LeadingWS)
	}

	// For comment entities, just output the original text
	if entity.Type == ast.EntityComment {
		if entity.OriginalText != "" {
			result.WriteString(entity.OriginalText)
		}
		return nil
	}

	// Build the complete signature with modifiers
	signature := f.buildCompleteSignature(entity)
	if signature != "" {
		// Add proper indentation
		indent := f.getIndent(depth)

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
	}

	// Handle different entity types
	switch entity.Type {
	case ast.EntityNamespace:
		f.formatNamespaceBody(entity, result, depth)
	case ast.EntityClass, ast.EntityStruct:
		f.formatClassBody(entity, result, depth)
	case ast.EntityEnum:
		f.formatEnumBody(entity, result, depth)
	case ast.EntityFunction, ast.EntityMethod, ast.EntityConstructor, ast.EntityDestructor:
		f.formatFunctionBody(entity, result, depth)
	default:
		// For other types, just add newline if signature was written
		if signature != "" {
			result.WriteString("\n")
		}
	}

	// Add trailing whitespace if preserved
	if f.preserveWhitespace && entity.TrailingWS != "" {
		result.WriteString(entity.TrailingWS)
	}

	return nil
}

// formatChildren formats all children of an entity in source order
func (f *Formatter) formatChildren(entity *ast.Entity, result *strings.Builder, depth int) error {
	// Sort children by their source position to maintain original order
	children := make([]*ast.Entity, len(entity.Children))
	copy(children, entity.Children)

	sort.Slice(children, func(i, j int) bool {
		// Comments inserted by document layer may have empty source ranges
		// Place them at the correct position based on their position in the children array
		if children[i].SourceRange.Start.Line == 0 && children[j].SourceRange.Start.Line != 0 {
			// Find the original index to maintain insertion order
			for idx, child := range entity.Children {
				if child == children[i] {
					// Check if this comment should come before the next entity
					if idx+1 < len(entity.Children) && entity.Children[idx+1] == children[j] {
						return true // comment should come before the entity
					}
				}
			}
		}
		if children[j].SourceRange.Start.Line == 0 && children[i].SourceRange.Start.Line != 0 {
			// Similar logic for the reverse case
			for idx, child := range entity.Children {
				if child == children[j] {
					if idx+1 < len(entity.Children) && entity.Children[idx+1] == children[i] {
						return false // entity should come after the comment
					}
				}
			}
		}

		// Both have valid source positions - sort by line and column
		if children[i].SourceRange.Start.Line != children[j].SourceRange.Start.Line {
			return children[i].SourceRange.Start.Line < children[j].SourceRange.Start.Line
		}
		return children[i].SourceRange.Start.Column < children[j].SourceRange.Start.Column
	})

	// Format each child
	for _, child := range children {
		err := f.formatEntity(child, result, depth+1)
		if err != nil {
			return err
		}
	}

	return nil
}

// formatComment formats a comment entity
func (f *Formatter) formatComment(entity *ast.Entity, result *strings.Builder) error {
	if entity.OriginalText != "" {
		result.WriteString(entity.OriginalText)
	}
	return nil
}

// formatNamespace formats a namespace entity
func (f *Formatter) formatNamespace(entity *ast.Entity, result *strings.Builder, depth int) error {
	// Write the namespace declaration
	if entity.OriginalText != "" {
		result.WriteString(entity.OriginalText)
	} else {
		// Reconstruct from entity data
		result.WriteString("namespace ")
		result.WriteString(entity.Name)
		result.WriteString(" {\n")

		// Format children
		err := f.formatChildren(entity, result, depth)
		if err != nil {
			return err
		}

		result.WriteString("}\n")
	}
	return nil
}

// formatClass formats a class or struct entity
func (f *Formatter) formatClass(entity *ast.Entity, result *strings.Builder, depth int) error {
	if entity.OriginalText != "" {
		result.WriteString(entity.OriginalText)
	}
	return nil
}

// formatFunction formats a function or method entity
func (f *Formatter) formatFunction(entity *ast.Entity, result *strings.Builder, depth int) error {
	if entity.OriginalText != "" {
		result.WriteString(entity.OriginalText)
	}
	return nil
}

// formatVariable formats a variable or field entity
func (f *Formatter) formatVariable(entity *ast.Entity, result *strings.Builder, depth int) error {
	if entity.OriginalText != "" {
		result.WriteString(entity.OriginalText)
	}
	return nil
}

// formatEnum formats an enum entity
func (f *Formatter) formatEnum(entity *ast.Entity, result *strings.Builder, depth int) error {
	if entity.OriginalText != "" {
		result.WriteString(entity.OriginalText)
	}
	return nil
}

// formatTypedef formats a typedef or using entity
func (f *Formatter) formatTypedef(entity *ast.Entity, result *strings.Builder, depth int) error {
	if entity.OriginalText != "" {
		result.WriteString(entity.OriginalText)
	}
	return nil
}

// buildCompleteSignature constructs the declaration from entity properties
func (f *Formatter) buildCompleteSignature(entity *ast.Entity) string {
	var parts []string

	// Add modifiers based on entity properties
	if entity.IsExtern {
		parts = append(parts, "extern")
	}
	if entity.IsStatic {
		parts = append(parts, "static")
	}
	if entity.IsInline {
		parts = append(parts, "inline")
	}
	if entity.IsVirtual {
		parts = append(parts, "virtual")
	}
	if entity.IsConstexpr {
		parts = append(parts, "constexpr")
	} else if entity.IsConst && (entity.Type == ast.EntityVariable || entity.Type == ast.EntityField) {
		parts = append(parts, "const")
	}

	// Construct the declaration based on entity type
	switch entity.Type {
	case ast.EntityNamespace:
		parts = append(parts, "namespace", entity.Name)

	case ast.EntityClass:
		parts = append(parts, "class", entity.Name)

	case ast.EntityStruct:
		parts = append(parts, "struct", entity.Name)

	case ast.EntityEnum:
		parts = append(parts, "enum", entity.Name)

	case ast.EntityAccessSpecifier:
		return entity.Name + ":"

	case ast.EntityComment:
		// Comments should use their original text
		return ""

	default:
		// For other types (fields, functions, etc.), we need more information
		// For now, fall back to the signature if available
		if entity.Signature != "" {
			parts = append(parts, entity.Signature)
		} else {
			// Try to construct basic declaration
			if entity.Name != "" {
				parts = append(parts, entity.Name)
			}
		}
	}

	return strings.Join(parts, " ")
}

// getIndent returns the proper indentation string for the given depth
func (f *Formatter) getIndent(depth int) string {
	if depth <= 0 {
		return ""
	}

	indentChar := " "
	if !f.useSpaces {
		indentChar = "\t"
		return strings.Repeat(indentChar, depth)
	}

	return strings.Repeat(indentChar, depth*f.indentSize)
}

// formatNamespaceBody formats the body of a namespace
func (f *Formatter) formatNamespaceBody(entity *ast.Entity, result *strings.Builder, depth int) {
	indent := f.getIndent(depth)

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
			f.formatEntity(child, result, depth+1)
		}

		// Add closing brace with namespace comment
		result.WriteString(indent + "} // namespace " + entity.Name + "\n")
	} else {
		// Empty namespace, just add opening and closing braces
		if !strings.HasSuffix(strings.TrimSpace(entity.Signature), "{") {
			if strings.Contains(entity.Signature, "\n") {
				result.WriteString("\n" + indent + "{\n" + indent + "} // namespace " + entity.Name + "\n")
			} else {
				result.WriteString(" {\n" + indent + "} // namespace " + entity.Name + "\n")
			}
		} else {
			result.WriteString("\n" + indent + "} // namespace " + entity.Name + "\n")
		}
	}
}

// formatClassBody formats the body of a class or struct
func (f *Formatter) formatClassBody(entity *ast.Entity, result *strings.Builder, depth int) {
	indent := f.getIndent(depth)

	// Always emit body braces if the signature represents a definition (ends with '{')
	trimmedSig := strings.TrimSpace(entity.Signature)
	if strings.HasSuffix(trimmedSig, "{") || len(entity.Children) > 0 {
		// If signature does not already include opening brace, add it inline
		if !strings.HasSuffix(trimmedSig, "{") {
			result.WriteString(" {")
		}
		result.WriteString("\n")

		// Emit children (if any)
		for _, child := range entity.Children {
			f.formatEntity(child, result, depth+1)
		}

		// Closing brace
		result.WriteString(indent + "}")
	}

	// Add semicolon for class/struct (always required in C++)
	if entity.Type == ast.EntityClass || entity.Type == ast.EntityStruct {
		result.WriteString(";")
	}

	result.WriteString("\n")
}

// formatEnumBody formats the body of an enum
func (f *Formatter) formatEnumBody(entity *ast.Entity, result *strings.Builder, depth int) {
	indent := f.getIndent(depth)

	// Handle enum body similar to class
	trimmedSig := strings.TrimSpace(entity.Signature)
	if strings.HasSuffix(trimmedSig, "{") || len(entity.Children) > 0 {
		if !strings.HasSuffix(trimmedSig, "{") {
			result.WriteString(" {")
		}
		result.WriteString("\n")

		// Emit children (enum values)
		for _, child := range entity.Children {
			f.formatEntity(child, result, depth+1)
		}

		result.WriteString(indent + "}")
	}

	// Enums need semicolon
	result.WriteString(";\n")
}

// formatFunctionBody formats a function declaration
func (f *Formatter) formatFunctionBody(entity *ast.Entity, result *strings.Builder, depth int) {
	// Functions typically just end with semicolon or have body
	// For now, just add newline (function body parsing would be more complex)
	result.WriteString("\n")
}
