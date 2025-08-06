#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
# File name: ollama_doxygen_update.py
# Script to automate the addition of Doxygen-style comments to C++ code files using Ollama LLM.
# This script uses the DoxLLM-IT tool for parsing and updating C++ files.

import os
import subprocess
import requests
import argparse
import json
import tempfile
import glob
from pathlib import Path

# === CONFIGURATION ===
OLLAMA_URL = os.getenv("OLLAMA_URL", "http://localhost:11434/api/generate")
MODEL_NAME = os.getenv("MODEL_NAME", "codellama:13b")
BRANCH_NAME = os.getenv("BRANCH_NAME", "doxygen-docs")

PROMPT_TEMPLATE = """You are a C++ documentation expert. Generate a comprehensive Doxygen comment for the following C++ entity.

Instructions:
- Use proper Doxygen tags (@brief, @param, @return, @throws, etc.)
- Include detailed descriptions for complex functions
- Document all parameters and return values clearly
- Add relevant @see references if applicable
- Use @since, @warning, or @note tags when appropriate
- Focus on the TARGET ENTITY, but use the context to understand its purpose
- Generate ONLY the Doxygen comment block (starting with /** and ending with */)
- Do not include any code, explanations, or markdown formatting

Context:
```cpp
{context}
```

Generate a comprehensive Doxygen comment for the target entity: {entity_name}

Response format: Start with /** and end with */"""

# === HELPER FUNCTIONS ===

def run_command(cmd, check=True, capture_output=True):
    """Run a shell command and return the result."""
    result = subprocess.run(
        cmd, 
        shell=True, 
        capture_output=capture_output, 
        text=True, 
        check=check
    )
    return result

def create_git_branch(branch_name):
    """Create and checkout a git branch for documentation changes."""
    try:
        # Check if branch exists
        result = run_command(f"git branch --list {branch_name}", check=False)
        if result.stdout.strip():
            print(f"Checking out existing branch: {branch_name}")
            run_command(f"git checkout {branch_name}")
        else:
            print(f"Creating new branch: {branch_name}")
            run_command(f"git checkout -b {branch_name}")
    except subprocess.CalledProcessError as e:
        print(f"Git operation failed: {e}")
        return False
    return True

def find_cpp_files(directory):
    """Find all C++ header files in the given directory."""
    cpp_patterns = ["**/*.hpp", "**/*.h", "**/*.hxx"]
    files = []
    
    for pattern in cpp_patterns:
        files.extend(glob.glob(os.path.join(directory, pattern), recursive=True))
    
    # Filter out build directories and vendor code
    excluded_dirs = ["build", "vendor", "third_party", ".git", "node_modules"]
    filtered_files = []
    
    for file in files:
        if not any(excluded in file for excluded in excluded_dirs):
            filtered_files.append(file)
    
    return filtered_files

def parse_file_entities(filepath, doxyllm_tool):
    """Parse a C++ file and return undocumented entities."""
    try:
        result = run_command(f"{doxyllm_tool} parse -f json '{filepath}'")
        data = json.loads(result.stdout)
        
        # Filter undocumented entities
        undocumented = []
        for entity in data.get("entities", []):
            if not entity.get("hasComment", False):
                undocumented.append(entity["fullName"])
        
        return undocumented
    except (subprocess.CalledProcessError, json.JSONDecodeError) as e:
        print(f"Error parsing {filepath}: {e}")
        return []

def extract_entity_context(filepath, entity_path, doxyllm_tool):
    """Extract context for a specific entity using DoxLLM-IT."""
    try:
        result = run_command(f"{doxyllm_tool} extract -p -s '{filepath}' '{entity_path}'")
        return result.stdout
    except subprocess.CalledProcessError as e:
        print(f"Error extracting context for {entity_path}: {e}")
        return None

