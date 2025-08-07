package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/formatter"
	"doxyllm-it/pkg/llm"
	"doxyllm-it/pkg/parser"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// DoxyllmConfig represents the structure of a .doxyllm configuration file
type DoxyllmConfig struct {
	Global string            `yaml:"global,omitempty"`
	Files  map[string]string `yaml:"files,omitempty"`
	Ignore []string          `yaml:"ignore,omitempty"`
	// Group configuration
	Groups map[string]GroupConfig `yaml:"groups,omitempty"`
}

// GroupConfig defines configuration for Doxygen groups
type GroupConfig struct {
	Name             string   `yaml:"name"`             // Group name (for @defgroup/@ingroup)
	Title            string   `yaml:"title"`            // Group title/brief description
	Description      string   `yaml:"description"`      // Detailed group description
	Files            []string `yaml:"files"`            // Files that belong to this group
	GenerateDefgroup bool     `yaml:"generateDefgroup"` // Whether to generate @defgroup in header files
}

var ollamaCmd = &cobra.Command{
	Use:   "ollama [flags] <file_or_directory>",
	Short: "Generate Doxygen comments using Ollama LLM",
	Long: `Generate Doxygen comments for undocumented C++ entities using Ollama LLM.

This command integrates with a local or remote Ollama instance to automatically
generate comprehensive Doxygen documentation for C++ code. It identifies
undocumented entities, extracts their context, generates appropriate comments
using the specified LLM model, and updates the source files.

Examples:
  # Process single file with default settings
  doxyllm-it ollama examples/example.hpp

  # Process directory with custom model
  doxyllm-it ollama --model codellama:13b src/

  # Process with custom Ollama URL
  doxyllm-it ollama --url http://remote:11434 --model deepseek-coder:6.7b .

  # Dry run to see what would be processed
  doxyllm-it ollama --dry-run src/

  # Limit entities per file for testing
  doxyllm-it ollama --max-entities 3 examples/`,
	Args: cobra.ExactArgs(1),
	Run:  runOllamaV2,
}

var (
	ollamaURL    string
	ollamaModel  string
	temperature  float64
	topP         float64
	numCtx       int
	timeout      int
	maxEntities  int
	dryRun       bool
	backup       bool
	formatOutput bool
	excludeDirs  []string
)

func init() {
	rootCmd.AddCommand(ollamaCmd)

	// Ollama configuration flags
	ollamaCmd.Flags().StringVarP(&ollamaURL, "url", "u", getEnvOrDefault("OLLAMA_URL", "http://10.19.4.106:11434/api/generate"), "Ollama API URL")
	ollamaCmd.Flags().StringVarP(&ollamaModel, "model", "m", getEnvOrDefault("MODEL_NAME", "deepseek-coder:6.7b"), "Ollama model name")
	ollamaCmd.Flags().Float64Var(&temperature, "temperature", 0.1, "LLM temperature (0.0-1.0)")
	ollamaCmd.Flags().Float64Var(&topP, "top-p", 0.9, "LLM top-p value (0.0-1.0)")
	ollamaCmd.Flags().IntVar(&numCtx, "context", 4096, "Context window size")
	ollamaCmd.Flags().IntVar(&timeout, "timeout", 120, "Request timeout in seconds")

	// Processing flags
	ollamaCmd.Flags().IntVar(&maxEntities, "max-entities", 0, "Maximum entities to process per file (0 = unlimited)")
	ollamaCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be processed without making changes")
	ollamaCmd.Flags().BoolVarP(&backup, "backup", "b", false, "Create backup files before updating")
	ollamaCmd.Flags().BoolVarP(&formatOutput, "format", "f", false, "Format updated files with clang-format")
	ollamaCmd.Flags().StringSliceVar(&excludeDirs, "exclude", []string{"build", "vendor", "third_party", ".git", "node_modules"}, "Directories to exclude")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runOllamaV2(cmd *cobra.Command, args []string) {
	target := args[0]

	// Create LLM configuration
	llmConfig := &llm.Config{
		Provider:    "ollama",
		URL:         ollamaURL,
		Model:       ollamaModel,
		Temperature: temperature,
		TopP:        topP,
		NumCtx:      numCtx,
		Timeout:     time.Duration(timeout) * time.Second,
	}

	// Create LLM provider
	provider, err := llm.NewProvider(llmConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create LLM provider: %v\n", err)
		os.Exit(1)
	}

	// Create documentation service
	docService := llm.NewDocumentationService(provider)

	// Test LLM connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := docService.TestConnection(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Cannot connect to LLM provider: %v\n", err)
		os.Exit(1)
	}

	modelInfo := docService.GetModelInfo()
	fmt.Printf("🤖 Connected to %s\n", llmConfig.URL)
	fmt.Printf("📚 Using model: %s\n", modelInfo.Name)

	// Find files to process
	files, err := findCppFiles(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("❌ No C++ header files found")
		os.Exit(1)
	}

	fmt.Printf("📂 Found %d C++ header files\n", len(files))

	if dryRun {
		fmt.Println("\n🔍 Dry run mode - showing what would be processed:")
	}

	// Process files
	totalUpdates := 0
	updatedFiles := []string{}

	for _, file := range files {
		// Always use the current working directory as root for configuration
		rootPath, _ := os.Getwd()

		updates := processFileWithService(file, docService, rootPath)
		if updates > 0 {
			totalUpdates += updates
			updatedFiles = append(updatedFiles, file)
		}
	}

	// Summary
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  Files processed: %d\n", len(files))
	fmt.Printf("  Files updated: %d\n", len(updatedFiles))
	fmt.Printf("  Total entities documented: %d\n", totalUpdates)

	if dryRun {
		fmt.Println("\n💡 Run without --dry-run to apply changes")
	} else if len(updatedFiles) > 0 {
		fmt.Println("\n🎉 Documentation generation complete!")
	} else {
		fmt.Println("\n✅ All files already have complete documentation")
	}
}

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

