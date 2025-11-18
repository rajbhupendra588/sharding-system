#!/bin/bash

set -e

# Get the project root directory (parent of scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🚀 Starting Sharding System Backend Services..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "Please install Go 1.21+ from https://go.dev"
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo "❌ Error: Docker is not running"
    echo "Please start Docker Desktop or Docker daemon"
    exit 1
fi

# Install backend dependencies
echo "📦 Installing backend dependencies..."
go mod download
go mod tidy

# Start etcd
echo "🔧 Starting etcd..."
docker-compose up -d etcd
echo "⏳ Waiting for etcd to be ready..."
sleep 5

# Build backend
echo "🔨 Building backend services..."
go build -o bin/router ./cmd/router
go build -o bin/manager ./cmd/manager

# Start router
echo "🌐 Starting router on port 8080..."
./bin/router > logs/router.log 2>&1 &
ROUTER_PID=$!
echo "Router started with PID: $ROUTER_PID"
echo $ROUTER_PID > .router.pid

# Start manager
echo "⚙️  Starting manager on port 8081..."
./bin/manager > logs/manager.log 2>&1 &
MANAGER_PID=$!
echo "Manager started with PID: $MANAGER_PID"
echo $MANAGER_PID > .manager.pid

sleep 2

echo ""
echo "✅ Backend services started successfully!"
echo ""
echo "📊 Services:"
echo "   - Router:    http://localhost:8080"
echo "   - Manager:   http://localhost:8081"
echo "   - etcd:      http://localhost:2379"
echo ""
echo "📝 Logs:"
echo "   - Router:    tail -f logs/router.log"
echo "   - Manager:   tail -f logs/manager.log"
echo ""
echo "🛑 To stop services: $SCRIPT_DIR/stop-backend.sh"
echo ""