def call_ollama(context, entity_name):
    """Call Ollama to generate Doxygen comment."""
    prompt = PROMPT_TEMPLATE.format(context=context, entity_name=entity_name)
    
    try:
        response = requests.post(OLLAMA_URL, json={
            "model": MODEL_NAME,
            "prompt": prompt,
            "stream": False,
            "options": {
                "temperature": 0.1,  # Low temperature for consistent output
                "top_p": 0.9,
                "num_ctx": 4096
            }
        }, timeout=120)
        
        response.raise_for_status()
        generated_comment = response.json()["response"].strip()
        
        # Clean up the response
        if generated_comment.startswith("```"):
            lines = generated_comment.split('\n')
            generated_comment = '\n'.join(lines[1:-1]) if len(lines) > 2 else generated_comment
        
        # Ensure it starts with /** and ends with */
        if not generated_comment.startswith("/**"):
            generated_comment = "/**\n * " + generated_comment.lstrip("* ")
        if not generated_comment.endswith("*/"):
            generated_comment = generated_comment.rstrip() + "\n */"
        
        return generated_comment
        
    except (requests.RequestException, KeyError) as e:
        print(f"Error calling Ollama: {e}")
        return None

def update_entity_comment(filepath, entity_path, comment, doxyllm_tool):
    """Update a single entity with the generated comment."""
    try:
        # Create temporary file with the comment
        with tempfile.NamedTemporaryFile(mode='w', suffix='.txt', delete=False) as tmp:
            tmp.write(comment)
            tmp_path = tmp.name
        
        try:
            # Update the file using DoxLLM-IT
            result = run_command(f"{doxyllm_tool} update -i -b '{filepath}' '{entity_path}' '{tmp_path}'")
            return True
        finally:
            os.unlink(tmp_path)
            
    except subprocess.CalledProcessError as e:
        print(f"Error updating {entity_path} in {filepath}: {e}")
        return False

def process_file(filepath, doxyllm_tool, max_entities=None):
    """Process a single C++ file and add documentation to undocumented entities."""
    print(f"\n📁 Processing: {filepath}")
    
    # Get undocumented entities
    undocumented = parse_file_entities(filepath, doxyllm_tool)
    
    if not undocumented:
        print("  ✅ All entities already documented")
        return 0
    
    print(f"  📋 Found {len(undocumented)} undocumented entities")
    
    # Limit entities if specified
    if max_entities:
        undocumented = undocumented[:max_entities]
        print(f"  🔢 Processing first {len(undocumented)} entities")
    
    successful_updates = 0
    
    for i, entity_path in enumerate(undocumented, 1):
        print(f"  📝 ({i}/{len(undocumented)}) Documenting: {entity_path}")
        
        # Extract context
        context = extract_entity_context(filepath, entity_path, doxyllm_tool)
        if not context:
            print(f"    ❌ Failed to extract context")
            continue
        
        # Generate comment using Ollama
        print(f"    🤖 Generating comment with {MODEL_NAME}...")
        comment = call_ollama(context, entity_path)
        if not comment:
            print(f"    ❌ Failed to generate comment")
            continue
        
        # Update the file
        if update_entity_comment(filepath, entity_path, comment, doxyllm_tool):
            print(f"    ✅ Successfully updated")
            successful_updates += 1
        else:
            print(f"    ❌ Failed to update file")
    
    print(f"  📊 Updated {successful_updates}/{len(undocumented)} entities")
    return successful_updates

def format_files(filepaths, doxyllm_tool):
    """Format updated files using clang-format."""
    print("\n🎨 Formatting updated files...")
    for filepath in filepaths:
        try:
            run_command(f"{doxyllm_tool} format -c '{filepath}' > '{filepath}.formatted'")
            run_command(f"mv '{filepath}.formatted' '{filepath}'")
            print(f"  ✅ Formatted: {filepath}")
        except subprocess.CalledProcessError:
            print(f"  ⚠️  Could not format: {filepath} (clang-format not available?)")