func isCppHeader(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".hpp" || ext == ".h" || ext == ".hxx"
}

func processFileWithService(filePath string, service *llm.DocumentationService, rootPath string) int {
	fmt.Printf("\n📁 Processing: %s\n", filePath)

	// Check if file should be ignored
	_, shouldIgnore := readDoxyllmContext(filePath, rootPath)
	if shouldIgnore {
		fmt.Println("  ⏭️  File ignored per .doxyllm configuration")
		return 0
	}

	// Read .doxyllm configuration
	doxyllmConfig := &DoxyllmConfig{}
	if configContent, err := os.ReadFile(filepath.Join(rootPath, ".doxyllm.yaml")); err == nil {
		yaml.Unmarshal(configContent, doxyllmConfig)
	}

	// Check if we need to add @defgroup at file level
	defgroupAdded := false
	shouldGenerate, group := shouldGenerateDefgroup(filePath, rootPath, doxyllmConfig)
	if shouldGenerate {
		wasAdded, err := addDefgroupToFile(filePath, group)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to add @defgroup: %v\n", err)
		} else if wasAdded {
			fmt.Printf("  📑 Added @defgroup %s\n", group.Name)
			defgroupAdded = true
		}
	}

	// Parse file to get entities that need updates (undocumented OR missing @ingroup)
	entitiesToUpdate, err := getEntitiesNeedingUpdates(filePath, group)
	if err != nil {
		fmt.Printf("  ❌ Error parsing file: %v\n", err)
		if defgroupAdded {
			return 1 // Still count as updated if we added defgroup
		}
		return 0
	}

	if len(entitiesToUpdate) == 0 {
		fmt.Println("  ✅ All entities already documented and grouped")
		if defgroupAdded {
			return 1 // Count as updated if we added defgroup
		}
		return 0
	}

	fmt.Printf("  📋 Found %d entities needing updates\n", len(entitiesToUpdate))

	// Limit entities if specified
	if maxEntities > 0 && len(entitiesToUpdate) > maxEntities {
		entitiesToUpdate = entitiesToUpdate[:maxEntities]
		fmt.Printf("  🔢 Processing first %d entities\n", len(entitiesToUpdate))
	}

	if dryRun {
		for i, entity := range entitiesToUpdate {
			fmt.Printf("  📝 (%d/%d) Would document: %s\n", i+1, len(entitiesToUpdate), entity)
		}
		dryRunUpdates := len(entitiesToUpdate)
		if defgroupAdded {
			dryRunUpdates++ // Count the defgroup addition
		}
		return dryRunUpdates
	}

	successfulUpdates := 0

	for i, entityPath := range entitiesToUpdate {
		fmt.Printf("  📝 (%d/%d) Processing: %s\n", i+1, len(entitiesToUpdate), entityPath)

		// Check if entity is completely undocumented or just missing @ingroup
		isUndocumented := isEntityUndocumented(filePath, entityPath)
		
		if isUndocumented {
			// Generate new documentation for undocumented entity
			if err := generateDocumentationForEntity(filePath, entityPath, service, rootPath, group); err != nil {
				fmt.Printf("    ❌ Failed to generate documentation: %v\n", err)
				continue
			}
			fmt.Println("    ✅ Generated new documentation")
		} else {
			// Just add @ingroup to existing documentation
			if err := addIngroupToExistingComment(filePath, entityPath, group); err != nil {
				fmt.Printf("    ❌ Failed to add @ingroup: %v\n", err)
				continue
			}
			fmt.Println("    ✅ Added @ingroup to existing documentation")
		}

		successfulUpdates++
	}

	fmt.Printf("  📊 Updated %d/%d entities\n", successfulUpdates, len(entitiesToUpdate))
	
	// Return total updates including defgroup if added
	totalUpdates := successfulUpdates
	if defgroupAdded {
		totalUpdates++ // Count the defgroup addition
	}
	return totalUpdates
}

