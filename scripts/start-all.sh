#!/bin/bash

set -e

# Get the project root directory (parent of scripts/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "🚀 Starting Sharding System (Backend + Frontend)"
echo ""

# Create logs directory
mkdir -p logs

# Start backend in background
echo "📦 Starting backend services..."
"$SCRIPT_DIR/start-backend.sh"

# Wait a bit for backend to be ready
sleep 3

# Start frontend (this will block)
echo ""
echo "🎨 Starting frontend UI..."
echo ""

# Run frontend in a new terminal if possible, otherwise run in background
if command -v osascript &> /dev/null; then
    # macOS - open new terminal
    osascript -e "tell application \"Terminal\" to do script \"cd '$PROJECT_ROOT' && '$SCRIPT_DIR/start-frontend.sh'\""
    echo "✅ Frontend opened in new terminal window"
    echo ""
    echo "📊 Access the application:"
    echo "   - Frontend UI: http://localhost:3000"
    echo "   - Router API:   http://localhost:8080"
    echo "   - Manager API: http://localhost:8081"
else
    # Linux/Other - run in background
    "$SCRIPT_DIR/start-frontend.sh" &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > .frontend.pid
    echo "✅ Frontend started with PID: $FRONTEND_PID"
    echo ""
    echo "📊 Access the application:"
    echo "   - Frontend UI: http://localhost:3000"
    echo "   - Router API:   http://localhost:8080"
    echo "   - Manager API: http://localhost:8081"
    echo ""
    echo "🛑 To stop all services: $SCRIPT_DIR/stop-all.sh"
fi

