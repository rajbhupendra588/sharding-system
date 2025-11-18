#!/bin/bash

# Get the project root directory (parent of scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🛑 Stopping Sharding System Backend Services..."

# Stop router
if [ -f .router.pid ]; then
    ROUTER_PID=$(cat .router.pid)
    if ps -p $ROUTER_PID > /dev/null 2>&1; then
        echo "Stopping router (PID: $ROUTER_PID)..."
        kill $ROUTER_PID
        rm .router.pid
    fi
fi

# Stop manager
if [ -f .manager.pid ]; then
    MANAGER_PID=$(cat .manager.pid)
    if ps -p $MANAGER_PID > /dev/null 2>&1; then
        echo "Stopping manager (PID: $MANAGER_PID)..."
        kill $MANAGER_PID
        rm .manager.pid
    fi
fi

# Stop etcd
echo "Stopping etcd..."
docker-compose stop etcd

echo "✅ Backend services stopped"