// Helper functions that are reused from the original implementation

// readDoxyllmContext reads a .doxyllm file from the root target directory
func readDoxyllmContext(filePath, rootPath string) (string, bool) {
	var rootDir string
	if info, err := os.Stat(rootPath); err == nil && info.IsDir() {
		rootDir = rootPath
	} else {
		rootDir = filepath.Dir(rootPath)
	}
	doxyllmPath := filepath.Join(rootDir, ".doxyllm.yaml")

	content, err := os.ReadFile(doxyllmPath)
	if err != nil {
		return "", false
	}

	var config DoxyllmConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return strings.TrimSpace(string(content)), false
	}

	fileName := filepath.Base(filePath)
	for _, ignorePattern := range config.Ignore {
		if matched, _ := filepath.Match(ignorePattern, fileName); matched {
			return "", true
		}
	}

	var contextParts []string
	if config.Global != "" {
		contextParts = append(contextParts, config.Global)
	}

	configDir := filepath.Dir(doxyllmPath)
	relPath, _ := filepath.Rel(configDir, filePath)

	var fileContext string
	var exists bool

	if fileContext, exists = config.Files[relPath]; !exists {
		fileContext, exists = config.Files[fileName]
	}

	if exists && fileContext != "" {
		contextIdentifier := relPath
		if _, exists := config.Files[relPath]; !exists {
			contextIdentifier = fileName
		}
		contextParts = append(contextParts, fmt.Sprintf("SPECIFIC TO %s:\n%s", contextIdentifier, fileContext))
	}

	return strings.Join(contextParts, "\n\n"), false
}

// getGroupForFile determines which group a file belongs to based on configuration
func getGroupForFile(filePath, rootPath string, config *DoxyllmConfig) *GroupConfig {
	if config.Groups == nil {
		return nil
	}

	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		relPath = filepath.Base(filePath)
	}

	for _, group := range config.Groups {
		for _, groupFile := range group.Files {
			matched, err := filepath.Match(groupFile, relPath)
			if err == nil && matched {
				return &group
			}
			if groupFile == relPath || groupFile == filepath.Base(filePath) {
				return &group
			}
		}
	}

	return nil
}

// shouldGenerateDefgroup determines if a file should contain @defgroup
func shouldGenerateDefgroup(filePath, rootPath string, config *DoxyllmConfig) (bool, *GroupConfig) {
	group := getGroupForFile(filePath, rootPath, config)
	if group == nil {
		return false, nil
	}
	return group.GenerateDefgroup, group
}

// generateGroupComment creates a @defgroup comment for a file
func generateGroupComment(group *GroupConfig) string {
	var comment strings.Builder
	comment.WriteString("/**\n")
	comment.WriteString(fmt.Sprintf(" * @defgroup %s %s\n", group.Name, group.Title))

	if group.Description != "" {
		comment.WriteString(" * @{\n")
		comment.WriteString(" *\n")
		lines := strings.Split(group.Description, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				comment.WriteString(fmt.Sprintf(" * %s\n", line))
			} else {
				comment.WriteString(" *\n")
			}
		}
		comment.WriteString(" * @}\n")
	}

	comment.WriteString(" */")
	return comment.String()
}

