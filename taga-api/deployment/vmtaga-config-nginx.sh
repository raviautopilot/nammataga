#!/usr/bin/env bash
set -euo pipefail

LOGFILE="/tmp/nginx-setup.log"
HELPFILE="/usr/local/share/nginx-proxy-help.txt"
ROLLBACK_ACTIONS=()

log_error()   { echo "[ERROR] ${*:-Unknown}" | tee -a "$LOGFILE"; }
log_info()    { echo "[INFO] $*"; }
rollback()    { for a in "${ROLLBACK_ACTIONS[@]}"; do eval "$a" 2>/dev/null||true; done; exit 1; }
trap rollback ERR

[ "$EUID" -eq 0 ] || { log_error "Script must be run as root"; exit 1; }

# Ensure nginx installed
if ! command -v nginx >/dev/null; then
  log_info "Installing nginx..."
  apt update -y
  apt install -y nginx
  ROLLBACK_ACTIONS+=("apt remove -y nginx")
fi

systemctl enable nginx
systemctl start nginx

# Domain → Port mapping
declare -A DOMAINS
DOMAINS["dev.nammataga.com"]=1701
DOMAINS["tst.nammataga.com"]=1702
DOMAINS["stg.nammataga.com"]=1703
DOMAINS["devapi.nammataga.com"]=1801
DOMAINS["tstapi.nammataga.com"]=1802
DOMAINS["stgapi.nammataga.com"]=1803
DOMAINS["api.nammataga.com"]=8080
DOMAINS["nammataga.com"]=80   # could serve static files here if needed

log_info "Creating nginx reverse proxy configs..."

for domain in "${!DOMAINS[@]}"; do
  port=${DOMAINS[$domain]}

  mkdir -p "/var/www/$domain"
  echo "<h1>$domain is set up on port $port</h1>" > "/var/www/$domain/index.html"

  cat > "/etc/nginx/sites-available/$domain" <<EOF
server {
    listen 80;
    server_name $domain;

    # Root domain (static) vs others (proxy)
    $(if [[ $port -eq 80 ]]; then
        echo "root /var/www/$domain;"
        echo "index index.html;"
    else
        cat <<INNER
location / {
    proxy_pass http://127.0.0.1:$port;
    proxy_http_version 1.1;
    proxy_set_header Upgrade \$http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host \$host;
    proxy_cache_bypass \$http_upgrade;
}
INNER
    fi)
}
EOF

  ln -sf "/etc/nginx/sites-available/$domain" "/etc/nginx/sites-enabled/$domain"

  ROLLBACK_ACTIONS+=(
    "rm -f /etc/nginx/sites-available/$domain"
    "rm -f /etc/nginx/sites-enabled/$domain"
    "rm -rf /var/www/$domain"
  )
done

nginx -t || { log_error "nginx config test failed"; exit 1; }
systemctl reload nginx

# Help file
cat <<EOF > "$HELPFILE"
# 🌐 Nginx Reverse Proxy Configuration

Domains configured:
- nammataga.com → static files at /var/www/nammataga.com (listen 80)
- dev.nammataga.com → proxy to localhost:1701
- tst.nammataga.com → proxy to localhost:1702
- stg.nammataga.com → proxy to localhost:1703
- uatapi.nammataga.com → proxy to localhost:1801
- api.nammataga.com → proxy to localhost:8080

Nginx config files:
- /etc/nginx/sites-available/<domain>
- Enabled via symlink in /etc/nginx/sites-enabled/

Reload nginx after changes:
  sudo nginx -t && sudo systemctl reload nginx

Cloudflare SSL:
- Currently serving HTTP only
- Cloudflare provides SSL termination
- For Full (Strict) mode, add Origin certs under /etc/nginx/ssl/<domain>

EOF

log_info "✅ All reverse proxies created"
log_info "📖 Help file: $HELPFILE"