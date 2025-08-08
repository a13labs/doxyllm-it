package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"doxyllm-it/pkg/llm"
	"doxyllm-it/pkg/parser"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [flags] <directory>",
	Short: "Initialize a .doxyllm.yaml.yaml configuration file by analyzing the codebase",
	Long: `Initialize a .doxyllm.yaml.yaml configuration file by automatically analyzing the codebase structure.

This command uses an LLM to analyze each C++ file in the specified directory and generate:
1. Individual file summaries describing the purpose and contents of each file
2. A global project summary based on all file summaries
3. A properly formatted .doxyllm.yaml.yaml YAML configuration file

The generated configuration provides contextual information that helps the LLM generate
better, more accurate documentation when running the 'ollama' command.

Examples:
  # Initialize .doxyllm.yaml.yaml for current directory
  doxyllm-it init .

  # Initialize with custom model
  doxyllm-it init --model deepseek-coder:6.7b src/

  # Initialize with custom Ollama URL
  doxyllm-it init --url http://remote:11434 --model codellama:13b .`,
	Args: cobra.ExactArgs(1),
	Run:  runInit,
}

var (
	initOllamaURL   string
	initOllamaModel string
	initTemperature float64
	initTopP        float64
	initNumCtx      int
	initTimeout     int
	overwrite       bool
)

func init() {
	rootCmd.AddCommand(initCmd)

	// Ollama configuration flags
	initCmd.Flags().StringVarP(&initOllamaURL, "url", "u", getEnvOrDefault("OLLAMA_URL", "http://10.19.4.106:11434/api/generate"), "Ollama API URL")
	initCmd.Flags().StringVarP(&initOllamaModel, "model", "m", getEnvOrDefault("MODEL_NAME", "codellama:13b"), "Ollama model name")
	initCmd.Flags().Float64Var(&initTemperature, "temperature", 0.1, "LLM temperature (0.0-1.0)")
	initCmd.Flags().Float64Var(&initTopP, "top-p", 0.9, "LLM top-p value (0.0-1.0)")
	initCmd.Flags().IntVar(&initNumCtx, "context", 4096, "Context window size")
	initCmd.Flags().IntVar(&initTimeout, "timeout", 120, "Request timeout in seconds")

	// Processing flags
	initCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing .doxyllm.yaml.yaml file if it exists")
}

