#!/bin/bash

# Kubernetes Deployment Script for Sharding System
# This script deploys all components of the sharding system to Kubernetes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="sharding-system"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${GREEN}=== Sharding System Kubernetes Deployment ===${NC}\n"

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

# Step 1: Create namespace
echo -e "${YELLOW}[1/9] Creating namespace...${NC}"
kubectl apply -f "${SCRIPT_DIR}/namespace.yaml"
echo -e "${GREEN}✓ Namespace created${NC}\n"

# Step 2: Create secrets (check if exists first)
echo -e "${YELLOW}[2/9] Setting up secrets...${NC}"
if kubectl get secret sharding-secrets -n "${NAMESPACE}" &> /dev/null; then
    echo -e "${YELLOW}  Secret 'sharding-secrets' already exists, skipping...${NC}"
    echo -e "${YELLOW}  To update secrets, edit k8s/secrets.yaml and run:${NC}"
    echo -e "${YELLOW}  kubectl apply -f k8s/secrets.yaml${NC}"
else
    # Check if JWT_SECRET is set in environment
    if [ -z "$JWT_SECRET" ]; then
        echo -e "${YELLOW}  JWT_SECRET not set in environment${NC}"
        echo -e "${YELLOW}  Generating a random JWT secret...${NC}"
        JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || head -c 32 /dev/urandom | base64)
    fi
    
    # Create secret from template
    if [ -f "${SCRIPT_DIR}/secrets.yaml" ]; then
        # Replace placeholder if JWT_SECRET is set
        if [ ! -z "$JWT_SECRET" ]; then
            # Escape special characters in JWT_SECRET for sed
            ESCAPED_JWT=$(echo "$JWT_SECRET" | sed 's/[[\.*^$()+?{|]/\\&/g')
            sed "s|change-me-to-a-secure-random-string-at-least-32-characters-long|${ESCAPED_JWT}|g" \
                "${SCRIPT_DIR}/secrets.yaml" | kubectl apply -f -
        else
            kubectl apply -f "${SCRIPT_DIR}/secrets.yaml"
            echo -e "${RED}  WARNING: Using default JWT secret from secrets.yaml${NC}"
            echo -e "${RED}  Please update it with: kubectl edit secret sharding-secrets -n ${NAMESPACE}${NC}"
        fi
    else
        # Create secret directly
        kubectl create secret generic sharding-secrets \
            --from-literal=jwt-secret="${JWT_SECRET:-change-me-to-a-secure-random-string-at-least-32-characters-long}" \
            -n "${NAMESPACE}" \
            --dry-run=client -o yaml | kubectl apply -f -
    fi
fi
echo -e "${GREEN}✓ Secrets configured${NC}\n"

# Step 3: Create ConfigMap
echo -e "${YELLOW}[3/9] Creating ConfigMap...${NC}"
kubectl apply -f "${SCRIPT_DIR}/configmap.yaml"
echo -e "${GREEN}✓ ConfigMap created${NC}\n"

# Step 4: Create RBAC (ServiceAccount, ClusterRole, ClusterRoleBinding)
echo -e "${YELLOW}[4/9] Setting up RBAC for Kubernetes discovery...${NC}"
kubectl apply -f "${SCRIPT_DIR}/rbac-discovery.yaml"
echo -e "${GREEN}✓ RBAC configured${NC}\n"

# Step 5: Deploy etcd
echo -e "${YELLOW}[5/9] Deploying etcd cluster...${NC}"
kubectl apply -f "${SCRIPT_DIR}/etcd-deployment.yaml"
echo -e "${GREEN}✓ etcd StatefulSet created${NC}\n"
echo -e "${YELLOW}  Waiting for etcd to be ready (this may take a minute)...${NC}"
kubectl wait --for=condition=ready --timeout=300s pod/etcd-0 -n "${NAMESPACE}" || true
kubectl wait --for=condition=ready --timeout=300s pod/etcd-1 -n "${NAMESPACE}" || true
kubectl wait --for=condition=ready --timeout=300s pod/etcd-2 -n "${NAMESPACE}" || true

