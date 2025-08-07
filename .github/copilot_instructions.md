# GitHub Copilot Instructions for DoxLLM-IT

## Project Overview

DoxLLM-IT is a CLI tool written in Go that parses C++ header files and creates a tree structure of documentable entities while preserving original formatting. It's designed to enable feeding specific code contexts to LLMs for generating Doxygen comments without losing code integrity.

## Project Architecture

### Core Components

1. **`pkg/ast/`** - Abstract Syntax Tree definitions
   - Defines entity types (namespace, class, function, etc.)
   - Manages hierarchical relationships and scope
   - Handles Doxygen comment structures

2. **`pkg/parser/`** - C++ parsing engine
   - Uses regex-based parsing (not full AST compilation)
   - Identifies documentable entities and their signatures
   - Preserves original text and formatting
   - Tracks source positions and ranges

3. **`pkg/document/`** - High-level document abstraction
   - Provides intuitive API for manipulating C++ header files
   - Encapsulates parser and AST complexity behind simple methods
   - Supports batch operations and modification tracking
   - Built-in validation and documentation analysis
   - Entity caching for fast lookup operations

4. **`pkg/formatter/`** - Code reconstruction and formatting
   - Reconstructs original code from parsed tree
   - Handles Doxygen comment formatting
   - Integrates with clang-format for consistency

5. **`pkg/utils/`** - Utility functions
   - Doxygen comment parsing helpers
   - Path manipulation utilities
   - C++ identifier validation

6. **`cmd/`** - CLI commands using Cobra framework
   - `parse` - Parse files and output entity structure
   - `extract` - Extract context for specific entities
   - `format` - Reconstruct and format code
   - `update` - Update files with LLM-generated comments
   - `batch-update` - Batch update multiple entities
   - `llm` - Built-in LLM integration with context-aware documentation (supports Ollama, OpenAI, Anthropic)

### Key Design Principles

- **Non-destructive parsing**: Never lose original code structure
- **Scope-aware**: Understand C++ scope rules and hierarchies with access level tracking
- **LLM-optimized**: Provide targeted context without overwhelming, enhanced with project-specific context
- **Format-preserving**: Maintain code style and formatting
- **Context-aware**: Support for `.doxyllm` configuration files with global and file-specific contexts
- **Modern C++ support**: Enhanced parser for constexpr macros, template functions, and C++20 features
- **Modular design**: Separate parsing, AST, and formatting concerns
- **Comprehensive testing**: Unit tests for all parser components and regex patterns

## Code Patterns and Conventions

### Entity Types
```go
type EntityType int

const (
    EntityNamespace EntityType = iota
    EntityClass
    EntityStruct
    EntityEnum
    EntityFunction
    EntityMethod
    EntityConstructor
    EntityDestructor
    EntityVariable
    EntityField
    EntityTypedef
    EntityUsing
)
```

### Entity Structure
```go
type Entity struct {
    Type         EntityType
    Name         string
    FullName     string           // Fully qualified name
    Signature    string           // Complete declaration
    AccessLevel  AccessLevel      // public/protected/private
    Children     []*Entity        // Child entities
    Parent       *Entity          // Parent entity
    SourceRange  Range           // Position in source
    Comment      *DoxygenComment // Associated documentation
    OriginalText string          // Preserve formatting
}
```

### Common Patterns

1. **Document Creation and Manipulation** (High-level API):
```go
// Create document from file or content
doc, err := document.NewFromFile("header.hpp")
doc, err := document.NewFromContent("header.hpp", content)

// Find entities
entity := doc.FindEntity("Namespace::Class::method")
undocumented := doc.GetUndocumentedEntities()
classes := doc.FindEntitiesByType(ast.EntityClass)

// Add documentation
err := doc.SetEntityBrief("Class::method", "Brief description")
err := doc.AddEntityParam("Class::method", "param", "Parameter description")
err := doc.SetEntityReturn("Class::method", "Return description")

// Batch operations
updates := []document.BatchUpdate{{
    EntityPath: "Class::method",
    Brief:      &brief,
    Params:     map[string]string{"param": "description"},
}}
err := doc.ApplyBatchUpdates(updates)

// Analysis and validation
stats := doc.GetDocumentationStats()
issues := doc.Validate()
```

2. **Tree Navigation** (Low-level AST):
```go
func (e *Entity) GetPath() []string
func (e *Entity) FindByPath(path []string) *Entity
func (tree *ScopeTree) FindEntity(path string) *Entity
```

