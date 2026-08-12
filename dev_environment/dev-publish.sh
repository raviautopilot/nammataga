#!/bin/bash
# ==============================================================================
# Script Name: dev-publish.sh
# Description: Smart & efficient dev build & deployment script.
#              1. Detects git changes in taga-api and taga-web.
#              2. Leverages Docker layer caching for fast builds.
#              3. Builds ONLY changed services (or both if --force).
#              4. Ships images, compose files, and deploy scripts to VPS.
# ==============================================================================

set -euo pipefail

# VPS details
SSH_TARGET="sys-taga@taga-prod"
REMOTE_PATH="/apps/taga-api/dev"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DIST_DIR="$PROJECT_ROOT/dist"
WEB_TAR="$DIST_DIR/taga-web-dev.tar.gz"
API_TAR="$DIST_DIR/taga-api-dev.tar.gz"

FORCE_BUILD=false
if [[ "${1:-}" == "--force" ]]; then
    FORCE_BUILD=true
fi

echo "=================================================="
echo "🚀 Starting Development Docker Build & Publish"
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

BUILD_API=false
BUILD_WEB=false

if [ "$FORCE_BUILD" = true ]; then
    echo "⚠️ '--force' flag detected. Building both backend and frontend..."
    BUILD_API=true
    BUILD_WEB=true
else
    # Check git diff against working tree / uncommitted changes / last commit
    CHANGES=$(git status --porcelain 2>/dev/null; git diff --name-only HEAD 2>/dev/null; git diff --name-only HEAD~1 HEAD 2>/dev/null || true)

    if echo "$CHANGES" | grep -q "^taga-api/"; then
        BUILD_API=true
    fi

    if echo "$CHANGES" | grep -q "^taga-web/"; then
        BUILD_WEB=true
    fi

    # If no specific service changes are found in git diff, but archives are missing, build them.
    # Otherwise, if archives exist and no code changed, skip building!
    if [ "$BUILD_API" = false ] && [ "$BUILD_WEB" = false ]; then
        if [ ! -f "$API_TAR" ] || [ ! -f "$WEB_TAR" ]; then
            echo "ℹ️ Missing image archives detected. Building both..."
            BUILD_API=true
            BUILD_WEB=true
        else
            echo "✨ No files changed in taga-api/ or taga-web/. Skipping rebuilding!"
        fi
    fi
fi

# Ensure archives exist if skipped from build
if [ "$BUILD_API" = false ] && [ ! -f "$API_TAR" ]; then
    echo "⚠️ $API_TAR missing! Forcing backend build..."
    BUILD_API=true
fi

if [ "$BUILD_WEB" = false ] && [ ! -f "$WEB_TAR" ]; then
    echo "⚠️ $WEB_TAR missing! Forcing frontend build..."
    BUILD_WEB=true
fi

# 1. Build & Export Backend (taga-api) Image if changed
if [ "$BUILD_API" = true ]; then
    if [ -d "$PROJECT_ROOT/taga-api" ]; then
        echo "⚙️ Building backend Docker image (taga-api:dev & taga-api:latest)..."
        docker build -t taga-api:dev -t taga-api:latest "$PROJECT_ROOT/taga-api"
        echo "📦 Exporting backend image to $API_TAR..."
        docker save taga-api:dev taga-api:latest | gzip > "$API_TAR"
    else
        echo "❌ Error: $PROJECT_ROOT/taga-api directory not found."
        exit 1
    fi
else
    echo "⏭️ Skipping backend (taga-api) build - no changes detected. Using existing $API_TAR."
fi

# 2. Build & Export Frontend (taga-web) Image if changed
if [ "$BUILD_WEB" = true ]; then
    if [ -d "$PROJECT_ROOT/taga-web" ]; then
        echo "⚙️ Building frontend Docker image (taga-web:dev & taga-web:latest)..."
        docker build -t taga-web:dev -t taga-web:latest "$PROJECT_ROOT/taga-web"
        echo "📦 Exporting frontend image to $WEB_TAR..."
        docker save taga-web:dev taga-web:latest | gzip > "$WEB_TAR"
    else
        echo "❌ Error: $PROJECT_ROOT/taga-web directory not found."
        exit 1
    fi
else
    echo "⏭️ Skipping frontend (taga-web) build - no changes detected. Using existing $WEB_TAR."
fi

# Gather git and user deployment log details
DEPLOY_TIME=$(date "+%Y-%m-%d %H:%M:%S")
DEPLOY_USER=$(git config user.name || whoami)
BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "non-git")
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "no-commit")

LOG_ENTRY="[$DEPLOY_TIME] User: $DEPLOY_USER | Branch: $BRANCH_NAME | Commit: $COMMIT_HASH | Status: SHIPPED | Built: (API:$BUILD_API WEB:$BUILD_WEB)"

# 3. Upload archives and configs to VPS via a staging directory (since target path requires sudo)
# Using absolute path to home directory to prevent SCP/SFTP variable expansion issues
TMP_UPLOAD="/home/sys-taga/taga-dev-upload"
echo "🚚 Creating staging directory on VPS ($SSH_TARGET)..."
ssh "$SSH_TARGET" "mkdir -p $TMP_UPLOAD/dist $TMP_UPLOAD/nginx"

echo "🚚 Transferring files to VPS staging directory..."
scp "$WEB_TAR" "$API_TAR" "$SSH_TARGET:$TMP_UPLOAD/dist/"
scp "$SCRIPT_DIR/docker-compose.dev.yml" "$SSH_TARGET:$TMP_UPLOAD/"
scp "$SCRIPT_DIR/dev-deploy-docker.sh" "$SCRIPT_DIR/dev-wipe-docker.sh" "$SSH_TARGET:$TMP_UPLOAD/"
scp "$SCRIPT_DIR/nginx/dev.nammataga.com" "$SCRIPT_DIR/nginx/devapi.nammataga.com" "$SSH_TARGET:$TMP_UPLOAD/nginx/"

echo "🌐 Deploying files to final path with sudo privileges..."
ssh -t "$SSH_TARGET" "
  sudo mkdir -p $REMOTE_PATH/dist $REMOTE_PATH/nginx &&
  sudo cp -r $TMP_UPLOAD/* $REMOTE_PATH/ &&
  sudo chown -R sys-taga:sys-taga $REMOTE_PATH &&
  sudo chmod -R 755 $REMOTE_PATH &&
  sudo bash -c \"echo '$LOG_ENTRY' >> $REMOTE_PATH/deploy.log\" &&
  rm -rf $TMP_UPLOAD
"

echo "=================================================="
echo "✅ Development docker image archives and configs shipped to VPS ($SSH_TARGET)!"
echo "👉 SSH into your VPS and run: sudo bash $REMOTE_PATH/dev-deploy-docker.sh"
echo "=================================================="
