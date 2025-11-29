#!/bin/bash
set -e

echo "Building router..."
go build -o bin/router ./cmd/router

echo "Building manager..."
go build -o bin/manager ./cmd/manager

echo "Build complete."
