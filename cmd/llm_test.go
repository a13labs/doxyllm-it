package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"doxyllm-it/pkg/document"
	"doxyllm-it/pkg/llm"
)

// mockLLMProvider implements a mock LLM provider for testing
type mockLLMProvider struct {
	generateFunc func(ctx context.Context, req llm.CommentRequest) (*llm.CommentResponse, error)
	connected    bool
	modelName    string
}

func (m *mockLLMProvider) GenerateComment(ctx context.Context, req llm.CommentRequest) (*llm.CommentResponse, error) {
	if m.generateFunc != nil {
		return m.generateFunc(ctx, req)
	}
	return &llm.CommentResponse{
		Description: "/** @brief Mock comment for " + req.EntityName + " */",
		Metadata:    make(map[string]string),
	}, nil
}

func (m *mockLLMProvider) TestConnection(ctx context.Context) error {
	if !m.connected {
		return &llm.ProviderError{Provider: "mock", Message: "Mock connection failed"}
	}
	return nil
}

func (m *mockLLMProvider) GetModelInfo() llm.ModelInfo {
	return llm.ModelInfo{Name: m.modelName}
}

// mockLLMDocumentationService wraps the mock provider
type mockLLMDocumentationService struct {
	provider llm.Provider
}

func (m *mockLLMDocumentationService) GenerateDocumentation(ctx context.Context, req llm.DocumentationRequest) (*llm.DocumentationResult, error) {
	response, err := m.provider.GenerateComment(ctx, llm.CommentRequest{
		EntityName:        req.EntityName,
		EntityType:        req.EntityType,
		Context:           req.Context,
		AdditionalContext: req.AdditionalContext,
	})
	if err != nil {
		return nil, err
	}

	return &llm.DocumentationResult{
		Comment:     response.Description,
		Description: response.Description,
		Metadata:    response.Metadata,
	}, nil
}

func (m *mockLLMDocumentationService) TestConnection(ctx context.Context) error {
	return m.provider.TestConnection(ctx)
}

func (m *mockLLMDocumentationService) GetModelInfo() llm.ModelInfo {
	return m.provider.GetModelInfo()
}

func TestLLMCommandFlags(t *testing.T) {
	// Test that the llm command has all expected flags
	cmd := llmCmd

	// Check that the command is properly configured
	if cmd.Use != "llm [flags] <file_or_directory>" {
		t.Errorf("Expected llm command Use to be 'llm [flags] <file_or_directory>', got '%s'", cmd.Use)
	}

	// Check that flags are available
	flags := cmd.Flags()

	expectedFlags := []string{
		"provider", "url", "model", "api-key", "temperature", "top-p", "context", "timeout",
		"max-entities", "dry-run", "backup", "format", "exclude",
	}

	for _, flagName := range expectedFlags {
		if flags.Lookup(flagName) == nil {
			t.Errorf("Expected flag '%s' to be defined", flagName)
		}
	}
}

func TestFindCppFilesForLLM(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "doxyllm-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"test.hpp",
		"test.h",
		"test.hxx",
		"test.cpp",       // Should be ignored
		"test.txt",       // Should be ignored
		"build/test.hpp", // Should be ignored (in excluded dir)
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		if strings.Contains(file, "/") {
			os.MkdirAll(filepath.Dir(filePath), 0755)
		}
		os.WriteFile(filePath, []byte("// test content"), 0644)
	}

	// Test directory processing
	files, err := findCppFiles(tempDir)
	if err != nil {
		t.Fatalf("findCppFiles failed: %v", err)
	}

	// Should find 3 header files (excluding .cpp, .txt, and build directory)
	expectedCount := 3
	if len(files) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(files))
		for _, file := range files {
			if !isCppHeader(file) {
				t.Errorf("Non-header file found: %s", file)
			}
		}
	}

	// Test single file processing
	singleFile := filepath.Join(tempDir, "single.hpp")
	os.WriteFile(singleFile, []byte("// single file"), 0644)
	files, err = findCppFiles(singleFile)
	if err != nil {
		t.Fatalf("findCppFiles failed for single file: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file for single file input, got %d", len(files))
	}
}

