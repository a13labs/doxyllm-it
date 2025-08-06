# .doxyllm Configuration File Examples

The `.doxyllm` file supports two formats:

## 1. YAML Format (Recommended for multi-file projects)

```yaml
# Global context applied to all files in this directory
global: |
  This is a header-only implementation of C++20's std::span class.
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
    Platform detection and compatibility utilities:
    - Compiler feature detection macros
    - Cross-platform type definitions
```

## 2. Plain Text Format (Backward compatible)

```
Simple context that applies to all files in this directory.
This format is automatically detected when YAML parsing fails.
```

## Benefits of YAML Format

1. **Global Context**: Share common information across all files
2. **File-Specific Context**: Provide targeted information for specific files
3. **Scalability**: Works well with projects containing many files
4. **Organization**: Clear separation between general and specific documentation needs

## Usage

Place a `.doxyllm` file in the same directory as your source files. The tool will:

1. Try to parse as YAML first
2. Fall back to plain text if YAML parsing fails
3. Combine global + file-specific context for each file
4. Use the enhanced context to generate better documentation

## Example Output

The LLM prompt will include:
- Global context (if present)
- File-specific context (if present for the current file)
- Code context from the actual source file
- Entity-specific information
