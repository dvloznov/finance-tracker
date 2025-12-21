#!/bin/bash

# Start API server in background
echo "Starting API server on port 8080..."
go run cmd/api/main.go -port 8080 -bucket finance-tracker-dev &
API_PID=$!

# Start Next.js frontend
echo "Starting frontend on port 3000..."
cd web
npm run dev &
FRONTEND_PID=$!

# Function to cleanup on exit
cleanup() {
    echo "\nStopping services..."
    kill $API_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    exit
}

# Trap Ctrl+C
trap cleanup INT

echo "\n✓ Services started!"
echo "  API: http://localhost:8080"
echo "  Frontend: http://localhost:3000"
echo "\nPress Ctrl+C to stop both services"

# Wait for both processes
wait