3. **Context Extraction**:
```go
func (f *Formatter) ExtractEntityContext(entity *Entity, includeParent, includeSiblings bool) string
```

4. **Context Configuration**:
```go
type DoxyllmConfig struct {
    Global string            `yaml:"global,omitempty"`
    Files  map[string]string `yaml:"files,omitempty"`
}
func readDoxyllmContext(filePath string) string
```

5. **Enhanced Parser Patterns**:
```go
// Access level tracking with stack
type accessStackItem struct {
    level AccessLevel
    line  int
}
var accessStack []accessStackItem

// Enhanced regex for modern C++
functionRegex = regexp.MustCompile(`^\s*(?:TCB_SPAN_CONSTEXPR11|TCB_SPAN_NODISCARD|TCB_SPAN_ARRAY_CONSTEXPR|\w+\s+)*(\w+)\s*\([^)]*\)\s*(?:const\s*)?(?:noexcept\s*)?(?:=\s*(?:default|delete)\s*)?;?\s*$`)
```

6. **Safe Name Conversion** for file paths:
```bash
safe_name=$(echo "$entity" | tr ':' '_' | tr ' ' '_')
```

## Development Guidelines

### When Using Document Abstraction (Recommended)
1. Use `document.NewFromFile()` or `document.NewFromContent()` for high-level operations
2. Prefer batch operations with `ApplyBatchUpdates()` for multiple changes
3. Use `GetUndocumentedEntities()` to find documentation gaps
4. Call `GetDocumentationStats()` to track progress
5. Use `Validate()` to check documentation quality
6. Always check `IsModified()` before saving
7. Handle errors gracefully with meaningful messages

### When Working with Document Package
- Use entity caching through `FindEntity()` for fast lookups
- Leverage built-in validation with `Validate()` for quality checks
- Use `GetEntitySummary()` for detailed entity analysis
- Set the `Raw` field when creating comments programmatically
- Use type-specific methods like `AddEntityParam()` for functions
- Combine related updates in batch operations for efficiency

### When Adding New Entity Types
1. Add to `EntityType` enum in `ast/ast.go`
2. Update `String()` method for the enum
3. Add parsing logic in `parser/parser.go`
4. Update regex patterns if needed
5. Add to `GetDocumentableEntities()` if documentable
6. Update document package methods if specific handling needed

### When Adding New Commands
1. Create new file in `cmd/` directory
2. Follow Cobra command structure
3. Add command to `root.go` init function
4. Include proper flag handling and validation
5. Add comprehensive help text
6. Consider using document abstraction for complex operations

### When Adding New LLM Providers
1. Implement the `llm.Provider` interface in `pkg/llm/`
2. Add provider initialization in `llm.NewProvider()`
3. Update command flags and help text in `cmd/llm.go`
4. Add comprehensive tests with mock providers
5. Update documentation with provider-specific examples
6. Ensure proper error handling and connection testing

### When Modifying Parser
- Always preserve `OriginalText` and formatting
- Update both `SourceRange` and `HeaderRange`
- Handle scope stack properly for nested entities
- Update access level tracking with `accessStack`
- Test regex patterns with comprehensive unit tests
- Handle modern C++ constructs (constexpr macros, templates)
- Test with complex C++ constructs

### When Adding Context Features
- Update `DoxyllmConfig` struct if needed
- Maintain backward compatibility with plain text files
- Test YAML parsing and fallback mechanisms
- Ensure global + file-specific context combination works
- Update prompt templates to include new context sections

### Enhanced Parser Testing
- Add unit tests for new regex patterns in `regex_test.go`
- Test access level tracking scenarios
- Validate modern C++ construct detection
- Test scope management with complex nesting
- Ensure entity deduplication works properly

### Error Handling
- Use `fmt.Errorf()` for error wrapping
- Provide context in error messages
- Handle edge cases gracefully
- Validate inputs before processing

## Testing Approach

### Test Files Structure
- `examples/example.hpp` - Comprehensive test case
- `examples/span.hpp` - Real-world complex C++ with templates and macros
- `examples/.doxyllm` - YAML context configuration for testing
- `examples/document_demo/` - Document abstraction usage examples
- `pkg/parser/parser_test.go` - Core parser functionality testing
- `pkg/parser/regex_test.go` - Regex pattern validation tests
- `pkg/document/document_test.go` - Document abstraction layer tests
- `workflow.sh` - Basic workflow demonstration
- `llm_workflow.sh` - Complete LLM integration example

