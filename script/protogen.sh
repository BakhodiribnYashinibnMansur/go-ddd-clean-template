#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}Error: protoc not found${NC}"
    echo "Please install protobuf compiler:"
    echo "  brew install protobuf"
    echo "  or visit: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

echo -e "${BLUE}Starting proto generation...${NC}\n"

# Base directories
PROTO_BASE_DIR="docs/protobuf/proto"
OUTPUT_BASE_DIR="docs/protobuf/genproto"

# Create output base directory if it doesn't exist
mkdir -p "$OUTPUT_BASE_DIR"

TOTAL_GENERATED=0
TOTAL_ERRORS=0

# Detect OS for sed syntax
SED_INPLACE=""
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    SED_INPLACE="-i ''"
    echo -e "${BLUE}Detected OS: macOS${NC}"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    SED_INPLACE="-i"
    echo -e "${BLUE}Detected OS: Linux${NC}"
else
    # Default to macOS syntax
    SED_INPLACE="-i ''"
    echo -e "${YELLOW}Unknown OS, using macOS sed syntax${NC}"
fi
echo ""

# Loop through all version directories (v1, v2, v3, etc.)
for VERSION_DIR in "$PROTO_BASE_DIR"/v*; do
    if [ ! -d "$VERSION_DIR" ]; then
        continue
    fi
    
    VERSION=$(basename "$VERSION_DIR")
    echo -e "${YELLOW}Processing $VERSION...${NC}"
    
    # Create output directory for this version
    OUTPUT_DIR="$OUTPUT_BASE_DIR/$VERSION"
    mkdir -p "$OUTPUT_DIR"
    
    # Find all .proto files in this version directory
    PROTO_FILES=$(find "$VERSION_DIR" -name "*.proto" 2>/dev/null)
    
    if [ -z "$PROTO_FILES" ]; then
        echo -e "${YELLOW}  No proto files found in $VERSION, skipping...${NC}"
        continue
    fi
    
    # Count proto files
    FILE_COUNT=$(echo "$PROTO_FILES" | wc -l | tr -d ' ')
    echo -e "  Found ${GREEN}$FILE_COUNT${NC} proto file(s)"
    
    # Generate proto files
    echo -e "  Generating..."
    
    if protoc --proto_path="$PROTO_BASE_DIR" \
        --go_out="$OUTPUT_BASE_DIR" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$OUTPUT_BASE_DIR" \
        --go-grpc_opt=paths=source_relative \
        $PROTO_FILES 2>&1; then
        echo -e "  ${GREEN}✓${NC} Successfully generated proto files for $VERSION"
        
        # Remove omitempty tags from generated files
        echo -e "  Removing omitempty tags..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS sed syntax
            for goFile in $(find "$OUTPUT_DIR" -name "*.pb.go" -type f); do
                sed -i '' -e "s/,omitempty//" "$goFile"
            done
        else
            # Linux sed syntax
            for goFile in $(find "$OUTPUT_DIR" -name "*.pb.go" -type f); do
                sed -i -e "s/,omitempty//" "$goFile"
            done
        fi
        echo -e "  ${GREEN}✓${NC} Removed omitempty tags"
        
        TOTAL_GENERATED=$((TOTAL_GENERATED + FILE_COUNT))
    else
        echo -e "  ${RED}✗${NC} Failed to generate proto files for $VERSION"
        TOTAL_ERRORS=$((TOTAL_ERRORS + 1))
    fi
    
    echo ""
done

# Summary
echo -e "${BLUE}================================${NC}"
if [ $TOTAL_ERRORS -eq 0 ]; then
    echo -e "${GREEN}✓ Proto generation completed successfully!${NC}"
    echo -e "  Total files processed: ${GREEN}$TOTAL_GENERATED${NC}"
else
    echo -e "${YELLOW}⚠ Proto generation completed with errors${NC}"
    echo -e "  Files processed: $TOTAL_GENERATED"
    echo -e "  Errors: ${RED}$TOTAL_ERRORS${NC}"
    exit 1
fi

