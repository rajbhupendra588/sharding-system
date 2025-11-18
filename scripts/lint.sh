#!/bin/bash
set -e

echo "Running linters..."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found. Installing..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
fi

# Run linter
golangci-lint run ./...

echo "Linting complete!"

