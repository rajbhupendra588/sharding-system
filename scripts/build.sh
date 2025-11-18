#!/bin/bash
set -e

echo "Building sharding-system..."

# Build router
echo "Building router..."
go build -o bin/router ./cmd/router

# Build manager
echo "Building manager..."
go build -o bin/manager ./cmd/manager

echo "Build complete!"
echo "Binaries are in bin/ directory"

