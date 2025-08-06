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

3. **`pkg/formatter/`** - Code reconstruction and formatting
   - Reconstructs original code from parsed tree
   - Handles Doxygen comment formatting
   - Integrates with clang-format for consistency

4. **`pkg/utils/`** - Utility functions
   - Doxygen comment parsing helpers
   - Path manipulation utilities
   - C++ identifier validation

5. **`cmd/`** - CLI commands using Cobra framework
   - `parse` - Parse files and output entity structure
   - `extract` - Extract context for specific entities
   - `format` - Reconstruct and format code
   - `update` - Update files with LLM-generated comments
   - `batch-update` - Batch update multiple entities

### Key Design Principles

- **Non-destructive parsing**: Never lose original code structure
- **Scope-aware**: Understand C++ scope rules and hierarchies  
- **LLM-optimized**: Provide targeted context without overwhelming
- **Format-preserving**: Maintain code style and formatting
- **Modular design**: Separate parsing, AST, and formatting concerns

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

1. **Tree Navigation**:
```go
func (e *Entity) GetPath() []string
func (e *Entity) FindByPath(path []string) *Entity
func (tree *ScopeTree) FindEntity(path string) *Entity
```

2. **Context Extraction**:
```go
func (f *Formatter) ExtractEntityContext(entity *Entity, includeParent, includeSiblings bool) string
```

3. **Safe Name Conversion** for file paths:
```bash
safe_name=$(echo "$entity" | tr ':' '_' | tr ' ' '_')
```

## Development Guidelines

### When Adding New Entity Types
1. Add to `EntityType` enum in `ast/ast.go`
2. Update `String()` method for the enum
3. Add parsing logic in `parser/parser.go`
4. Update regex patterns if needed
5. Add to `GetDocumentableEntities()` if documentable

### When Adding New Commands
1. Create new file in `cmd/` directory
2. Follow Cobra command structure
3. Add command to `root.go` init function
4. Include proper flag handling and validation
5. Add comprehensive help text

### When Modifying Parser
- Always preserve `OriginalText` and formatting
- Update both `SourceRange` and `HeaderRange`
- Handle scope stack properly for nested entities
- Test with complex C++ constructs

### Error Handling
- Use `fmt.Errorf()` for error wrapping
- Provide context in error messages
- Handle edge cases gracefully
- Validate inputs before processing

## Testing Approach

### Test Files Structure
- `examples/example.hpp` - Comprehensive test case
- `test_parser.go` - Basic functionality testing
- `workflow.sh` - Basic workflow demonstration
- `llm_workflow.sh` - Complete LLM integration example

### Test Scenarios
1. **Entity Recognition**: All C++ construct types
2. **Scope Handling**: Nested namespaces and classes
3. **Context Extraction**: Parent/sibling relationships
4. **Code Reconstruction**: Lossless formatting preservation
5. **Update Operations**: Comment insertion and formatting

## LLM Integration Patterns

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
- **Template parsing**: Basic recognition only, no complex parameter parsing
- **Preprocessor directives**: Limited handling of #define, #ifdef
- **Complex inheritance**: Simple inheritance only
- **Modern C++ features**: May need updates for newest C++20/23 features

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
├── pkg/                    # Core packages
├── examples/               # Test files
└── .github/               # Project metadata
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- Standard Go libraries for parsing and formatting

## Future Enhancement Areas

1. **Enhanced Template Support**: Better template parameter parsing
2. **Include Analysis**: Parse #include dependencies
3. **C++20 Features**: Concepts, modules, coroutines
4. **IDE Integration**: VS Code extension
5. **Web Interface**: Browser-based tool for teams
6. **Comment Quality**: AI-powered comment quality scoring

## Integration Examples

### Python API Wrapper
```python
import subprocess
import json

class DoxLLM:
    def parse(self, header_file):
        result = subprocess.run(['./doxyllm-it', 'parse', '-f', 'json', header_file])
        return json.loads(result.stdout)
    
    def extract_context(self, header_file, entity_path):
        result = subprocess.run(['./doxyllm-it', 'extract', '-p', '-s', header_file, entity_path])
        return result.stdout
```

This tool bridges C++ code analysis with modern LLM-powered documentation generation, maintaining code integrity while enabling automated documentation workflows.
