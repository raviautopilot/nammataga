#!/usr/bin/env bash
set -e

echo "=== 1. Installing certutil (libnss3-tools) & mkcert dependencies ==="
sudo apt-get update -qq
sudo apt-get install -y -qq libnss3-tools wget curl

echo ""
echo "=== 2. Downloading & Installing mkcert (Local CA Generator) ==="
if ! command -v mkcert &> /dev/null; then
    wget -q https://dl.filippo.io/mkcert/latest?for=linux/amd64 -O /tmp/mkcert
    sudo chmod +x /tmp/mkcert
    sudo mv /tmp/mkcert /usr/local/bin/mkcert
fi

echo "mkcert version: $(mkcert -version)"

echo ""
echo "=== 3. Creating & Installing Local Root Certificate Authority (CA) ==="
# Install local CA into system and browser trust stores
mkcert -install

echo ""
echo "=== 4. Generating Trusted Certificates for nammataga.com & api.nammataga.com ==="
sudo mkdir -p /etc/ssl/certs /etc/ssl/private

# Generate certificate signed by our new local CA
sudo mkcert -cert-file /etc/ssl/certs/nammataga.crt \
            -key-file /etc/ssl/private/nammataga.key \
            nammataga.com api.nammataga.com 127.0.0.1 localhost ::1

sudo chmod 644 /etc/ssl/certs/nammataga.crt
sudo chmod 600 /etc/ssl/private/nammataga.key

echo ""
echo "=== 5. Applying Nginx HTTPS Configuration ==="
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
sudo cp "$SCRIPT_DIR/nammataga_https.conf" /etc/nginx/sites-available/nammataga
sudo rm -f /etc/nginx/sites-enabled/default
sudo ln -sf /etc/nginx/sites-available/nammataga /etc/nginx/sites-enabled/nammataga

echo ""
echo "=== 6. Reloading Nginx ==="
sudo nginx -t
sudo service nginx reload || sudo nginx -s reload

echo ""
echo "=== COMPLETE! TRUSTED HTTPS IS READY ==="
echo "You can now open https://nammataga.com without any certificate errors."
