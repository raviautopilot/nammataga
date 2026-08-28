# ==============================================================================
# Nginx Configuration for Nammataga Production Frontend Application
# Domain: nammataga.com
# Root Path: /var/www/prd/nammataga.com
# ==============================================================================

# ==============================================================================
# DEFAULT SERVER - Catch all unknown domains
# Prevents other domains from being served by this configuration
# ==============================================================================
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    listen 443 ssl default_server;
    listen [::]:443 ssl default_server;

    server_name _;

    # Use self-signed certificate for default server to avoid SSL errors
    ssl_certificate /etc/nginx/ssl/default.crt;
    ssl_certificate_key /etc/nginx/ssl/default.key;

    # Log unknown domain attempts
    access_log /var/log/nginx/default.access.log combined;
    error_log /var/log/nginx/default.error.log warn;

    # Return 444 to close connection without response
    return 444;
}

# ==============================================================================
# HTTP to HTTPS Redirect (Forces www.nammataga.com)
# ==============================================================================
server {
    listen 80;
    listen [::]:80;
    server_name nammataga.com www.nammataga.com;

    # Security headers for HTTP (required before redirect)
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Redirect all HTTP traffic to HTTPS www.nammataga.com
    return 301 https://www.nammataga.com$request_uri;
}

# ==============================================================================
# HTTPS Redirect: non-www to www
# ==============================================================================
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name nammataga.com;

    # SSL Configuration (needed to encrypt before redirecting)
    ssl_certificate /etc/letsencrypt/live/nammataga.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/nammataga.com/privkey.pem;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # Redirect non-www HTTPS to www HTTPS
    return 301 https://www.nammataga.com$request_uri;
}

# ==============================================================================
# HTTPS Server (Production - serving www.nammataga.com only)
# ==============================================================================
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name www.nammataga.com;

    # ==========================================================================
    # SSL Configuration
    # ==========================================================================
    ssl_certificate /etc/letsencrypt/live/nammataga.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/nammataga.com/privkey.pem;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # SSL Protocols & Ciphers
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers on;

    # SSL Session Settings
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1440m;
    ssl_session_tickets off;

    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 1.1.1.1 8.8.8.8 valid=300s;
    resolver_timeout 5s;

    # ==========================================================================
    # Security Headers
    # ==========================================================================
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline' 'unsafe-eval'; frame-ancestors 'self';" always;

    # Prevent Cloudflare from caching 404s and other errors
    add_header Cache-Control "no-cache, no-store, must-revalidate" always;
    add_header Pragma "no-cache" always;
    add_header Expires "0" always;

    # ==========================================================================
    # Logging
    # ==========================================================================
    access_log /var/log/nginx/nammataga.com.access.log combined;
    error_log /var/log/nginx/nammataga.com.error.log warn;

    # Hide Nginx version
    server_tokens off;

    # ==========================================================================
    # Root Directory & Index (Commented for Docker Proxy mode)
    # ==========================================================================
    # root /var/www/prd/nammataga.com;
    # index index.html;

    # Prevent directory listing
    autoindex off;

    # ==========================================================================
    # Health Check Endpoint
    # ==========================================================================
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }

    # ==========================================================================
    # Maintenance Mode
    # ==========================================================================
    set $maintenance 0;
    if (-f /var/www/maintenance/maintenance.enable) {
        set $maintenance 1;
    }

    # ==========================================================================
    # Main Location Block - SPA Support via Docker Container Proxy
    # ==========================================================================
    location / {
        if ($maintenance) {
            return 503;
        }

        # Proxy requests to the frontend Docker container
        proxy_pass http://127.0.0.1:8081;
        proxy_http_version 1.1;

        # WebSocket support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Proxy headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
    }

    # ==========================================================================
    # Maintenance Page
    # ==========================================================================
    error_page 503 /maintenance.html;
    location = /maintenance.html {
        root /var/www/maintenance;
        internal;
    }

    # ==========================================================================
    # Deny access to hidden files
    # ==========================================================================
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }

    # ==========================================================================
    # Deny access to sensitive files
    # ==========================================================================
    location ~* (?:\.(?:bak|config|sql|fla|psd|ini|log|sh|inc|swp|dist)|~)$ {
        deny all;
        access_log off;
        log_not_found off;
    }