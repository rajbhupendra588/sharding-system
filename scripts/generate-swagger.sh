#!/bin/bash

# Generate Swagger documentation for Router and Manager APIs

set -e

echo "Generating Swagger documentation..."

# Ensure swag is installed
if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Add go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Generate Router API docs
echo "Generating Router API documentation..."
swag init -g cmd/router/main.go -o docs/swagger/router --parseDependency --parseInternal --instanceName router

# Generate Manager API docs
echo "Generating Manager API documentation..."
swag init -g cmd/manager/main.go -o docs/swagger/manager --parseDependency --parseInternal --instanceName manager

echo "Swagger documentation generated successfully!"
echo ""
echo "Router API docs: docs/swagger/router/"
echo "Manager API docs: docs/swagger/manager/"
echo ""
echo "Access Swagger UI:"
echo "  Router: http://localhost:8080/swagger/index.html"
echo "  Manager: http://localhost:8081/swagger/index.html"