// addDefgroupToFile adds a @defgroup comment at the beginning of a file
// Returns (wasAdded, error) where wasAdded indicates if the defgroup was actually added
func addDefgroupToFile(filePath string, group *GroupConfig) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(content), "\n")

	// Check if @defgroup already exists
	for _, line := range lines {
		if strings.Contains(line, "@defgroup") && strings.Contains(line, group.Name) {
			return false, nil // Already exists, no error, but not added
		}
	}

	// Find the insertion point - after file header comments and header guards
	insertIndex := 0
	inComment := false
	foundHeaderGuard := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the beginning
		if trimmed == "" && !foundHeaderGuard {
			continue
		}

		// Track comment blocks
		if strings.HasPrefix(trimmed, "/**") {
			inComment = true
			continue
		}
		if strings.HasSuffix(trimmed, "*/") && inComment {
			inComment = false
			continue
		}
		if inComment || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// Skip header guards
		if strings.HasPrefix(trimmed, "#ifndef") || strings.HasPrefix(trimmed, "#define") {
			foundHeaderGuard = true
			continue
		}

		// Skip other preprocessor directives and single-line comments
		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			continue
		}

		// If we reach here, we found the first actual code line
		insertIndex = i
		break
	}

	defgroupComment := generateGroupComment(group)
	commentLines := strings.Split(defgroupComment, "\n")

	newLines := make([]string, len(lines)+len(commentLines)+1)
	copy(newLines[:insertIndex], lines[:insertIndex])
	copy(newLines[insertIndex:insertIndex+len(commentLines)], commentLines)
	newLines[insertIndex+len(commentLines)] = ""
	copy(newLines[insertIndex+len(commentLines)+1:], lines[insertIndex:])

	updatedContent := strings.Join(newLines, "\n")
	err = os.WriteFile(filePath, []byte(updatedContent), 0644)
	return true, err // Successfully added
}

func getUndocumentedEntities(filepath string) ([]string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	p := parser.New()
	scopeTree, err := p.Parse(filepath, string(content))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	var undocumented []string
	documentedEntities := make(map[string]bool)

	var traverse func(*ast.Entity)
	traverse = func(entity *ast.Entity) {
		if shouldSkipEntity(entity) {
			return
		}

		hasComment := entity.Comment != nil || hasCommentBefore(lines, entity.SourceRange.Start.Line)

		if !hasComment && !documentedEntities[entity.FullName] {
			if entity.Type == ast.EntityConstructor {
				parentName := getParentEntityName(entity.FullName)
				if documentedEntities[parentName] {
					return
				}
			}

			undocumented = append(undocumented, entity.FullName)
			documentedEntities[entity.FullName] = true
		}

		for _, child := range entity.Children {
			traverse(child)
		}
	}

	for _, entity := range scopeTree.Root.Children {
		traverse(entity)
	}

	return undocumented, nil
}

func hasCommentBefore(lines []string, entityLine int) bool {
	if entityLine <= 1 {
		return false
	}

	foundCommentEnd := -1
	maxLookback := 10

	for i := entityLine - 2; i >= 0 && i >= entityLine-maxLookback; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "template") ||
			strings.HasPrefix(line, "inline") ||
			strings.HasPrefix(line, "static") ||
			strings.HasPrefix(line, "const ") ||
			strings.HasPrefix(line, "constexpr") ||
			strings.HasPrefix(line, "[[") {
			continue
		}

		if strings.HasSuffix(line, "*/") && !strings.HasSuffix(line, "**/") {
			if !strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "/**") {
				foundCommentEnd = i
				break
			}
		}

		if strings.HasPrefix(line, "///") || strings.HasPrefix(line, "//!") {
			return true
		}

		if !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "/*") &&
			!strings.HasPrefix(line, "*") && !strings.HasSuffix(line, "*/") {
			break
		}
	}

	if foundCommentEnd >= 0 {
		for j := foundCommentEnd; j >= 0; j-- {
			commentLine := strings.TrimSpace(lines[j])
			if commentLine == "" {
				continue
			}

			if strings.HasPrefix(commentLine, "/**") {
				return true
			}

			if !strings.HasPrefix(commentLine, "*") && !strings.HasSuffix(commentLine, "*/") {
				break
			}
		}
	}

	return false
}

