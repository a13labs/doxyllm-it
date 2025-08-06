#!/bin/bash

# Complete LLM Integration Workflow Example
# This script demonstrates the full workflow from parsing to updating with LLM-generated comments

set -e

HEADER_FILE="${1:-examples/example.hpp}"
WORK_DIR="llm_workflow"

echo "=== Complete LLM Integration Workflow ==="
echo "Processing: $HEADER_FILE"
echo

# Create working directory
mkdir -p "$WORK_DIR"

# Step 1: Parse and identify undocumented entities
echo "Step 1: Identifying undocumented entities..."
./doxyllm-it parse -f json "$HEADER_FILE" > "$WORK_DIR/entities.json"

# Extract undocumented entities using jq
undocumented=$(jq -r '.entities[] | select(.hasComment == false) | .fullName' "$WORK_DIR/entities.json")

echo "Found undocumented entities:"
echo "$undocumented" | while read -r entity; do
    echo "  - $entity"
done
echo

# Step 2: Extract contexts for each undocumented entity
echo "Step 2: Extracting contexts for LLM processing..."
mkdir -p "$WORK_DIR/contexts"

echo "$undocumented" | while read -r entity; do
    if [ -n "$entity" ]; then
        safe_name=$(echo "$entity" | tr ':' '_' | tr ' ' '_')
        echo "  Extracting context for: $entity"
        ./doxyllm-it extract -p -s "$HEADER_FILE" "$entity" > "$WORK_DIR/contexts/${safe_name}.txt" 2>/dev/null || {
            echo "    ⚠ Failed to extract context for $entity"
        }
    fi
done

# Step 3: Generate LLM prompts
echo
echo "Step 3: Generating LLM prompts..."
mkdir -p "$WORK_DIR/prompts"

for context_file in "$WORK_DIR/contexts"/*.txt; do
    if [ -f "$context_file" ]; then
        base_name=$(basename "$context_file" .txt)
        entity_name=$(echo "$base_name" | tr '_' ':')
        
        cat > "$WORK_DIR/prompts/${base_name}_prompt.txt" << EOF
Generate a comprehensive Doxygen comment for the following C++ entity.

Instructions:
- Use proper Doxygen tags (@brief, @param, @return, @throws, etc.)
- Include detailed descriptions for complex functions
- Document all parameters and return values
- Add relevant @see references if applicable
- Use @since, @warning, or @note tags when appropriate

Context:
\`\`\`cpp
$(cat "$context_file")
\`\`\`

Please generate only the Doxygen comment block (starting with /** and ending with */) for the target entity: $entity_name
EOF
    fi
done

echo "✓ Generated prompts for $(ls "$WORK_DIR/prompts"/*.txt 2>/dev/null | wc -l) entities"

# Step 4: Simulate LLM responses (in real workflow, you'd call your LLM API here)
echo
echo "Step 4: Simulating LLM responses..."
mkdir -p "$WORK_DIR/responses"

# Create some example responses
cat > "$WORK_DIR/responses/Graphics_shutdownGraphics_response.txt" << 'EOF'
/**
 * @brief Shuts down the graphics system and releases all resources
 * @details This function performs a complete cleanup of the graphics subsystem,
 * including releasing GPU memory, destroying rendering contexts, and cleaning up
 * any remaining graphics objects. It should be called exactly once during
 * application shutdown, after all rendering operations have completed.
 * 
 * @warning This function must not be called while any rendering operations
 * are in progress. Ensure all rendering threads have finished before calling.
 * 
 * @warning Calling this function multiple times or calling other graphics
 * functions after shutdown results in undefined behavior.
 * 
 * @see createRenderer() for graphics system initialization
 * @since 1.0.0
 */
EOF

cat > "$WORK_DIR/responses/Graphics_Renderer2D_drawSprite_response.txt" << 'EOF'
/**
 * @brief Draws a 2D sprite at the specified screen coordinates
 * @details Renders a textured quad using the specified texture resource.
 * The sprite is drawn at the given pixel coordinates with its original
 * size and no transformations applied. This is a convenience method for
 * simple 2D sprite rendering.
 * 
 * @param texture The texture identifier or file path for the sprite image
 * @param x The horizontal position in screen coordinates (pixels from left edge)
 * @param y The vertical position in screen coordinates (pixels from top edge)
 * 
 * @throws std::runtime_error if the renderer is not initialized
 * @throws std::invalid_argument if the texture cannot be found or loaded
 * @throws std::out_of_range if coordinates are outside the valid screen area
 * 
 * @note The coordinate system origin (0,0) is at the top-left corner
 * @note This method assumes the texture has already been loaded into memory
 * 
 * @see Renderer2D() for renderer initialization
 * @since 1.0.0
 */
