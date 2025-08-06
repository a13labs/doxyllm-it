# DoxLLM-IT

A C++ Doxygen comments parser for LLM integration.

## Overview

DoxLLM-IT is a CLI tool that parses C++ header files and creates a tree structure of documentable entities (namespaces, classes, functions, etc.) while preserving the original formatting. It enables feeding specific code contexts to LLMs for generating Doxygen comments without losing the original code structure.

## Features

- **Parse C++ headers**: Extract documentable entities from C++ header files with enhanced regex patterns
- **Tree structure**: Create hierarchical representation of code structure with improved scope handling
- **Preserve formatting**: Maintain original code formatting and comments
- **Scope extraction**: Extract specific scopes (namespaces, classes, functions) for LLM context
- **Code reconstruction**: Reconstruct original code from parsed tree
- **Update entities**: Add or update Doxygen comments for specific entities
- **Batch processing**: Update multiple entities using JSON configuration
- **Ollama integration**: Built-in command for automatic documentation with Ollama LLM
- **Context-aware documentation**: Support for `.doxyllm` configuration files with global and file-specific contexts
- **YAML configuration**: Structured context files for multi-file projects
- **Backward compatibility**: Plain text `.doxyllm` files still supported
- **Enhanced parser**: Improved detection of modern C++ constructs (constexpr macros, template functions)
- **Comprehensive testing**: Unit tests for parser components and regex patterns
- **Clang-format integration**: Format output using clang-format
- **JSON output**: Export parsed structure as JSON for further processing

## Installation

### Prerequisites

- Go 1.21 or later
- clang-format (optional, for code formatting)

### Build

```bash
go mod tidy
go build -o doxyllm-it .
```

## Usage

### Parse a C++ header file

```bash
# Human-readable output
./doxyllm-it parse header.hpp

# JSON output
./doxyllm-it parse -f json header.hpp

# Show all entities (including undocumented)
./doxyllm-it parse -a header.hpp
```

### Extract context for specific entity

```bash
# Extract context for a specific function
./doxyllm-it extract header.hpp "MyNamespace::MyClass::myMethod"

# Include parent and sibling context
./doxyllm-it extract -p -s header.hpp "MyNamespace::MyClass::myMethod"

# Extract only the entity scope
./doxyllm-it extract --scope header.hpp "MyNamespace::MyClass"
```

### Format code

```bash
# Reconstruct and display formatted code
./doxyllm-it format header.hpp

# Apply clang-format
./doxyllm-it format -c header.hpp
```

### Update with LLM-generated comments

```bash
# Update a single entity with a comment from file
./doxyllm-it update header.hpp "MyNamespace::MyClass::myMethod" comment.txt

# Update a single entity with comment from stdin
echo "/** @brief My function description */" | ./doxyllm-it update header.hpp "MyFunction" -

# Update in-place with backup
./doxyllm-it update -i -b header.hpp "MyClass::myMethod" comment.txt

# Batch update multiple entities from JSON
./doxyllm-it batch-update updates.json -i -f
```

### Ollama Integration (Built-in)

```bash
# Auto-generate documentation with Ollama LLM
./doxyllm-it ollama --in-place --backup examples/example.hpp

# Process entire directory with custom model
./doxyllm-it ollama --model deepseek-coder:6.7b --in-place --format src/

# Dry run to see what would be processed
./doxyllm-it ollama --dry-run --max-entities 3 .

# Use remote Ollama server
./doxyllm-it ollama --url http://remote:11434/api/generate --model codellama:34b src/

# Custom temperature and context settings
./doxyllm-it ollama --temperature 0.1 --context 8192 --model codegemma:7b header.hpp
```

### Context Configuration with .doxyllm Files

Create a `.doxyllm` file in your project directory to enhance LLM documentation quality:

#### YAML Format (Recommended for multi-file projects)
```yaml
# Global context applied to all files in this directory
global: |
  This is a C++20 implementation providing span functionality.
  Key design principles:
  - Zero-overhead abstraction
  - Type safety through bounds checking
  - Cross-compiler compatibility

# File-specific contexts (optional)
files:
  span.hpp: |
    Main implementation file containing:
    - Template class span<ElementType, Extent>
    - Helper classes and type traits
    - SFINAE-based overload resolution
  
  platform.hpp: |
    Platform detection and compatibility utilities
```

#### Plain Text Format (Backward compatible)
```
Simple context that applies to all files in this directory.
This format is automatically detected when YAML parsing fails.
```

## Architecture

### Core Components

