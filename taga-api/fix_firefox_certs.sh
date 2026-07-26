#!/usr/bin/env bash
# ==============================================================================
# Script Name: fix_firefox_certs.sh
# Description: Generates matching local certificates, copies them to /etc/ssl,
#              imports the Root CA into all Firefox/Chrome profile databases,
#              and reloads Nginx.
# ==============================================================================

set -euo pipefail

# Ensure script is run with sudo
if [ "$EUID" -ne 0 ]; then
    echo "❌ Error: Please run this script with sudo:"
    echo "   sudo $0"
    exit 1
fi

REAL_USER="${SUDO_USER:-ubuntu}"
REAL_HOME=$(eval echo "~$REAL_USER")
PROJECT_ROOT="/home/ubuntu/code/github/nammataga/taga-api"
CAROOT_DIR="$PROJECT_ROOT/mkcert_ca"
ROOT_CA_PATH="$PROJECT_ROOT/rootCA.crt"
CERT_OUT="$PROJECT_ROOT/nammataga.crt"
KEY_OUT="$PROJECT_ROOT/nammataga.key"

echo "=================================================="
echo "🛡️  Running as user: $REAL_USER (Home: $REAL_HOME)"
echo "=================================================="

# 1. Install certutil (libnss3-tools) if not installed
echo "=== 1. Checking libnss3-tools ==="
if ! command -v certutil &> /dev/null; then
    echo "Installing libnss3-tools..."
    apt-get update -qq
    apt-get install -y -qq libnss3-tools
fi
echo "✅ certutil is available."

# 2. Re-generate site certificates under the correct user CA
echo ""
echo "=== 2. Generating site certificates signed by $REAL_USER CA ==="
# We run mkcert as the real user to use their mkcert_ca
sudo -u "$REAL_USER" CAROOT="$CAROOT_DIR" mkcert \
    -cert-file "$CERT_OUT" \
    -key-file "$KEY_OUT" \
    nammataga.com api.nammataga.com 127.0.0.1 localhost ::1

# Verify the issuer is correct
ISSUER=$(openssl x509 -in "$CERT_OUT" -noout -issuer)
echo "✅ Site certificate issuer: $ISSUER"

# 3. Copy to system SSL directories and set permissions
echo ""
echo "=== 3. Copying certificates to system SSL directory ==="
mkdir -p /etc/ssl/certs /etc/ssl/private
cp "$CERT_OUT" /etc/ssl/certs/nammataga.crt
cp "$KEY_OUT" /etc/ssl/private/nammataga.key
chmod 644 /etc/ssl/certs/nammataga.crt
chmod 600 /etc/ssl/private/nammataga.key
echo "✅ Certificates copied successfully."

# 4. Import the root CA certificate to all Firefox / Chrome profiles
echo ""
echo "=== 4. Importing Root CA into NSS Databases ==="
# Search in real user's home directory for cert9.db
echo "Searching for Firefox/Chrome cert9.db files in $REAL_HOME..."
find "$REAL_HOME" -name "cert9.db" 2>/dev/null | while read -r db_file; do
    db_dir=$(dirname "$db_file")
    echo "Found NSS database: $db_dir"
    
    # Import and set CT,c,c trust for SSL/TLS and Email
    sudo -u "$REAL_USER" certutil -A -n "mkcert development CA" -t "CT,c,c" -i "$ROOT_CA_PATH" -d "sql:$db_dir"
    echo "   ↳ Successfully imported and trusted."
done

# 5. Reload Nginx to pick up the new certificate
echo ""
echo "=== 5. Reloading Nginx ==="
if nginx -t; then
    service nginx reload || nginx -s reload
    echo "✅ Nginx reloaded successfully!"
else
    echo "❌ Error: Nginx configuration test failed. Not reloading Nginx."
    exit 1
fi

echo ""
echo "=================================================="
echo "🎉 SUCCESS: Local HTTPS setup for nammataga.com is complete!"
echo "Please restart Firefox and navigate to https://nammataga.com"
echo "=================================================="
