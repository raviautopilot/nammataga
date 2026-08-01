#!/bin/bash

# Ensure script executes in taga-test directory where go.mod resides
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "========================================="
echo " Starting E2E API Test Suite             "
echo "========================================="

# 1. Export a shared timestamp for reports
export E2E_RUN_TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")

# 2. Run API test suite
echo "Running API test suite..."
go test -v ./tests/api/... "$@"
TEST_EXIT_CODE=$?

echo "========================================="
echo " API Tests finished with exit code $TEST_EXIT_CODE"
echo "========================================="

exit $TEST_EXIT_CODE