func runInit(cmd *cobra.Command, args []string) {
	targetDir := args[0]

	// Create LLM provider using factory
	provider, err := llm.NewProvider(&llm.Config{
		Provider:    "ollama",
		URL:         initOllamaURL,
		Model:       initOllamaModel,
		Temperature: initTemperature,
		TopP:        initTopP,
		NumCtx:      initNumCtx,
		Timeout:     time.Duration(initTimeout) * time.Second,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create Ollama provider: %v\n", err)
		os.Exit(1)
	}

	// Test connectivity
	ctx := context.Background()
	if err := provider.TestConnection(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Cannot connect to Ollama: %v\n", err)
		os.Exit(1)
	}

	modelInfo := provider.GetModelInfo()
	fmt.Printf("🤖 Connected to %s\n", modelInfo.Provider)
	fmt.Printf("📚 Using model: %s\n", modelInfo.Name)

	// Check if .doxyllm.yaml.yaml already exists
	doxyllmPath := filepath.Join(targetDir, ".doxyllm.yaml.yaml")
	if _, err := os.Stat(doxyllmPath); err == nil && !overwrite {
		fmt.Printf("❌ .doxyllm.yaml.yaml file already exists at: %s\n", doxyllmPath)
		fmt.Println("   Use --overwrite flag to replace it")
		os.Exit(1)
	}

	// Find C++ files
	files, err := findCppFilesForInit(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("❌ No C++ header files found")
		os.Exit(1)
	}

	fmt.Printf("📂 Found %d C++ header files to analyze\n", len(files))

	// Generate file summaries
	fmt.Println("\n📝 Generating file summaries...")
	fileSummaries := make(map[string]string)

	for i, file := range files {
		relPath, _ := filepath.Rel(targetDir, file)
		fmt.Printf("  📄 (%d/%d) Analyzing: %s\n", i+1, len(files), relPath)

		summary, err := generateFileSummary(file, provider)
		if err != nil {
			fmt.Printf("    ❌ Failed to generate summary: %v\n", err)
			continue
		}

		// Use relative path as key to preserve directory structure
		fileSummaries[relPath] = summary
		fmt.Printf("    ✅ Summary generated\n")
	}

	if len(fileSummaries) == 0 {
		fmt.Println("❌ No file summaries could be generated")
		os.Exit(1)
	}

	// Generate global summary
	fmt.Println("\n🌍 Generating global project summary...")
	globalSummary, err := generateGlobalSummary(fileSummaries, provider)
	if err != nil {
		fmt.Printf("❌ Failed to generate global summary: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Global summary generated")

	// Create .doxyllm.yaml.yaml configuration
	doxyllmConfig := DoxyllmConfig{
		Global: globalSummary,
		Files:  fileSummaries,
		Ignore: []string{}, // Empty by default, user can add files manually
	}

	// Write to file
	fmt.Printf("\n💾 Writing .doxyllm.yaml.yaml configuration to: %s\n", doxyllmPath)
	if err := writeDoxyllmConfig(doxyllmPath, &doxyllmConfig); err != nil {
		fmt.Printf("❌ Failed to write configuration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n🎉 Successfully initialized .doxyllm.yaml.yaml configuration!\n")
	fmt.Printf("📊 Generated summaries for %d files\n", len(fileSummaries))
	fmt.Println("\n💡 You can now run 'doxyllm-it ollama' to generate documentation with enhanced context")
}

func findCppFilesForInit(targetDir string) ([]string, error) {
	var files []string

	// Default excluded directories
	excludeDirs := []string{"build", "vendor", "third_party", ".git", "node_modules"}

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() {
			for _, exclude := range excludeDirs {
				if strings.Contains(path, exclude) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if isCppHeader(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

const fileSummaryPrompt = `You are a C++ code analysis expert. Analyze the provided C++ header file and generate a concise summary that describes:

1. The main purpose and functionality of the file
2. Key classes, functions, or components defined
3. The role this file plays in the broader codebase
4. Any important patterns, utilities, or abstractions provided

Keep the summary focused and technical, suitable for helping an AI generate accurate documentation.
Limit the summary to 2-3 sentences that capture the essence of the file.

File content:
` + "```cpp\n%s\n```" + `

Generate a concise technical summary of this file's purpose and contents:`

const globalSummaryPrompt = `You are a C++ project analysis expert. Based on the individual file summaries provided below, generate a comprehensive project overview that describes:

1. The overall purpose and domain of this C++ project
2. Main architectural patterns and design approaches used
3. Key functional areas or modules represented
4. The general coding style and conventions observed

This overview will be used to provide context for AI-generated documentation, so focus on technical aspects that would help understand the codebase's structure and intent.

File summaries:
%s

Generate a comprehensive project overview based on these file summaries:`

func generateFileSummary(filePath string, provider llm.Provider) (string, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Parse file to get a better understanding of its structure
	p := parser.New()
	_, err = p.Parse(filePath, string(content))
	if err != nil {
		// If parsing fails, still try to generate summary from raw content
		fmt.Printf("    ⚠️  Parsing failed, using raw content\n")
	}

	// Limit content size to avoid overwhelming the LLM
	contentStr := string(content)
	if len(contentStr) > 8000 {
		contentStr = contentStr[:8000] + "\n// ... (file truncated for analysis)"
	}

	prompt := fmt.Sprintf(fileSummaryPrompt, contentStr)

	// Use the LLM provider to generate the summary
	ctx := context.Background()
	request := llm.CommentRequest{
		EntityName:        filepath.Base(filePath),
		EntityType:        "file",
		Context:           contentStr,
		AdditionalContext: prompt,
	}

	response, err := provider.GenerateComment(ctx, request)
	if err != nil {
		return "", err
	}

	// Clean and trim the response
	summary := strings.TrimSpace(response.Description)

	// Remove any markdown formatting
	if strings.HasPrefix(summary, "```") {
		lines := strings.Split(summary, "\n")
		if len(lines) > 2 {
			summary = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	summary = strings.TrimPrefix(summary, "```")
	summary = strings.TrimSuffix(summary, "```")
	summary = strings.TrimSpace(summary)

	return summary, nil
}

func generateGlobalSummary(fileSummaries map[string]string, provider llm.Provider) (string, error) {
	// Build the summaries text
	var summariesText strings.Builder
	for filename, summary := range fileSummaries {
		summariesText.WriteString(fmt.Sprintf("**%s:**\n%s\n\n", filename, summary))
	}

	prompt := fmt.Sprintf(globalSummaryPrompt, summariesText.String())

	// Use the LLM provider to generate the global summary
	ctx := context.Background()
	request := llm.CommentRequest{
		EntityName:        "project",
		EntityType:        "global",
		Context:           summariesText.String(),
		AdditionalContext: prompt,
	}

	response, err := provider.GenerateComment(ctx, request)
	if err != nil {
		return "", err
	}

	// Clean and trim the response
	summary := strings.TrimSpace(response.Description)

	// Remove any markdown formatting
	if strings.HasPrefix(summary, "```") {
		lines := strings.Split(summary, "\n")
		if len(lines) > 2 {
			summary = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	summary = strings.TrimPrefix(summary, "```")
	summary = strings.TrimSuffix(summary, "```")
	summary = strings.TrimSpace(summary)

	return summary, nil
}

func writeDoxyllmConfig(filePath string, config *DoxyllmConfig) error {
	// Create a custom YAML format with proper multiline strings
	var content strings.Builder

	// Add header comment
	content.WriteString(`# .doxyllm.yaml.yaml configuration file
# Generated by doxyllm-it init command
#
# This file provides context to help the LLM generate better documentation.
# You can modify these summaries or add file-specific context as needed.
#
# Structure:
# - global: Overall project context used for all files
# - files: File-specific context for individual files
# - ignore: List of file patterns to skip during documentation generation

`)

	// Write global section with proper multiline formatting
	if config.Global != "" {
		content.WriteString("global: |\n")
		for _, line := range strings.Split(config.Global, "\n") {
			content.WriteString("  " + line + "\n")
		}
		content.WriteString("\n")
	}

	// Write files section
	if len(config.Files) > 0 {
		content.WriteString("files:\n")
		for filename, summary := range config.Files {
			content.WriteString("  " + filename + ": |\n")
			for _, line := range strings.Split(summary, "\n") {
				content.WriteString("    " + line + "\n")
			}
			content.WriteString("\n")
		}
	}

	// Write ignore section
	if len(config.Ignore) > 0 {
		content.WriteString("ignore:\n")
		for _, pattern := range config.Ignore {
			content.WriteString("  - '" + pattern + "'\n")
		}
	} else {
		content.WriteString("ignore: []\n")
	}

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}
