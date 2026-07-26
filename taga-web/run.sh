#!/bin/bash

# Default to local if no argument provided
ENV=${1:-local}

ENV_FILE="infra/env.$ENV"
PID_FILE="taga.$ENV.pid"
LOG_FILE="taga.$ENV.log"

# Check if environment file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "Environment file $ENV_FILE not found!"
    exit 1
fi

# Read values safely (ignore comments/spaces)
API_BASE_URL=$(grep '^VITE_API_BASE_URL=' "$ENV_FILE" | cut -d '=' -f2-)
PORT=$(grep '^VITE_DEV_SERVER_PORT=' "$ENV_FILE" | cut -d '=' -f2-)

# Default port fallback
PORT=${PORT:-1701}

# Export env vars
export PORT
export VITE_API_BASE_URL="$API_BASE_URL"

echo "Starting $ENV server on port $PORT..."
echo "Using API_BASE_URL: $API_BASE_URL"

# Kill existing process if PID file exists
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if kill -0 "$OLD_PID" 2>/dev/null; then
        echo "Stopping existing process (PID $OLD_PID)"
        kill "$OLD_PID"
    fi
fi

# Start server
nohup npm run local > "$LOG_FILE" 2>&1 &

# Save PID
echo $! > "$PID_FILE"

echo "Server started with PID $(cat "$PID_FILE")"
echo "Logs: $LOG_FILE"
echo "URL: http://localhost:$PORT"