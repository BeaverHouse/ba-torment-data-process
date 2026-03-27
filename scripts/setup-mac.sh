#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR/.."

echo "Building batorment for macOS..."
cd "$PROJECT_DIR"
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o batorment .
echo "Build complete. Run './batorment' from this directory (with .env file)."
