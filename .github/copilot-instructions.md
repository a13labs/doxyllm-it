# GitHub Copilot Instructions for DoxLLM-IT

## Project Overview

DoxLLM-IT is a CLI tool written in Go that parses C++ header files and creates a tree structure of documentable entities while preserving original formatting. It's designed to enable feeding specific code contexts to LLMs for generating Doxygen comments without losing code integrity.

## Project Architecture

### Core Components

1. **`pkg/ast/`** - Abstract Syntax Tree definitions
   - Defines entity types (namespace, class, function, etc.)
   - Manages hierarchical relationships and scope
   - Handles Doxygen comment structures

2. **`pkg/parser/`** - C++ parsing engine with streaming tokenizer
   - Uses streaming token-based parsing following single responsibility principle
   - Streaming tokenizer provides tokens on-demand with O(1) memory complexity
   - Small 3-token lookahead buffer for parser needs without pre-tokenizing entire files
   - Identifies documentable entities and their signatures
   - Preserves original text and formatting
   - Tracks source positions and ranges
   - Backward compatibility layer maintains existing parser interface

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

6. **`pkg/llm/`** - LLM integration layer
   - Provider interface and factory for multiple LLM providers
   - Ollama provider implementation with full API support
   - Comment builder for structured Doxygen comment generation
   - Service layer for high-level LLM operations

7. **`cmd/`** - CLI commands using Cobra framework
   - `parse` - Parse files and output entity structure
   - `extract` - Extract context for specific entities
   - `format` - Reconstruct and format code
   - `update` - Update files with LLM-generated comments
   - `batch-update` - Batch update multiple entities
   - `llm` - Built-in LLM integration with context-aware documentation (supports Ollama, OpenAI, Anthropic)
   - `init` - Initialize .doxyllm.yaml.yaml configuration files by analyzing codebase structure

### Key Design Principles

- **Non-destructive parsing**: Never lose original code structure
- **Scope-aware**: Understand C++ scope rules and hierarchies with access level tracking
- **LLM-optimized**: Provide targeted context without overwhelming, enhanced with project-specific context
- **Format-preserving**: Maintain code style and formatting
- **Context-aware**: Support for `.doxyllm.yaml.yaml` configuration files with global and file-specific contexts
- **Modern C++ support**: Enhanced parser for constexpr macros, template functions, and C++20 features
- **Modular design**: Separate parsing, AST, and formatting concerns
- **Streaming architecture**: Tokenizer provides O(1) memory complexity with lazy evaluation
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

### When Working with Init Command
- Use the `init` command to automatically generate `.doxyllm.yaml` configuration files
- The command analyzes the codebase structure using LLM providers
- It generates individual file summaries and global project summaries
- The resulting YAML configuration provides contextual information for better documentation
- Supports overwrite flag to replace existing configurations

### When Adding New LLM Providers
1. Implement the `llm.Provider` interface in `pkg/llm/`
2. Add provider initialization in `llm.NewProvider()`
3. Update command flags and help text in `cmd/llm.go`
4. Add comprehensive tests with mock providers
5. Update documentation with provider-specific examples
6. Ensure proper error handling and connection testing

### When Modifying Parser
- **Parser is now modularized** into focused files in `pkg/parser/`:
  - `parser_core.go` - Main parsing logic and top-level dispatcher
  - `lexer_helpers.go` - Token navigation utilities
  - `scope_helpers.go` - Scope management functions  
  - `preprocessor.go` - Macro and define handling
  - `comments.go`, `doxygen.go` - Comment parsing
  - `classes.go`, `functions.go`, `variables.go`, `enums.go` - Language construct parsers
  - `templates.go`, `namespaces.go`, `using_typedef.go` - Specialized parsing
  - `access.go` - Access specifier handling
- **Adding new parsing logic**: Choose the appropriate file or create a new one following the pattern
- **Modifying existing parsers**: Update the specific module file rather than monolithic parser
- Always preserve `OriginalText` and formatting
- Update both `SourceRange` and `HeaderRange`
- Handle scope stack properly for nested entities
- Update access level tracking with `accessStack`
- Test regex patterns with comprehensive unit tests
- Handle modern C++ constructs (constexpr macros, templates)
- Test with complex C++ constructs
- Use streaming tokenizer architecture: tokens generated on-demand via NextToken()
- Maintain backward compatibility layer for existing parser interface
- Ensure tokenizer follows single responsibility principle

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
- Test streaming tokenizer with large files
- Validate memory efficiency of streaming approach
- Ensure backward compatibility layer works correctly

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
# .doxyllm.yaml.yaml file structure
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

