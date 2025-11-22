#!/bin/bash

# Demo script showing how Sharding System helps Java Spring Boot Application

echo "=========================================="
echo "Sharding System Benefits Demo"
echo "=========================================="
echo ""

echo "1. CHECKING SHARDING SYSTEM STATUS"
echo "-----------------------------------"
echo "Router (Data Plane):"
curl -s http://localhost:8080/health && echo ""
echo ""
echo "Manager (Control Plane):"
curl -s http://localhost:8081/health && echo ""
echo ""

echo "2. SEEING SHARD ROUTING IN ACTION"
echo "-----------------------------------"
echo "Testing different user IDs to see shard distribution:"
echo ""
for key in "user-alice" "user-bob" "user-charlie" "user-diana"; do
    shard=$(curl -s "http://localhost:8080/v1/shard-for-key?key=$key" | grep -o '"shard_id":"[^"]*"' | cut -d'"' -f4)
    echo "  $key → Shard: $shard"
done
echo ""

echo "3. JAVA APP USING SHARDING SYSTEM"
echo "-----------------------------------"
echo "When Java app creates a user, Sharding System automatically routes to correct shard:"
echo ""
echo "Creating user via Java API..."
response=$(curl -s -X POST http://localhost:8082/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "id": "demo-user-001",
    "username": "demo_user",
    "email": "demo@example.com",
    "fullName": "Demo User"
  }')
echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
echo ""

echo "4. CHECKING WHICH SHARD THIS USER IS ON"
echo "----------------------------------------"
shard_info=$(curl -s "http://localhost:8080/v1/shard-for-key?key=demo-user-001")
echo "$shard_info" | python3 -m json.tool 2>/dev/null || echo "$shard_info"
echo ""

echo "5. BENEFITS SUMMARY"
echo "-------------------"
echo "✅ Java app doesn't need to know about shards"
echo "✅ Sharding System handles routing automatically"
echo "✅ Queries hit only ONE shard (fast!)"
echo "✅ Can add more shards without changing Java code"
echo "✅ Data distributed evenly across shards"
echo ""
echo "=========================================="
echo "Demo Complete!"
echo "=========================================="

