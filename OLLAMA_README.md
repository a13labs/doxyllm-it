# Ollama Integration for DoxLLM-IT

This directory contains scripts for integrating DoxLLM-IT with Ollama to automatically generate Doxygen comments for C++ code.

## Files

- `ollama_doxygen_update.py` - Main Python script for Ollama integration
- `ollama_doxygen.sh` - Convenient shell wrapper with dependency checking
- `OLLAMA_README.md` - This file

## Quick Start

1. **Install and start Ollama:**
   ```bash
   # Install Ollama (see https://ollama.ai)
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # Start Ollama
   ollama serve
   
   # Pull a code model (in another terminal)
   ollama pull codellama:13b
   ```

2. **Install Python dependencies:**
   ```bash
   pip3 install requests
   ```

3. **Build DoxLLM-IT tool:**
   ```bash
   go build -o doxyllm-it .
   ```

4. **Run the integration:**
   ```bash
   # Simple usage - process current directory
   ./ollama_doxygen.sh
   
   # Test mode - quick run to see how it works
   ./ollama_doxygen.sh --test
   
   # Use different model
   ./ollama_doxygen.sh --model deepseek-coder:6.7b
   
   # Process specific directory
   ./ollama_doxygen.sh /path/to/cpp/project
   ```

## Configuration

Environment variables:
- `OLLAMA_URL` - Ollama API endpoint (default: http://localhost:11434/api/generate)
- `MODEL_NAME` - Model to use (default: codellama:13b)
- `BRANCH_NAME` - Git branch for changes (default: doxygen-docs)

## Recommended Models

For C++ documentation generation:

1. **CodeLlama** (Recommended)
   ```bash
   ollama pull codellama:13b      # Good balance of quality and speed
   ollama pull codellama:34b      # Better quality, slower
   ```

2. **DeepSeek Coder**
   ```bash
   ollama pull deepseek-coder:6.7b   # Fast and good for code
   ollama pull deepseek-coder:33b    # Higher quality
   ```

3. **Code-specific models**
   ```bash
   ollama pull magicoder:7b           # Microsoft's code model
   ollama pull phind-codellama:34b    # Optimized for coding tasks
   ```

## Examples

### Basic Usage
```bash
# Process current directory with default settings
./ollama_doxygen.sh

# Process specific directory
./ollama_doxygen.sh /path/to/cpp/project

# Use different model
MODEL_NAME=deepseek-coder:6.7b ./ollama_doxygen.sh
```

### Advanced Usage
```bash
# Process without committing (review first)
./ollama_doxygen.sh --no-commit

# Process specific files only
./ollama_doxygen.sh --files src/main.hpp include/utils.h

# Limit entities per file (for testing)
./ollama_doxygen.sh --limit 5

# Custom Ollama server
OLLAMA_URL=http://remote-server:11434/api/generate ./ollama_doxygen.sh
```

### Test Mode
```bash
# Quick test with limited processing
./ollama_doxygen.sh --test
```

## Workflow

The integration follows this process:

1. **Parse** - Use DoxLLM-IT to identify undocumented C++ entities
2. **Extract** - Get context for each undocumented entity
3. **Generate** - Call Ollama to generate Doxygen comments
4. **Update** - Use DoxLLM-IT to insert comments into source files
5. **Format** - Apply clang-format for consistent style
6. **Commit** - Create git commit with changes

## Output Example

```
=== Ollama DoxLLM-IT Integration ===
Model: codellama:13b
URL: http://localhost:11434/api/generate
Branch: doxygen-docs

📁 Processing: src/graphics.hpp
📋 Found 3 undocumented entities
📝 (1/3) Documenting: Graphics::shutdownGraphics
🤖 Generating comment with codellama:13b...
✅ Successfully updated

📊 Summary:
Files processed: 5
Files updated: 3
Total entities documented: 12

✅ Changes committed to git
🎉 Documentation complete! Check branch 'doxygen-docs'
```

## Comparison with Original Script

| Feature | Original Python Script | DoxLLM-IT + Ollama |
|---------|----------------------|-------------------|
| Parsing | cxxheaderparser | Robust regex-based parser |
| Entity Detection | Basic | Comprehensive (classes, functions, namespaces, etc.) |
| Context Extraction | Full file | Precise entity context |
| Comment Insertion | Text replacement | Structured insertion with validation |
| Code Validation | Simple diff check | AST-aware updates |
| Formatting | External clang-format | Integrated formatting |
| Error Handling | Basic | Comprehensive with backups |
| Scalability | Limited | Handles large codebases |

## Troubleshooting

### Common Issues

1. **"Cannot connect to Ollama"**
   ```bash
   # Start Ollama service
   ollama serve
   
   # Check if running
   curl http://localhost:11434/api/tags
   ```

2. **"Model not found"**
   ```bash
   # Pull the model
   ollama pull codellama:13b
   
   # List available models
   ollama list
   ```

3. **"DoxLLM-IT tool not found"**
   ```bash
   # Build the tool
   go build -o doxyllm-it .
   ```

4. **Python import errors**
   ```bash
   # Install required packages
   pip3 install requests
   ```

### Performance Tips

1. **Use appropriate model size**
   - 7B models: Fast, good for simple documentation
   - 13B models: Balanced quality and speed
   - 34B+ models: Best quality, slower

2. **Batch processing**
   - The script processes files sequentially for stability
   - Use `--limit` for testing large codebases

3. **Network optimization**
   - Run Ollama locally for best performance
   - Increase timeout for large files

## Integration with CI/CD

```yaml
# GitHub Actions example
name: Auto Documentation
on:
  push:
    branches: [ main ]
    paths: [ '**/*.hpp', '**/*.h' ]

jobs:
  document:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Ollama
        run: |
          curl -fsSL https://ollama.ai/install.sh | sh
          ollama serve &
          sleep 10
          ollama pull codellama:13b
      
      - name: Build DoxLLM-IT
        run: go build -o doxyllm-it .
      
      - name: Generate Documentation
        run: ./ollama_doxygen.sh --no-commit
      
      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v5
        with:
          title: "docs: Auto-generated Doxygen comments"
          branch: auto-documentation
```

This integration provides a robust, scalable solution for automatically generating high-quality Doxygen documentation using local LLMs!