func shouldSkipEntity(entity *ast.Entity) bool {
	// Skip single-letter entities (likely template parameters)
	if len(entity.Name) == 1 && ((entity.Name >= "A" && entity.Name <= "Z") || (entity.Name >= "a" && entity.Name <= "z")) {
		return true
	}

	if entity.Type == ast.EntityVariable {
		localVarNames := map[string]bool{
			"msg": true, "result": true, "temp": true, "i": true, "j": true, "k": true,
			"it": true, "iter": true, "val": true, "value": true, "ret": true,
			"data": true, "ptr": true, "len": true, "size": true, "count": true,
			"txt": true, "str": true, "buffer": true, "instance": true,
			"binary": true, "in": true, "out": true, "trunc": true, "npos": true,
			"end_index": true, "start_index": true, "srcIndex": true, "timepoint": true,
		}

		if localVarNames[entity.Name] {
			return true
		}

		if len(entity.Name) <= 3 && strings.ToLower(entity.Name) == entity.Name {
			return true
		}

		if strings.Contains(entity.Signature, "static const") || strings.Contains(entity.Signature, "const ") {
			return true
		}

		if strings.Contains(entity.Signature, " = ") &&
			(strings.HasSuffix(entity.Signature, ");") ||
				strings.HasSuffix(entity.Signature, ",") ||
				strings.Contains(entity.Signature, entity.Name+" = ")) {
			return true
		}

		if strings.Contains(entity.Signature, "<<") ||
			strings.Contains(entity.Signature, ">>") ||
			strings.Contains(entity.Signature, "->") ||
			strings.Contains(entity.Signature, ".") {
			return true
		}
	}

	if entity.Type == ast.EntityFunction {
		if strings.ToUpper(entity.Name) == entity.Name && strings.Contains(entity.Name, "_") {
			return true
		}

		if strings.Contains(entity.Signature, "std::string "+entity.Name) ||
			strings.Contains(entity.Signature, "auto "+entity.Name) ||
			strings.Contains(entity.Signature, "const "+entity.Name) ||
			strings.Contains(entity.Signature, "int "+entity.Name) ||
			strings.Contains(entity.Signature, "float "+entity.Name) ||
			strings.Contains(entity.Signature, "double "+entity.Name) ||
			strings.Contains(entity.Signature, "bool "+entity.Name) {
			return true
		}

		if !strings.Contains(entity.Signature, "(") || strings.HasPrefix(entity.Signature, entity.Name+"(") {
			return true
		}

		assertMacros := map[string]bool{
			"MGL_CORE_ASSERT": true, "assert": true, "ASSERT": true,
			"CHECK": true, "VERIFY": true, "EXPECT": true,
		}

		if assertMacros[entity.Name] {
			return true
		}

		if len(entity.Name) <= 3 && strings.ToLower(entity.Name) == entity.Name {
			return true
		}
	}

	if len(entity.Name) > 1 && strings.ToUpper(entity.Name) == entity.Name &&
		(strings.Contains(entity.Name, "_") || entity.Type == ast.EntityFunction) {
		return true
	}

	commonTemplateParams := map[string]bool{
		"T": true, "U": true, "V": true, "E": true, "N": true, "S": true,
		"Container": true, "ElementType": true, "OtherElementType": true,
		"Count": true, "Offset": true, "Extent": true, "OtherExtent": true,
	}

	if commonTemplateParams[entity.Name] {
		return true
	}

	if len(entity.Name) <= 2 && strings.ToUpper(entity.Name) == entity.Name {
		return true
	}

	systemEntities := map[string]bool{
		"std":       true,
		"__gnu_cxx": true,
		"__detail":  true,
	}

	if systemEntities[entity.Name] {
		return true
	}

	if entity.Type == ast.EntityTypedef || entity.Type == ast.EntityUsing || entity.Type == ast.EntityVariable {
		typeAliases := map[string]bool{
			"value_type": true, "size_type": true, "difference_type": true,
			"pointer": true, "const_pointer": true, "reference": true, "const_reference": true,
			"iterator": true, "reverse_iterator": true, "element_type": true,
		}

		if typeAliases[entity.Name] {
			return true
		}
	}

	return false
}

func getParentEntityName(fullName string) string {
	parts := strings.Split(fullName, "::")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "::")
}

func extractEntityContext(filepath, entityPath string) (string, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	p := parser.New()
	scopeTree, err := p.Parse(filepath, string(content))
	if err != nil {
		return "", err
	}

	entity := findEntityByPath(scopeTree, entityPath)
	if entity == nil {
		return "", fmt.Errorf("entity not found: %s", entityPath)
	}

	f := formatter.New()

	switch entity.Type {
	case ast.EntityNamespace:
		return extractNamespaceContext(entity), nil
	case ast.EntityClass, ast.EntityStruct:
		return extractClassContext(entity), nil
	case ast.EntityFunction, ast.EntityMethod, ast.EntityConstructor, ast.EntityDestructor:
		return f.ExtractEntityContext(entity, false, false), nil
	default:
		return f.ExtractEntityContext(entity, false, false), nil
	}
}

