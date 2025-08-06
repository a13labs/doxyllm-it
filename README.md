# DoxLLM-IT

A C++ Doxygen comments parser for LLM integration.

## Overview

DoxLLM-IT is a CLI tool that parses C++ header files and creates a tree structure of documentable entities (namespaces, classes, functions, etc.) while preserving the original formatting. It enables feeding specific code contexts to LLMs for generating Doxygen comments without losing the original code structure.

## Features

- **Parse C++ headers**: Extract documentable entities from C++ header files
- **Tree structure**: Create hierarchical representation of code structure
- **Preserve formatting**: Maintain original code formatting and comments
- **Scope extraction**: Extract specific scopes (namespaces, classes, functions) for LLM context
- **Code reconstruction**: Reconstruct original code from parsed tree
- **Update entities**: Add or update Doxygen comments for specific entities
- **Batch processing**: Update multiple entities using JSON configuration
- **Ollama integration**: Built-in command for automatic documentation with Ollama LLM
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
```

## Architecture

### Core Components

1. **Parser** (`pkg/parser`): Parses C++ code and identifies documentable entities
2. **AST** (`pkg/ast`): Defines data structures for the abstract syntax tree
3. **Formatter** (`pkg/formatter`): Handles code reconstruction and formatting
4. **CLI** (`cmd/`): Command-line interface

### Entity Types

- Namespaces
- Classes and Structs
- Enums
- Functions and Methods
- Constructors and Destructors
- Variables and Fields
- Typedefs and Using declarations

### Workflow

1. **Parse**: Read C++ header and create AST
2. **Extract**: Get specific entity contexts for LLM processing
3. **Generate**: Use LLM to generate/update Doxygen comments
4. **Update**: Apply LLM-generated comments back to the original code
5. **Format**: Use clang-format to ensure consistent formatting

## Complete LLM Integration Example

```bash
# 1. Parse and identify undocumented entities
./doxyllm-it parse -f json header.hpp | jq '.entities[] | select(.hasComment == false)'

# 2. Extract context for each entity
./doxyllm-it extract -p -s header.hpp "MyClass::myMethod" > context.txt

# 3. Send context to LLM (pseudocode)
llm_response=$(send_to_llm "Generate doxygen comment for: $(cat context.txt)")

# 4. Update the original file
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

- **Documentation automation**: Generate comprehensive Doxygen comments
- **Code analysis**: Understand code structure and dependencies
- **Context extraction**: Provide targeted code context to LLMs
- **Code refactoring**: Maintain documentation during code changes
- **API documentation**: Generate consistent API documentation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
