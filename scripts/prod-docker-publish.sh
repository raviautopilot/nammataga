#!/bin/bash
# ==============================================================================
# Script Name: prod-docker-publish.sh (Unified Version)
# Description: Triggers production builds, logs metadata, and transfers images.
# WARNING:     ONLY FOR PRODUCTION DEPLOYMENT.
# ==============================================================================

set -euo pipefail

# VPS details (imported from dev-publish.bat)
REMOTE_HOST="31.97.62.187"
REMOTE_USER="dev-taga"
REMOTE_PATH="/apps/taga-api/prd"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Locate configuration repository path
if [ -d "$PROJECT_ROOT/taga-config" ]; then
    CONFIG_DIR="$(cd "$PROJECT_ROOT/taga-config" && pwd)"
else
    echo "❌ Error: taga-config directory not found at $PROJECT_ROOT/taga-config"
    exit 1
fi

BUILD_SCRIPT="$CONFIG_DIR/docker/prd-build-docker.sh"

if [ ! -f "$BUILD_SCRIPT" ]; then
    echo "❌ Error: Central build script not found at $BUILD_SCRIPT"
    exit 1
fi

# Run build script to generate production images
bash "$BUILD_SCRIPT"

# Gather git and user deployment log details
DEPLOY_TIME=$(date "+%Y-%m-%d %H:%M:%S")
DEPLOY_USER=$(git config user.name || whoami)
BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "non-git")
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "no-commit")

LOG_ENTRY="[$DEPLOY_TIME] User: $DEPLOY_USER | Branch: $BRANCH_NAME | Commit: $COMMIT_HASH | Status: SHIPPED"

echo ""
echo "📝 Logging deployment data..."
echo "$LOG_ENTRY" >> "$CONFIG_DIR/docker/deploy.log"

# Upload packages to VPS
echo "🚚 Shipping production images to VPS ($REMOTE_HOST)..."
ssh "$REMOTE_USER@$REMOTE_HOST" "mkdir -p $REMOTE_PATH/dist"
ssh "$REMOTE_USER@$REMOTE_HOST" "mkdir -p $REMOTE_PATH/nginx"

# Transfer images, compose configs, deployment, and wipe scripts to VPS
scp -r "$CONFIG_DIR/docker/dist/"* "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/dist/"
scp "$CONFIG_DIR/docker/prd-deploy-docker.sh" "$CONFIG_DIR/docker/prd-wipe-docker.sh" "$CONFIG_DIR/docker/install-docker.sh" "$CONFIG_DIR/docker/docker-compose.prod.yml" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/"
scp "$CONFIG_DIR/nginx/prd/nammataga.com" "$CONFIG_DIR/nginx/prd/api.nammataga.com" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/nginx/"

# Log execution on VPS
ssh "$REMOTE_USER@$REMOTE_HOST" "echo \"$LOG_ENTRY\" >> $REMOTE_PATH/deploy.log"

echo "=================================================="
echo "✅ Production deployment packages shipped to VPS ($REMOTE_HOST)!"
echo "👉 SSH into your VPS and run: sudo bash $REMOTE_PATH/prd-deploy-docker.sh"
echo "=================================================="
