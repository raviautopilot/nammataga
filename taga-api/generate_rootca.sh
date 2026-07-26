#!/usr/bin/env bash
set -e

echo "=== Creating dedicated Root CA and copying to workspace ==="
export CAROOT="/home/ubuntu/code/github/nammataga/taga-api/mkcert_ca"
mkdir -p "$CAROOT"

# Generate new local CA
/usr/local/bin/mkcert -install

echo "=== Copying rootCA.pem to rootCA.crt for easy download ==="
cp "$CAROOT/rootCA.pem" /home/ubuntu/code/github/nammataga/taga-api/rootCA.crt

echo "=== Regenerating site certificate with this Root CA ==="
sudo /usr/local/bin/mkcert -cert-file /etc/ssl/certs/nammataga.crt \
            -key-file /etc/ssl/private/nammataga.key \
            nammataga.com api.nammataga.com 127.0.0.1 localhost ::1

sudo chmod 644 /etc/ssl/certs/nammataga.crt
sudo chmod 600 /etc/ssl/private/nammataga.key

echo "=== Reloading Nginx ==="
sudo service nginx reload || sudo nginx -s reload

echo "Done!"
