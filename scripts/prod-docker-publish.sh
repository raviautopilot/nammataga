#!/bin/bash
# ==============================================================================
# Script Name: prod-docker-publish.sh
# Description: Smart production build & deployment script.
#              1. Detects git changes in taga-api and taga-web.
#              2. Builds, exports, and uploads ONLY the changed service(s).
#              3. Use '--force' flag to force building and shipping both services.
# ==============================================================================

set -euo pipefail

# VPS details
REMOTE_HOST="31.97.62.187"
REMOTE_USER="dev-taga"
REMOTE_PATH="/apps/taga-api/prd"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DIST_DIR="$PROJECT_ROOT/dist"
WEB_TAR="$DIST_DIR/taga-web-prd.tar.gz"
API_TAR="$DIST_DIR/taga-api-prd.tar.gz"

FORCE_BUILD=false
if [[ "${1:-}" == "--force" ]]; then
    FORCE_BUILD=true
fi

echo "=================================================="
echo "🚀 Starting Production Docker Build & Publish"
echo "=================================================="

# Check if docker is installed and daemon is responsive
if ! command -v docker &> /dev/null; then
    echo "❌ Error: 'docker' CLI is not installed or not in PATH."
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "❌ Error: Docker daemon is not running or current user lacks permissions."
    exit 1
fi

# Ensure output dist directory exists
mkdir -p "$DIST_DIR"

# Determine which directories changed in git (compared to HEAD~1 or uncommitted changes)
BUILD_API=false
BUILD_WEB=false

if [ "$FORCE_BUILD" = true ]; then
    echo "⚠️ '--force' flag detected. Building both backend and frontend..."
    BUILD_API=true
    BUILD_WEB=true
else
    # Check git status for uncommitted changes or difference with HEAD~1
    COMPARE_REF="HEAD~1"
    if ! git rev-parse --verify "$COMPARE_REF" &>/dev/null; then
        COMPARE_REF="HEAD"
    fi

    CHANGES=$(git diff --name-only $COMPARE_REF HEAD 2>/dev/null; git status --porcelain 2>/dev/null)

    if echo "$CHANGES" | grep -q "^taga-api/"; then
        BUILD_API=true
    fi

    if echo "$CHANGES" | grep -q "^taga-web/"; then
        BUILD_WEB=true
    fi

    # If neither directory registered changes (e.g. first run or initial commit), build both as safety default
    if [ "$BUILD_API" = false ] && [ "$BUILD_WEB" = false ]; then
        echo "ℹ️ No recent single-folder changes detected in git. Defaulting to building both..."
        BUILD_API=true
        BUILD_WEB=true
    fi
fi

FILES_TO_UPLOAD=()

# 1. Build & Export Backend (taga-api) Image if changed
if [ "$BUILD_API" = true ]; then
    if [ -d "$PROJECT_ROOT/taga-api" ]; then
        echo "⚙️ Building backend Docker image (taga-api:latest) without cache..."
        docker build --no-cache --progress=plain -t taga-api:latest "$PROJECT_ROOT/taga-api"
        echo "📦 Exporting backend image to $API_TAR..."
        docker save taga-api:latest | gzip > "$API_TAR"
        FILES_TO_UPLOAD+=("$API_TAR")
    else
        echo "❌ Error: $PROJECT_ROOT/taga-api directory not found."
        exit 1
    fi
else
    echo "⏭️ Skipping backend (taga-api) - no changes detected."
fi

# 2. Build & Export Frontend (taga-web) Image if changed
if [ "$BUILD_WEB" = true ]; then
    if [ -d "$PROJECT_ROOT/taga-web" ]; then
        echo "⚙️ Building frontend Docker image (taga-web:latest) without cache..."
        docker build --no-cache --progress=plain -t taga-web:latest "$PROJECT_ROOT/taga-web"
        echo "📦 Exporting frontend image to $WEB_TAR..."
        docker save taga-web:latest | gzip > "$WEB_TAR"
        FILES_TO_UPLOAD+=("$WEB_TAR")
    else
        echo "❌ Error: $PROJECT_ROOT/taga-web directory not found."
        exit 1
    fi
else
    echo "⏭️ Skipping frontend (taga-web) - no changes detected."
fi

if [ ${#FILES_TO_UPLOAD[@]} -eq 0 ]; then
    echo "ℹ️ No images were built or queued for upload."
    exit 0
fi

# Gather git and user deployment log details
DEPLOY_TIME=$(date "+%Y-%m-%d %H:%M:%S")
DEPLOY_USER=$(git config user.name || whoami)
BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "non-git")
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "no-commit")

LOG_ENTRY="[$DEPLOY_TIME] User: $DEPLOY_USER | Branch: $BRANCH_NAME | Commit: $COMMIT_HASH | Status: SHIPPED | Built: (API:$BUILD_API WEB:$BUILD_WEB)"

# 3. Upload archives to VPS
echo "🚚 Shipping production image archive(s) to VPS ($REMOTE_HOST)..."
ssh "$REMOTE_USER@$REMOTE_HOST" "mkdir -p $REMOTE_PATH/dist"
scp "${FILES_TO_UPLOAD[@]}" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PATH/dist/"

# Log execution on VPS
ssh "$REMOTE_USER@$REMOTE_HOST" "echo \"$LOG_ENTRY\" >> $REMOTE_PATH/deploy.log"

echo "=================================================="
echo "✅ Production docker image archive(s) shipped to VPS ($REMOTE_HOST)!"
echo "👉 SSH into your VPS and run: sudo bash $REMOTE_PATH/prd-deploy-docker.sh"
echo "=================================================="

# ==============================================================================
# USAGE & COMMANDS REFERENCE GUIDE
# ==============================================================================
#
# 1. Standard Smart Publish (Default):
#    Detects git changes in taga-api and taga-web, then builds & ships ONLY changed services.
#    $ ./scripts/prod-docker-publish.sh
#
# 2. Force Rebuild & Publish All:
#    Ignores git diff and forces a full rebuild and upload of both frontend & backend images.
#    $ ./scripts/prod-docker-publish.sh --force
#
# 3. VPS Deployment Execution (On Remote Server):
#    After running this script, SSH to the VPS and run the deployment script to apply changes:
#    $ ssh dev-taga@31.97.62.187
#    $ sudo bash /apps/taga-api/prd/prd-deploy-docker.sh
#
# ==============================================================================
