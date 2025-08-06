#!/bin/bash

# Ollama DoxLLM-IT Integration Script
# Convenience wrapper for the Python Ollama integration

set -e

# Default configuration
export OLLAMA_URL="${OLLAMA_URL:-http://localhost:11434/api/generate}"
export MODEL_NAME="${MODEL_NAME:-codellama:13b}"
export BRANCH_NAME="${BRANCH_NAME:-doxygen-docs}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

show_help() {
    cat << EOF
Ollama DoxLLM-IT Integration

USAGE:
    $0 [OPTIONS] [DIRECTORY]

DESCRIPTION:
    Automatically generates Doxygen comments for C++ files using Ollama LLM
    and the DoxLLM-IT tool for parsing and updating.

OPTIONS:
    -h, --help          Show this help message
    -m, --model MODEL   Ollama model to use (default: codellama:13b)
    -u, --url URL       Ollama API URL (default: http://localhost:11434/api/generate)
    -b, --branch NAME   Git branch name (default: doxygen-docs)
    -n, --no-commit     Don't commit changes to git
    -f, --no-format     Don't format files with clang-format
    -l, --limit N       Limit entities per file (for testing)
    -t, --test          Test mode: process only first 3 files with max 2 entities each
    --files FILE...     Process specific files instead of directory

EXAMPLES:
    # Process current directory with default model
    $0

    # Use a different model
    $0 --model deepseek-coder:6.7b /path/to/cpp/project

    # Test mode - quick run to see how it works
    $0 --test

    # Process specific files only
    $0 --files src/header1.hpp include/header2.h

    # Don't commit changes (review first)
    $0 --no-commit

ENVIRONMENT VARIABLES:
    OLLAMA_URL      Ollama API endpoint
    MODEL_NAME      Default model name
    BRANCH_NAME     Default git branch name

REQUIREMENTS:
    - DoxLLM-IT tool built (./doxyllm-it)
    - Ollama running and accessible
    - Python 3 with requests library
    - Git repository (for commit functionality)
    - clang-format (optional, for formatting)

EOF
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check DoxLLM-IT tool
    if [ ! -f "./doxyllm-it" ]; then
        log_error "DoxLLM-IT tool not found. Please build it first:"
        echo "  go build -o doxyllm-it ."
        exit 1
    fi
    log_success "DoxLLM-IT tool found"
    
    # Check Python
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 not found"
        exit 1
    fi
    log_success "Python 3 found"
    
    # Check requests library
    if ! python3 -c "import requests" 2>/dev/null; then
        log_warning "Python requests library not found. Installing..."
        pip3 install requests || {
            log_error "Failed to install requests library"
            exit 1
        }
    fi
    log_success "Python requests library available"
    
    # Check Ollama connectivity
    if curl -s "$OLLAMA_URL" > /dev/null 2>&1 || curl -s "${OLLAMA_URL%/*}/api/tags" > /dev/null 2>&1; then
        log_success "Ollama is accessible at $OLLAMA_URL"
    else
        log_error "Cannot connect to Ollama at $OLLAMA_URL"
        echo "  Please ensure Ollama is running:"
        echo "  ollama serve"
        exit 1
    fi
    
    # Check if model is available
    if curl -s "${OLLAMA_URL%/*}/api/tags" | grep -q "\"name\":\"$MODEL_NAME\"" 2>/dev/null; then
        log_success "Model $MODEL_NAME is available"
    else
        log_warning "Model $MODEL_NAME might not be available"
        echo "  Available models:"
        curl -s "${OLLAMA_URL%/*}/api/tags" 2>/dev/null | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    for model in data.get('models', []):
        print(f\"    - {model['name']}\")
except:
    print('    Could not fetch model list')
" || echo "    Could not fetch model list"
        echo "  To install the model: ollama pull $MODEL_NAME"
    fi
}

# Parse command line arguments
ARGS=()
NO_COMMIT=false
NO_FORMAT=false
TEST_MODE=false
LIMIT=""
FILES=()

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -m|--model)
            export MODEL_NAME="$2"
            shift 2
            ;;
        -u|--url)
            export OLLAMA_URL="$2"
            shift 2
            ;;
        -b|--branch)
            export BRANCH_NAME="$2"
            shift 2
            ;;
        -n|--no-commit)
            NO_COMMIT=true
            shift
            ;;
        -f|--no-format)
            NO_FORMAT=true
            shift
            ;;
        -l|--limit)
            LIMIT="$2"
            shift 2
            ;;
        -t|--test)
            TEST_MODE=true
            shift
            ;;
        --files)
            shift
            while [[ $# -gt 0 && ! "$1" =~ ^- ]]; do
                FILES+=("$1")
                shift
            done
            ;;
        -*)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
        *)
            ARGS+=("$1")
            shift
            ;;
    esac
done

# Set directory
DIRECTORY="${ARGS[0]:-}"

# Build Python command
PYTHON_CMD="python3 ollama_doxygen_update.py"

if [ -n "$DIRECTORY" ]; then
    PYTHON_CMD="$PYTHON_CMD --dir '$DIRECTORY'"
fi

if [ "$NO_COMMIT" = true ]; then
    PYTHON_CMD="$PYTHON_CMD --no-commit"
fi

if [ "$NO_FORMAT" = true ]; then
    PYTHON_CMD="$PYTHON_CMD --no-format"
fi

if [ -n "$LIMIT" ]; then
    PYTHON_CMD="$PYTHON_CMD --max-entities $LIMIT"
fi

if [ "$TEST_MODE" = true ]; then
    PYTHON_CMD="$PYTHON_CMD --max-entities 2 --no-commit"
    log_info "Test mode: limiting to 2 entities per file, no commits"
fi

if [ ${#FILES[@]} -gt 0 ]; then
    PYTHON_CMD="$PYTHON_CMD --files ${FILES[*]}"
fi

# Show configuration
echo "=== Ollama DoxLLM-IT Integration ==="
echo "Model: $MODEL_NAME"
echo "URL: $OLLAMA_URL"
echo "Branch: $BRANCH_NAME"
echo "Directory: ${DIRECTORY:-current directory}"
if [ ${#FILES[@]} -gt 0 ]; then
    echo "Files: ${FILES[*]}"
fi
echo

# Check dependencies
check_dependencies

echo
log_info "Starting documentation generation..."

# Run the Python script
eval "$PYTHON_CMD"

echo
log_success "Documentation process complete!"

if [ "$TEST_MODE" = true ]; then
    echo
    log_info "This was a test run. To process all files and commit:"
    echo "  $0 ${ARGS[*]}"
fi
