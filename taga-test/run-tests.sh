#!/bin/bash

# Ensure script executes in taga-test directory where go.mod resides
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Configuration
PORT=9515

echo "========================================="
echo " Starting Chromedriver & E2E Test Suite "
echo "========================================="

# 1. Start chromedriver in the background
chromedriver --port=$PORT > /dev/null 2>&1 &
CHROMEDRIVER_PID=$!

# 2. Setup automatic cleanup on exit (trap)
cleanup() {
    echo "Cleaning up: stopping chromedriver (PID: $CHROMEDRIVER_PID)..."
    kill $CHROMEDRIVER_PID 2>/dev/null
    wait $CHROMEDRIVER_PID 2>/dev/null
    echo "Cleanup complete."
}
trap cleanup EXIT

# 3. Wait for Chromedriver to become responsive
echo "Waiting for Chromedriver to start on port $PORT..."
for i in {1..10}; do
    if curl -s http://localhost:$PORT/status | grep -q '"ready":true'; then
        echo "Chromedriver is ready!"
        break
    fi
    if [ $i -eq 10 ]; then
        echo "Error: Chromedriver failed to start on port $PORT."
        exit 1
    fi
    sleep 0.5
done

# 3.5. Export a shared timestamp so all test packages run under the same run folder
export E2E_RUN_TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")

# 4. Run the Go E2E tests (respecting config.json headless setting or E2E_HEADLESS env var)
echo "Running E2E test suite..."
go test -v -timeout 30m ./tests/ui/... ./tests/api/... "$@"
TEST_EXIT_CODE=$?

echo "========================================="
echo " Tests finished with exit code $TEST_EXIT_CODE"
HTML_REPORT="$SCRIPT_DIR/evidence/run-$E2E_RUN_TIMESTAMP/reports/report.html"
if [ -f "$HTML_REPORT" ]; then
    echo "📊 HTML Report: $HTML_REPORT"
fi
echo "========================================="

exit $TEST_EXIT_CODE

