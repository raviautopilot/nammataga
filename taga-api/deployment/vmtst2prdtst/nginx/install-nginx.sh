#!/usr/bin/env bash
set -euo pipefail
LOGFILE="/tmp/nginx-install.log"
HELPFILE="/usr/local/share/nginx-help.txt"
ROLLBACK_ACTIONS=()

log_error()   { echo "[ERROR] ${*:-Unknown}" | tee -a "$LOGFILE"; }
log_info()    { echo "[INFO] $*"; }
rollback()    { for a in "${ROLLBACK_ACTIONS[@]}"; do eval "$a" 2>/dev/null||true; done; exit 1; }
trap rollback ERR

[ "$EUID" -eq 0 ] || { log_error "must run as root"; exit 1; }

log_info "Installing Nginx..."
apt install -y nginx
ROLLBACK_ACTIONS+=("apt remove -y nginx")

systemctl enable nginx
systemctl start nginx
command -v nginx >/dev/null || log_error "nginx missing"

cat <<EOF > "$HELPFILE"
# Nginx Web Server

- Installed via apt → available for all users at /usr/sbin/nginx
- Configs: /etc/nginx/sites-available/
- Enable site: ln -s /etc/nginx/sites-available/<site> /etc/nginx/sites-enabled/
- Test config: nginx -t
- Reload: systemctl reload nginx
- Default docroot: /var/www/html

### Status commands
- Check status: systemctl status nginx
- Start: systemctl start nginx
- Stop: systemctl stop nginx
- Restart: systemctl restart nginx
EOF

log_info "✅ Nginx installed and available for all users."