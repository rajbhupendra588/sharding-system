#!/bin/bash

# Get the project root directory (parent of scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🛑 Stopping Sharding System (Backend + Frontend + Client Apps)"
echo ""

# Stop Java E-Commerce Service
if [ -f .ecommerce.pid ]; then
    ECOMMERCE_PID=$(cat .ecommerce.pid)
    if ps -p $ECOMMERCE_PID > /dev/null 2>&1; then
        echo "Stopping Java E-Commerce Service (PID: $ECOMMERCE_PID)..."
        kill $ECOMMERCE_PID 2>/dev/null || true
        sleep 2
        # Force kill if still running
        if ps -p $ECOMMERCE_PID > /dev/null 2>&1; then
            kill -9 $ECOMMERCE_PID 2>/dev/null || true
        fi
        rm .ecommerce.pid
        echo "✅ Java E-Commerce Service stopped"
    else
        rm .ecommerce.pid
    fi
fi

# Stop Quarkus Service
if [ -f .quarkus.pid ]; then
    QUARKUS_PID=$(cat .quarkus.pid)
    if ps -p $QUARKUS_PID > /dev/null 2>&1; then
        echo "Stopping Quarkus Service (PID: $QUARKUS_PID)..."
        kill $QUARKUS_PID 2>/dev/null || true
        sleep 2
        # Force kill if still running
        if ps -p $QUARKUS_PID > /dev/null 2>&1; then
            kill -9 $QUARKUS_PID 2>/dev/null || true
        fi
        rm .quarkus.pid
        echo "✅ Quarkus Service stopped"
    else
        rm .quarkus.pid
    fi
fi

# Stop frontend
if [ -f .frontend.pid ]; then
    FRONTEND_PID=$(cat .frontend.pid)
    if ps -p $FRONTEND_PID > /dev/null 2>&1; then
        echo "Stopping frontend (PID: $FRONTEND_PID)..."
        kill $FRONTEND_PID 2>/dev/null || true
        sleep 2
        # Force kill if still running
        if ps -p $FRONTEND_PID > /dev/null 2>&1; then
            kill -9 $FRONTEND_PID 2>/dev/null || true
        fi
        rm .frontend.pid
        echo "✅ Frontend stopped"
    else
        rm .frontend.pid
    fi
fi

# Stop backend services (router, manager, etcd)
echo ""
echo "Stopping backend services..."
"$SCRIPT_DIR/stop-backend.sh"

# Also check for any processes on common ports and kill them
echo ""
echo "Checking for processes on service ports..."

# Check port 8082 (E-Commerce Service)
LSOF_OUTPUT=$(lsof -ti :8082 2>/dev/null || true)
if [ ! -z "$LSOF_OUTPUT" ]; then
    echo "Stopping processes on port 8082..."
    echo "$LSOF_OUTPUT" | xargs kill 2>/dev/null || true
fi

# Check port 8083 (Quarkus Service)
LSOF_OUTPUT=$(lsof -ti :8083 2>/dev/null || true)
if [ ! -z "$LSOF_OUTPUT" ]; then
    echo "Stopping processes on port 8083..."
    echo "$LSOF_OUTPUT" | xargs kill 2>/dev/null || true
fi

# Check port 3000 (Frontend)
LSOF_OUTPUT=$(lsof -ti :3000 2>/dev/null || true)
if [ ! -z "$LSOF_OUTPUT" ]; then
    echo "Stopping processes on port 3000..."
    echo "$LSOF_OUTPUT" | xargs kill 2>/dev/null || true
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All services stopped"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

