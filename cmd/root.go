package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "doxyllm-it",
	Short: "A C++ Doxygen comments parser for LLM integration",
	Long: `doxyllm-it is a CLI tool that parses C++ header files and creates
a tree structure of documentable entities (namespaces, classes, functions, etc.)
while preserving original formatting. It enables feeding specific code contexts
to LLMs for generating Doxygen comments.`,
	Version: getVersionString(),
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("doxyllm-it %s\n", getVersionString())
		fmt.Printf("  Version: %s\n", version)
		fmt.Printf("  Commit:  %s\n", commit)
		fmt.Printf("  Date:    %s\n", date)
	},
}

func getVersionString() string {
	if version == "dev" {
		return fmt.Sprintf("%s (%s)", version, commit)
	}
	return version
}

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
	rootCmd.Version = getVersionString()
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
	rootCmd.AddCommand(versionCmd)
}
