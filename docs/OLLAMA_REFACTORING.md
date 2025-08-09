# Ollama Command Refactoring with Document Abstraction

## Overview

This document describes the refactoring of the `ollama` command to use the new document abstraction layer, following the single responsibility principle and improving testability.

## Architecture Changes

### Before (Original Implementation)

The original `ollama` command in `cmd/ollama.go` contained:

- **Direct file manipulation**: Reading, parsing, and writing files directly
- **Mixed responsibilities**: LLM integration, file processing, AST manipulation, and formatting all in one place
- **Hard to test**: Tightly coupled components made unit testing difficult
- **Complex error handling**: Error handling scattered throughout the code

### After (Refactored Implementation)

The refactored implementation follows a layered architecture:

1. **Command Layer** (`cmd/ollama_v2.go`): CLI interface and orchestration
2. **Service Layer** (`pkg/document/service.go`): High-level business logic
3. **Document Layer** (`pkg/document/document.go`): Document abstraction
4. **LLM Layer** (`pkg/llm/*`): LLM provider abstraction

## Key Benefits

### 1. Single Responsibility Principle

Each component now has a single, well-defined responsibility:

- **DocumentationService**: Process entities and manage documentation workflow
- **Document**: Represent and manipulate C++ files with their AST
- **LLM Service**: Handle LLM communication and response processing
- **Command**: Handle CLI interaction and orchestration

### 2. Improved Testability

- **Mock interfaces**: Easy to mock LLM services for testing
- **Isolated components**: Each layer can be tested independently
- **Dependency injection**: Services accept interfaces, enabling test doubles
- **Unit tests**: Comprehensive test coverage for each component

### 3. Better Error Handling

- **Structured errors**: Clear error types and messages
- **Graceful degradation**: Failures in one entity don't stop processing others
- **Error collection**: Non-fatal errors are collected and reported

### 4. Enhanced Maintainability

- **Clear boundaries**: Well-defined interfaces between components
- **Extensibility**: Easy to add new features or modify existing ones
- **Code reuse**: Components can be reused across different commands

## Implementation Details

### Document Service Layer

```go
type DocumentationService struct {
    llmService LLMService
}

func (s *DocumentationService) ProcessUndocumentedEntities(
    ctx context.Context, 
    doc *Document, 
    opts ProcessingOptions
) (*ProcessingResult, error)
```

**Responsibilities**:
- Identify undocumented entities
- Generate documentation using LLM service
- Apply group configurations (@ingroup tags)
- Handle batch processing with error collection

### Document Abstraction

```go
type Document struct {
    filename    string
    content     string
    tree        *ast.ScopeTree
    // ... other fields
}

func (d *Document) SetEntityComment(entityPath string, comment *ast.DoxygenComment) error
func (d *Document) GetUndocumentedEntities() []*ast.Entity
func (d *Document) Save() error
```

**Responsibilities**:
- Represent C++ files with parsed AST
- Provide high-level operations for documentation manipulation
- Handle file I/O and formatting
- Maintain consistency between AST and file content

### Command Layer

```go
func runOllamaNew(cmd *cobra.Command, args []string) {
    // 1. Setup LLM service
    // 2. Process each file using DocumentationService
    // 3. Report results
}

func processFileV2(filePath string, docService *DocumentationService, rootPath string) *ProcessingResult {
    // 1. Load document
    // 2. Process entities using service
    // 3. Save results
}
```

**Responsibilities**:
- Handle CLI arguments and flags
- Orchestrate the documentation process
- Provide user feedback and progress reporting
- Manage file operations and backups

## Testing Strategy

### Unit Tests

Each component has comprehensive unit tests:

1. **Service Tests** (`pkg/document/service_test.go`):
   - Mock LLM service for fast, isolated testing
   - Test entity processing logic
   - Verify error handling and edge cases

2. **Document Tests** (`pkg/document/document_test.go`):
   - Test document manipulation operations
   - Verify AST consistency
   - Test file I/O operations

3. **Command Tests** (`cmd/ollama_v2_test.go`):
   - Test CLI argument handling
   - Test file discovery and filtering
   - Integration tests with mock services

### Integration Tests

The tests demonstrate real-world usage:

```go
func TestProcessFileV2_Integration(t *testing.T) {
    // Create test C++ file
    // Setup mock LLM service
    // Process file and verify results
}
```

### Benchmarks

Performance benchmarks ensure the refactoring doesn't impact performance:

```go
func BenchmarkProcessFileV2(b *testing.B) {
    // Benchmark file processing with 100 entities
}
```

## Usage Examples

### Basic Usage

```bash
# Process a single file
doxyllm-it ollama-new examples/test.hpp

# Process directory with dry run
doxyllm-it ollama-new --dry-run --max-entities 5 src/

# Use custom model and settings
doxyllm-it ollama-new --model codellama:13b --temperature 0.2 src/
```

### Configuration

The refactored command uses the same `.doxyllm.yaml` configuration:

```yaml
global: "Project-wide context"
groups:
  mygroup:
    name: "mygroup"
    title: "My Component Group"
    files: ["*.hpp"]
    generateDefgroup: true
```

## Migration Path

### For Users

1. **Current command remains unchanged**: The original `ollama` command continues to work
2. **New command available**: `ollama-new` provides the refactored implementation
3. **Same configuration**: Uses the existing `.doxyllm.yaml` files
4. **Same output**: Produces the same documentation results

### For Developers

1. **Study the new architecture**: Understand the layered approach
2. **Use document abstraction**: Leverage `pkg/document` for new features
3. **Write tests**: Follow the testing patterns established
4. **Extend services**: Add new functionality through service interfaces

## Future Enhancements

The new architecture enables several future improvements:

1. **Multiple LLM providers**: Easy to add OpenAI, Anthropic, etc.
2. **Batch processing**: Process multiple files in parallel
3. **Custom documentation templates**: Configurable comment styles
4. **Documentation validation**: Verify generated comments meet standards
5. **IDE integration**: Provide language server protocol support

## Performance Considerations

The refactored implementation maintains performance while improving structure:

- **Lazy loading**: Documents are only parsed when needed
- **Efficient caching**: Entity lookups use hash maps for O(1) access
- **Minimal memory overhead**: Smart AST management prevents memory leaks
- **Parallel processing**: Ready for concurrent file processing

## Conclusion

The refactoring successfully achieves the goals of:
- ✅ **Single Responsibility**: Each component has a clear, focused purpose
- ✅ **Improved Testability**: Comprehensive test coverage with fast unit tests
- ✅ **Better Architecture**: Clean separation of concerns and dependency injection
- ✅ **Enhanced Maintainability**: Easy to understand, modify, and extend

The new implementation serves as a model for future command implementations and demonstrates best practices for Go application architecture.
