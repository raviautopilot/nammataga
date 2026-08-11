#!/bin/bash
# ==============================================================================
# Script Name: dev-deploy-docker.sh
# Description: Deploys the frontend and backend Docker containers on the dev VPS.
# Location:    dev_environment/dev-deploy-docker.sh
# ==============================================================================

set -euo pipefail

if [ "$EUID" -ne 0 ]; then 
  echo "❌ Error: This script must be run with sudo privileges." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="$SCRIPT_DIR/dist"

WEB_TAR="$DIST_DIR/taga-web-dev.tar.gz"
API_TAR="$DIST_DIR/taga-api-dev.tar.gz"

DATA_DIR="/apps/taga-api/dev/data"
AUDIT_DIR="/apps/taga-api/dev/audit-logs"

echo "=================================================="
echo "🚀 Starting Development Container Deployment"
echo "=================================================="

# 1. Check for image archives
if [ ! -f "$WEB_TAR" ] || [ ! -f "$API_TAR" ]; then
  echo "❌ Error: Docker image archives not found in $DIST_DIR!"
  echo "Ensure taga-web-dev.tar.gz and taga-api-dev.tar.gz are transferred first."
  exit 1
fi

# 2. Ensure directories exist on the host VPS (Preserving existing data)
echo "📁 Checking host storage volume at $DATA_DIR..."
mkdir -p "$DATA_DIR"
chown -R www-data:www-data "$DATA_DIR"
chmod 755 "$DATA_DIR"
echo "✅ Host storage volume verified. Existing files remain intact."

echo "📁 Checking host audit logs volume at $AUDIT_DIR..."
mkdir -p "$AUDIT_DIR"
chown -R www-data:www-data "$AUDIT_DIR"
chmod 755 "$AUDIT_DIR"
echo "✅ Host audit logs volume verified."

# 3. Load Docker images
echo "📥 Loading frontend Docker image..."
docker load -i "$WEB_TAR"

echo "📥 Loading backend Docker image..."
docker load -i "$API_TAR"

# 4. Stop and delete old development containers via Docker Compose
echo "🧹 Stopping and removing old development containers..."
docker compose -f "$SCRIPT_DIR/docker-compose.dev.yml" down 2>/dev/null || true

# 5. Start Services via Docker Compose
echo "🚀 Starting containers via Docker Compose..."
docker compose -f "$SCRIPT_DIR/docker-compose.dev.yml" up -d

# 7. Apply Nginx configurations from the deployment bundle
echo "🌐 Updating host Nginx configurations..."
if [ -d "$SCRIPT_DIR/nginx" ]; then
  cp "$SCRIPT_DIR/nginx/dev.nammataga.com" "$SCRIPT_DIR/nginx/devapi.nammataga.com" /etc/nginx/sites-available/
  ln -sf /etc/nginx/sites-available/dev.nammataga.com /etc/nginx/sites-enabled/
  ln -sf /etc/nginx/sites-available/devapi.nammataga.com /etc/nginx/sites-enabled/
  echo "✅ Nginx configurations copied and enabled."
else
  echo "⚠️ Warning: Nginx configurations not found in deployment bundle. Skipping update."
fi

# 8. Reload Nginx on host to pick up proxy routes
echo "🌐 Testing and reloading host Nginx proxy..."
if nginx -t; then
  systemctl reload nginx
  echo "✅ Nginx proxy reloaded successfully!"
else
  echo "⚠️ Warning: Nginx test failed. Host routing might need inspection."
fi

# 8. Clean up tarballs to free VPS space
echo "🧹 Cleaning up temporary image tar archives..."
rm -f "$WEB_TAR" "$API_TAR"
echo "✅ Archives removed."

echo "=================================================="
echo "🎉 Development Deployment Successful!"
echo "🌐 Web Interface: https://dev.nammataga.com"
echo "🌐 API Endpoint : https://devapi.nammataga.com"
echo "=================================================="
