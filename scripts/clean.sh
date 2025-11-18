#!/bin/bash
set -e

echo "Cleaning build artifacts..."

# Remove binaries
rm -rf bin/

# Remove test coverage files
rm -f coverage.out coverage.html

# Remove temporary files
find . -type f -name "*.tmp" -delete
find . -type f -name "*.log" -delete

echo "Clean complete!"