1. **Parser** (`pkg/parser`): Enhanced C++ parser with improved regex patterns for modern C++ constructs
2. **AST** (`pkg/ast`): Defines data structures for the abstract syntax tree with access level tracking
3. **Formatter** (`pkg/formatter`): Handles code reconstruction and formatting with context extraction
4. **CLI** (`cmd/`): Command-line interface with comprehensive Ollama integration
5. **Testing** (`pkg/parser/*_test.go`): Comprehensive unit tests for parser components

### Enhanced Parser Features

- **Modern C++ Support**: Detects constexpr macros like `TCB_SPAN_CONSTEXPR11`, `TCB_SPAN_NODISCARD`
- **Access Level Tracking**: Properly tracks public/private/protected sections with stack-based management
- **Template Detection**: Improved recognition of template functions and classes
- **Scope Management**: Enhanced scope handling for nested namespaces and classes
- **Entity Deduplication**: Prevents duplicate entity detection with improved filtering

### Context System

- **Global Context**: Shared project-wide documentation context
- **File-Specific Context**: Targeted context for individual files
- **YAML Configuration**: Structured configuration with backward compatibility
- **Smart Context Loading**: Automatic format detection and fallback

### Entity Types

- Namespaces
- Classes and Structs
- Enums
- Functions and Methods
- Constructors and Destructors
- Variables and Fields
- Typedefs and Using declarations

### Workflow

1. **Parse**: Read C++ header and create AST with enhanced entity detection
2. **Context**: Load project context from `.doxyllm` configuration files
3. **Extract**: Get specific entity contexts for LLM processing with project-aware context
4. **Generate**: Use LLM to generate/update Doxygen comments with enhanced prompts
5. **Update**: Apply LLM-generated comments back to the original code
6. **Format**: Use clang-format to ensure consistent formatting

## Ollama Integration Workflow

```bash
# 1. Create context configuration
cat > .doxyllm << EOF
# Global context applied to all files
global: |
  This project implements C++20 span functionality with backward compatibility.
  Design principles: zero-overhead, type safety, cross-platform support.

files:
  span.hpp: |
    Main implementation with template metaprogramming and SFINAE constraints.
EOF

# 2. Auto-generate documentation with context
./doxyllm-it ollama --model codegemma:7b --backup --in-place span.hpp

# 3. Process multiple files with different contexts
./doxyllm-it ollama --model deepseek-coder:6.7b --max-entities 5 src/

# 4. Review changes and format
./doxyllm-it format -c span.hpp
```

## Complete LLM Integration Example

```bash
# 1. Create project context configuration
cat > .doxyllm << EOF
global: |
  C++20 span implementation with C++11 compatibility.
  Features: bounds checking, type safety, zero overhead.
files:
  header.hpp: |
    Template metaprogramming with SFINAE constraints.
    Cross-compiler compatibility macros included.
EOF

# 2. Parse and identify undocumented entities
./doxyllm-it parse -f json header.hpp | jq '.entities[] | select(.hasComment == false)'

# 3. Auto-generate with Ollama (recommended)
./doxyllm-it ollama --model codegemma:7b --backup header.hpp

# 4. Manual workflow (for custom LLMs)
./doxyllm-it extract -p -s header.hpp "MyClass::myMethod" > context.txt
llm_response=$(send_to_llm "Generate doxygen comment for: $(cat context.txt)")
echo "$llm_response" | ./doxyllm-it update -i -b header.hpp "MyClass::myMethod"

# 5. Format the result
./doxyllm-it format -c header.hpp
```

For a complete automated workflow, see `./llm_workflow.sh`

## Example

Given a C++ header file:

```cpp
namespace Graphics {
    class Renderer {
    public:
        void initialize();
        void render(const Scene& scene);
    private:
        bool initialized;
    };
}
```

The tool can:

1. Parse the structure and identify entities
2. Extract context for `Graphics::Renderer::render` method
3. Provide this context to an LLM for documentation generation
4. Update the original code with generated comments
5. Preserve all original formatting and structure

## Use Cases

- **Documentation automation**: Generate comprehensive Doxygen comments with project-specific context
- **Code analysis**: Understand code structure and dependencies with enhanced parser
- **Context extraction**: Provide targeted code context to LLMs with global and file-specific information
- **Multi-file projects**: Efficiently document large codebases with structured context configuration
- **Modern C++ support**: Handle complex template constructs and constexpr macros
- **Team workflows**: Consistent documentation generation with shared context files
- **Code refactoring**: Maintain documentation during code changes
- **API documentation**: Generate consistent API documentation with enhanced quality

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
