package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "doxyllm-it",
	Short: "A C++ Doxygen comments parser for LLM integration",
	Long: `doxyllm-it is a CLI tool that parses C++ header files and creates
a tree structure of documentable entities (namespaces, classes, functions, etc.)
while preserving original formatting. It enables feeding specific code contexts
to LLMs for generating Doxygen comments.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(extractCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(batchUpdateCmd)
}