func extractNamespaceContext(entity *ast.Entity) string {
	var context strings.Builder

	context.WriteString(entity.Signature)
	context.WriteString("\n")

	if len(entity.Children) > 0 {
		context.WriteString("  // Contains:\n")
		for _, child := range entity.Children {
			switch child.Type {
			case ast.EntityClass, ast.EntityStruct:
				context.WriteString(fmt.Sprintf("  %s %s;\n", child.Type.String(), child.Name))
			case ast.EntityFunction:
				context.WriteString(fmt.Sprintf("  %s\n", child.Signature))
			case ast.EntityEnum:
				context.WriteString(fmt.Sprintf("  enum %s;\n", child.Name))
			case ast.EntityNamespace:
				context.WriteString(fmt.Sprintf("  namespace %s;\n", child.Name))
			}
		}
	}

	return context.String()
}

func extractClassContext(entity *ast.Entity) string {
	var context strings.Builder

	context.WriteString(entity.Signature)
	context.WriteString("\n")

	if len(entity.Children) > 0 {
		context.WriteString("  // Public interface:\n")
		publicMethods := 0
		publicFields := 0

		for _, child := range entity.Children {
			if child.AccessLevel == ast.AccessPublic || child.AccessLevel == ast.AccessUnknown {
				switch child.Type {
				case ast.EntityMethod, ast.EntityFunction, ast.EntityConstructor, ast.EntityDestructor:
					if publicMethods < 5 {
						context.WriteString(fmt.Sprintf("  %s\n", child.Signature))
					}
					publicMethods++
				case ast.EntityField, ast.EntityVariable:
					if publicFields < 3 {
						context.WriteString(fmt.Sprintf("  %s\n", child.Signature))
					}
					publicFields++
				}
			}
		}

		if publicMethods > 5 {
			context.WriteString(fmt.Sprintf("  // ... and %d more methods\n", publicMethods-5))
		}
		if publicFields > 3 {
			context.WriteString(fmt.Sprintf("  // ... and %d more fields\n", publicFields-3))
		}
	}

	return context.String()
}

func findEntityByPath(scopeTree *ast.ScopeTree, fullName string) *ast.Entity {
	var find func(*ast.Entity) *ast.Entity
	find = func(entity *ast.Entity) *ast.Entity {
		if entity.FullName == fullName {
			return entity
		}
		for _, child := range entity.Children {
			if result := find(child); result != nil {
				return result
			}
		}
		return nil
	}

	return find(scopeTree.Root)
}

func getEntityTypeFromName(fullName string) string {
	if strings.Contains(fullName, "::") {
		parts := strings.Split(fullName, "::")
		lastPart := parts[len(parts)-1]

		if strings.HasSuffix(lastPart, "_") || strings.Contains(lastPart, "variable") {
			return "variable/field"
		}
		if strings.Contains(lastPart, "()") || strings.HasPrefix(lastPart, "get") || strings.HasPrefix(lastPart, "set") {
			return "method/function"
		}
		if isCapitalized(lastPart) {
			return "class/namespace"
		}
		return "member"
	}

	if isCapitalized(fullName) {
		return "class/namespace/type"
	}

	return "function/variable"
}

func isCapitalized(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := rune(s[0])
	return first >= 'A' && first <= 'Z'
}

