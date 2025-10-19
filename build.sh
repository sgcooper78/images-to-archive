#!/bin/bash

# Build script for CBZ WebP Converter
# This script builds the application for multiple platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building CBZ WebP Converter for multiple platforms...${NC}"

# Create build directory
mkdir -p build
cd build

# Build for different platforms
platforms=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    
    echo -e "${YELLOW}Building for $os/$arch...${NC}"
    
    if [ "$os" = "windows" ]; then
        ext=".exe"
    else
        ext=""
    fi
    
    GOOS=$os GOARCH=$arch go build -ldflags="-s -w" -o "cbz-converter-$os-$arch$ext" ../main.go
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built cbz-converter-$os-$arch$ext${NC}"
    else
        echo -e "${RED}✗ Failed to build for $os/$arch${NC}"
        exit 1
    fi
done

# Create checksums
echo -e "${YELLOW}Creating checksums...${NC}"
sha256sum * > checksums.txt

echo -e "${GREEN}Build complete! Files are in the build/ directory${NC}"
echo -e "${YELLOW}To create a release:${NC}"
echo "1. Commit and push your changes"
echo "2. Create and push a tag: git tag v1.0.0 && git push origin v1.0.0"
echo "3. GitHub Actions will automatically create a release with the built binaries"
