# ==============================================================================
# Nginx Configuration for Nammataga Development API Reverse Proxy
# Domain: devapi.nammataga.com
# Proxy Destination: 127.0.0.1:1801 (via upstream block)
# ==============================================================================

# ==============================================================================
# Rate Limiting Zones
# ==============================================================================
limit_req_zone $binary_remote_addr zone=dev_api_limit:10m rate=10r/s;
limit_conn_zone $binary_remote_addr zone=dev_api_conn:10m;

# ==============================================================================
# Backend upstream group
# ==============================================================================
upstream dev_backend {
    server 127.0.0.1:1801 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

# ==============================================================================
# HTTP to HTTPS Redirect (API)
# ==============================================================================
server {
    listen 80;
    listen [::]:80;
    server_name devapi.nammataga.com;

    # Security headers for HTTP (required before redirect)
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Redirect all HTTP traffic to HTTPS
    return 301 https://$host$request_uri;
}

# ==============================================================================
# HTTPS Server (API)
# ==============================================================================
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name devapi.nammataga.com;

    # ==========================================================================
    # SSL Configuration
    # ==========================================================================
    ssl_certificate /etc/letsencrypt/live/dev.nammataga.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/dev.nammataga.com/privkey.pem;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # SSL Protocols & Ciphers
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384';

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
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'; frame-ancestors 'self' https://dev.nammataga.com;" always;

    # API-specific headers
    add_header Cache-Control "no-cache, no-store, must-revalidate" always;
    add_header Pragma "no-cache" always;

    # ==========================================================================
    # Logging
    # ==========================================================================
    access_log /var/log/nginx/devapi.nammataga.com.access.log combined;
    error_log /var/log/nginx/devapi.nammataga.com.error.log warn;

    # Hide Nginx version
    server_tokens off;

    # ==========================================================================
    # Health Check Endpoint
    # ==========================================================================
    location /health {
        access_log off;
        proxy_pass http://dev_backend/health;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # ==========================================================================
    # PDF Documents Proxy (/docs)
    # ==========================================================================
    location /docs {
        proxy_pass http://dev_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
    }

    # ==========================================================================
    # Main API Proxy
    # ==========================================================================
    location / {
        # Rate limiting
        limit_req zone=dev_api_limit burst=20 nodelay;
        limit_conn dev_api_conn 100;

        # Proxy to backend API
        proxy_pass http://dev_backend;
        proxy_http_version 1.1;

        # WebSockets support
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Proxy Headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;

        # Timeout Configuration
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # Buffer Configuration
        proxy_buffer_size 128k;
        proxy_buffers 4 256k;
        proxy_busy_buffers_size 256k;

        # Disable buffering for streaming APIs / SSE
        proxy_buffering off;

        # Retry Configuration
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503 http_504;
        proxy_next_upstream_tries 2;
        proxy_next_upstream_timeout 30s;
    }

    # ==========================================================================
    # Gzip Compression
    # ==========================================================================
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml application/json application/javascript application/xml+rss application/atom+xml image/svg+xml;
}
