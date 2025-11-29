#!/bin/bash

set -e

echo "=========================================="
echo "Building and Testing All Client Applications"
echo "=========================================="

# Build Go applications
echo ""
echo "Building Go applications..."
echo "---------------------------"

cd go-app-1
echo "Building go-app-1 (E-commerce)..."
go build -o /tmp/ecommerce-app main.go
echo "✓ go-app-1 built successfully"
cd ..

cd go-app-2
echo "Building go-app-2 (Users)..."
go build -o /tmp/users-app main.go
echo "✓ go-app-2 built successfully"
cd ..

cd go-app-3
echo "Building go-app-3 (Orders)..."
go build -o /tmp/orders-app main.go
echo "✓ go-app-3 built successfully"
cd ..

# Build Java applications
echo ""
echo "Building Java applications..."
echo "-----------------------------"

cd java-app-1
echo "Building java-app-1 (Products)..."
mvn clean package -DskipTests -q
echo "✓ java-app-1 built successfully"
cd ..

cd java-app-2
echo "Building java-app-2 (Inventory)..."
mvn clean package -DskipTests -q
echo "✓ java-app-2 built successfully"
cd ..

echo ""
echo "=========================================="
echo "All applications built successfully!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Build Docker images: ./deploy-all.sh"
echo "2. Deploy to Kubernetes: kubectl apply -f <app>/k8s/deployment.yaml"
echo "3. Register cluster and scan: POST /api/v1/clusters/scan"
echo "4. View metrics: GET /metrics"

