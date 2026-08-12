# ==============================================================================
# Nginx Configuration for Nammataga Development Frontend Application
# Domain: dev.nammataga.com
# Proxy Destination: 127.0.0.1:1701
# ==============================================================================

# ==============================================================================
# HTTP to HTTPS Redirect
# ==============================================================================
server {
    listen 80;
    listen [::]:80;
    server_name dev.nammataga.com;

    # Security headers for HTTP (required before redirect)
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Redirect all HTTP traffic to HTTPS
    return 301 https://dev.nammataga.com$request_uri;
}

# ==============================================================================
# HTTPS Server (Development - dev.nammataga.com)
# ==============================================================================
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name dev.nammataga.com;

    # ==========================================================================
    # SSL Configuration
    # ==========================================================================
    ssl_certificate /etc/letsencrypt/live/dev.nammataga.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dev.nammataga.com/privkey.pem;
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
    access_log /var/log/nginx/dev.nammataga.com.access.log combined;
    error_log /var/log/nginx/dev.nammataga.com.error.log warn;

    # Hide Nginx version
    server_tokens off;

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
    if (-f /var/www/maintenance/maintenance-dev.enable) {
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
        proxy_pass http://127.0.0.1:1701;
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
}