func TestIsCppHeader(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.hpp", true},
		{"test.h", true},
		{"test.hxx", true},
		{"test.cpp", false},
		{"test.cc", false},
		{"test.txt", false},
		{"test", false},
		{"TEST.HPP", true}, // Should handle uppercase
	}

	for _, tt := range tests {
		result := isCppHeader(tt.filename)
		if result != tt.expected {
			t.Errorf("isCppHeader(%s) = %v, expected %v", tt.filename, result, tt.expected)
		}
	}
}

func TestLoadDoxyllmConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "doxyllm-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test configuration file
	configContent := `
global: "Global context for the project"
files:
  test.hpp: "Specific context for test.hpp"
ignore:
  - "*.tmp"
  - "temp_*"
groups:
  testgroup:
    name: "testgroup"
    title: "Test Group"
    description: "A test group for documentation"
    files:
      - "*.hpp"
    generateDefGroup: true
`

	configPath := filepath.Join(tempDir, ".doxyllm.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test loading the configuration
	testFile := filepath.Join(tempDir, "test.hpp")
	config := loadDoxyllmConfig(testFile, tempDir)

	if config == nil {
		t.Fatal("Expected configuration to be loaded")
	}

	if config.Global != "Global context for the project" {
		t.Errorf("Expected global context to be loaded, got: %s", config.Global)
	}

	if len(config.Ignore) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(config.Ignore))
	}

	if config.Groups == nil || len(config.Groups) != 1 {
		t.Error("Expected groups to be loaded")
	}

	if group := config.Groups["testgroup"]; group == nil {
		t.Error("Expected testgroup to be loaded")
	} else {
		if group.Name != "testgroup" {
			t.Errorf("Expected group name 'testgroup', got '%s'", group.Name)
		}
		if !group.GenerateDefgroup {
			t.Error("Expected generateDefgroup to be true")
		}
	}
}

func TestShouldIgnoreFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "doxyllm-ignore-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a configuration file with ignore patterns
	configContent := `
ignore:
  - "temp_*"
  - "*.tmp"
  - "ignored.hpp"
`

	configPath := filepath.Join(tempDir, ".doxyllm.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.hpp", false},
		{"temp_file.hpp", true},
		{"file.tmp", true},
		{"ignored.hpp", true},
		{"normal.h", false},
	}

	for _, tt := range tests {
		filePath := filepath.Join(tempDir, tt.filename)
		result := shouldIgnoreFile(filePath, tempDir)
		if result != tt.expected {
			t.Errorf("shouldIgnoreFile(%s) = %v, expected %v", tt.filename, result, tt.expected)
		}
	}
}

func TestGetGroupForFile(t *testing.T) {
	config := &DoxyllmConfig{
		Groups: map[string]*document.GroupConfig{
			"headers": {
				Name:             "headers",
				Title:            "Header Files",
				Files:            []string{"*.hpp", "*.h"},
				GenerateDefgroup: true,
			},
			"utils": {
				Name:  "utils",
				Title: "Utility Files",
				Files: []string{"utils/*"},
			},
		},
	}

	tests := []struct {
		filePath     string
		expectedName string
	}{
		{"test.hpp", "headers"},
		{"test.h", "headers"},
		{"utils/helper.hpp", "utils"},
		{"src/main.cpp", ""},
	}

	for _, tt := range tests {
		group := getGroupForFile(tt.filePath, ".", config)
		if tt.expectedName == "" {
			if group != nil {
				t.Errorf("Expected no group for %s, got %s", tt.filePath, group.Name)
			}
		} else {
			if group == nil {
				t.Errorf("Expected group %s for %s, got nil", tt.expectedName, tt.filePath)
			} else if group.Name != tt.expectedName {
				t.Errorf("Expected group %s for %s, got %s", tt.expectedName, tt.filePath, group.Name)
			}
		}
	}
}

