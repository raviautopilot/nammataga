#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$SCRIPT_DIR/nammataga_http.conf"

echo "=== 1. Updating Nginx Site Configuration ==="
sudo cp "$CONFIG_FILE" /etc/nginx/sites-available/nammataga

echo "=== 2. Ensuring Symlink in sites-enabled ==="
sudo rm -f /etc/nginx/sites-enabled/default
sudo ln -sf /etc/nginx/sites-available/nammataga /etc/nginx/sites-enabled/nammataga

echo "=== 3. Testing Nginx Syntax ==="
sudo nginx -t

echo "=== 4. Reloading Nginx Service ==="
sudo service nginx reload || sudo nginx -s reload

echo ""
echo "=== 5. Testing HTTP Endpoints ==="
echo "Testing http://nammataga.com ..."
curl -sIL http://nammataga.com | head -n 10

echo ""
echo "Testing http://api.nammataga.com ..."
curl -sIL http://api.nammataga.com | head -n 10

echo ""
echo "=== RECONFIGURATION COMPLETE! ==="
