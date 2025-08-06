package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"doxyllm-it/pkg/ast"
	"doxyllm-it/pkg/formatter"
	"doxyllm-it/pkg/parser"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// OllamaRequest represents the request structure for Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents the response structure from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// OllamaConfig holds configuration for Ollama integration
type OllamaConfig struct {
	URL         string
	Model       string
	Temperature float64
	TopP        float64
	NumCtx      int
	Timeout     time.Duration
}

// DoxyllmConfig represents the structure of a .doxyllm configuration file
type DoxyllmConfig struct {
	Global string            `yaml:"global,omitempty"`
	Files  map[string]string `yaml:"files,omitempty"`
}

// readDoxyllmContext reads a .doxyllm file from the same directory as the file being processed
// Returns the context content (global + file-specific), or empty string if no file is found
func readDoxyllmContext(filePath string) string {
	dir := filepath.Dir(filePath)
	doxyllmPath := filepath.Join(dir, ".doxyllm")

	content, err := os.ReadFile(doxyllmPath)
	if err != nil {
		// File doesn't exist or can't be read, return empty string
		return ""
	}

	// Try to parse as YAML first
	var config DoxyllmConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		// If YAML parsing fails, treat as plain text (backward compatibility)
		return strings.TrimSpace(string(content))
	}

	// Build context from YAML structure
	var contextParts []string

	// Add global context if present
	if config.Global != "" {
		contextParts = append(contextParts, config.Global)
	}

	// Add file-specific context if present
	fileName := filepath.Base(filePath)
	if fileContext, exists := config.Files[fileName]; exists && fileContext != "" {
		contextParts = append(contextParts, fmt.Sprintf("SPECIFIC TO %s:\n%s", fileName, fileContext))
	}

	return strings.Join(contextParts, "\n\n")
}

const promptTemplate = `You are a C++ documentation expert. Generate a comprehensive Doxygen comment for ONLY the specific entity requested.

CRITICAL INSTRUCTIONS:
- Document ONLY the target entity: %s
- Do NOT document any child entities, member functions, or other entities shown in the context
- Use proper Doxygen tags (@brief, @param, @return, @throws, etc.)
- For namespaces: Focus on the purpose and scope of the namespace
- For classes: Focus on the class responsibility and main purpose
- For functions: Document parameters, return value, and behavior
- Generate ONLY the Doxygen comment block (starting with /** and ending with */)
- Do not include any code, explanations, or markdown formatting

%s

Context for understanding:
` + "```cpp\n%s\n```" + `

TARGET ENTITY TO DOCUMENT: %s
Type: %s

Generate a focused Doxygen comment for ONLY this specific entity.
Response format: /** ... */`

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
	Run:  runOllama,
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
	ollamaCmd.Flags().StringVarP(&ollamaURL, "url", "u", getEnvOrDefault("OLLAMA_URL", "http://localhost:11434/api/generate"), "Ollama API URL")
	ollamaCmd.Flags().StringVarP(&ollamaModel, "model", "m", getEnvOrDefault("MODEL_NAME", "codellama:13b"), "Ollama model name")
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

