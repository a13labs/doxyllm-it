# Release Process

## Creating a Release

To create a new release, follow these steps:

1. **Ensure all changes are committed and tests pass:**
   ```bash
   go test ./...
   go build
   ```

2. **Create and push a semantic version tag:**
   ```bash
   # For a new major version
   git tag v1.0.0
   
   # For a minor version  
   git tag v1.1.0
   
   # For a patch version
   git tag v1.0.1
   
   # Push the tag to trigger the release workflow
   git push origin v1.0.0
   ```

3. **The GitHub workflow will automatically:**
   - Build binaries for Linux, macOS, and Windows (both AMD64 and ARM64)
   - Run all tests
   - Create a GitHub release with auto-generated release notes
   - Upload all binaries as release assets
   - Create a Docker image (optional)

## Supported Platforms

The release workflow builds binaries for:

### Linux
- `doxyllm-it-linux-amd64` - Linux x86_64
- `doxyllm-it-linux-arm64` - Linux ARM64

### macOS  
- `doxyllm-it-darwin-amd64` - macOS Intel
- `doxyllm-it-darwin-arm64` - macOS Apple Silicon

### Windows
- `doxyllm-it-windows-amd64.exe` - Windows x86_64
- `doxyllm-it-windows-arm64.exe` - Windows ARM64

## Docker Image

A Docker image is also created and pushed to GitHub Container Registry:
```bash
docker pull ghcr.io/username/doxyllm-it:latest
docker pull ghcr.io/username/doxyllm-it:v1.0.0
```

## Version Information

The binary includes embedded version information that can be displayed:
```bash
./doxyllm-it version
./doxyllm-it --version
```

## Installation from Release

1. **Download the appropriate binary** from the GitHub releases page
2. **Make it executable** (Linux/macOS):
   ```bash
   chmod +x doxyllm-it-*
   ```
3. **Move to PATH** (optional):
   ```bash
   sudo mv doxyllm-it-* /usr/local/bin/doxyllm-it
   ```
4. **Verify installation**:
   ```bash
   doxyllm-it version
   ```

## Development Builds

For development, the version will show as `dev (commit-hash)`:
```bash
go build -ldflags="-X main.version=dev -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```
