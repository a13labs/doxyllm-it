package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/formatter"
	"doxyllm-it/pkg/parser"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [file] [entity-path] [comment-file]",
	Short: "Update a C++ file with new Doxygen comments",
	Long: `Update a C++ header file by inserting or replacing Doxygen comments for specific entities.
The comment can be provided via a file or stdin. The tool will parse the file, locate the entity,
insert the new comment, and output the updated file.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		entityPath := args[1]

		var commentContent string
		var err error

		// Get comment content from file or stdin
		if len(args) >= 3 {
			// Read from comment file
			commentFile := args[2]
			contentBytes, err := os.ReadFile(commentFile)
			if err != nil {
				return fmt.Errorf("failed to read comment file %s: %w", commentFile, err)
			}
			commentContent = string(contentBytes)
		} else {
			// Read from stdin
			stdinContent, err := readFromStdin()
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			commentContent = stdinContent
		}

		// Parse the original file
		originalContent, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		p := parser.New()
		tree, err := p.Parse(filename, string(originalContent))
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", filename, err)
		}

		// Find the target entity
		entity := tree.FindEntity(entityPath)
		if entity == nil {
			return fmt.Errorf("entity not found: %s", entityPath)
		}

		// Parse the new comment
		doxygenComment := parser.ParseDoxygenComment(commentContent)
		if doxygenComment == nil {
			return fmt.Errorf("invalid doxygen comment format")
		}

		// Update the entity with new comment
		entity.Comment = doxygenComment

		// Get output options
		inPlace, _ := cmd.Flags().GetBool("in-place")
		outputFile, _ := cmd.Flags().GetString("output")
		useClangFormat, _ := cmd.Flags().GetBool("format")
		backup, _ := cmd.Flags().GetBool("backup")

		// Generate updated content
		updatedContent, err := generateUpdatedFile(tree, entity, commentContent)
		if err != nil {
			return fmt.Errorf("failed to generate updated file: %w", err)
		}

		// Apply clang-format if requested
		if useClangFormat {
			f := formatter.New()
			formatted, err := f.FormatWithClang(updatedContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: clang-format failed: %v\n", err)
			} else {
				updatedContent = formatted
			}
		}

		// Output the result
		if inPlace {
			return writeInPlace(filename, updatedContent, backup)
		} else if outputFile != "" {
			return writeToFile(outputFile, updatedContent)
		} else {
			fmt.Print(updatedContent)
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().BoolP("in-place", "i", false, "Update the file in place")
	updateCmd.Flags().StringP("output", "o", "", "Write output to specific file")
	updateCmd.Flags().BoolP("format", "f", false, "Apply clang-format to the result")
	updateCmd.Flags().BoolP("backup", "b", false, "Create backup when updating in place")
}

// generateUpdatedFile creates the updated file content with the new comment
func generateUpdatedFile(tree *ast.ScopeTree, entity *ast.Entity, newComment string) (string, error) {
	lines := strings.Split(tree.Content, "\n")

	// Find the line where the entity starts
	entityLine := entity.SourceRange.Start.Line - 1 // Convert to 0-based

	// Look for existing doxygen comment before the entity
	commentStartLine := entityLine
	for i := entityLine - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])

		// Stop if we hit a non-comment, non-empty line
		if line != "" && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "*") &&
			!strings.HasPrefix(line, "/*") && !strings.HasSuffix(line, "*/") {
			break
		}

		// Check if this is a doxygen comment
		if strings.HasPrefix(line, "/**") || strings.HasPrefix(line, "///") || strings.HasPrefix(line, "//!") {
			commentStartLine = i
			break
		}

		// If we find the start of a comment block, include it
		if strings.HasPrefix(line, "/**") {
			commentStartLine = i
			break
		}
	}

	// Find the end of existing comment if any
	if commentStartLine < entityLine {
		for i := commentStartLine; i < entityLine; i++ {
			line := strings.TrimSpace(lines[i])
			if strings.Contains(line, "*/") {
				break
			}
		}
	}

	// Prepare the new comment with proper indentation
	entityIndent := getLineIndentation(lines[entityLine])
	formattedComment := formatCommentForInsertion(newComment, entityIndent)

	// Build the new file content
	var result strings.Builder

	// Write lines before the comment
	for i := 0; i < commentStartLine; i++ {
		result.WriteString(lines[i])
		result.WriteString("\n")
	}

	// Write the new comment
	result.WriteString(formattedComment)
	result.WriteString("\n")

	// Write lines from the entity onward (skip old comment)
	for i := entityLine; i < len(lines); i++ {
		result.WriteString(lines[i])
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// getLineIndentation returns the indentation (spaces/tabs) of a line
func getLineIndentation(line string) string {
	indent := ""
	for _, char := range line {
		if char == ' ' || char == '\t' {
			indent += string(char)
		} else {
			break
		}
	}
	return indent
}

// formatCommentForInsertion formats a comment with proper indentation
func formatCommentForInsertion(comment, indent string) string {
	comment = strings.TrimSpace(comment)

	// If it's already a properly formatted doxygen comment, use as is
	if strings.HasPrefix(comment, "/**") || strings.HasPrefix(comment, "///") {
		lines := strings.Split(comment, "\n")
		var result strings.Builder

		for i, line := range lines {
			if i > 0 {
				result.WriteString("\n")
			}
			result.WriteString(indent + line)
		}

		return result.String()
	}

	// If it's plain text, format it as a doxygen comment
	lines := strings.Split(comment, "\n")
	var result strings.Builder

	result.WriteString(indent + "/**")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result.WriteString("\n" + indent + " * " + line)
		} else {
			result.WriteString("\n" + indent + " *")
		}
	}

	result.WriteString("\n" + indent + " */")

	return result.String()
}

// readFromStdin reads content from standard input
func readFromStdin() (string, error) {
	var content strings.Builder
	buffer := make([]byte, 1024)

	for {
		n, err := os.Stdin.Read(buffer)
		if n > 0 {
			content.Write(buffer[:n])
		}
		if err != nil {
			break
		}
	}

	return content.String(), nil
}

// writeInPlace writes content to a file, optionally creating a backup
func writeInPlace(filename, content string, backup bool) error {
	if backup {
		backupFile := filename + ".bak"
		originalContent, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read original file for backup: %w", err)
		}

		if err := os.WriteFile(backupFile, originalContent, 0644); err != nil {
			return fmt.Errorf("failed to create backup file: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Backup created: %s\n", backupFile)
	}

	return os.WriteFile(filename, []byte(content), 0644)
}

// writeToFile writes content to a specific file
func writeToFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// BatchUpdateCommand handles batch updates from JSON input
type BatchUpdateInput struct {
	SourceFile string `json:"sourceFile"`
	Updates    []struct {
		EntityPath string `json:"entityPath"`
		Comment    string `json:"comment"`
	} `json:"updates"`
}

var batchUpdateCmd = &cobra.Command{
	Use:   "batch-update [json-file]",
	Short: "Update multiple entities with Doxygen comments from JSON input",
	Long: `Update multiple entities in a C++ header file with Doxygen comments.
