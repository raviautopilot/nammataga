#!/bin/bash

set -euo pipefail

LOG_FILE="/apps/taga-api/tst/tst-publish.log"
mkdir -p /apps/taga-api/tst
exec > >(tee -a "$LOG_FILE") 2>&1

echo "===== TST DEPLOYMENT STARTED ====="

# Ensure root
if [ "$EUID" -ne 0 ]; then
  echo "❌ Script must be run as root"
  exit 1
fi

SRC_DIR="/apps/taga-api/dev"
DEST_DIR="/apps/taga-api/tst"
FINAL_DIR="/var/www/tst/tstapi.nammataga.com"
LOG_DIR="/var/www/tst/tstapi.nammataga.com/logs"
DIST_DIR="/apps/taga-ui/tst"

BIN_SRC="$SRC_DIR/taga-api"
BIN_DEST="$DEST_DIR/tst-api.bin"

CONFIG_SRC="$SRC_DIR/config.json"
CONFIG_DEST="$DEST_DIR/config.json"

echo "👉 Creating destination directory"
mkdir -p "$DEST_DIR" "$LOG_DIR"

echo "👉 Copying binary and config"
cp "$BIN_SRC" "$BIN_DEST"
cp "$CONFIG_SRC" "$CONFIG_DEST"

echo "👉 Setting binary executable"
chmod +x "$BIN_DEST"

echo "👉 Updating config.json values"
sed -i 's/"port":[ ]*[0-9]*/"port": 1802/' "$CONFIG_DEST"
sed -i 's/"environment":[ ]*".*"/"environment": "test"/' "$CONFIG_DEST"
sed -i 's#"reset_password_url":[ ]*".*"#"reset_password_url": "https://tst.nammataga.com"#' "$CONFIG_DEST"
sed -i 's/"jwt_secret":[ ]*".*"/"jwt_secret": "ilovetesttaga"/' "$CONFIG_DEST"

echo "👉 Creating final directory"
mkdir -p "$FINAL_DIR"

echo "👉 Copying files to final location"
cp "$BIN_DEST" "$FINAL_DIR/tst-api.bin"
cp "$CONFIG_DEST" "$FINAL_DIR/config.json"

echo "👉 Ensuring data directory exists"
mkdir -p "$FINAL_DIR/data"

echo "👉 Setting proper permissions"

sudo chown -R www-data:www-data "$FINAL_DIR"

# Directory access
sudo chmod 755 "$FINAL_DIR"

# Binary executable
sudo chmod 755 "$FINAL_DIR/tst-api.bin"

# Config secure
sudo chmod 640 "$FINAL_DIR/config.json"

# Data directory (app only)
chmod 750 "$FINAL_DIR/data"

# Log directory 
sudo chmod 775 /var/www/tst/tstapi.nammataga.com/logs

echo "👉 Testing binary as www-data"
if sudo -u www-data test -x "$FINAL_DIR/tst-api.bin"; then
  echo "✅ Binary executable by www-data"
else
  echo "❌ Binary execution issue"
  exit 1
fi

echo "👉 Testing Nginx configuration"
if nginx -t; then
  echo "✅ Nginx config OK"
  systemctl reload nginx
else
  echo "❌ Nginx config failed"
  exit 1
fi

echo "👉 Cleaning UI folder"
if [ -d "$DIST_DIR" ]; then
  find "$DIST_DIR" -type f ! -name "tst-ops.sh" ! -name "tstapi-publish.sh" -delete
else
  echo "⚠️ DIST_DIR not found, skipping cleanup"
fi

echo "👉 Deployment verification"
if [ -f "$FINAL_DIR/tst-api.bin" ] && [ -f "$FINAL_DIR/config.json" ]; then
  echo "✅ Files deployed successfully"
else
  echo "❌ Deployment failed: missing files"
  exit 1
fi

echo "===== DEPLOYMENT COMPLETED SUCCESSFULLY ====="
echo "Logs: $LOG_FILE"
