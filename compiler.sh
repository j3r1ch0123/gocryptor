#!/bin/bash

# Check if the correct number of arguments is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <path-to-go-file>"
    exit 1
fi

# Get the Go file path from the argument
GO_FILE=$1

# Check if the provided file exists
if [ ! -f "$GO_FILE" ]; then
    echo "File not found: $GO_FILE"
    exit 1
fi

# Extract the filename without the extension
OUTPUT_FILE=$(basename "$GO_FILE" .go)

# Set the GOOS and GOARCH environment variables for cross-compiling
export GOOS=windows
export GOARCH=amd64

# Compile the Go file
go build -ldflags "-H=windowsgui" -o "$OUTPUT_FILE.exe" "$GO_FILE"

# Check if the compilation was successful
if [ $? -eq 0 ]; then
    echo "Compilation successful: $OUTPUT_FILE.exe"
else
    echo "Compilation failed"
    exit 1
fi

