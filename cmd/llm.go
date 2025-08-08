package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"doxyllm-it/pkg/document"
	"doxyllm-it/pkg/llm"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// llmCmd represents the LLM-based documentation generation command
var llmCmd = &cobra.Command{
	Use:   "llm [flags] <file_or_directory>",
	Short: "Generate Doxygen comments using LLM providers (Ollama, OpenAI, etc.)",
	Long: `Generate Doxygen comments for undocumented C++ entities using various LLM providers.

This command provides a unified interface for multiple LLM providers including Ollama (default),
OpenAI, Anthropic, and others. It uses the document abstraction layer for clean architecture,
better separation of concerns, and improved testability.

Examples:
  # Process single file with default settings (Ollama)
  doxyllm-it llm examples/example.hpp

  # Process directory with custom model
  doxyllm-it llm --model codellama:13b src/

  # Use specific provider
  doxyllm-it llm --provider ollama --model deepseek-coder:6.7b src/

  # Process with custom URL
  doxyllm-it llm --url http://remote:11434 --model deepseek-coder:6.7b .

  # Dry run to see what would be processed
  doxyllm-it llm --dry-run src/

  # Limit entities per file for testing
  doxyllm-it llm --max-entities 3 examples/`,
	Args: cobra.ExactArgs(1),
	Run:  runLLM,
}

// DoxyllmConfig represents the structure of a .doxyllm.yaml configuration file
type DoxyllmConfig struct {
	Global string                           `yaml:"global,omitempty"`
	Files  map[string]string                `yaml:"files,omitempty"`
	Ignore []string                         `yaml:"ignore,omitempty"`
	Groups map[string]*document.GroupConfig `yaml:"groups,omitempty"`
}

var (
	// Command-line flags for LLM
	llmProvider     string
	llmURL          string
	llmModel        string
	llmAPIKey       string
	llmTemperature  float64
	llmTopP         float64
	llmNumCtx       int
	llmTimeout      int
	llmMaxEntities  int
	llmDryRun       bool
	llmBackup       bool
	llmFormatOutput bool
	llmExcludeDirs  []string
)

func init() {
	rootCmd.AddCommand(llmCmd)

	// LLM Provider configuration flags
	llmCmd.Flags().StringVarP(&llmProvider, "provider", "p", "ollama", "LLM provider (ollama, openai, anthropic)")
	llmCmd.Flags().StringVarP(&llmURL, "url", "u", getEnvOrDefault("LLM_URL", "http://localhost:11434/api/generate"), "LLM API URL")
	llmCmd.Flags().StringVarP(&llmModel, "model", "m", getEnvOrDefault("LLM_MODEL", "deepseek-coder:6.7b"), "LLM model name")
	llmCmd.Flags().StringVar(&llmAPIKey, "api-key", getEnvOrDefault("LLM_API_KEY", ""), "API key for cloud providers")
	llmCmd.Flags().Float64Var(&llmTemperature, "temperature", 0.1, "LLM temperature (0.0-1.0)")
	llmCmd.Flags().Float64Var(&llmTopP, "top-p", 0.9, "LLM top-p value (0.0-1.0)")
	llmCmd.Flags().IntVar(&llmNumCtx, "context", 4096, "Context window size")
	llmCmd.Flags().IntVar(&llmTimeout, "timeout", 120, "Request timeout in seconds")

	// Processing flags
	llmCmd.Flags().IntVar(&llmMaxEntities, "max-entities", 0, "Maximum entities to process per file (0 = unlimited)")
	llmCmd.Flags().BoolVar(&llmDryRun, "dry-run", false, "Show what would be processed without making changes")
	llmCmd.Flags().BoolVarP(&llmBackup, "backup", "b", false, "Create backup files before updating")
	llmCmd.Flags().BoolVarP(&llmFormatOutput, "format", "f", false, "Format updated files with clang-format")
	llmCmd.Flags().StringSliceVar(&llmExcludeDirs, "exclude", []string{"build", "vendor", "third_party", ".git", "node_modules"}, "Directories to exclude")
}