# Step 6: Deploy Manager
echo -e "${YELLOW}[6/9] Deploying Manager...${NC}"
kubectl apply -f "${SCRIPT_DIR}/manager-deployment.yaml"
echo -e "${GREEN}✓ Manager deployment created${NC}\n"

# Step 7: Deploy Router
echo -e "${YELLOW}[7/9] Deploying Router...${NC}"
kubectl apply -f "${SCRIPT_DIR}/router-deployment.yaml"
echo -e "${GREEN}✓ Router deployment created${NC}\n"

# Step 8: Deploy Proxy
echo -e "${YELLOW}[8/9] Deploying Proxy...${NC}"
kubectl apply -f "${SCRIPT_DIR}/proxy-deployment.yaml"
echo -e "${GREEN}✓ Proxy deployment created${NC}\n"

# Step 9: Wait for deployments to be ready
echo -e "${YELLOW}[9/9] Waiting for deployments to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/manager -n "${NAMESPACE}" || true
kubectl wait --for=condition=available --timeout=300s deployment/router -n "${NAMESPACE}" || true
kubectl wait --for=condition=available --timeout=300s deployment/proxy -n "${NAMESPACE}" || true

echo -e "\n${GREEN}=== Deployment Complete ===${NC}\n"

# Show status
echo -e "${YELLOW}Deployment Status:${NC}"
kubectl get deployments -n "${NAMESPACE}"
echo ""
kubectl get services -n "${NAMESPACE}"
echo ""
kubectl get pods -n "${NAMESPACE}"

# Show service endpoints
echo -e "\n${YELLOW}Service Endpoints:${NC}"
MANAGER_IP=$(kubectl get service manager -n "${NAMESPACE}" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "Pending")
ROUTER_IP=$(kubectl get service router -n "${NAMESPACE}" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "Pending")

if [ "$MANAGER_IP" != "Pending" ] && [ ! -z "$MANAGER_IP" ]; then
    echo -e "  Manager API: http://${MANAGER_IP}:8081"
else
    echo -e "  Manager API: Use port-forward: kubectl port-forward svc/manager 8081:8081 -n ${NAMESPACE}"
fi

if [ "$ROUTER_IP" != "Pending" ] && [ ! -z "$ROUTER_IP" ]; then
    echo -e "  Router API: http://${ROUTER_IP}:8080"
else
    echo -e "  Router API: Use port-forward: kubectl port-forward svc/router 8080:8080 -n ${NAMESPACE}"
fi

PROXY_IP=$(kubectl get service proxy -n "${NAMESPACE}" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "Pending")
if [ "$PROXY_IP" != "Pending" ] && [ ! -z "$PROXY_IP" ]; then
    echo -e "  Proxy (PostgreSQL): ${PROXY_IP}:5432"
    echo -e "  Proxy Admin API: http://${PROXY_IP}:8082"
else
    echo -e "  Proxy (PostgreSQL): Use port-forward: kubectl port-forward svc/proxy 5432:5432 -n ${NAMESPACE}"
    echo -e "  Proxy Admin API: Use port-forward: kubectl port-forward svc/proxy 8082:8082 -n ${NAMESPACE}"
fi

echo -e "\n${YELLOW}Useful Commands:${NC}"
echo -e "  View logs: kubectl logs -f deployment/manager -n ${NAMESPACE}"
echo -e "  View pods: kubectl get pods -n ${NAMESPACE}"
echo -e "  Port forward: kubectl port-forward svc/manager 8081:8081 -n ${NAMESPACE}"
echo -e "  Delete deployment: kubectl delete -f ${SCRIPT_DIR}/ -n ${NAMESPACE}"

echo -e "\n${GREEN}✓ All done!${NC}\n"