### Test Scenarios
1. **Entity Recognition**: All C++ construct types including modern features
2. **Scope Handling**: Nested namespaces, classes, and access level tracking
3. **Context Extraction**: Parent/sibling relationships with project context
4. **Code Reconstruction**: Lossless formatting preservation
5. **Update Operations**: Comment insertion and formatting
6. **Context Configuration**: YAML parsing, fallback, and context combination
7. **Regex Patterns**: Modern C++ constructs like constexpr macros
8. **Access Level Tracking**: Public/private/protected section management
9. **Document Operations**: High-level API functionality and batch updates
10. **Validation and Analysis**: Documentation quality checks and statistics

## LLM Integration Patterns

### LLM Integration (Built-in)
```bash
# Basic usage with default provider (Ollama)
./doxyllm-it llm --model codegemma:7b --backup header.hpp

# Advanced configuration with Ollama
./doxyllm-it llm --provider ollama --url http://remote:11434 --model deepseek-coder:6.7b \
  --temperature 0.1 --context 8192 --max-entities 5 --dry-run src/

# Use different LLM providers
./doxyllm-it llm --provider openai --model gpt-4 --api-key YOUR_KEY src/
./doxyllm-it llm --provider anthropic --model claude-3-sonnet --api-key YOUR_KEY src/

# Process directories with context files
./doxyllm-it llm --format --exclude build,vendor .
```

### Context Configuration
```yaml
# .doxyllm file structure
global: |
  Project-wide context and design principles...
  
files:
  header.hpp: |
    File-specific implementation details...
  platform.hpp: |
    Platform-specific utilities and macros...
```

### Context Extraction
```bash
# Get minimal context for LLM
./doxyllm-it extract header.hpp "Class::method"

# Get context with parent and siblings
./doxyllm-it extract -p -s header.hpp "Class::method"
```

### Update Patterns
```bash
# Single update from file
./doxyllm-it update -i -b header.hpp "Entity::path" comment.txt

# Batch update from JSON
./doxyllm-it batch-update config.json -i -f
```

### LLM Documentation Patterns
```bash
# Basic LLM documentation with default provider (Ollama)
./doxyllm-it llm --backup header.hpp

# Dry run to preview changes
./doxyllm-it llm --dry-run --max-entities 3 src/

# Use specific providers
./doxyllm-it llm --provider ollama --model deepseek-coder:6.7b src/
./doxyllm-it llm --provider openai --model gpt-4 --api-key YOUR_KEY src/
./doxyllm-it llm --provider anthropic --model claude-3-sonnet --api-key YOUR_KEY src/

# Advanced configuration
./doxyllm-it llm --temperature 0.1 --context 8192 --max-entities 5 \
  --backup --format --exclude build,vendor src/
```
```

### JSON Structure for Batch Updates
```json
{
  "sourceFile": "path/to/header.hpp",
  "updates": [
    {
      "entityPath": "namespace::class::method",
      "comment": "/**\n * @brief Description\n */"
    }
  ]
}
```

## Common Issues and Solutions

### Parser Limitations
- **Template parsing**: Basic recognition with improved support for constexpr templates
- **Preprocessor directives**: Limited handling of #define, #ifdef but improved macro detection
- **Complex inheritance**: Simple inheritance only
- **Modern C++ features**: Enhanced support for C++20 features, may need updates for newest C++23 features

### Enhanced Parser Features
- **Constexpr Macro Support**: Detects `TCB_SPAN_CONSTEXPR11`, `TCB_SPAN_NODISCARD`, `TCB_SPAN_ARRAY_CONSTEXPR`
- **Access Level Tracking**: Stack-based tracking of public/private/protected sections
- **Template Function Detection**: Improved recognition of template function declarations
- **Entity Deduplication**: Prevents duplicate detection with enhanced filtering logic

### Context System
- **YAML Configuration**: Structured context with global and file-specific sections
- **Backward Compatibility**: Plain text files automatically detected and supported
- **Context Combination**: Smart merging of global and file-specific contexts
- **Fallback Handling**: Graceful degradation when context files are missing or malformed

### Scope Resolution
- Use `::` for global scope
- Handle anonymous namespaces carefully
- Track access level changes in classes

### Comment Formatting
- Preserve original indentation
- Handle both `/** */` and `///` styles
- Escape JSON properly for batch updates

## File Organization

