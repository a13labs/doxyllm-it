# Enhanced Doxygen Tag Management

This enhancement introduces intelligent Doxygen tag management that reduces reliance on LLMs for generating structural tags while improving consistency and accuracy.

## Key Improvements

### 1. **Enhanced Parser with Doxygen Comment Detection**

The parser now:
- Detects existing Doxygen comments (`/**`, `///`, `//!`)
- Associates comments with their corresponding entities
- Tracks Doxygen tags including group-related ones (`@defgroup`, `@ingroup`, `@addtogroup`)

### 2. **Structured Comment Generation**

Instead of asking the LLM to generate complete Doxygen comments, the system now:
- Asks LLM only for descriptive content (brief and detailed descriptions)
- Automatically generates appropriate tags (`@brief`, `@param`, `@return`, `@ingroup`)
- Ensures consistent formatting and structure

### 3. **Group Management**

New `.doxyllm.yaml` configuration supports:
```yaml
groups:
  group_name:
    name: "group_identifier"
    title: "Group Title"
    description: "Detailed group description"
    files:
      - "pattern/*.hpp"
      - "specific_file.hpp"
    generateDefgroup: true/false
```

### 4. **Automatic Tag Generation**

The system automatically:
- Adds `@ingroup` tags to entities based on file group membership
- Generates `@defgroup` comments at file level when `generateDefgroup: true`
- Extracts function parameters and adds `@param` placeholders
- Detects return types and adds `@return` placeholders

## Benefits

1. **Consistency**: All comments follow the same structure
2. **Efficiency**: LLM focuses only on descriptive content
3. **Maintainability**: Group membership managed declaratively
4. **Accuracy**: Automatic parameter detection prevents mismatches
5. **Scalability**: Easy to manage large codebases with multiple groups

## Usage Example

With the enhanced system, given this function:
```cpp
void processData(const std::vector<int>& input, std::string& output, bool verbose = false);
```

The LLM only needs to provide:
> "Processes input data and generates formatted output with optional verbose logging"

The system automatically generates:
```cpp
/**
 * @brief Processes input data and generates formatted output with optional verbose logging
 * @param input 
 * @param output 
 * @param verbose 
 * @ingroup data_processing
 */
```

## Configuration Migration

Existing `.doxyllm.yaml` files remain compatible. To use group features:

1. Rename `.doxyllm.yaml` to `.doxyllm.yaml`
2. Add `groups:` section
3. Configure group membership and `generateDefgroup` flags

The system maintains backward compatibility with plain text `.doxyllm.yaml` files.
