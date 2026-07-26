#!/usr/bin/env bash
set -e

CERT_PATH="/etc/ssl/certs/nammataga.crt"

if [ ! -f "$CERT_PATH" ]; then
    echo "Error: Certificate $CERT_PATH not found!"
    exit 1
fi

echo "=== Adding $CERT_PATH to Linux System Trust Store ==="

if [ -d "/usr/local/share/ca-certificates" ]; then
    echo "Copying cert to /usr/local/share/ca-certificates/nammataga.crt ..."
    sudo cp "$CERT_PATH" /usr/local/share/ca-certificates/nammataga.crt
    echo "Updating ca-certificates..."
    sudo update-ca-certificates
elif [ -d "/etc/pki/ca-trust/source/anchors" ]; then
    echo "Copying cert to /etc/pki/ca-trust/source/anchors/nammataga.crt ..."
    sudo cp "$CERT_PATH" /etc/pki/ca-trust/source/anchors/nammataga.crt
    echo "Updating ca-trust..."
    sudo update-ca-trust
fi

echo ""
echo "=== Adding certificate to Chrome/Firefox NSS Databases (if libnss3-tools installed) ==="

if command -v certutil &> /dev/null; then
    echo "certutil found. Adding certificate to user NSS databases..."
    
    # Target all cert9.db databases in home directory (Firefox & Chrome profiles)
    find "$HOME" -name "cert9.db" 2>/dev/null | while read -r db_file; do
        db_dir=$(dirname "$db_file")
        echo "Adding cert to NSS database in: $db_dir"
        certutil -A -n "nammataga-local-ca" -t "TCu,cu,cu" -i "$CERT_PATH" -d "sql:$db_dir" || true
    done
else
    echo "Tip: Install libnss3-tools ('sudo apt install libnss3-tools') to automatically insert local certs into Chrome/Firefox browser profile stores."
fi

echo ""
echo "=== How to trust in Browsers manually if warning persists ==="
echo "1. Chrome / Edge:"
echo "   - Open chrome://settings/certificates"
echo "   - Click 'Authorities' tab -> Click 'Import'"
echo "   - Select: $CERT_PATH"
echo "   - Check 'Trust this certificate for identifying websites'"
echo ""
echo "2. Firefox:"
echo "   - Open about:preferences#privacy"
echo "   - Scroll down to 'Certificates' -> Click 'View Certificates'"
echo "   - Click 'Authorities' tab -> Click 'Import'"
echo "   - Select: $CERT_PATH"
echo "   - Check 'Trust this CA to identify websites'"
echo ""
echo "3. Quick browser bypass (Chrome):"
echo "   - Click anywhere on the warning page and type: thisisunsafe"
echo ""