The input should be a JSON file with the following structure:
{
  "sourceFile": "path/to/header.hpp",
  "updates": [
    {
      "entityPath": "namespace::class::method",
      "comment": "/**\n * Brief description\n */"
    }
  ]
}`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonFile := args[0]

		// Read and parse JSON input
		jsonContent, err := os.ReadFile(jsonFile)
		if err != nil {
			return fmt.Errorf("failed to read JSON file: %w", err)
		}

		var input BatchUpdateInput
		if err := json.Unmarshal(jsonContent, &input); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Read and parse the source file
		sourceContent, err := os.ReadFile(input.SourceFile)
		if err != nil {
			return fmt.Errorf("failed to read source file: %w", err)
		}

		p := parser.New()
		tree, err := p.Parse(input.SourceFile, string(sourceContent))
		if err != nil {
			return fmt.Errorf("failed to parse source file: %w", err)
		}

		// Apply all updates
		updatedContent := string(sourceContent)
		successCount := 0

		for _, update := range input.Updates {
			entity := tree.FindEntity(update.EntityPath)
			if entity == nil {
				fmt.Fprintf(os.Stderr, "Warning: entity not found: %s\n", update.EntityPath)
				continue
			}

			// For batch updates, we need to be more careful about line number changes
			// This is a simplified approach - for production use, you'd want more sophisticated handling
			newContent, err := generateUpdatedFile(tree, entity, update.Comment)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update %s: %v\n", update.EntityPath, err)
				continue
			}

			updatedContent = newContent
			successCount++

			// Re-parse the updated content for next iteration
			tree, err = p.Parse(input.SourceFile, updatedContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to re-parse after updating %s: %v\n", update.EntityPath, err)
				break
			}
		}

		// Output options
		inPlace, _ := cmd.Flags().GetBool("in-place")
		outputFile, _ := cmd.Flags().GetString("output")
		useClangFormat, _ := cmd.Flags().GetBool("format")
		backup, _ := cmd.Flags().GetBool("backup")

		// Apply clang-format if requested
		if useClangFormat {
			f := formatter.New()
			formatted, err := f.FormatWithClang(updatedContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: clang-format failed: %v\n", err)
			} else {
				updatedContent = formatted
			}
		}

		// Output the result
		if inPlace {
			err = writeInPlace(input.SourceFile, updatedContent, backup)
		} else if outputFile != "" {
			err = writeToFile(outputFile, updatedContent)
		} else {
			fmt.Print(updatedContent)
		}

		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Successfully updated %d out of %d entities\n", successCount, len(input.Updates))
		return nil
	},
}

func init() {
	batchUpdateCmd.Flags().BoolP("in-place", "i", false, "Update the file in place")
	batchUpdateCmd.Flags().StringP("output", "o", "", "Write output to specific file")
	batchUpdateCmd.Flags().BoolP("format", "f", false, "Apply clang-format to the result")
	batchUpdateCmd.Flags().BoolP("backup", "b", false, "Create backup when updating in place")
}