def commit_changes():
    """Commit changes to git."""
    try:
        run_command("git add .")
        run_command('git commit -m "docs: Add Doxygen comments via Ollama LLM automation\\n\\n- Generated comprehensive documentation for undocumented entities\\n- Used DoxLLM-IT tool for parsing and updating\\n- Applied consistent formatting"')
        print("✅ Changes committed to git")
        return True
    except subprocess.CalledProcessError as e:
        print(f"❌ Failed to commit changes: {e}")
        return False

def main():
    parser = argparse.ArgumentParser(
        description="Automate Doxygen-style comments for C++ files using Ollama and DoxLLM-IT.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Process current directory
  python3 ollama_doxygen_update.py

  # Process specific directory with custom model
  MODEL_NAME=deepseek-coder:6.7b python3 ollama_doxygen_update.py --dir /path/to/cpp/project

  # Process without committing, limit entities per file
  python3 ollama_doxygen_update.py --no-commit --max-entities 5

Environment Variables:
  OLLAMA_URL     - Ollama API URL (default: http://localhost:11434/api/generate)
  MODEL_NAME     - Model to use (default: codellama:13b)
  BRANCH_NAME    - Git branch name (default: doxygen-docs)
"""
    )
    
    parser.add_argument("--dir", type=str, default=".", 
                       help="Directory to process C++ files (default: current directory)")
    parser.add_argument("--no-commit", action="store_true", 
                       help="Skip committing changes to git")
    parser.add_argument("--no-format", action="store_true",
                       help="Skip formatting files with clang-format")
    parser.add_argument("--max-entities", type=int,
                       help="Maximum entities to process per file (for testing)")
    parser.add_argument("--doxyllm-tool", type=str, default="./doxyllm-it",
                       help="Path to DoxLLM-IT tool (default: ./doxyllm-it)")
    parser.add_argument("--files", nargs="+", 
                       help="Specific files to process (overrides --dir)")
    
    args = parser.parse_args()
    
    # Check if DoxLLM-IT tool exists
    if not os.path.exists(args.doxyllm_tool):
        print(f"❌ DoxLLM-IT tool not found at: {args.doxyllm_tool}")
        print("   Please build the tool first: go build -o doxyllm-it .")
        return 1
    
    # Check Ollama connectivity
    try:
        response = requests.get(OLLAMA_URL.replace("/api/generate", "/api/tags"), timeout=5)
        print(f"🤖 Connected to Ollama at: {OLLAMA_URL}")
        print(f"📚 Using model: {MODEL_NAME}")
    except requests.RequestException:
        print(f"❌ Cannot connect to Ollama at: {OLLAMA_URL}")
        print("   Please ensure Ollama is running and accessible")
        return 1
    
    # Change to target directory
    os.chdir(args.dir)
    
    # Create git branch
    if not args.no_commit:
        if not create_git_branch(BRANCH_NAME):
            return 1
    
    # Find files to process
    if args.files:
        cpp_files = [f for f in args.files if f.endswith(('.hpp', '.h', '.hxx'))]
    else:
        cpp_files = find_cpp_files(".")
    
    if not cpp_files:
        print("❌ No C++ header files found")
        return 1
    
    print(f"📂 Found {len(cpp_files)} C++ header files")
    
    # Process files
    total_updates = 0
    updated_files = []
    
    for filepath in cpp_files:
        updates = process_file(filepath, args.doxyllm_tool, args.max_entities)
        if updates > 0:
            total_updates += updates
            updated_files.append(filepath)
    
    # Format files if requested
    if updated_files and not args.no_format:
        format_files(updated_files, args.doxyllm_tool)
    
    # Commit changes
    if updated_files:
        print(f"\n📊 Summary:")
        print(f"  Files processed: {len(cpp_files)}")
        print(f"  Files updated: {len(updated_files)}")
        print(f"  Total entities documented: {total_updates}")
        
        if not args.no_commit:
            commit_changes()
            print(f"\n🎉 Documentation complete! Check branch '{BRANCH_NAME}'")
        else:
            print(f"\n🎉 Documentation complete! Review changes before committing.")
    else:
        print("\n✅ All files already have complete documentation")
    
    return 0

if __name__ == "__main__":
    exit(main())
