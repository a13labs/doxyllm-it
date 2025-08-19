package cmd

import (
	"fmt"
	"os"

	"doxyllm-it/pkg/formatter"
	"doxyllm-it/pkg/parser"

	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract [file] [entity-path]",
	Short: "Extract code context for a specific entity",
	Long: `Extract code context for a specific entity that can be fed to an LLM.
The entity-path should be in the format namespace::class::method or similar.
Use :: for global scope entities.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		entityPath := args[1]

		// Read and parse the file
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		p := parser.New()
		tree, err := p.Parse(filename, string(content))
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", filename, err)
		}

		// Find the entity
		entity := tree.FindEntity(entityPath)
		if entity == nil {
			return fmt.Errorf("entity not found: %s", entityPath)
		}

		// Get options
		scopeOnly, _ := cmd.Flags().GetBool("scope")

		// Extract context
		f := formatter.New()

		var output string
		if scopeOnly {
			output = f.ReconstructScope(entity, -1, true)
		} else {
			output = f.ExtractEntityContext(entity)
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	extractCmd.Flags().BoolP("scope", "", false, "Extract only the entity scope (no context)")
}
