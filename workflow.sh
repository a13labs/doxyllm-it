#!/bin/bash

# DoxLLM-IT Workflow Example
# This script demonstrates a typical workflow for using the tool with LLM integration

set -e

HEADER_FILE="${1:-test/example.hpp}"
OUTPUT_DIR="output"

echo "=== DoxLLM-IT Workflow Example ==="
echo "Processing: $HEADER_FILE"
echo

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Step 1: Parse the file and identify undocumented entities
echo "Step 1: Parsing C++ header file..."
./doxyllm-it parse "$HEADER_FILE" > "$OUTPUT_DIR/parse_results.txt"
echo "✓ Parse results saved to $OUTPUT_DIR/parse_results.txt"

# Step 2: Generate JSON for programmatic processing
echo
echo "Step 2: Generating JSON structure..."
./doxyllm-it parse -f json "$HEADER_FILE" > "$OUTPUT_DIR/entities.json"
echo "✓ JSON structure saved to $OUTPUT_DIR/entities.json"

# Step 3: Extract context for specific entities (example)
echo
echo "Step 3: Extracting contexts for LLM processing..."

# Extract some function contexts
functions=(
    "Graphics::Renderer::render"
    "Graphics::createRenderer"
    "runApplication"
)

for func in "${functions[@]}"; do
    safe_name=$(echo "$func" | tr ':' '_')
    echo "  Extracting context for: $func"
    
    # Extract with full context for LLM
    ./doxyllm-it extract -p -s "$HEADER_FILE" "$func" > "$OUTPUT_DIR/context_${safe_name}.txt" 2>/dev/null || {
        echo "    ⚠ Function $func not found, skipping..."
        continue
    }
    echo "    ✓ Context saved to $OUTPUT_DIR/context_${safe_name}.txt"
done

# Step 4: Extract class scopes
echo
echo "Step 4: Extracting class scopes..."

classes=(
    "Graphics::Renderer"
    "AppConfig"
)

for class in "${classes[@]}"; do
    safe_name=$(echo "$class" | tr ':' '_')
    echo "  Extracting scope for: $class"
    
    ./doxyllm-it extract --scope "$HEADER_FILE" "$class" > "$OUTPUT_DIR/scope_${safe_name}.txt" 2>/dev/null || {
        echo "    ⚠ Class $class not found, skipping..."
        continue
    }
    echo "    ✓ Scope saved to $OUTPUT_DIR/scope_${safe_name}.txt"
done

# Step 5: Demonstrate code reconstruction
echo
echo "Step 5: Testing code reconstruction..."
./doxyllm-it format "$HEADER_FILE" > "$OUTPUT_DIR/reconstructed.hpp"
echo "✓ Reconstructed code saved to $OUTPUT_DIR/reconstructed.hpp"

# Summary
echo
echo "=== Workflow Complete ==="
echo
echo "Generated files in $OUTPUT_DIR/:"
echo "  - parse_results.txt     : Human-readable parse results"
echo "  - entities.json         : JSON structure for programmatic use"
echo "  - context_*.txt         : Function contexts for LLM input"
echo "  - scope_*.txt           : Class scopes for LLM input"
echo "  - reconstructed.hpp     : Reconstructed code"
echo
echo "Next steps for LLM integration:"
echo "1. Use context files as input to your LLM"
echo "2. Generate doxygen comments for undocumented entities"
echo "3. Update the original header file with generated comments"
echo "4. Use 'doxyllm-it format -c' to apply clang-format"
echo
echo "Example LLM prompt:"
echo "\"Generate a comprehensive doxygen comment for the following C++ function:\""
echo "\"$(head -5 $OUTPUT_DIR/context_Graphics_Renderer_render.txt 2>/dev/null || echo 'Context file not available')\""
