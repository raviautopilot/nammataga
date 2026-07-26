#!/bin/bash

set -euo pipefail

APP_NAME="tst-taga"
APP_DIR="/var/www/tst/tstapi.nammataga.com"
APP_BINARY="$APP_DIR/tst-api.bin"
PID_FILE="$APP_DIR/logs/taga-api.pid"
LOG_FILE="$APP_DIR/logs/tst-taga.log"
RUN_USER="www-data"

log() { echo -e "$1"; }

run_as_user() {
    if [[ "$(id -u)" -eq 0 ]]; then
        sudo -u "$RUN_USER" bash -c "$1"
    else
        bash -c "$1"
    fi
}

is_running() {
    [[ -f "$PID_FILE" ]] && ps -p "$(cat "$PID_FILE")" > /dev/null 2>&1
}

start_app() {
    if is_running; then
        log "Already running (PID: $(cat $PID_FILE))"
        return
    fi

    if [[ ! -f "$APP_BINARY" ]]; then
        log "Binary not found: $APP_BINARY"
        exit 1
    fi

    log "Starting $APP_NAME..."

    run_as_user "
        cd $APP_DIR
        nohup ./tst-api.bin > $LOG_FILE 2>&1 &
        echo \$! > $PID_FILE
    "

    sleep 2

    if is_running; then
        log "Started (PID: $(cat $PID_FILE))"
    else
        log "Failed to start. Check logs: $LOG_FILE"
        exit 1
    fi
}

stop_app() {
    if ! is_running; then
        log "Not running"
        return
    fi

    pid=$(cat "$PID_FILE")
    log "Stopping $APP_NAME (PID: $pid)"
    kill "$pid" || true
    sleep 2

    if ps -p "$pid" > /dev/null 2>&1; then
        kill -9 "$pid"
    fi

    rm -f "$PID_FILE"
    log "Stopped"
}

status_app() {
    if is_running; then
        log "RUNNING PID: $(cat $PID_FILE)"
    else
        log "STOPPED"
    fi
}

case "${1:-}" in
    start) start_app ;;
    stop) stop_app ;;
    restart) stop_app; sleep 1; start_app ;;
    status) status_app ;;
    *) echo "Usage: $0 {start|stop|restart|status}" ;;
esac