EOF

echo "✓ Created example LLM responses"

# Step 5: Prepare batch update JSON
echo
echo "Step 5: Preparing batch update configuration..."

# Create the batch update JSON
cat > "$WORK_DIR/batch_update.json" << EOF
{
  "sourceFile": "$HEADER_FILE",
  "updates": [
EOF

# Add updates for each response
first=true
for response_file in "$WORK_DIR/responses"/*.txt; do
    if [ -f "$response_file" ]; then
        base_name=$(basename "$response_file" _response.txt)
        entity_name=$(echo "$base_name" | tr '_' ':')
        
        if [ "$first" = true ]; then
            first=false
        else
            echo "," >> "$WORK_DIR/batch_update.json"
        fi
        
        echo "    {" >> "$WORK_DIR/batch_update.json"
        echo "      \"entityPath\": \"$entity_name\"," >> "$WORK_DIR/batch_update.json"
        echo -n "      \"comment\": " >> "$WORK_DIR/batch_update.json"
        # Escape the comment content for JSON
        python3 -c "
import json
with open('$response_file', 'r') as f:
    content = f.read()
print(json.dumps(content))
" >> "$WORK_DIR/batch_update.json"
        echo -n "    }" >> "$WORK_DIR/batch_update.json"
    fi
done

echo >> "$WORK_DIR/batch_update.json"
echo "  ]" >> "$WORK_DIR/batch_update.json"
echo "}" >> "$WORK_DIR/batch_update.json"

echo "✓ Batch update configuration created"

# Step 6: Apply updates
echo
echo "Step 6: Applying LLM-generated comments..."

# Create a copy of the original file for updating
cp "$HEADER_FILE" "$WORK_DIR/updated_header.hpp"

# Update the batch configuration to point to our copy
sed -i "s|\"sourceFile\": \".*\"|\"sourceFile\": \"$WORK_DIR/updated_header.hpp\"|" "$WORK_DIR/batch_update.json"

# Apply the batch update
./doxyllm-it batch-update "$WORK_DIR/batch_update.json" -i -b -f

echo "✓ Comments applied and formatted"

# Step 7: Verification
echo
echo "Step 7: Verification..."

# Parse the updated file to check documentation status
./doxyllm-it parse "$WORK_DIR/updated_header.hpp" > "$WORK_DIR/updated_parse_results.txt"

# Count documented vs undocumented
total_entities=$(grep "Total entities:" "$WORK_DIR/updated_parse_results.txt" | awk '{print $3}')
documented_info=$(grep "Documented:" "$WORK_DIR/updated_parse_results.txt")

echo "Results:"
echo "  Total entities: $total_entities"
echo "  $documented_info"

# Show the difference
echo
echo "Comparing original vs updated files:"
echo "Original file: $HEADER_FILE"
echo "Updated file: $WORK_DIR/updated_header.hpp"
echo "Backup file: $WORK_DIR/updated_header.hpp.bak"

# Step 8: Summary and next steps
echo
echo "=== Workflow Complete ==="
echo
echo "Generated files in $WORK_DIR/:"
echo "  - entities.json              : Parsed entities structure"
echo "  - contexts/                  : Entity contexts for LLM input"
echo "  - prompts/                   : Ready-to-use LLM prompts"
echo "  - responses/                 : Example LLM responses (normally from API)"
echo "  - batch_update.json          : Batch update configuration"
echo "  - updated_header.hpp         : Final updated header file"
echo "  - updated_header.hpp.bak     : Backup of original"
echo "  - updated_parse_results.txt  : Verification parse results"
echo
echo "Integration points for real LLM workflow:"
echo "1. Replace Step 4 with actual LLM API calls:"
echo "   - Send prompts from prompts/ directory to your LLM"
echo "   - Save responses to responses/ directory"
echo "   - Use the same naming convention"
echo
echo "2. Automated pipeline example:"
echo "   for prompt in \$WORK_DIR/prompts/*.txt; do"
echo "     response=\$(call_llm_api \"\$prompt\")"
echo "     echo \"\$response\" > \"\${prompt/_prompt.txt/_response.txt}\""
echo "   done"
echo
echo "3. Quality control:"
echo "   - Review generated comments before applying"
echo "   - Use git diff to see changes"
echo "   - Test compilation after updates"
echo
echo "Example API integration (Python):"
echo "  import openai"
echo "  response = openai.ChatCompletion.create("
echo "    model=\"gpt-4\","
echo "    messages=[{\"role\": \"user\", \"content\": prompt}]"
echo "  )"
echo "  generated_comment = response.choices[0].message.content"
