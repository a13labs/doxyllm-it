package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	document "doxyllm-it/pkg/document_new"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect C++ header files and entities",
	Long: `Inspect C++ header files to explore their structure and get detailed
information about specific entities like namespaces, classes, functions, etc.`,
}

var listCmd = &cobra.Command{
	Use:   "list [file]",
	Short: "List all instruction paths in a C++ header file",
	Long: `List all documentable entities (instructions) in a C++ header file.
This includes namespaces, classes, functions, variables, etc. but excludes
comments and access specifiers.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		// Load and parse the document
		doc, err := document.NewFromFile(filename)
		if err != nil {
			return fmt.Errorf("failed to load document: %w", err)
		}

		// Get all instruction paths
		instructions := doc.ListInstructions()

		if len(instructions) == 0 {
			fmt.Println("No documentable entities found.")
			return nil
		}

		// Print each instruction path
		for _, path := range instructions {
			fmt.Println(path)
		}

		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get [file] [path]",
	Short: "Get detailed information about a specific entity",
	Long: `Get detailed information about a specific entity in YAML format.
The path should be the full qualified name (e.g., "MyNamespace::MyClass::myMethod").
Use the list command to see available paths.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		entityPath := args[1]

		// Load and parse the document
		doc, err := document.NewFromFile(filename)
		if err != nil {
			return fmt.Errorf("failed to load document: %w", err)
		}

		// Get the specific instruction
		entity := doc.GetInstruction(entityPath)
		if entity == nil {
			return fmt.Errorf("entity not found or not documentable: %s", entityPath)
		}

		// Create a simplified structure for YAML output
		entityInfo := map[string]interface{}{
			"type":        entity.Type.String(),
			"name":        entity.Name,
			"full_name":   entity.FullName,
			"signature":   entity.Signature,
			"access":      entity.AccessLevel.String(),
			"is_template": entity.IsTemplate,
			"is_forward":  entity.IsForwardDeclaration,
			"scope":       entity.GetScope(),
			"path":        strings.Join(entity.GetPath(), "::"),
		}

		// Add defines if present
		if len(entity.Defines) > 0 {
			entityInfo["defines"] = entity.Defines
		}
		// Add body if present (full, not truncated)
		if entity.Body != "" {
			entityInfo["body"] = entity.Body
		}

		// Add children count if any
		if len(entity.Children) > 0 {
			childPaths := make([]string, len(entity.Children))
			for i, child := range entity.Children {
				childPaths[i] = child.GetFullPath()
			}
			entityInfo["children"] = childPaths
		}

		// Convert to YAML and print
		yamlData, err := yaml.Marshal(entityInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal entity to YAML: %w", err)
		}

		fmt.Print(string(yamlData))
		return nil
	},
}

func init() {
	// Add subcommands to inspect
	inspectCmd.AddCommand(listCmd)
	inspectCmd.AddCommand(getCmd)
}
