#!/bin/bash

set -e

# Get the project root directory (parent of scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🚀 Starting Sharding System (Backend + Frontend + Client Apps)"
echo ""

# Check if services are already running
if [ -f .router.pid ] || [ -f .manager.pid ] || [ -f .frontend.pid ] || [ -f .ecommerce.pid ] || [ -f .quarkus.pid ]; then
    echo "⚠️  Warning: Some services may already be running."
    echo "   Consider running '$SCRIPT_DIR/stop-everything.sh' first to ensure a clean start."
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
    echo ""
fi

# Create logs directory
mkdir -p logs
mkdir -p examples/java-ecommerce-service/logs
mkdir -p examples/quarkus-service/logs

# Step 1: Start backend services
echo "📦 Step 1/4: Starting backend services (etcd, router, manager)..."
"$SCRIPT_DIR/start-backend.sh"

# Wait for backend to be ready
echo ""
echo "⏳ Waiting for backend services to be ready..."
sleep 5

# Step 2: Start frontend
echo ""
echo "🎨 Step 2/4: Starting frontend UI..."
if command -v osascript &> /dev/null; then
    # macOS - open new terminal
    osascript -e "tell application \"Terminal\" to do script \"cd '$PROJECT_ROOT' && '$SCRIPT_DIR/start-frontend.sh'\""
    echo "✅ Frontend opened in new terminal window"
else
    # Linux/Other - run in background
    "$SCRIPT_DIR/start-frontend.sh" &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > .frontend.pid
    echo "✅ Frontend started with PID: $FRONTEND_PID"
fi

# Wait a bit for frontend to start
sleep 3

# Step 3: Start Java E-Commerce Service
echo ""
echo "🛒 Step 3/4: Starting Java E-Commerce Service..."
cd "$PROJECT_ROOT/examples/java-ecommerce-service"

# Check if Maven is installed
if ! command -v mvn &> /dev/null; then
    echo "⚠️  Warning: Maven not found. Skipping Java E-Commerce Service."
    echo "   Install Maven to start this service: https://maven.apache.org/install.html"
else
    # Check if already built
    if [ ! -f "target/java-ecommerce-service-1.0.0.jar" ]; then
        echo "   Building Java E-Commerce Service..."
        mvn clean install -DskipTests > logs/build.log 2>&1
    fi
    
    # Set environment variable for sharding router
    export SHARDING_ROUTER_URL="http://localhost:8080"
    
    # Check if port 8082 is already in use
    if lsof -ti :8082 > /dev/null 2>&1; then
        echo "⚠️  Warning: Port 8082 is already in use. Skipping Java E-Commerce Service."
        echo "   Stop the process using port 8082 or modify the port configuration."
    else
        # Start the service
        echo "   Starting on port 8082..."
        mvn spring-boot:run > logs/ecommerce-service.log 2>&1 &
        ECOMMERCE_PID=$!
        echo $ECOMMERCE_PID > "$PROJECT_ROOT/.ecommerce.pid"
        echo "✅ Java E-Commerce Service started with PID: $ECOMMERCE_PID"
    fi
fi

cd "$PROJECT_ROOT"

# Step 4: Start Quarkus Service
echo ""
echo "⚡ Step 4/4: Starting Quarkus Service..."
cd "$PROJECT_ROOT/examples/quarkus-service"

# Check if Maven is installed
if ! command -v mvn &> /dev/null; then
    echo "⚠️  Warning: Maven not found. Skipping Quarkus Service."
else
    # Check if port 8083 is already in use
    if lsof -ti :8083 > /dev/null 2>&1; then
        echo "⚠️  Warning: Port 8083 is already in use. Skipping Quarkus Service."
        echo "   Stop the process using port 8083 or modify the port configuration."
    else
        # Start Quarkus in dev mode on port 8083
        echo "   Starting Quarkus in dev mode on port 8083..."
        QUARKUS_HTTP_PORT=8083 mvn quarkus:dev -Dquarkus.http.port=8083 > logs/quarkus-service.log 2>&1 &
        QUARKUS_PID=$!
        echo $QUARKUS_PID > "$PROJECT_ROOT/.quarkus.pid"
        echo "✅ Quarkus Service started with PID: $QUARKUS_PID (port 8083)"
    fi
fi

cd "$PROJECT_ROOT"

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All services started successfully!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📊 Service URLs:"
echo "   - Frontend UI:              http://localhost:3000"
echo "   - Router API:               http://localhost:8080"
echo "   - Manager API:              http://localhost:8081"
echo "   - Manager Swagger:          http://localhost:8081/swagger/"
echo "   - Java E-Commerce Service:  http://localhost:8082"
echo "   - E-Commerce Swagger:       http://localhost:8082/swagger-ui.html"
echo "   - Quarkus Service:          http://localhost:8083"
echo ""
echo "📝 Logs:"
echo "   - Router:                   tail -f logs/router.log"
echo "   - Manager:                  tail -f logs/manager.log"
echo "   - E-Commerce Service:       tail -f examples/java-ecommerce-service/logs/ecommerce-service.log"
echo "   - Quarkus Service:          tail -f examples/quarkus-service/logs/quarkus-service.log"
echo ""
echo "🛑 To stop all services: $SCRIPT_DIR/stop-everything.sh"
echo ""

