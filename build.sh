#!/usr/bin/env bash

# Build script for Heimdall CLI
# Always builds to ./build directory

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Create build directory if it doesn't exist
mkdir -p build

# Build the binary
echo -e "${YELLOW}Building heimdall...${NC}"

if go build -o build/heimdall cmd/heimdall/main.go; then
    echo -e "${GREEN}✓ Build successful!${NC}"
    echo -e "${GREEN}Binary location: ./build/heimdall${NC}"
    
    # Show binary size
    SIZE=$(du -h build/heimdall | cut -f1)
    echo -e "${GREEN}Binary size: $SIZE${NC}"
else
    echo -e "${RED}✗ Build failed!${NC}"
    exit 1
fi