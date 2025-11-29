#!/bin/bash

# Comprehensive Kubernetes Deployment Script
# Deploys sharding-system and all client applications to Kubernetes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CLIENTS_DIR="${PROJECT_ROOT}/clients"

echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     Kubernetes Deployment - Sharding System + Clients        ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}\n"

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if we can connect to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Error: Cannot connect to Kubernetes cluster${NC}"
    echo "Please ensure kubectl is configured correctly"
    exit 1
fi

echo -e "${GREEN}✓ Kubernetes cluster connection verified${NC}\n"

# Function to deploy client application
deploy_client_app() {
    local app_name=$1
    local app_dir=$2
    
    echo -e "${BLUE}[Client] Deploying ${app_name}...${NC}"
    
    if [ ! -d "${app_dir}/k8s" ]; then
        echo -e "${YELLOW}  Warning: No k8s directory found for ${app_name}, skipping...${NC}"
        return
    fi
    
    # Deploy all YAML files in the k8s directory
    for yaml_file in "${app_dir}"/k8s/*.yaml; do
        if [ -f "$yaml_file" ]; then
            kubectl apply -f "$yaml_file"
        fi
    done
    
    echo -e "${GREEN}  ✓ ${app_name} deployed${NC}\n"
}

# Step 1: Deploy Sharding System
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}Step 1: Deploying Sharding System Components${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}\n"

cd "${SCRIPT_DIR}"
bash deploy.sh

echo -e "\n${GREEN}✓ Sharding System deployed${NC}\n"

# Step 2: Deploy Client Applications
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}Step 2: Deploying Client Applications${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}\n"

cd "${CLIENTS_DIR}"

# Deploy Go applications
if [ -d "go-app-1" ]; then
    deploy_client_app "go-app-1 (E-Commerce)" "go-app-1"
fi

if [ -d "go-app-2" ]; then
    deploy_client_app "go-app-2 (Users)" "go-app-2"
fi

if [ -d "go-app-3" ]; then
    deploy_client_app "go-app-3 (Orders)" "go-app-3"
fi

# Deploy Java applications
if [ -d "java-app-1" ]; then
    deploy_client_app "java-app-1 (Products)" "java-app-1"
fi

if [ -d "java-app-2" ]; then
    deploy_client_app "java-app-2 (Inventory)" "java-app-2"
fi

if [ -d "java-app-3" ]; then
    deploy_client_app "java-app-3 (User Service)" "java-app-3"
fi

# Step 3: Wait for all deployments
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}Step 3: Waiting for all deployments to be ready...${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}\n"

echo -e "${BLUE}Waiting for sharding-system components...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/manager -n sharding-system || true
kubectl wait --for=condition=available --timeout=300s deployment/router -n sharding-system || true
kubectl wait --for=condition=available --timeout=300s deployment/proxy -n sharding-system || true

echo -e "\n${BLUE}Waiting for client applications...${NC}"
# Wait for client app deployments (with error handling)
for namespace in ecommerce-ns users-ns orders-ns products-ns inventory-ns users-cluster-3; do
    if kubectl get namespace "$namespace" &> /dev/null; then
        for deployment in $(kubectl get deployments -n "$namespace" -o name 2>/dev/null | cut -d'/' -f2); do
            kubectl wait --for=condition=available --timeout=180s "deployment/$deployment" -n "$namespace" || true
        done
    fi
done

# Step 4: Display Status
echo -e "\n${GREEN}╔═══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Deployment Complete                        ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════╝${NC}\n"

echo -e "${YELLOW}Sharding System Status:${NC}"
kubectl get all -n sharding-system

echo -e "\n${YELLOW}Client Applications Status:${NC}"
for namespace in ecommerce-ns users-ns orders-ns products-ns inventory-ns users-cluster-3; do
    if kubectl get namespace "$namespace" &> /dev/null; then
        echo -e "\n${BLUE}Namespace: ${namespace}${NC}"
        kubectl get deployments,services -n "$namespace" 2>/dev/null || echo "  No resources found"
    fi
done

# Step 5: Display Service Endpoints
echo -e "\n${YELLOW}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}Service Endpoints:${NC}"
echo -e "${YELLOW}═══════════════════════════════════════════════════════════════${NC}\n"

echo -e "${BLUE}Sharding System:${NC}"
MANAGER_IP=$(kubectl get service manager -n sharding-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
ROUTER_IP=$(kubectl get service router -n sharding-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
PROXY_IP=$(kubectl get service proxy -n sharding-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")

if [ ! -z "$MANAGER_IP" ] && [ "$MANAGER_IP" != "Pending" ]; then
    echo -e "  Manager API:    http://${MANAGER_IP}:8081"
else
    echo -e "  Manager API:    kubectl port-forward svc/manager 8081:8081 -n sharding-system"
fi

if [ ! -z "$ROUTER_IP" ] && [ "$ROUTER_IP" != "Pending" ]; then
    echo -e "  Router API:     http://${ROUTER_IP}:8080"
else
    echo -e "  Router API:     kubectl port-forward svc/router 8080:8080 -n sharding-system"
fi

if [ ! -z "$PROXY_IP" ] && [ "$PROXY_IP" != "Pending" ]; then
    echo -e "  Proxy (PostgreSQL): ${PROXY_IP}:5432"
    echo -e "  Proxy Admin:    http://${PROXY_IP}:8082"
else
    echo -e "  Proxy (PostgreSQL): kubectl port-forward svc/proxy 5432:5432 -n sharding-system"
    echo -e "  Proxy Admin:    kubectl port-forward svc/proxy 8082:8082 -n sharding-system"
fi

echo -e "\n${BLUE}Client Applications:${NC}"
for namespace in ecommerce-ns users-ns orders-ns products-ns inventory-ns users-cluster-3; do
    if kubectl get namespace "$namespace" &> /dev/null; then
        for service in $(kubectl get services -n "$namespace" -o name 2>/dev/null | cut -d'/' -f2 | grep -v postgres); do
            SERVICE_IP=$(kubectl get service "$service" -n "$namespace" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
            PORT=$(kubectl get service "$service" -n "$namespace" -o jsonpath='{.spec.ports[0].port}' 2>/dev/null || echo "8080")
            if [ ! -z "$SERVICE_IP" ] && [ "$SERVICE_IP" != "Pending" ]; then
                echo -e "  ${service} (${namespace}): http://${SERVICE_IP}:${PORT}"
            else
                echo -e "  ${service} (${namespace}): kubectl port-forward svc/${service} ${PORT}:${PORT} -n ${namespace}"
            fi
        done
    fi
done

echo -e "\n${YELLOW}Useful Commands:${NC}"
echo -e "  View all pods:        kubectl get pods --all-namespaces"
echo -e "  View sharding logs:   kubectl logs -f deployment/manager -n sharding-system"
echo -e "  View router logs:     kubectl logs -f deployment/router -n sharding-system"
echo -e "  View proxy logs:      kubectl logs -f deployment/proxy -n sharding-system"
echo -e "  Delete all:           kubectl delete -f ${SCRIPT_DIR}/ && kubectl delete namespace ecommerce-ns users-ns orders-ns products-ns inventory-ns users-cluster-3"

echo -e "\n${GREEN}✓ All components deployed successfully!${NC}\n"