func runOllama(cmd *cobra.Command, args []string) {
	target := args[0]

	// Create Ollama configuration
	config := &OllamaConfig{
		URL:         ollamaURL,
		Model:       ollamaModel,
		Temperature: temperature,
		TopP:        topP,
		NumCtx:      numCtx,
		Timeout:     time.Duration(timeout) * time.Second,
	}

	// Test Ollama connectivity
	if !testOllamaConnection(config) {
		os.Exit(1)
	}

	fmt.Printf("🤖 Connected to Ollama at: %s\n", config.URL)
	fmt.Printf("📚 Using model: %s\n", config.Model)

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
		updates := processFileWithOllama(file, config)
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

func testOllamaConnection(config *OllamaConfig) bool {
	// Test with /api/tags endpoint first
	tagsURL := strings.Replace(config.URL, "/api/generate", "/api/tags", 1)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(tagsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Cannot connect to Ollama at: %s\n", config.URL)
		fmt.Fprintf(os.Stderr, "   Please ensure Ollama is running and accessible\n")
		fmt.Fprintf(os.Stderr, "   Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "❌ Ollama responded with status: %d\n", resp.StatusCode)
		return false
	}

	return true
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

func processFileWithOllama(filepath string, config *OllamaConfig) int {
	fmt.Printf("\n📁 Processing: %s\n", filepath)

	// Parse file to get undocumented entities
	undocumented, err := getUndocumentedEntities(filepath)
	if err != nil {
		fmt.Printf("  ❌ Error parsing file: %v\n", err)
		return 0
	}

	if len(undocumented) == 0 {
		fmt.Println("  ✅ All entities already documented")
		return 0
	}

	fmt.Printf("  📋 Found %d undocumented entities\n", len(undocumented))

	// Limit entities if specified
	if maxEntities > 0 && len(undocumented) > maxEntities {
		undocumented = undocumented[:maxEntities]
		fmt.Printf("  🔢 Processing first %d entities\n", len(undocumented))
	}

	if dryRun {
		for i, entity := range undocumented {
			fmt.Printf("  📝 (%d/%d) Would document: %s\n", i+1, len(undocumented), entity)
		}
		return len(undocumented)
	}

	successfulUpdates := 0

	for i, entityPath := range undocumented {
		fmt.Printf("  📝 (%d/%d) Documenting: %s\n", i+1, len(undocumented), entityPath)

		// Extract context
		context, err := extractEntityContext(filepath, entityPath)
		if err != nil {
			fmt.Printf("    ❌ Failed to extract context: %v\n", err)
			continue
		}

		// Generate comment using Ollama
		fmt.Printf("    🤖 Generating comment with %s...\n", config.Model)
		comment, err := generateComment(context, entityPath, filepath, config)
		if err != nil {
			fmt.Printf("    ❌ Failed to generate comment: %v\n", err)
			continue
		}

		// Update the file
		if err := updateEntityComment(filepath, entityPath, comment); err != nil {
			fmt.Printf("    ❌ Failed to update file: %v\n", err)
			continue
		}

		fmt.Println("    ✅ Successfully updated")
		successfulUpdates++
	}

	fmt.Printf("  📊 Updated %d/%d entities\n", successfulUpdates, len(undocumented))
	return successfulUpdates
}

func getUndocumentedEntities(filepath string) ([]string, error) {
	// Read the file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Parse the file
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
		// Skip template parameters and single-letter entities
		if shouldSkipEntity(entity) {
			return
		}

		// Check if entity has a comment or if there's a comment immediately before it
		hasComment := entity.Comment != nil || hasCommentBefore(lines, entity.SourceRange.Start.Line)

		if !hasComment && !documentedEntities[entity.FullName] {
			// For constructors, check if parent class is already documented or will be documented
			if entity.Type == ast.EntityConstructor {
				parentName := getParentEntityName(entity.FullName)
				if documentedEntities[parentName] {
					return // Skip constructor if parent class is documented
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

// hasCommentBefore checks if there's a Doxygen comment immediately before the given line
func hasCommentBefore(lines []string, entityLine int) bool {
	if entityLine <= 1 {
		return false
	}

	// Look back from the entity line to find the last non-empty line
	for i := entityLine - 2; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue // Skip empty lines
		}

		// Check if this line ends a Doxygen comment
		if strings.HasSuffix(line, "*/") {
			// Look backwards to find the start of the comment block
			for j := i; j >= 0; j-- {
				commentLine := strings.TrimSpace(lines[j])
				if strings.HasPrefix(commentLine, "/**") {
					return true
				}
				if !strings.HasPrefix(commentLine, "*") && !strings.HasPrefix(commentLine, "*/") {
					break
				}
			}
		}

		// If we hit a non-comment line, stop looking
		if !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "/*") && !strings.HasPrefix(line, "*") {
			break
		}
	}

	return false
}

// shouldSkipEntity determines if an entity should be skipped during documentation
func shouldSkipEntity(entity *ast.Entity) bool {
	// Skip single-letter entities (likely template parameters)
	if len(entity.Name) == 1 && entity.Name >= "A" && entity.Name <= "Z" {
		return true
	}
	if len(entity.Name) == 1 && entity.Name >= "a" && entity.Name <= "z" {
		return true
	}

	// Skip common template parameter names
	commonTemplateParams := map[string]bool{
		"T": true, "U": true, "V": true, "E": true, "N": true, "S": true,
		"Container": true, "ElementType": true, "OtherElementType": true,
		"Count": true, "Offset": true, "Extent": true, "OtherExtent": true,
	}

	if commonTemplateParams[entity.Name] {
		return true
	}

	// Skip very short names that are likely template params
	if len(entity.Name) <= 2 && strings.ToUpper(entity.Name) == entity.Name {
		return true
	}

	// Skip standard library namespaces and common system entities
	systemEntities := map[string]bool{
		"std":       true,
		"__gnu_cxx": true,
		"__detail":  true,
	}

	if systemEntities[entity.Name] {
		return true
	}

	// Skip type aliases that are clearly template-related or system-level
	if entity.Type == ast.EntityTypedef || entity.Type == ast.EntityUsing || entity.Type == ast.EntityVariable {
		// Skip common type aliases like value_type, size_type, etc. within templates
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

// getParentEntityName extracts the parent entity name from a full entity path
func getParentEntityName(fullName string) string {
	parts := strings.Split(fullName, "::")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "::")
}

func extractEntityContext(filepath, entityPath string) (string, error) {
	// Read the file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// Parse the file
	p := parser.New()
	scopeTree, err := p.Parse(filepath, string(content))
	if err != nil {
		return "", err
	}

	entity := findEntityByPath(scopeTree, entityPath)
	if entity == nil {
		return "", fmt.Errorf("entity not found: %s", entityPath)
	}

	// Use different context extraction strategies based on entity type
	f := formatter.New()

	switch entity.Type {
	case ast.EntityNamespace:
		// For namespaces, just show the namespace declaration and immediate children signatures
		return extractNamespaceContext(entity), nil

	case ast.EntityClass, ast.EntityStruct:
		// For classes/structs, show class declaration and public interface
		return extractClassContext(entity), nil

	case ast.EntityFunction, ast.EntityMethod, ast.EntityConstructor, ast.EntityDestructor:
		// For functions, extract minimal surrounding context
		return f.ExtractEntityContext(entity, false, false), nil

	default:
		// For other entities (variables, enums, etc.), use minimal context
		return f.ExtractEntityContext(entity, false, false), nil
	}
}

// extractNamespaceContext creates a focused context for namespace documentation
func extractNamespaceContext(entity *ast.Entity) string {
	var context strings.Builder

	// Show the namespace declaration
	context.WriteString(entity.Signature)
	context.WriteString("\n")

	// Show immediate children (classes, functions, etc.) as signatures only
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

// extractClassContext creates a focused context for class documentation
func extractClassContext(entity *ast.Entity) string {
	var context strings.Builder

	// Show the class declaration
	context.WriteString(entity.Signature)
	context.WriteString("\n")

	// Show public interface summary
	if len(entity.Children) > 0 {
		context.WriteString("  // Public interface:\n")
		publicMethods := 0
		publicFields := 0

		for _, child := range entity.Children {
			if child.AccessLevel == ast.AccessPublic || child.AccessLevel == ast.AccessUnknown {
				switch child.Type {
				case ast.EntityMethod, ast.EntityFunction, ast.EntityConstructor, ast.EntityDestructor:
					if publicMethods < 5 { // Limit to first 5 methods
						context.WriteString(fmt.Sprintf("  %s\n", child.Signature))
					}
					publicMethods++
				case ast.EntityField, ast.EntityVariable:
					if publicFields < 3 { // Limit to first 3 fields
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

func generateComment(context, entityName, filePath string, config *OllamaConfig) (string, error) {
	// Get entity type for better prompt context
	entityType := getEntityTypeFromName(entityName)

	// Read additional context from .doxyllm file if it exists
	additionalContext := readDoxyllmContext(filePath)
	var contextSection string
	if additionalContext != "" {
		contextSection = fmt.Sprintf("ADDITIONAL PROJECT CONTEXT:\n%s\n", additionalContext)
	}

	prompt := fmt.Sprintf(promptTemplate, entityName, contextSection, context, entityName, entityType)

	reqBody := OllamaRequest{
		Model:  config.Model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": config.Temperature,
			"top_p":       config.TopP,
			"num_ctx":     config.NumCtx,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: config.Timeout}
	resp, err := client.Post(config.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	comment := strings.TrimSpace(ollamaResp.Response)

	// Clean up the response - remove any markdown formatting
	if strings.HasPrefix(comment, "```") {
		lines := strings.Split(comment, "\n")
		if len(lines) > 2 {
			comment = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}

	// Remove any leading/trailing code block markers
	comment = strings.TrimPrefix(comment, "```cpp")
	comment = strings.TrimPrefix(comment, "```c++")
	comment = strings.TrimPrefix(comment, "```")
	comment = strings.TrimSuffix(comment, "```")
	comment = strings.TrimSpace(comment)

	// Extract only the first Doxygen comment block if multiple are present
	comment = extractFirstDoxygenComment(comment)

	// Ensure proper Doxygen format
	if !strings.HasPrefix(comment, "/**") {
		comment = "/**\n * " + strings.TrimPrefix(comment, "* ")
	}
	if !strings.HasSuffix(comment, "*/") {
		comment = strings.TrimSuffix(comment, " ") + "\n */"
	}

	// Clean up common formatting issues
	comment = strings.ReplaceAll(comment, "**/", "*/")
	comment = strings.ReplaceAll(comment, "  **/", " */")

	return comment, nil
}

// extractFirstDoxygenComment extracts only the first Doxygen comment from potentially multiple comments
func extractFirstDoxygenComment(text string) string {
	lines := strings.Split(text, "\n")
	var commentLines []string
	inComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "/**") {
			if inComment {
				// If we're already in a comment and find another /**, stop at the previous one
				break
			}
			inComment = true
			commentLines = append(commentLines, line)
		} else if inComment {
			commentLines = append(commentLines, line)
			if strings.HasSuffix(trimmed, "*/") {
				// End of comment block
				break
			}
		}
	}

	if len(commentLines) > 0 {
		return strings.Join(commentLines, "\n")
	}

	// If no proper Doxygen comment found, return cleaned up version of original
	return strings.TrimSpace(text)
}

// getEntityTypeFromName determines the entity type from the full name
func getEntityTypeFromName(fullName string) string {
	if strings.Contains(fullName, "::") {
		parts := strings.Split(fullName, "::")
		lastPart := parts[len(parts)-1]

		// Heuristics to determine type
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

// isCapitalized checks if a string starts with a capital letter
func isCapitalized(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := rune(s[0])
	return first >= 'A' && first <= 'Z'
}

func updateEntityComment(filepath, entityPath, comment string) error {
	// Read the file
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Parse the file
	p := parser.New()
	scopeTree, err := p.Parse(filepath, string(content))
	if err != nil {
		return err
	}

	entity := findEntityByPath(scopeTree, entityPath)
	if entity == nil {
		return fmt.Errorf("entity not found: %s", entityPath)
	}

	// Create backup if requested
	if backup {
		backupPath := filepath + ".bak"
		if err := os.WriteFile(backupPath, content, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
	}

	// Insert comment before the entity
	lines := strings.Split(string(content), "\n")
	entityLine := entity.SourceRange.Start.Line - 1 // Convert to 0-based index

	// Insert the comment lines before the entity
	commentLines := strings.Split(comment, "\n")

	// Build new content
	var newLines []string
	newLines = append(newLines, lines[:entityLine]...)
	newLines = append(newLines, commentLines...)
	newLines = append(newLines, lines[entityLine:]...)

	updatedContent := strings.Join(newLines, "\n")

	// Format if requested
	if formatOutput {
		f := formatter.New()
		if formatted, err := f.FormatWithClang(updatedContent); err == nil {
			updatedContent = formatted
		}
	}

	// Write updated content
	return os.WriteFile(filepath, []byte(updatedContent), 0644)
}
