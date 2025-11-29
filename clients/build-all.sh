#!/bin/bash
set -e

echo "Building all client applications..."

# Build Go applications
echo "Building Go applications..."
cd go-app-1
go mod tidy
docker build -t go-app-1:latest .
cd ..

cd go-app-2
go mod tidy
docker build -t go-app-2:latest .
cd ..

cd go-app-3
go mod tidy
docker build -t go-app-3:latest .
cd ..

# Build Java applications
echo "Building Java applications..."
cd java-app-1
mvn clean package -DskipTests
docker build -t java-app-1:latest .
cd ..

cd java-app-2
mvn clean package -DskipTests
docker build -t java-app-2:latest .
cd ..

echo "All applications built successfully!"

