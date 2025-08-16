package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/cppparser"

	"github.com/spf13/cobra"
)

var parseCmd = &cobra.Command{
	Use:   "parse [file]",
	Short: "Parse a C++ header file and output the AST structure",
	Long: `Parse a C++ header file and create a tree structure of documentable entities.
The output can be in JSON format for further processing or human-readable format.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		// Read the file
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		// Parse the file
		p := cppparser.New()
		tree, err := p.Parse(filename, string(content))
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", filename, err)
		}

		// Get output format
		format, _ := cmd.Flags().GetString("format")
		showAll, _ := cmd.Flags().GetBool("all")

		switch format {
		case "json":
			return outputJSON(tree, showAll)
		default:
			return outputHuman(tree, showAll)
		}
	},
}

func init() {
	parseCmd.Flags().StringP("format", "f", "human", "Output format (human, json)")
	parseCmd.Flags().BoolP("all", "a", false, "Show all entities including undocumented ones")
}

func outputJSON(tree *ast.ScopeTree, showAll bool) error {
	// Create a simplified structure for JSON output
	type JSONEntity struct {
		Type                 string       `json:"type"`
		Name                 string       `json:"name"`
		FullName             string       `json:"fullName"`
		Signature            string       `json:"signature"`
		AccessLevel          string       `json:"accessLevel,omitempty"`
		IsForwardDeclaration bool         `json:"isForwardDeclaration,omitempty"`
		IsTemplate           bool         `json:"isTemplate,omitempty"`
		Children             []JSONEntity `json:"children,omitempty"`
	}

	var convertEntity func(*ast.Entity) JSONEntity
	convertEntity = func(e *ast.Entity) JSONEntity {
		je := JSONEntity{
			Type:                 e.Type.String(),
			Name:                 e.Name,
			FullName:             e.FullName,
			Signature:            e.Signature,
			IsForwardDeclaration: e.IsForwardDeclaration,
			IsTemplate:           e.IsTemplate,
		}

		if e.AccessLevel != ast.AccessUnknown {
			je.AccessLevel = e.AccessLevel.String()
		}

		for _, child := range e.Children {
			je.Children = append(je.Children, convertEntity(child))
		}

		return je
	}

	// Convert only the direct children of root (no duplicates)
	var jsonEntities []JSONEntity
	for _, entity := range tree.Root.Children {
		jsonEntities = append(jsonEntities, convertEntity(entity))
	}

	output := map[string]interface{}{
		"filename": tree.Filename,
		"entities": jsonEntities,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputHuman(tree *ast.ScopeTree, showAll bool) error {
	fmt.Printf("Parsed file: %s\n", tree.Filename)
	fmt.Printf("=====================================\n\n")

	// Print the tree hierarchically starting from root children
	for _, entity := range tree.Root.Children {
		printEntity(entity, 0)
		fmt.Println()
	}

	// Summary
	fmt.Printf("Summary:\n")
	fmt.Printf("--------\n")

	// Get all entities for counting, but from root traversal to avoid duplicates
	allEntities := tree.Root.GetAllEntities()
	// Filter out the root entity itself
	var entities []*ast.Entity
	entities = append(entities, allEntities...)

	fmt.Printf("Total entities: %d\n", len(entities))

	return nil
}

func printEntity(entity *ast.Entity, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	fmt.Printf("%s%s:", indent, entity.Type.String())

	if entity.AccessLevel != ast.AccessUnknown {
		fmt.Printf(" [%s]", entity.AccessLevel.String())
	}

	if entity.IsForwardDeclaration {
		fmt.Printf(" [forward]")
	}

	fmt.Printf("\n%s  Signature: %s\n", indent, entity.Signature)

	// Print children
	for _, child := range entity.Children {
		printEntity(child, depth+1)
	}
}
