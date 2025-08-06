# Ollama Integration (Built-in Command)

The DoxLLM-IT tool now includes a built-in `ollama` command that integrates directly with Ollama for automatic Doxygen comment generation. This eliminates the need for external Python scripts and provides a seamless, native experience.

## Quick Start

### 1. Install and Start Ollama

```bash
# Install Ollama (see https://ollama.ai)
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama server
ollama serve

# Pull a code model (in another terminal)
ollama pull codellama:13b
```

### 2. Use the Built-in Command

```bash
# Build DoxLLM-IT (if not already built)
go build -o doxyllm-it .

# Test with dry run
./doxyllm-it ollama --dry-run examples/example.hpp

# Process a single file
./doxyllm-it ollama --in-place --backup examples/example.hpp

# Process entire directory
./doxyllm-it ollama --in-place --backup --format src/
```

## Command Options

### Basic Usage
```bash
doxyllm-it ollama [flags] <file_or_directory>
```

### Configuration Flags
- `--url, -u` - Ollama API URL (default: http://localhost:11434/api/generate)
- `--model, -m` - Model name (default: codellama:13b)
- `--temperature` - LLM temperature 0.0-1.0 (default: 0.1)
- `--top-p` - LLM top-p value 0.0-1.0 (default: 0.9)
- `--context` - Context window size (default: 4096)
- `--timeout` - Request timeout in seconds (default: 120)

### Processing Flags
- `--dry-run` - Show what would be processed without making changes
- `--max-entities` - Maximum entities to process per file (0 = unlimited)
- `--exclude` - Directories to exclude (default: build,vendor,third_party,.git,node_modules)

### Output Flags
- `--in-place, -i` - Update files in place
- `--backup, -b` - Create backup files before updating
- `--format, -f` - Format updated files with clang-format

### Environment Variables
- `OLLAMA_URL` - Default Ollama API URL
- `MODEL_NAME` - Default model name

## Examples

### 1. Basic Processing
```bash
# Process single file with defaults
./doxyllm-it ollama --in-place --backup examples/example.hpp

# Process directory
./doxyllm-it ollama --in-place --backup src/
```

### 2. Using Different Models
```bash
# Use DeepSeek Coder
./doxyllm-it ollama --model deepseek-coder:6.7b --in-place src/

# Use larger CodeLlama model
./doxyllm-it ollama --model codellama:34b --in-place include/
```

### 3. Testing and Development
```bash
# Dry run to see what would be processed
./doxyllm-it ollama --dry-run src/

# Process only first 3 entities per file
./doxyllm-it ollama --max-entities 3 --in-place examples/

# Use custom temperature for more creative output
./doxyllm-it ollama --temperature 0.3 --in-place src/
```

### 4. Remote Ollama Server
```bash
# Connect to remote Ollama instance
./doxyllm-it ollama --url http://remote-server:11434/api/generate \
  --model codellama:13b --in-place src/

# Using environment variables
export OLLAMA_URL="http://remote-server:11434/api/generate"
export MODEL_NAME="deepseek-coder:6.7b"
./doxyllm-it ollama --in-place src/
```

### 5. Production Workflow
```bash
# Full processing with backup and formatting
./doxyllm-it ollama \
  --in-place \
  --backup \
  --format \
  --model codellama:13b \
  --exclude build,vendor,third_party,.git,test \
  .
```

## Workflow Comparison

### Before (Python Script)
```bash
# Required Python dependencies and external script
pip3 install requests
python3 ollama_doxygen_update.py --dir . --no-commit
```

### After (Built-in Command)
```bash
# Native Go integration, no external dependencies
./doxyllm-it ollama --in-place --backup .
```

## Recommended Models

### For C++ Documentation:

1. **CodeLlama (Recommended)**
   ```bash
   ollama pull codellama:7b       # Fast, good quality
   ollama pull codellama:13b      # Balanced (default)
   ollama pull codellama:34b      # Best quality
   ```

2. **DeepSeek Coder**
   ```bash
   ollama pull deepseek-coder:1.3b   # Very fast
   ollama pull deepseek-coder:6.7b   # Good balance
   ollama pull deepseek-coder:33b    # High quality
   ```

3. **Other Code Models**
   ```bash
   ollama pull magicoder:7b           # Microsoft's model
   ollama pull phind-codellama:34b    # Optimized for coding
   ollama pull starcoder2:15b         # Latest Hugging Face model
   ```

## Sample Output

```bash
$ ./doxyllm-it ollama --dry-run examples/example.hpp

🤖 Connected to Ollama at: http://localhost:11434/api/generate
📚 Using model: codellama:13b
📂 Found 1 C++ header files

🔍 Dry run mode - showing what would be processed:

📁 Processing: examples/example.hpp
📋 Found 3 undocumented entities
📝 (1/3) Would document: Graphics::shutdownGraphics
📝 (2/3) Would document: Graphics::Renderer2D::drawSprite  
📝 (3/3) Would document: Graphics::Utils::calculateFPS

📊 Summary:
  Files processed: 1
  Files updated: 1
  Total entities documented: 3

💡 Run without --dry-run to apply changes
```

## Real Processing Example

```bash
$ ./doxyllm-it ollama --in-place --backup --max-entities 1 examples/example.hpp

🤖 Connected to Ollama at: http://localhost:11434/api/generate
📚 Using model: codellama:13b
📂 Found 1 C++ header files

📁 Processing: examples/example.hpp
📋 Found 3 undocumented entities
🔢 Processing first 1 entities
📝 (1/1) Documenting: Graphics::shutdownGraphics
🤖 Generating comment with codellama:13b...
✅ Successfully updated
📊 Updated 1/1 entities

📊 Summary:
  Files processed: 1
  Files updated: 1
  Total entities documented: 1

🎉 Documentation generation complete!
```

## Error Handling

The command provides comprehensive error handling:

- **Connection Issues**: Clear messages about Ollama connectivity
- **Model Issues**: Validation of model availability
- **File Issues**: Proper error messages for file access problems
- **Parsing Issues**: Detailed error reporting for parsing failures
- **Backup Safety**: Automatic backup creation before modifications

## Performance Tips

1. **Model Selection**: Use smaller models (7B) for speed, larger (34B) for quality
2. **Batch Size**: Use `--max-entities` to limit processing for testing
3. **Exclusions**: Use `--exclude` to skip unnecessary directories
4. **Local Ollama**: Run Ollama locally for best performance
5. **Context Size**: Adjust `--context` based on model capabilities

## Integration with Git

```bash
# Create a branch for documentation
git checkout -b add-documentation

# Process files
./doxyllm-it ollama --in-place --backup src/

# Review changes
git diff

# Commit if satisfied
git add .
git commit -m "docs: Add comprehensive Doxygen comments

- Generated using DoxLLM-IT ollama integration
- Used codellama:13b model for documentation
- Applied to all undocumented entities"
```

## Advantages of Built-in Integration

1. **No External Dependencies**: Pure Go implementation
2. **Better Performance**: Direct integration with parsing/formatting
3. **Consistent Interface**: Uses same flags and patterns as other commands
4. **Better Error Handling**: Native Go error handling and reporting
5. **Type Safety**: Compile-time checking of all integrations
6. **Memory Efficiency**: No process spawning or temporary files
7. **Unified Workflow**: Single tool for all documentation needs

This built-in integration makes DoxLLM-IT a complete, self-contained solution for C++ documentation automation!
