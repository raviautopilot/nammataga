#!/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WEB_DIR="${PROJECT_ROOT}/taga-web"

cd "$WEB_DIR"

# ===== REQUIRED ENV FILES =====
SOURCE_DIR="dist"
REQUIRED_FILES=(".env" ".env.development" ".env.production")

# ===== VALIDATION =====
for file in "${REQUIRED_FILES[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "❌ Missing required file: $file"
    exit 1
  fi

  if ! grep -q '^VITE_API_BASE_URL=' "$file"; then
    echo "❌ VITE_API_BASE_URL not defined in $file"
    exit 1
  fi

  VALUE=$(grep '^VITE_API_BASE_URL=' "$file" | cut -d '=' -f2- | tr -d '"')

  if [[ -z "$VALUE" ]]; then
    echo "❌ VITE_API_BASE_URL is empty in $file"
    exit 1
  fi
done

echo "✅ Env validation passed"

# ===== CONFIG =====
REMOTE_HOST="taga-prod"
REMOTE_USER="dev-taga"
REMOTE_PATH="/apps/taga-web/dev"

# ===== BUILD =====
npm install
npm run build:dev

ENV_FILE="dist/env-config.js"
cp env.config.js "$ENV_FILE"

if [[ -f "$ENV_FILE" ]]; then
  sed -i 's|"http://localhost:1801/api"|"https://devapi.nammataga.com/api"|g' "$ENV_FILE"
else
  echo "❌ $ENV_FILE not found"
  exit 1
fi

# ===== DEPLOY =====
rsync -az --delete \
  -e "ssh" \
  $SOURCE_DIR/ \
  "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_PATH}/"

echo "✅ Deployment completed"
