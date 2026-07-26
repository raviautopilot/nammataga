#!/bin/bash

# SSL Certificate Setup for All Domains

set -e  # Exit on error

# List of all non-prod your domains
DOMAINS=(
    "dev.nammataga.com"
    "tst.nammataga.com" 
    "stg.nammataga.com"
    "devapi.nammataga.com"
    "tstapi.nammataga.com"
    "stgapi.nammataga.com"
)

echo "Starting SSL certificate setup for all domains..."

# Check if Certbot is installed
if ! command -v certbot >/dev/null 2>&1; then
    echo "Certbot is not installed. Please install it first."
    exit 1
fi

# Get SSL certificates for all domains using a single certificate
# This is more efficient than individual certificates
echo "Obtaining SSL certificate covering all domains..."

sudo certbot --nginx --agree-tos --no-eff-email --expand \
    -d dev.nammataga.com \
    -d tst.nammataga.com \
    -d stg.nammataga.com \
    -d devapi.nammataga.com \
    -d tstapi.nammataga.com \
    -d stgapi.nammataga.com \

# Alternatively, if you want to be prompted for email (remove --no-eff-email)
# sudo certbot --nginx --agree-tos --expand \
#     -d dev.nammataga.com \
#     -d tst.nammataga.com \
#     -d stg.nammataga.com \
#     -d devapi.nammataga.com \
#     -d tstapi.nammataga.com \
#     -d stgapi.nammataga.com \

# Set up automatic renewal
echo "Setting up automatic certificate renewal..."

# Test the renewal process
sudo certbot renew --dry-run

if [ $? -eq 0 ]; then
    echo "✓ Certificate renewal test successful"
    
    # Check if cron job already exists
    if ! crontab -l | grep -q "certbot renew"; then
        echo "Adding cron job for automatic renewal..."
        # Add a cron job to run renewal twice daily
        (crontab -l 2>/dev/null; echo "0 0,12 * * * /usr/bin/certbot renew --quiet") | crontab -
        echo "✓ Automatic renewal cron job added"
    else
        echo "✓ Automatic renewal cron job already exists"
    fi
else
    echo "✗ Certificate renewal test failed"
    echo "Please check your configuration and try again."
fi

echo ""
echo "SSL setup completed!"
echo "All domains now have SSL certificates:"
for domain in "${DOMAINS[@]}"; do
    echo "✓ https://$domain"
done

echo ""
echo "Certificate information:"
sudo certbot certificates
