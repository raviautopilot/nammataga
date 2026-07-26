#!/bin/bash
#
# run.sh - Operational script for taga-api service management
#
# Usage: ./run.sh {start|stop|restart|status|kill|logs|tail}
#

set -euo pipefail

# Configuration
APP_NAME="taga-api"
APP_BINARY="./taga-api"
PID_FILE="./taga-api.pid"
LOG_FILE="./dev-app.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if process is running
is_running() {
    if [[ -f "$PID_FILE" ]]; then
        local pid=$(cat "$PID_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            return 0
        else
            # PID file exists but process is dead
            rm -f "$PID_FILE"
            return 1
        fi
    fi
    return 1
}

# Get PID if running
get_pid() {
    if [[ -f "$PID_FILE" ]]; then
        cat "$PID_FILE"
    else
        echo ""
    fi
}

# Start the application
start_app() {
    if is_running; then
        log_warning "$APP_NAME is already running (PID: $(get_pid))"
        return 1
    fi

    if [[ ! -x "$APP_BINARY" ]]; then
        log_error "Binary $APP_BINARY not found or not executable"
        return 1
    fi

    log_info "Starting $APP_NAME..."
    nohup "$APP_BINARY" > "$LOG_FILE" 2>&1 &
    local pid=$!
    echo "$pid" > "$PID_FILE"
    
    # Wait a moment and check if it's still running
    sleep 2
    if is_running; then
        log_success "$APP_NAME started successfully (PID: $pid)"
        log_info "Logs: $LOG_FILE"
    else
        log_error "$APP_NAME failed to start. Check logs: $LOG_FILE"
        return 1
    fi
}

# Stop the application gracefully
stop_app() {
    if ! is_running; then
        log_warning "$APP_NAME is not running"
        return 1
    fi

    local pid=$(get_pid)
    log_info "Stopping $APP_NAME (PID: $pid)..."
    
    kill "$pid"
    
    # Wait for graceful shutdown (max 10 seconds)
    local count=0
    while is_running && [[ $count -lt 10 ]]; do
        sleep 1
        ((count++))
    done
    
    if is_running; then
        log_warning "$APP_NAME didn't stop gracefully, force killing..."
        kill -9 "$pid"
        sleep 1
    fi
    
    rm -f "$PID_FILE"
    log_success "$APP_NAME stopped"
}

# Force kill the application
kill_app() {
    if ! is_running; then
        log_warning "$APP_NAME is not running"
        return 1
    fi

    local pid=$(get_pid)
    log_info "Force killing $APP_NAME (PID: $pid)..."
    kill -9 "$pid"
    rm -f "$PID_FILE"
    log_success "$APP_NAME killed"
}

# Restart the application
restart_app() {
    log_info "Restarting $APP_NAME..."
    stop_app || true
    sleep 2
    start_app
}

# Show application status
show_status() {
    echo "========================================="
    echo "🔍 $APP_NAME Status"
    echo "========================================="
    
    if is_running; then
        local pid=$(get_pid)
        echo -e "Status: ${GREEN}RUNNING${NC}"
        echo "PID: $pid"
        echo "Binary: $APP_BINARY"
        echo "Log file: $LOG_FILE"
        echo "PID file: $PID_FILE"
        
        # Show process info
        echo ""
        echo "Process Info:"
        ps -p "$pid" -o pid,ppid,cmd,etime,pcpu,pmem --no-headers || true
        
        # Show log file size
        if [[ -f "$LOG_FILE" ]]; then
            local log_size=$(du -h "$LOG_FILE" | cut -f1)
            echo "Log size: $log_size"
        fi
    else
        echo -e "Status: ${RED}STOPPED${NC}"
        echo "Binary: $APP_BINARY"
        echo "Log file: $LOG_FILE"
    fi
    echo "========================================="
}

# Show logs
show_logs() {
    if [[ ! -f "$LOG_FILE" ]]; then
        log_error "Log file $LOG_FILE not found"
        return 1
    fi
    
    log_info "Showing logs from $LOG_FILE"
    echo "========================================="
    cat "$LOG_FILE"
}

# Tail logs
tail_logs() {
    if [[ ! -f "$LOG_FILE" ]]; then
        log_error "Log file $LOG_FILE not found"
        return 1
    fi
    
    log_info "Tailing logs from $LOG_FILE (Ctrl+C to exit)"
    echo "========================================="
    tail -f "$LOG_FILE"
}

# Main script logic
case "${1:-}" in
    start)
        start_app
        ;;
    stop)
        stop_app
        ;;
    restart)
        restart_app
        ;;
    status)
        show_status
        ;;
    kill)
        kill_app
        ;;
    logs)
        show_logs
        ;;
    tail)
        tail_logs
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|kill|logs|tail}"
        echo ""
        echo "Commands:"
        echo "  start   - Start the application"
        echo "  stop    - Stop the application gracefully"
        echo "  restart - Restart the application"
        echo "  status  - Show application status"
        echo "  kill    - Force kill the application"
        echo "  logs    - Show all logs"
        echo "  tail    - Tail logs in real-time"
        exit 1
        ;;
esac