// getEnvOrDefault gets an environment variable value or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runLLM(cmd *cobra.Command, args []string) {
	target := args[0]

	fmt.Println("🚀 Starting Doxygen documentation generation with LLM")

	// Create LLM configuration
	llmConfig := &llm.Config{
		Provider:    llmProvider,
		URL:         llmURL,
		Model:       llmModel,
		Temperature: llmTemperature,
		TopP:        llmTopP,
		NumCtx:      llmNumCtx,
		Timeout:     time.Duration(llmTimeout) * time.Second,
		Options:     make(map[string]interface{}),
	}

	// Add API key to options if provided
	if llmAPIKey != "" {
		llmConfig.Options["api_key"] = llmAPIKey
	}

	// Create LLM provider
	provider, err := llm.NewProvider(llmConfig)
	if err != nil {
		fmt.Printf("❌ Failed to create LLM provider: %v\n", err)
		os.Exit(1)
	}

	// Create LLM documentation service
	llmService := llm.NewDocumentationService(provider)

	// Create document service
	docService := document.NewDocumentationService(llmService)

	// Check if dry-run first
	if llmDryRun {
		fmt.Println("🔍 Dry run mode - no files will be modified")
		// Skip connection test for dry-run
	} else {
		// Test LLM connectivity only for actual runs
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = llmService.TestConnection(ctx)
		if err != nil {
			fmt.Printf("❌ Failed to connect to LLM: %v\n", err)
			os.Exit(1)
		}
	}

	modelInfo := llmService.GetModelInfo()
	fmt.Printf("🤖 Connected to %s provider\n", llmProvider)
	fmt.Printf("📚 Using model: %s\n", modelInfo.Name)
	if llmURL != "" {
		fmt.Printf("🔗 API URL: %s\n", llmURL)
	}

	// Find files to process
	files, err := findCppFiles(target)
	if err != nil {
		fmt.Printf("❌ Failed to find files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("⚠️  No C++ header files found")
		return
	}

	fmt.Printf("📂 Found %d C++ header files\n", len(files))

	if llmDryRun {
		fmt.Println("🔍 Dry run mode - no files will be modified")
	}

	// Process files
	totalUpdates := 0
	updatedFiles := []string{}

	for _, file := range files {
		result := processFile(file, docService, target)
		if result.EntitiesUpdated > 0 {
			updatedFiles = append(updatedFiles, file)
		}
		totalUpdates += result.EntitiesUpdated
	}

	// Summary
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  Files processed: %d\n", len(files))
	fmt.Printf("  Files updated: %d\n", len(updatedFiles))
	fmt.Printf("  Total entities documented: %d\n", totalUpdates)

	if llmDryRun {
		fmt.Println("  ℹ️  This was a dry run - no changes were made")
	} else if len(updatedFiles) > 0 {
		fmt.Println("  ✅ Documentation generation completed successfully!")
		if llmFormatOutput {
			fmt.Println("  📝 Files were formatted with clang-format")
		}
	} else {
		fmt.Println("  ℹ️  No entities needed documentation")
	}
}