func updateEntityComment(filepath, entityPath, comment string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	p := parser.New()
	scopeTree, err := p.Parse(filepath, string(content))
	if err != nil {
		return err
	}

	entity := findEntityByPath(scopeTree, entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	if backup {
		backupPath := filepath + ".bak"
		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
	}

	lines := strings.Split(string(content), "\n")
	entityLine := entity.SourceRange.Start.Line - 1

	commentLines := strings.Split(comment, "\n")

	var newLines []string
	newLines = append(newLines, lines[:entityLine]...)
	newLines = append(newLines, commentLines...)
	newLines = append(newLines, lines[entityLine:]...)

	updatedContent := strings.Join(newLines, "\n")

	if formatOutput {
		f := formatter.New()
		if formatted, err := f.FormatWithClang(updatedContent); err == nil {
			updatedContent = formatted
		}
	}

	return os.WriteFile(filepath, []byte(updatedContent), 0644)
}

// getEntitiesNeedingUpdates finds entities that either:
// 1. Are completely undocumented, OR 
// 2. Are documented but missing @ingroup tag (when group is specified)
func getEntitiesNeedingUpdates(filePath string, group *GroupConfig) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	p := parser.New()
	scopeTree, err := p.Parse(filePath, string(content))
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var entitiesToUpdate []string
	processedEntities := make(map[string]bool)

	var traverse func(*ast.Entity)
	traverse = func(entity *ast.Entity) {
		if shouldSkipEntity(entity) {
			return
		}

		hasComment := entity.Comment != nil || hasCommentBefore(lines, entity.SourceRange.Start.Line)

		// Check if entity needs updates
		needsUpdate := false
		
		if !hasComment {
			// Case 1: Completely undocumented
			needsUpdate = true
		} else if group != nil {
			// Case 2: Documented but missing @ingroup tag
			needsUpdate = !hasIngroupTag(lines, entity, group.Name)
		}

		if needsUpdate && !processedEntities[entity.FullName] {
			if entity.Type == ast.EntityConstructor {
				parentName := getParentEntityName(entity.FullName)
				if processedEntities[parentName] {
					return
				}
			}

			entitiesToUpdate = append(entitiesToUpdate, entity.FullName)
			processedEntities[entity.FullName] = true
		}

		for _, child := range entity.Children {
			traverse(child)
		}
	}

	for _, entity := range scopeTree.Root.Children {
		traverse(entity)
	}

	return entitiesToUpdate, nil
}

// hasIngroupTag checks if an entity's comment contains the correct @ingroup tag
func hasIngroupTag(lines []string, entity *ast.Entity, groupName string) bool {
	// Check the entity's immediate comment
	if entity.Comment != nil {
		commentText := entity.Comment.Raw
		return strings.Contains(commentText, "@ingroup") && strings.Contains(commentText, groupName)
	}

	// Check for comment before the entity
	entityLine := entity.SourceRange.Start.Line
	if entityLine <= 1 {
		return false
	}

	// Look backwards for comment blocks
	maxLookback := 15
	for i := entityLine - 2; i >= 0 && i >= entityLine-maxLookback; i-- {
		line := strings.TrimSpace(lines[i])
		
		// Skip empty lines and non-comment lines that might be between comment and entity
		if line == "" || 
		   strings.HasPrefix(line, "template") ||
		   strings.HasPrefix(line, "inline") ||
		   strings.HasPrefix(line, "static") ||
		   strings.HasPrefix(line, "const ") ||
		   strings.HasPrefix(line, "constexpr") ||
		   strings.HasPrefix(line, "[[") {
			continue
		}

		// If we found a comment line, check for @ingroup
		if strings.Contains(line, "/**") || strings.Contains(line, "*/") || 
		   strings.Contains(line, "*") || strings.HasPrefix(line, "///") {
			if strings.Contains(line, "@ingroup") && strings.Contains(line, groupName) {
				return true
			}
			continue
		}

		// If we hit actual code, stop looking
		break
	}

	return false
}

// postProcessComment adds structural Doxygen tags to LLM-generated comments
// This keeps the LLM service focused purely on content generation
func postProcessComment(rawComment string, group *GroupConfig) string {
	if group == nil {
		return rawComment
	}

	// Check if @ingroup is already present
	if strings.Contains(rawComment, "@ingroup") {
		return rawComment
	}

	// Find the insertion point for @ingroup (before the closing */)
	lines := strings.Split(rawComment, "\n")
	
	// Look for the last line with content before */
	insertIndex := len(lines) - 1
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "*/" {
			insertIndex = i
			break
		}
	}

	// Insert @ingroup before the closing */
	ingroupLine := fmt.Sprintf(" * @ingroup %s", group.Name)
	
	newLines := make([]string, len(lines)+1)
	copy(newLines[:insertIndex], lines[:insertIndex])
	newLines[insertIndex] = ingroupLine
	copy(newLines[insertIndex+1:], lines[insertIndex:])

	return strings.Join(newLines, "\n")
}

// isEntityUndocumented checks if an entity has no documentation at all
func isEntityUndocumented(filePath, entityPath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return true // Assume undocumented if can't read file
	}

	p := parser.New()
	scopeTree, err := p.Parse(filePath, string(content))
	if err != nil {
		return true // Assume undocumented if can't parse
	}

	entity := findEntityByPath(scopeTree, entityPath)
	if entity == nil {
		return true // Entity not found, assume undocumented
	}

	lines := strings.Split(string(content), "\n")
	hasComment := entity.Comment != nil || hasCommentBefore(lines, entity.SourceRange.Start.Line)
	
	return !hasComment
}

