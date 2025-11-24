#!/bin/bash

# Start Manager Service
# This script starts the manager service for local development

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "Starting Sharding System Manager..."
echo ""

# Check if etcd is running (optional check)
if ! nc -z localhost 2379 2>/dev/null; then
    echo "⚠️  Warning: etcd might not be running on localhost:2379"
    echo "   The manager will try to connect but may fail if etcd is not available"
    echo ""
fi

# Set default config path if not set
export CONFIG_PATH="${CONFIG_PATH:-configs/manager.json}"

# Set JWT secret if not set (development only)
if [ -z "$JWT_SECRET" ]; then
    export JWT_SECRET="development-secret-not-for-production-use-min-32-chars-please-change"
    echo "⚠️  Using development JWT_SECRET. Set JWT_SECRET environment variable for production."
    echo ""
fi

# Start manager
echo "Manager will be available at: http://localhost:8081"
echo "API docs: http://localhost:8081/swagger/"
echo ""
echo "Press Ctrl+C to stop"
echo ""

exec ./bin/manager