// processFile processes a single file using the document abstraction
func processFile(filePath string, docService *document.DocumentationService, rootPath string) *document.ProcessingResult {
	fmt.Printf("\n📁 Processing: %s\n", filePath)

	// Load the document
	doc, err := document.NewFromFile(filePath)
	if err != nil {
		fmt.Printf("  ❌ Failed to load document: %v\n", err)
		return &document.ProcessingResult{}
	}

	// Check if file should be ignored
	if shouldIgnoreFile(filePath, rootPath) {
		fmt.Printf("  ⏭️  File ignored by configuration\n")
		return &document.ProcessingResult{}
	}

	// Load configuration
	config := loadDoxyllmConfig(filePath, rootPath)
	var group *document.GroupConfig

	// Determine group for this file
	if config != nil && config.Groups != nil {
		group = getGroupForFile(filePath, rootPath, config)
	}

	// Add @defgroup if needed
	if group != nil && group.GenerateDefgroup {
		err := docService.AddDefgroupToDocument(doc, group)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to add defgroup: %v\n", err)
		}
	}

	// Process undocumented entities
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	opts := document.ProcessingOptions{
		MaxEntities:  llmMaxEntities,
		DryRun:       llmDryRun,
		BackupFiles:  llmBackup,
		FormatOutput: llmFormatOutput,
		GroupConfig:  group,
	}

	result, err := docService.ProcessUndocumentedEntities(ctx, doc, opts)
	if err != nil {
		fmt.Printf("  ❌ Failed to process entities: %v\n", err)
		return &document.ProcessingResult{}
	}

	// Process entities needing group updates
	if group != nil {
		groupResult, err := docService.ProcessEntitiesNeedingGroupUpdate(ctx, doc, group)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to update groups: %v\n", err)
		} else {
			result.EntitiesUpdated += groupResult.EntitiesUpdated
			result.UpdatedEntities = append(result.UpdatedEntities, groupResult.UpdatedEntities...)
		}
	}

	// Report progress
	if result.EntitiesProcessed == 0 {
		fmt.Printf("  ✅ No entities need documentation\n")
	} else {
		fmt.Printf("  📋 Found %d entities needing updates\n", result.EntitiesProcessed)
		if llmDryRun {
			for i, entity := range result.UpdatedEntities {
				fmt.Printf("    📝 (%d/%d) Would document: %s\n", i+1, len(result.UpdatedEntities), entity)
			}
		} else {
			fmt.Printf("  📊 Updated %d/%d entities\n", result.EntitiesUpdated, result.EntitiesProcessed)
		}
	}

	// Report any errors
	if len(result.Errors) > 0 {
		fmt.Printf("  ⚠️  %d errors encountered:\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("    • %v\n", err)
		}
	}

	// Save the document if changes were made and not in dry run mode
	if !llmDryRun && doc.IsModified() {
		// Create backup if requested
		if llmBackup {
			backupPath := filePath + ".bak"
			originalContent, _ := os.ReadFile(filePath)
			err := os.WriteFile(backupPath, originalContent, 0644)
			if err != nil {
				fmt.Printf("  ⚠️  Failed to create backup: %v\n", err)
			}
		}

		// Save the document
		if llmFormatOutput {
			content, err := doc.SaveToStringFormatted()
			if err != nil {
				fmt.Printf("  ⚠️  Clang-format failed, using unformatted output: %v\n", err)
				content, err = doc.SaveToString()
			}
			if err == nil {
				err = os.WriteFile(filePath, []byte(content), 0644)
				if err != nil {
					fmt.Printf("  ❌ Failed to write formatted file: %v\n", err)
				}
			}
		} else {
			err = doc.Save()
		}

		if err != nil {
			fmt.Printf("  ❌ Failed to save document: %v\n", err)
		}
	}

	return result
}

// findCppFiles finds C++ header files in the target directory
func findCppFiles(target string) ([]string, error) {
	var files []string

	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		// Single file
		if isCppHeader(target) {
			files = append(files, target)
		}
		return files, nil
	}

	// Directory - walk recursively
	err = filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		for _, excludeDir := range llmExcludeDirs {
			if info.IsDir() && info.Name() == excludeDir {
				return filepath.SkipDir
			}
		}

		// Add C++ header files
		if !info.IsDir() && isCppHeader(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// isCppHeader checks if a file is a C++ header file
func isCppHeader(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".hpp" || ext == ".h" || ext == ".hxx"
}

// shouldIgnoreFile checks if a file should be ignored based on configuration
func shouldIgnoreFile(filePath, rootPath string) bool {
	config := loadDoxyllmConfig(filePath, rootPath)
	if config == nil {
		return false
	}

	fileName := filepath.Base(filePath)
	for _, ignorePattern := range config.Ignore {
		if matched, _ := filepath.Match(ignorePattern, fileName); matched {
			return true
		}
	}

	return false
}

// loadDoxyllmConfig loads the .doxyllm.yaml.yaml configuration file
func loadDoxyllmConfig(filePath, rootPath string) *DoxyllmConfig {
	var rootDir string
	if info, err := os.Stat(rootPath); err == nil && info.IsDir() {
		rootDir = rootPath
	} else {
		rootDir = filepath.Dir(rootPath)
	}

	doxyllmPath := filepath.Join(rootDir, ".doxyllm.yaml.yaml")
	content, err := os.ReadFile(doxyllmPath)
	if err != nil {
		return nil
	}

	var config DoxyllmConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		fmt.Printf("⚠️  Warning: Failed to parse .doxyllm.yaml.yaml: %v\n", err)
		return nil
	}

	return &config
}

// getGroupForFile determines which group a file belongs to
func getGroupForFile(filePath, rootPath string, config *DoxyllmConfig) *document.GroupConfig {
	if config.Groups == nil {
		return nil
	}

	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		relPath = filepath.Base(filePath)
	}

	for _, group := range config.Groups {
		for _, groupFile := range group.Files {
			if matched, _ := filepath.Match(groupFile, relPath); matched {
				return group
			}
		}
	}

	return nil
}