```
doxyllm-it/
├── main.go                 # Entry point
├── go.mod                  # Dependencies
├── cmd/                    # CLI commands
├── tmp/                    # Temporary files used during test and developement
├── pkg/                    # Core packages
│   ├── ast/               # AST definitions and structures
│   ├── parser/            # C++ parsing engine
│   ├── document/          # High-level document abstraction
│   ├── formatter/         # Code reconstruction
│   └── utils/             # Utility functions
├── examples/               # Test files and demos
│   └── document_demo/     # Document abstraction examples
└── .github/               # Project metadata
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `gopkg.in/yaml.v2` - YAML parsing for context configuration
- Standard Go libraries for parsing and formatting

## Future Enhancement Areas

1. **Enhanced Template Support**: Better template parameter parsing and constraint detection
2. **Include Analysis**: Parse #include dependencies and cross-file relationships
3. **C++20/23 Features**: Concepts, modules, coroutines, ranges
4. **IDE Integration**: VS Code extension with context-aware documentation
5. **Web Interface**: Browser-based tool for teams with collaborative context editing
6. **Comment Quality**: AI-powered comment quality scoring and suggestions
7. **Context Analytics**: Usage statistics and optimization suggestions for context files
8. **Multi-language Support**: Extend parser for other languages (C, Rust, etc.)
9. **Advanced Context**: Support for inheritance hierarchies and cross-file dependencies in context
10. **Performance Optimization**: Parallel processing for large codebases

## Integration Examples

### LLM Integration (Recommended)
```bash
# Create project context
cat > .doxyllm << 'EOF'
global: |
  C++20 span implementation with backward compatibility.
  Design: zero-overhead, type safety, cross-platform support.
files:
  span.hpp: |
    Template metaprogramming with SFINAE constraints.
    Extensive use of feature detection macros.
EOF

# Auto-generate documentation with default provider (Ollama)
./doxyllm-it llm --model codegemma:7b --backup span.hpp

# Use different providers
./doxyllm-it llm --provider openai --model gpt-4 --api-key YOUR_KEY span.hpp
./doxyllm-it llm --provider anthropic --model claude-3-sonnet --api-key YOUR_KEY span.hpp
```

### Python API Wrapper
```python
import subprocess
import json
import yaml

class DoxLLM:
    def parse(self, header_file):
        result = subprocess.run(['./doxyllm-it', 'parse', '-f', 'json', header_file])
        return json.loads(result.stdout)
    
    def extract_context(self, header_file, entity_path):
        result = subprocess.run(['./doxyllm-it', 'extract', '-p', '-s', header_file, entity_path])
        return result.stdout
    
    def generate_docs(self, header_file, model='codegemma:7b', provider='ollama'):
        subprocess.run(['./doxyllm-it', 'llm', '--provider', provider, '--model', model, '--backup', header_file])
    
    def create_context(self, directory, global_context, file_contexts=None):
        config = {'global': global_context}
        if file_contexts:
            config['files'] = file_contexts
        
        with open(f"{directory}/.doxyllm", 'w') as f:
            yaml.dump(config, f, default_flow_style=False)
```

### Go API Usage (Document Abstraction)
```go
package main

import (
    "fmt"
    "doxyllm-it/pkg/document"
)

func documentHeader(filename string) error {
    // Create document
    doc, err := document.NewFromFile(filename)
    if err != nil {
        return err
    }
    
    // Find undocumented entities
    undocumented := doc.GetUndocumentedEntities()
    fmt.Printf("Found %d undocumented entities\n", len(undocumented))
    
    // Document systematically
    for _, entity := range undocumented {
        path := entity.GetFullPath()
        switch entity.Type {
        case ast.EntityClass:
            doc.SetEntityBrief(path, "TODO: Add class description")
        case ast.EntityMethod, ast.EntityFunction:
            doc.SetEntityBrief(path, "TODO: Add function description")
            doc.SetEntityReturn(path, "TODO: Add return description")
        }
    }
    
    // Get final statistics
    stats := doc.GetDocumentationStats()
    fmt.Printf("Coverage: %.1f%%\n", stats.DocumentationCoverage)
    
    return nil
}
```

This tool bridges C++ code analysis with modern LLM-powered documentation generation, maintaining code integrity while enabling automated documentation workflows. The enhanced context system and multi-provider LLM integration (Ollama, OpenAI, Anthropic) provide production-ready documentation automation for complex C++ projects.
