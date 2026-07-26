#!/bin/bash

set -euo pipefail

LOG_FILE="/apps/taga-ui/deploy-tst.log"

exec > >(tee -a "$LOG_FILE") 2>&1

echo "=========================================="
echo "🚀 Starting TST UI Deployment: $(date)"
echo "=========================================="

# Ensure running as root
if [[ "$EUID" -ne 0 ]]; then
  echo "❌ Must run as root"
  exit 1
fi

SOURCE="/apps/taga-ui/dev"
INTERMEDIATE="/apps/taga-ui/tst"
TARGET="/var/www/tst/tst.nammataga.com"

OLD_API="devapi.nammataga.com"
NEW_API="tstapi.nammataga.com"

echo "📂 Source: $SOURCE"
echo "📂 Intermediate: $INTERMEDIATE"
echo "🎯 Target: $TARGET"

# Validate source
if [[ ! -d "$SOURCE" ]]; then
  echo "❌ Source directory missing: $SOURCE"
  exit 1
fi

# Step 1: Prepare intermediate folder
echo "🧹 Cleaning intermediate directory..."
rm -rf "$INTERMEDIATE"
mkdir -p "$INTERMEDIATE"

echo "📋 Copying files to intermediate..."
cp -r "$SOURCE"/* "$INTERMEDIATE"

# Step 2: Replace API URLs
echo "🔄 Replacing API URLs..."
grep -rlZ "$OLD_API" "$INTERMEDIATE" | xargs -0 sed -i "s/$OLD_API/$NEW_API/g"

echo "✅ API replacement complete"

# Step 3: Deploy to nginx target
echo "🧹 Cleaning target directory..."
rm -rf "$TARGET"/*
mkdir -p "$TARGET"

echo "📋 Copying files to target..."
cp -r "$INTERMEDIATE"/* "$TARGET"

# Step 4: Permissions
echo "🔧 Setting permissions..."
chown -R www-data:www-data "$TARGET"
chmod -R 755 "$TARGET"
chown -R www-data:www-data "$TARGET"
chmod -R 755 "$TARGET"

# Step 5: Nginx test
echo "🧪 Testing nginx config..."
if nginx -t; then
  echo "✅ Nginx config OK"
else
  echo "❌ Nginx config failed"
  echo "👉 Fix nginx config before reload"
  exit 1
fi

# Step 6: Reload nginx
echo "🔄 Reloading nginx..."
if systemctl reload nginx; then
  echo "✅ Nginx reloaded"
else
  echo "❌ Failed to reload nginx"
  exit 1
fi

# Step 7: Verification
echo "🔍 Verifying deployment..."

if grep -r "$NEW_API" "$TARGET" >/dev/null; then
  echo "✅ API correctly updated"
else
  echo "❌ API replacement failed"
  echo "👉 Check sed replacement step"
  exit 1
fi

if [[ -f "$TARGET/index.html" ]]; then
  echo "✅ UI files deployed"
else
  echo "❌ Missing index.html"
  echo "👉 Build may be broken"
  exit 1
fi

echo "=========================================="
echo "✅ Deployment SUCCESS for TST"
echo "🌐 URL: https://tst.nammataga.com"
echo "📄 Logs: $LOG_FILE"
echo "=========================================="
