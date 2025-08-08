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
   - Uses token-based parsing (not full AST compilation, but enough to manipulate the doxygen comments)
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
- **Context-aware**: Support for `.doxyllm.yaml` configuration files with global and file-specific contexts
- **Modern C++ support**: Enhanced parser for constexpr macros, template functions, and C++20 features
- **Modular design**: Separate parsing, AST, and formatting concerns
- **Comprehensive testing**: Unit tests for all parser components and regex patterns

## Code Patterns and Conventions

### Entity Types

Can be found in `pkg/ast/ast.go`:

### Entity Structure

Can be found in `pkg/ast/ast.go`:

### Common Patterns

1. **Document Creation and Manipulation** (High-level API):

Can be found in `pkg/document/document.go`:

2. **Tree Navigation** (Low-level AST):

Can be found in `pkg/ast/ast.go`:

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

### Context Configuration
```yaml
# .doxyllm.yaml file structure
global: |
  Project-wide context and design principles...
  
files:
  header.hpp: |
    File-specific implementation details...
  platform.hpp: |
    Platform-specific utilities and macros...

# Doxygen group configuration
groups:
  group1:
    name: Doxygen Group Name
    title: Group Title
    description: |
      Group description with context and purpose.
    files:
      - header.hpp
      - platform.hpp
      - other_file.hpp
    generateDefgroup: true # Generate Doxygen group definition

# Files to ignore
ignore:
  - "test_*"
  - "*_test.hpp"

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

# Reasoning, Troubleshooting, and Debugging

All temporary files during reasoning, troubleshooting and debugging should be stored in the `tmp/` directory.