// generateDocumentationForEntity generates new documentation for an undocumented entity
func generateDocumentationForEntity(filePath, entityPath string, service *llm.DocumentationService, rootPath string, group *GroupConfig) error {
	// Extract context
	entityContext, err := extractEntityContext(filePath, entityPath)
	if err != nil {
		return fmt.Errorf("failed to extract context: %v", err)
	}

	// Get entity type
	entityType := getEntityTypeFromName(entityPath)

	// Read additional context
	additionalContext, _ := readDoxyllmContext(filePath, rootPath)

	// Create documentation request (no GroupInfo - handled by post-processor)
	docRequest := llm.DocumentationRequest{
		EntityName:        entityPath,
		EntityType:        entityType,
		Context:           entityContext,
		AdditionalContext: additionalContext,
	}

	// Generate documentation
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := service.GenerateDocumentation(ctx, docRequest)
	if err != nil {
		return fmt.Errorf("failed to generate comment: %v", err)
	}

	// Post-process the comment to add structural tags like @ingroup
	finalComment := postProcessComment(result.Comment, group)

	// Update the file
	return updateEntityComment(filePath, entityPath, finalComment)
}

// addIngroupToExistingComment adds @ingroup tag to an existing comment
func addIngroupToExistingComment(filePath, entityPath string, group *GroupConfig) error {
	if group == nil {
		return nil // Nothing to add
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	p := parser.New()
	scopeTree, err := p.Parse(filePath, string(content))
	if err != nil {
		return err
	}

	entity := findEntityByPath(scopeTree, entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	lines := strings.Split(string(content), "\n")
	
	// Find the existing comment
	if entity.Comment != nil {
		// Entity has a parsed comment - modify it directly
		return addIngroupToComment(filePath, entity.Comment.Raw, group.Name)
	} else if hasCommentBefore(lines, entity.SourceRange.Start.Line) {
		// Find and modify the comment before the entity
		return addIngroupToCommentBefore(filePath, entity.SourceRange.Start.Line, group.Name)
	}

	return fmt.Errorf("no existing comment found for entity: %s", entityPath)
}

// addIngroupToComment adds @ingroup to a specific comment block
func addIngroupToComment(filePath, commentText, groupName string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Find and replace the comment in the file
	updatedContent := strings.Replace(string(content), commentText, addIngroupToCommentText(commentText, groupName), 1)
	
	return os.WriteFile(filePath, []byte(updatedContent), 0644)
}

// addIngroupToCommentBefore adds @ingroup to a comment block before a specific line
func addIngroupToCommentBefore(filePath string, entityLine int, groupName string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	
	// Find the comment block before the entity
	commentEnd := -1
	commentStart := -1
	
	// Look backwards for comment end
	for i := entityLine - 2; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		if strings.HasSuffix(line, "*/") {
			commentEnd = i
			break
		}
		
		// If we hit non-comment code, stop
		if !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "//") {
			break
		}
	}
	
	if commentEnd == -1 {
		return fmt.Errorf("comment end not found")
	}
	
	// Find comment start
	for i := commentEnd; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "/**") {
			commentStart = i
			break
		}
	}
	
	if commentStart == -1 {
		return fmt.Errorf("comment start not found")
	}
	
	// Insert @ingroup before the closing */
	ingroupLine := fmt.Sprintf(" * @ingroup %s", groupName)
	
	newLines := make([]string, len(lines)+1)
	copy(newLines[:commentEnd], lines[:commentEnd])
	newLines[commentEnd] = ingroupLine
	copy(newLines[commentEnd+1:], lines[commentEnd:])
	
	updatedContent := strings.Join(newLines, "\n")
	return os.WriteFile(filePath, []byte(updatedContent), 0644)
}

// addIngroupToCommentText adds @ingroup to a comment text string
func addIngroupToCommentText(commentText, groupName string) string {
	if strings.Contains(commentText, "@ingroup") {
		return commentText // Already has @ingroup
	}

	lines := strings.Split(commentText, "\n")
	
	// Find insertion point (before closing */)
	insertIndex := len(lines) - 1
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "*/" {
			insertIndex = i
			break
		}
	}

	// Insert @ingroup
	ingroupLine := fmt.Sprintf(" * @ingroup %s", groupName)
	
	newLines := make([]string, len(lines)+1)
	copy(newLines[:insertIndex], lines[:insertIndex])
	newLines[insertIndex] = ingroupLine
	copy(newLines[insertIndex+1:], lines[insertIndex:])

	return strings.Join(newLines, "\n")
}