func TestProcessFile_Integration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "doxyllm-process-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test C++ file with undocumented entities
	testContent := `
namespace TestNamespace {
    class TestClass {
    public:
        void undocumentedMethod();
        
        /** @brief Already documented method */
        void documentedMethod();
        
        int undocumentedField;
    };
}
`

	testFile := filepath.Join(tempDir, "test.hpp")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a mock LLM service
	mockProvider := &mockLLMProvider{
		connected: true,
		modelName: "test-model",
		generateFunc: func(ctx context.Context, req llm.CommentRequest) (*llm.CommentResponse, error) {
			return &llm.CommentResponse{
				Description: "/** @brief Generated comment for " + req.EntityName + " */",
				Metadata:    make(map[string]string),
			}, nil
		},
	}

	mockLLMService := &mockLLMDocumentationService{provider: mockProvider}
	docService := document.NewDocumentationService(mockLLMService)

	// Set up test flags using the actual flag variables
	originalDryRun := llmDryRun
	originalMaxEntities := llmMaxEntities
	defer func() {
		llmDryRun = originalDryRun
		llmMaxEntities = originalMaxEntities
	}()

	llmDryRun = true // Use dry run to avoid file modification in tests
	llmMaxEntities = 0

	// Process the file
	result := processFile(testFile, docService, tempDir)

	// Verify the result
	if result.EntitiesProcessed == 0 {
		t.Error("Expected some entities to be processed")
	}

	if len(result.UpdatedEntities) == 0 {
		t.Error("Expected some entities to be updated")
	}

	// Verify that undocumented entities were identified
	found := false
	for _, entity := range result.UpdatedEntities {
		if strings.Contains(entity, "undocumentedMethod") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected undocumentedMethod to be in the update list")
	}
}

func TestProcessFile_WithErrors(t *testing.T) {
	// Create a mock LLM service that returns errors
	mockProvider := &mockLLMProvider{
		connected: true,
		modelName: "test-model",
		generateFunc: func(ctx context.Context, req llm.CommentRequest) (*llm.CommentResponse, error) {
			return nil, &llm.ProviderError{Provider: "mock", Message: "Mock generation error"}
		},
	}

	mockLLMService := &mockLLMDocumentationService{provider: mockProvider}
	docService := document.NewDocumentationService(mockLLMService)

	// Process a non-existent file
	result := processFile("/non/existent/file.hpp", docService, "/tmp")

	// Should handle the error gracefully
	if result.EntitiesProcessed != 0 {
		t.Error("Expected no entities to be processed for non-existent file")
	}
}

// Benchmark for performance testing
func BenchmarkProcessFile(b *testing.B) {
	// Create a temporary file for benchmarking
	tempDir, err := os.MkdirTemp("", "doxyllm-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a larger test file
	var contentBuilder strings.Builder
	contentBuilder.WriteString("namespace BenchmarkNamespace {\n")
	for i := 0; i < 100; i++ {
		contentBuilder.WriteString(fmt.Sprintf("    void function%d();\n", i))
	}
	contentBuilder.WriteString("}\n")

	testFile := filepath.Join(tempDir, "benchmark.hpp")
	err = os.WriteFile(testFile, []byte(contentBuilder.String()), 0644)
	if err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}

	// Create a fast mock LLM service
	mockProvider := &mockLLMProvider{
		connected: true,
		modelName: "benchmark-model",
		generateFunc: func(ctx context.Context, req llm.CommentRequest) (*llm.CommentResponse, error) {
			return &llm.CommentResponse{
				Description: "/** @brief Quick comment */",
				Metadata:    make(map[string]string),
			}, nil
		},
	}

	mockLLMService := &mockLLMDocumentationService{provider: mockProvider}
	docService := document.NewDocumentationService(mockLLMService)

	// Set up for benchmark
	originalDryRun := llmDryRun
	originalMaxEntities := llmMaxEntities
	defer func() {
		llmDryRun = originalDryRun
		llmMaxEntities = originalMaxEntities
	}()

	llmDryRun = true
	llmMaxEntities = 10 // Limit for faster benchmarks

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = processFile(testFile, docService, tempDir)
	}
}
