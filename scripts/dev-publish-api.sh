#!/bin/env bash
set -euo pipefail

REMOTE_HOST="taga-prod"
REMOTE_USER="dev-taga"
REMOTE_PATH="/apps/taga-api/dev"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
API_DIR="${PROJECT_ROOT}/taga-api"
BUILD_SCRIPT="${API_DIR}/build.sh"

# change this if your binary name is different
BINARY_NAME="taga-api"

echo "🚀 Running build script from ${API_DIR}..."

chmod +x "$BUILD_SCRIPT"
cd "$API_DIR"
"$BUILD_SCRIPT"

echo "📦 Uploading binary to server..."

rsync -az \
  -e "ssh" \
  "${API_DIR}/${BINARY_NAME}" \
  "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}/"

echo "✅ Binary deployment completed"
