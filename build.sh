#!/bin/bash

# Define application name and output directory
APP_NAME="go-print"
OUTPUT_DIR="build"

# Define target platforms
# PLATFORMS=("linux/amd64" "linux/arm64" "windows/amd64" "darwin/amd64" "darwin/arm64")
PLATFORMS=("windows/amd64" "linux/amd64")

# Clean up previous builds
echo "Cleaning up previous builds..."
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    IFS="/" read -r -a PLATFORM_SPLIT <<< "$PLATFORM"
    GOOS=${PLATFORM_SPLIT[0]}
    GOARCH=${PLATFORM_SPLIT[1]}
    OUTPUT_NAME="$OUTPUT_DIR/$APP_NAME-$GOOS-$GOARCH"

    if [ "$GOOS" == "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi

    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -o $OUTPUT_NAME .

    if [ $? -ne 0 ]; then
        echo "Failed to build for $GOOS/$GOARCH"
        exit 1
    fi
done

echo "Build completed. Binaries are located in the $OUTPUT_DIR directory."
