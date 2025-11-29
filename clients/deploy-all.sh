#!/bin/bash

set -e

echo "Building and deploying all client applications..."

# Build Go applications
echo "Building Go applications..."
cd go-app-1
docker build -t ecommerce-app:latest .
cd ..

cd go-app-2
docker build -t users-app:latest .
cd ..

cd go-app-3
docker build -t orders-app:latest .
cd ..

# Build Java applications
echo "Building Java applications..."
cd java-app-1
docker build -t products-app:latest .
cd ..

cd java-app-2
docker build -t inventory-app:latest .
cd ..

# Deploy to Kubernetes
echo "Deploying to Kubernetes..."

# Deploy Go apps
kubectl apply -f go-app-1/k8s/deployment.yaml
kubectl apply -f go-app-2/k8s/deployment.yaml
kubectl apply -f go-app-3/k8s/deployment.yaml

# Deploy Java apps
kubectl apply -f java-app-1/k8s/deployment.yaml
kubectl apply -f java-app-2/k8s/deployment.yaml

echo "Waiting for deployments to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/ecommerce-app -n ecommerce-ns || true
kubectl wait --for=condition=available --timeout=300s deployment/users-app -n users-ns || true
kubectl wait --for=condition=available --timeout=300s deployment/orders-app -n orders-ns || true
kubectl wait --for=condition=available --timeout=300s deployment/products-app -n products-ns || true
kubectl wait --for=condition=available --timeout=300s deployment/inventory-app -n inventory-ns || true

echo "All applications deployed!"
echo ""
echo "Applications are running in the following namespaces:"
echo "  - ecommerce-ns: ecommerce-app"
echo "  - users-ns: users-app"
echo "  - orders-ns: orders-app"
echo "  - products-ns: products-app"
echo "  - inventory-ns: inventory-app"

