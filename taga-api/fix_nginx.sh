#!/usr/bin/env bash
set -e

echo "=== 1. Disabling default Nginx site config ==="
if [ -f /etc/nginx/sites-enabled/default ]; then
    sudo rm -f /etc/nginx/sites-enabled/default
    echo "Removed /etc/nginx/sites-enabled/default"
else
    echo "Default site already disabled"
fi

echo ""
echo "=== 2. Stopping existing Nginx processes ==="
NGINX_PIDS=$(pgrep nginx || true)
if [ -n "$NGINX_PIDS" ]; then
    echo "Found running Nginx PIDs: $NGINX_PIDS"
    sudo kill -9 $NGINX_PIDS 2>/dev/null || true
    sleep 1
    echo "Killed active Nginx processes"
else
    echo "No running Nginx process found"
fi

echo ""
echo "=== 3. Testing Nginx Configuration ==="
sudo nginx -t

echo ""
echo "=== 4. Starting Nginx Service ==="
sudo systemctl restart nginx || sudo service nginx restart || sudo nginx

echo ""
echo "=== 5. Checking Listening Ports (80 & 443) ==="
sudo ss -tuln | grep -E ':80|:443' || sudo netstat -tuln | grep -E ':80|:443' || echo "No active listener on 80/443 found"

echo ""
echo "=== 6. Testing local requests to nammataga.com ==="
echo "--- Testing HTTP (Port 80) ---"
curl -sIL http://nammataga.com || echo "HTTP request failed"

echo ""
echo "--- Testing HTTPS (Port 443) ---"
curl -sIL -k https://nammataga.com || echo "HTTPS request failed"

echo ""
echo "=== Complete! ==="
