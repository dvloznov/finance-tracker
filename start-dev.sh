#!/bin/bash

# Default environment configuration (override by exporting before running)
export GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT:-studious-union-470122-v7}
export GOOGLE_CLOUD_LOCATION=${GOOGLE_CLOUD_LOCATION:-global}
export GOOGLE_GENAI_USE_VERTEXAI=${GOOGLE_GENAI_USE_VERTEXAI:-True}
export PUBSUB_PROJECT=${PUBSUB_PROJECT:-${GOOGLE_CLOUD_PROJECT}}
export PUBSUB_TOPIC=${PUBSUB_TOPIC:-projects/${GOOGLE_CLOUD_PROJECT}/topics/document_events}
export PUBSUB_SUBSCRIPTION=${PUBSUB_SUBSCRIPTION:-projects/${GOOGLE_CLOUD_PROJECT}/subscriptions/document_events-sub}

# Build debug binaries (no optimizations, no inlining)
echo "Building debug binaries..."
go build -gcflags="all=-N -l" -o /tmp/finance-api ./cmd/api
go build -gcflags="all=-N -l" -o /tmp/finance-worker ./cmd/worker

# Start API server
echo "Starting API server on port 8080..."
/tmp/finance-api -port 8080 -bucket ${GCS_BUCKET:-personal-tracker-finance-pdfs} &
API_PID=$!

# Start worker service (Pub/Sub subscriber)
if [[ -z "${PUBSUB_PROJECT:-${GOOGLE_CLOUD_PROJECT}}" || -z "${PUBSUB_SUBSCRIPTION}" ]]; then
        echo "Warning: PUBSUB_PROJECT/GOOGLE_CLOUD_PROJECT and PUBSUB_SUBSCRIPTION not set. Worker will fail to start."
fi
# Start worker service
echo "Starting worker service..."
/tmp/finance-worker \
  -pubsub-project "${PUBSUB_PROJECT:-${GOOGLE_CLOUD_PROJECT}}" \
  -pubsub-subscription "${PUBSUB_SUBSCRIPTION}" &
WORKER_PID=$!

# Start Next.js frontend
echo "Starting frontend on port 3000..."
cd web
npm run dev &
FRONTEND_PID=$!

# Function to cleanup on exit
cleanup() {
    echo "\nStopping services..."
    kill $API_PID 2>/dev/null
    kill $WORKER_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    exit
}

# Trap Ctrl+C
trap cleanup INT

echo "\n✓ Services started!"
echo "  API: http://localhost:8080"
echo "  Worker: running (Pub/Sub subscriber)"
echo "  Frontend: http://localhost:3000"
echo "\nPress Ctrl+C to stop both services"

# Wait for both processes
wait
