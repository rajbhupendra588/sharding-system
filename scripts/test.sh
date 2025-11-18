#!/bin/bash
set -e

echo "Running tests..."

# Run all tests with coverage
go test -v -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

echo "Test coverage report generated: coverage.html"

