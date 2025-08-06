package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/parser"

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
		p := parser.New()
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
	var entities []*ast.Entity

	if showAll {
		entities = tree.Root.GetAllEntities()
	} else {
		entities = tree.GetDocumentableEntities()
	}

	// Create a simplified structure for JSON output
	type JSONEntity struct {
		Type        string       `json:"type"`
		Name        string       `json:"name"`
		FullName    string       `json:"fullName"`
		Signature   string       `json:"signature"`
		AccessLevel string       `json:"accessLevel,omitempty"`
		IsStatic    bool         `json:"isStatic,omitempty"`
		IsVirtual   bool         `json:"isVirtual,omitempty"`
		IsConst     bool         `json:"isConst,omitempty"`
		HasComment  bool         `json:"hasComment"`
		Children    []JSONEntity `json:"children,omitempty"`
		Line        int          `json:"line"`
		Column      int          `json:"column"`
	}

	var convertEntity func(*ast.Entity) JSONEntity
	convertEntity = func(e *ast.Entity) JSONEntity {
		je := JSONEntity{
			Type:       e.Type.String(),
			Name:       e.Name,
			FullName:   e.FullName,
			Signature:  e.Signature,
			IsStatic:   e.IsStatic,
			IsVirtual:  e.IsVirtual,
			IsConst:    e.IsConst,
			HasComment: e.HasDoxygenComment(),
			Line:       e.SourceRange.Start.Line,
			Column:     e.SourceRange.Start.Column,
		}

		if e.AccessLevel != ast.AccessUnknown {
			je.AccessLevel = e.AccessLevel.String()
		}

		for _, child := range e.Children {
			je.Children = append(je.Children, convertEntity(child))
		}

		return je
	}

	var jsonEntities []JSONEntity
	for _, entity := range entities {
		if entity.Type != ast.EntityUnknown {
			jsonEntities = append(jsonEntities, convertEntity(entity))
		}
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

	var entities []*ast.Entity
	if showAll {
		entities = tree.Root.GetAllEntities()
	} else {
		entities = tree.GetDocumentableEntities()
	}

	for _, entity := range entities {
		if entity.Type == ast.EntityUnknown {
			continue
		}

		printEntity(entity, 0)
		fmt.Println()
	}

	// Summary
	fmt.Printf("Summary:\n")
	fmt.Printf("--------\n")
	fmt.Printf("Total entities: %d\n", len(entities))

	// Count by type
	typeCounts := make(map[ast.EntityType]int)
	documentedCount := 0

	for _, entity := range entities {
		if entity.Type != ast.EntityUnknown {
			typeCounts[entity.Type]++
			if entity.HasDoxygenComment() {
				documentedCount++
			}
		}
	}

	for entityType, count := range typeCounts {
		fmt.Printf("%s: %d\n", entityType.String(), count)
	}

	fmt.Printf("Documented: %d (%.1f%%)\n", documentedCount, float64(documentedCount)/float64(len(entities))*100)

	return nil
}

func printEntity(entity *ast.Entity, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	fmt.Printf("%s%s: %s", indent, entity.Type.String(), entity.Name)

	if entity.FullName != entity.Name {
		fmt.Printf(" (%s)", entity.FullName)
	}

	if entity.AccessLevel != ast.AccessUnknown {
		fmt.Printf(" [%s]", entity.AccessLevel.String())
	}

	if entity.IsStatic {
		fmt.Printf(" [static]")
	}
	if entity.IsVirtual {
		fmt.Printf(" [virtual]")
	}
	if entity.IsConst {
		fmt.Printf(" [const]")
	}

	if entity.HasDoxygenComment() {
		fmt.Printf(" [documented]")
	}

	fmt.Printf("\n%s  Signature: %s\n", indent, entity.Signature)
	fmt.Printf("%s  Location: Line %d, Column %d\n", indent, entity.SourceRange.Start.Line, entity.SourceRange.Start.Column)

	if entity.HasDoxygenComment() {
		fmt.Printf("%s  Documentation:\n", indent)
		if entity.Comment.Brief != "" {
			fmt.Printf("%s    Brief: %s\n", indent, entity.Comment.Brief)
		}
		if entity.Comment.Detailed != "" {
			fmt.Printf("%s    Details: %s\n", indent, entity.Comment.Detailed)
		}
	}

	// Print children
	for _, child := range entity.Children {
		printEntity(child, depth+1)
	}
}
