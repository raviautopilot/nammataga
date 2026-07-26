#!/bin/bash

# Default to local if no argument provided
ENV=${1:-local}

ENV_FILE="infra/env.$ENV"

# Check if environment file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "Environment file $ENV_FILE not found!"
    exit 1
fi

# Read the port from the environment file
PORT=$(grep 'VITE_DEV_SERVER_PORT' "$ENV_FILE" | cut -d '=' -f2)

# Default port if not set
if [ -z "$PORT" ]; then
    case $ENV in
        "local") PORT=1701 ;;
        *) PORT=1701 ;;
    esac
fi

echo "Stopping $ENV environment on port $PORT..."

# Find and kill the process using the port
PORT_PID=$(lsof -ti :$PORT)

if [ -z "$PORT_PID" ]; then
    echo "No process found running on port $PORT."
else
    echo "Stopping process (PID $PORT_PID) on port $PORT."
    kill $PORT_PID
fi

# Also attempt to kill the npm process from the .pid file if it exists
PID_FILE="app-taga-ui.$ENV.pid"
if [ -f "$PID_FILE" ]; then
    APP_PID=$(cat "$PID_FILE")
    echo "Stopping app process (PID $APP_PID) for $ENV."
    kill $APP_PID 2>/dev/null || true
    rm "$PID_FILE"
fi

echo "$ENV environment stopped."