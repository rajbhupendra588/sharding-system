#!/bin/bash

set -e

echo "Testing all client applications..."

# Test Go App 1
echo "Testing ecommerce-service..."
kubectl wait --for=condition=available deployment/ecommerce-service -n ecommerce-cluster-1 --timeout=60s || true
kubectl port-forward -n ecommerce-cluster-1 svc/ecommerce-service 8081:8080 &
PF1_PID=$!
sleep 3
curl -s http://localhost:8081/health | jq . || echo "Health check failed"
curl -s http://localhost:8081/api/products | jq . | head -20 || echo "Products API failed"
kill $PF1_PID 2>/dev/null || true

# Test Go App 2
echo "Testing inventory-service..."
kubectl wait --for=condition=available deployment/inventory-service -n inventory-cluster-2 --timeout=60s || true
kubectl port-forward -n inventory-cluster-2 svc/inventory-service 8082:8080 &
PF2_PID=$!
sleep 3
curl -s http://localhost:8082/health | jq . || echo "Health check failed"
curl -s http://localhost:8082/api/inventory | jq . | head -20 || echo "Inventory API failed"
kill $PF2_PID 2>/dev/null || true

# Test Java App 1
echo "Testing order-service..."
kubectl wait --for=condition=available deployment/order-service -n orders-cluster-1 --timeout=60s || true
kubectl port-forward -n orders-cluster-1 svc/order-service 8083:8080 &
PF3_PID=$!
sleep 3
curl -s http://localhost:8083/health | jq . || echo "Health check failed"
curl -s http://localhost:8083/api/orders | jq . | head -20 || echo "Orders API failed"
kill $PF3_PID 2>/dev/null || true

# Test Java App 2
echo "Testing payment-service..."
kubectl wait --for=condition=available deployment/payment-service -n payments-cluster-2 --timeout=60s || true
kubectl port-forward -n payments-cluster-2 svc/payment-service 8084:8080 &
PF4_PID=$!
sleep 3
curl -s http://localhost:8084/health | jq . || echo "Health check failed"
curl -s http://localhost:8084/api/payments | jq . | head -20 || echo "Payments API failed"
kill $PF4_PID 2>/dev/null || true

# Test Java App 3
echo "Testing user-service..."
kubectl wait --for=condition=available deployment/user-service -n users-cluster-3 --timeout=60s || true
kubectl port-forward -n users-cluster-3 svc/user-service 8085:8080 &
PF5_PID=$!
sleep 3
curl -s http://localhost:8085/health | jq . || echo "Health check failed"
curl -s http://localhost:8085/api/users | jq . | head -20 || echo "Users API failed"
kill $PF5_PID 2>/dev/null || true

echo ""
echo "All applications tested!"
echo ""
echo "To test manually:"
echo "  kubectl port-forward -n ecommerce-cluster-1 svc/ecommerce-service 8080:8080"
echo "  curl http://localhost:8080/health"
echo "  curl http://localhost:8080/api/products"

