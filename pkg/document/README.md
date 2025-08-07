# Document Package

The `document` package provides a high-level abstraction layer for manipulating C++ header files with Doxygen documentation. It hides the complexity of AST parsing and traversal behind a simple, intuitive API.

## Overview

The Document abstraction layer offers several key benefits:

- **Simplified API**: Hide parser complexity and AST traversal details
- **Stateful Operations**: Maintain parsing state and track changes
- **Convenient Entity Access**: Provide intuitive methods to find/modify entities
- **Change Tracking**: Keep track of modifications for efficient saving
- **Batch Operations**: Apply multiple documentation updates efficiently
- **Validation**: Built-in validation for common documentation issues

## Quick Start

### Creating a Document

```go
package main

import (
    "doxyllm-it/pkg/document"
)

func main() {
    // From file
    doc, err := document.NewFromFile("myheader.hpp")
    if err != nil {
        log.Fatal(err)
    }
    
    // From content
    content := `class MyClass { public: void method(); };`
    doc, err := document.NewFromContent("myheader.hpp", content)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Finding Entities

```go
// Find by full path
entity := doc.FindEntity("MyNamespace::MyClass::myMethod")

// Find by name (returns all matches)
entities := doc.FindEntitiesByName("myMethod")

// Find by type
classes := doc.FindEntitiesByType(ast.EntityClass)
methods := doc.FindEntitiesByType(ast.EntityMethod)

// Get all documentable entities
documentable := doc.GetDocumentableEntities()

// Get undocumented entities
undocumented := doc.GetUndocumentedEntities()
```

### Adding Documentation

```go
// Set brief description
err := doc.SetEntityBrief("MyClass::myMethod", "Brief description")

// Set detailed description  
err := doc.SetEntityDetailed("MyClass::myMethod", "Detailed description...")

// Add parameter documentation
err := doc.AddEntityParam("MyClass::myMethod", "param1", "Description of param1")

// Set return documentation
err := doc.SetEntityReturn("MyClass::myMethod", "Description of return value")

// Add to groups
err := doc.AddEntityGroup("MyClass::myMethod", "utility_functions")

// Mark as deprecated
err := doc.SetEntityDeprecated("MyClass::myMethod", "Use newMethod instead")

// Set custom tags
err := doc.SetEntityCustomTag("MyClass::myMethod", "complexity", "O(n)")
```

### Batch Operations

For efficiency when making many changes:

```go
brief := "Brief description"
detailed := "Detailed description"
returnDesc := "Return value description"

updates := []document.BatchUpdate{
    {
        EntityPath: "MyClass::myMethod",
        Brief:      &brief,
        Detailed:   &detailed,
        Params: map[string]string{
            "param1": "First parameter",
            "param2": "Second parameter",
        },
        Return:     &returnDesc,
        Groups:     []string{"utilities", "core"},
        CustomTags: map[string]string{
            "complexity": "O(1)",
            "thread_safety": "safe",
        },
    },
}

err := doc.ApplyBatchUpdates(updates)
```

### Documentation Analysis

```go
// Get documentation statistics
stats := doc.GetDocumentationStats()
fmt.Printf("Coverage: %.1f%%\n", stats.DocumentationCoverage)
fmt.Printf("Total: %d, Documented: %d, Undocumented: %d\n", 
    stats.TotalEntities, stats.DocumentedEntities, stats.UndocumentedEntities)

// Get entity summary
summary, err := doc.GetEntitySummary("MyClass::myMethod")
if err == nil {
    fmt.Printf("Has doc: %t, Has brief: %t, Param count: %d\n",
        summary.HasDoc, summary.HasBrief, summary.ParamCount)
}

// Validate documentation
issues := doc.Validate()
for _, issue := range issues {
    fmt.Printf("%s [%s]: %s\n", issue.EntityPath, issue.Severity, issue.Message)
}
```

### Modification Tracking

```go
// Check if document has been modified
if doc.IsModified() {
    fmt.Println("Document has unsaved changes")
}

// Save changes (requires integration with formatter package)
// err := doc.Save() // Save to original file
// err := doc.SaveAs("newfile.hpp") // Save to new file
// content, err := doc.SaveToString() // Get modified content as string
```

## API Reference

### Document Creation

- `NewFromFile(filename string) (*Document, error)` - Create from file
- `NewFromContent(name, content string) (*Document, error)` - Create from content

### Entity Lookup

- `FindEntity(path string) *ast.Entity` - Find by full path
- `FindEntitiesByName(name string) []*ast.Entity` - Find by name
- `FindEntitiesByType(entityType ast.EntityType) []*ast.Entity` - Find by type
- `GetAllEntities() []*ast.Entity` - Get all entities
- `GetDocumentableEntities() []*ast.Entity` - Get documentable entities
- `GetUndocumentedEntities() []*ast.Entity` - Get undocumented entities

### Documentation Manipulation

- `SetEntityComment(entityPath string, comment *ast.DoxygenComment) error`
- `SetEntityBrief(entityPath, brief string) error`
- `SetEntityDetailed(entityPath, detailed string) error`
- `AddEntityParam(entityPath, paramName, description string) error`
- `SetEntityReturn(entityPath, description string) error`
- `AddEntityGroup(entityPath, groupName string) error`
- `SetEntityDeprecated(entityPath, message string) error`
- `SetEntityCustomTag(entityPath, tagName, value string) error`

### Batch Operations

- `ApplyBatchUpdates(updates []BatchUpdate) error`

### Analysis

- `GetEntitySummary(entityPath string) (*EntitySummary, error)`
- `GetDocumentationStats() *DocumentationStats`
- `Validate() []ValidationIssue`

### State Management

- `GetFilename() string`
- `GetContent() string`
- `IsModified() bool`
- `GetTree() *ast.ScopeTree` - Access underlying AST (advanced)

## Examples

See `examples/document_demo/main.go` for a comprehensive example that demonstrates:

- Creating a document from content
- Finding undocumented entities
- Adding documentation systematically
- Using batch updates
- Validation and analysis
- Working with different entity types

## Integration with Other Packages

The document package integrates with:

- `pkg/ast` - Provides the AST structures
- `pkg/parser` - Handles C++ parsing
- `pkg/formatter` - Will handle content reconstruction (TODO)

## Future Enhancements

Planned features:

- **Save Functionality**: Integration with formatter package for saving changes
- **Diff Generation**: Show documentation changes
- **Template Support**: Documentation templates for different entity types
- **Plugin System**: Extensible validation and formatting rules
- **Incremental Parsing**: Efficient updates for large files

## Error Handling

The package provides clear error messages for common issues:

- Entity not found errors
- Type mismatches (e.g., adding parameters to non-functions)
- File I/O errors
- Parse errors

## Thread Safety

The Document type is not thread-safe. If you need concurrent access, you should synchronize access with mutexes or use separate Document instances per goroutine.
