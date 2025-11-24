#!/bin/bash

# Restart Manager Service
# This script stops and restarts the manager service

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🔄 Restarting Sharding System Manager..."
echo ""

# Find and kill existing manager process
if [ -f .manager.pid ]; then
    MANAGER_PID=$(cat .manager.pid)
    if ps -p $MANAGER_PID > /dev/null 2>&1; then
        echo "Stopping existing manager (PID: $MANAGER_PID)..."
        kill $MANAGER_PID 2>/dev/null || true
        sleep 2
        # Force kill if still running
        if ps -p $MANAGER_PID > /dev/null 2>&1; then
            kill -9 $MANAGER_PID 2>/dev/null || true
        fi
    fi
    rm -f .manager.pid
fi

# Also check for manager processes on port 8081
LSOF_OUTPUT=$(lsof -ti :8081 2>/dev/null || true)
if [ ! -z "$LSOF_OUTPUT" ]; then
    echo "Stopping processes on port 8081..."
    echo "$LSOF_OUTPUT" | xargs kill 2>/dev/null || true
    sleep 2
fi

# Wait a moment
sleep 1

# Start manager
echo "Starting manager..."
export CONFIG_PATH="${CONFIG_PATH:-configs/manager.json}"

# Set JWT secret if not set (development only)
if [ -z "$JWT_SECRET" ]; then
    export JWT_SECRET="development-secret-not-for-production-use-min-32-chars-please-change"
    echo "⚠️  Using development JWT_SECRET"
fi

./bin/manager > logs/manager.log 2>&1 &
MANAGER_PID=$!
echo $MANAGER_PID > .manager.pid

sleep 2

# Check if it's running
if ps -p $MANAGER_PID > /dev/null 2>&1; then
    echo "✅ Manager restarted successfully (PID: $MANAGER_PID)"
    echo ""
    echo "Manager API: http://localhost:8081"
    echo "API docs: http://localhost:8081/swagger/"
    echo "Discovery endpoint: http://localhost:8081/api/v1/client-apps/discover"
    echo ""
    echo "Logs: tail -f logs/manager.log"
else
    echo "❌ Failed to start manager. Check logs/manager.log"
    exit 1
fi