### Initialize Configuration
```bash
# Initialize .doxyllm.yaml for current directory
./doxyllm-it init .

# Initialize with custom model and overwrite existing
./doxyllm-it init --model deepseek-coder:6.7b --overwrite src/

# Initialize with specific provider settings
./doxyllm-it init --provider ollama --url http://remote:11434 --overwrite .
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
├── go.sum                  # Dependencies lock file
├── README.md               # Project documentation
├── IMPLEMENTATION.md       # Implementation details
├── cmd/                    # CLI commands
│   ├── extract.go         # Extract command
│   ├── format.go          # Format command
│   ├── init.go            # Initialize .doxyllm.yaml configuration
│   ├── llm.go             # LLM integration command
│   ├── llm_test.go        # LLM command tests
│   ├── parse.go           # Parse command
│   ├── root.go            # Root command and CLI setup
│   └── update.go          # Update and batch-update commands
├── docs/                   # Documentation files
│   ├── doxyllm-config.md  # Configuration documentation
│   ├── ENHANCED_TAG_MANAGEMENT.md
│   ├── OLLAMA_BUILTIN.md  # Ollama integration docs
│   ├── OLLAMA_REFACTORING.md
│   ├── REFACTORING_SUMMARY.md
│   ├── RELEASE.md         # Release documentation
│   └── WORKFLOW_SUMMARY.md
├── pkg/                    # Core packages
│   ├── ast/               # AST definitions and structures
│   │   └── ast.go
│   ├── document/          # High-level document abstraction
│   │   ├── document.go
│   │   ├── document_test.go
│   │   ├── formatter_integration_test.go
│   │   ├── README.md
│   │   ├── service.go
│   │   └── service_test.go
│   ├── formatter/         # Code reconstruction and formatting
│   │   ├── formatter.go
│   │   └── formatter_test.go
│   ├── llm/               # LLM integration layer
│   │   ├── comment_builder.go
│   │   ├── comment_builder_test.go
│   │   ├── factory.go
│   │   ├── interface.go
│   │   ├── ollama.go
│   │   ├── ollama_test.go
│   │   ├── service.go
│   │   └── service_test.go
│   ├── parser/            # C++ parsing engine
│   │   ├── parser.go
│   │   ├── parser_test.go
│   │   ├── tokenizer.go
│   │   ├── tokenizer_safeguard_test.go
│   │   └── tokenizer_test.go
│   └── utils/             # Utility functions
│       └── utils.go
├── examples/               # Test files and demos
│   └── document_demo/     # Document abstraction examples
│       └── main.go
├── scripts/               # Build and release scripts
│   └── build-release.sh
├── tmp/                   # Temporary files used during test and development
└── .github/               # Project metadata and CI/CD
```

# Documentation 

Documentation for the project can be found in the `docs/` directory. This includes guides, API documentation, and other relevant materials. If documentation is missing or incomplete, it should be added to the appropriate files in this directory. The `README.md` file provides an overview of the project, while `IMPLEMENTATION.md` contains detailed information about the architecture and design decisions.

# Changes to Architecture

Every change to the architecture should be documented in the `IMPLEMENTATION.md` file. This includes changes to the core components, design principles, and any new features or modifications to existing functionality. Also you should update the `README.md` file to reflect any significant changes that impact the overall understanding of the project. The ".github/copilot-instructions.md" file should also be updated to include any new patterns or conventions that have been established as a result of these changes.

# Coding Agent

Every architecture change should be reflected in the Coding agent's instructions. This includes updating the patterns, conventions, and guidelines for using the DoxLLM-IT project. The Coding agent should be able to understand the current architecture, design principles, and how to effectively use the project's features. This are stored in the `.github/copilot-instructions.md` file, which serves as a reference for the Coding agent to provide accurate and relevant suggestions.


# Reasoning, Troubleshooting, and Debugging

All temporary files during reasoning, troubleshooting and debugging should be stored in the `tmp/` directory. The agent should ensure that these files are cleaned up after use to maintain a tidy project structure. If any files in the `tmp/` directory are no longer needed, they should be deleted to prevent clutter. Also it must runs all tests before making any changes to ensure that the project remains stable and functional. If any tests fail, the agent should investigate the cause of the failure and address it before proceeding with further changes.

# Refactorings

Every time a refactoring is done the agent must:

First: Move the files that will be refactoring to 'tmp/' directory.
Second: The moved files must be deleted from the original place.
Third: The agent should use the old files as reference for the refactoring. The agent should always read the original code to understand what code can be reused and what needs to be changed.
Fourth: Create the new files in the original location.
Fifth: Use the old files as a reference.
Sixth: Always test the new implementation to ensure it works as expected.