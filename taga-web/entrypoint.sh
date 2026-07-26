#!/bin/sh

# Generate dynamic env-config.js file at container boot
cat <<EOF > /usr/share/nginx/html/env-config.js
window._env_ = {
  VITE_API_BASE_URL: "${VITE_API_BASE_URL:-https://api.nammataga.com/api}",
  VITE_RAZORPAY_KEY: "${VITE_RAZORPAY_KEY:-}"
};
EOF

# Start Nginx in foreground
exec nginx -g "daemon off;"
