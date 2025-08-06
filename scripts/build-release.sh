#!/bin/bash

# Local release build script for testing
# This simulates what the GitHub workflow does

set -e

VERSION=${1:-"dev"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

echo "Building DoxLLM-IT version: $VERSION"
echo "Commit: $COMMIT"
echo "Date: $DATE"

# Clean previous builds
rm -rf dist/
mkdir -p dist/

# Build for multiple platforms
platforms=(
    "linux/amd64"
    "linux/arm64"  
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

for platform in "${platforms[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    
    output_name="doxyllm-it-${GOOS}-${GOARCH}"
    if [ $GOOS = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    
    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
        -ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" \
        -o "dist/$output_name" .
        
    echo "  ✓ Built dist/$output_name"
done

echo ""
echo "Build complete! Binaries in dist/ directory:"
ls -la dist/

echo ""
echo "Test version info:"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    ./dist/doxyllm-it-linux-amd64 version
elif [[ "$OSTYPE" == "darwin"* ]]; then
    if [[ $(uname -m) == "arm64" ]]; then
        ./dist/doxyllm-it-darwin-arm64 version
    else
        ./dist/doxyllm-it-darwin-amd64 version
    fi
fi
