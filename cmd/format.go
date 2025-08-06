package cmd

import (
	"fmt"
	"os"

	"doxyllm-it/pkg/formatter"
	"doxyllm-it/pkg/parser"

	"github.com/spf13/cobra"
)

var formatCmd = &cobra.Command{
	Use:   "format [file]",
	Short: "Format a C++ file with clang-format",
	Long: `Format a C++ file using clang-format while preserving the tree structure.
This command parses the file, reconstructs it, and applies clang-format.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

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

		// Reconstruct the code
		f := formatter.New()
		reconstructed := f.ReconstructCode(tree)

		// Apply clang-format if requested
		useClang, _ := cmd.Flags().GetBool("clang-format")
		if useClang {
			formatted, err := f.FormatWithClang(reconstructed)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: clang-format failed: %v\n", err)
				fmt.Print(reconstructed)
			} else {
				fmt.Print(formatted)
			}
		} else {
			fmt.Print(reconstructed)
		}

		return nil
	},
}

func init() {
	formatCmd.Flags().BoolP("clang-format", "c", false, "Apply clang-format to the output")
}
