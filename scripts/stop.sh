#!/bin/bash

# Stop All Sharding System Services
# This script stops manager, router, and etcd services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🛑 Stopping Sharding System Services..."
echo ""

# Stop manager
echo "1️⃣ Stopping manager..."
if [ -f .manager.pid ]; then
    MANAGER_PID=$(cat .manager.pid)
    if ps -p $MANAGER_PID > /dev/null 2>&1; then
        echo "   Stopping manager (PID: $MANAGER_PID)..."
        kill $MANAGER_PID 2>/dev/null || true
        sleep 2
        if ps -p $MANAGER_PID > /dev/null 2>&1; then
            kill -9 $MANAGER_PID 2>/dev/null || true
        fi
        echo "   ✅ Manager stopped"
    else
        echo "   Manager not running"
    fi
    rm -f .manager.pid
else
    echo "   No manager PID file found"
fi

# Also kill any processes on port 8081
lsof -ti :8081 2>/dev/null | xargs kill 2>/dev/null || true

echo ""

# Stop router
echo "2️⃣ Stopping router..."
if [ -f .router.pid ]; then
    ROUTER_PID=$(cat .router.pid)
    if ps -p $ROUTER_PID > /dev/null 2>&1; then
        echo "   Stopping router (PID: $ROUTER_PID)..."
        kill $ROUTER_PID 2>/dev/null || true
        sleep 2
        if ps -p $ROUTER_PID > /dev/null 2>&1; then
            kill -9 $ROUTER_PID 2>/dev/null || true
        fi
        echo "   ✅ Router stopped"
    else
        echo "   Router not running"
    fi
    rm -f .router.pid
else
    echo "   No router PID file found"
fi

# Also kill any processes on port 8080
lsof -ti :8080 2>/dev/null | xargs kill 2>/dev/null || true

echo ""

# Stop frontend
echo "3️⃣ Stopping frontend..."
if [ -f .frontend.pid ]; then
    FRONTEND_PID=$(cat .frontend.pid)
    if ps -p $FRONTEND_PID > /dev/null 2>&1; then
        echo "   Stopping frontend (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID 2>/dev/null || true
        sleep 2
        if ps -p $FRONTEND_PID > /dev/null 2>&1; then
            kill -9 $FRONTEND_PID 2>/dev/null || true
        fi
        echo "   ✅ Frontend stopped"
    else
        echo "   Frontend not running"
    fi
    rm -f .frontend.pid
else
    echo "   No frontend PID file found"
fi

# Also kill any processes on port 3000
lsof -ti :3000 2>/dev/null | xargs kill 2>/dev/null || true

echo ""

# Stop etcd
echo "4️⃣ Stopping etcd..."
# Check if Docker is available and running
if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    if docker ps 2>/dev/null | grep -q sharding-etcd; then
        docker-compose stop etcd 2>/dev/null || docker stop sharding-etcd 2>/dev/null || true
        echo "   ✅ etcd stopped"
    else
        echo "   etcd container not running"
    fi
else
    echo "   Docker not available or not running - skipping etcd stop"
fi

echo ""
echo "✅ All services stopped!"
echo ""

