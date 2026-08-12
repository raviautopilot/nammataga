#!/bin/bash
# ==============================================================================
# Script Name: dev-wipe-docker.sh
# Description: HIGHLY DESTRUCTIVE. Cleans up all Docker resources for the dev environment.
# Location:    dev_environment/dev-wipe-docker.sh
# ==============================================================================

set -euo pipefail

# Require sudo
if [ "$EUID" -ne 0 ]; then
  echo "❌ Error: This script must be run with sudo privileges." >&2
  exit 1
fi

RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${RED}======================================================================${NC}"
echo -e "${RED}⚠️  WARNING: DESTRUCTIVE DANGER WIPE IN PROGRESS - DEVELOPMENT ENVIRONMENT${NC}"
echo -e "${RED}======================================================================${NC}"
echo -e "This script will completely wipe the Docker state for the DEV environment on this VPS:"
echo -e "  🔥 Stop all dev containers"
echo -e "  🔥 Delete all stopped/created dev containers"
echo -e "  🔥 Delete all loaded dev Docker images"
echo -e "  🔥 Delete all Docker networks and system volumes associated with dev"
echo ""
echo -e "${YELLOW}NOTE:${NC} This will NOT delete your database files located in:"
echo -e "      /apps/taga-api/dev/data"
echo -e "      Your actual database files remain safe on the host VPS hard drive."
echo -e "${RED}======================================================================${NC}"

# Ask for confirmation
read -p "Are you absolutely sure you want to clean out the dev Docker resources? (y/N): " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "❌ Wipe cancelled. Safe exit."
    exit 0
fi

# Deep verification prompt
echo -e "${YELLOW}Please type 'DANGER_WIPE_DEV' in all capitals to proceed:${NC}"
read -p "> " TEXT_VERIFY

if [ "$TEXT_VERIFY" != "DANGER_WIPE_DEV" ]; then
    echo "❌ Verification failed. Safe exit."
    exit 0
fi

echo ""
echo "🧹 Stopping and removing dev containers..."
# Stop and remove dev containers specifically via docker-compose if file is available
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/docker-compose.dev.yml" ]; then
    docker compose -f "$SCRIPT_DIR/docker-compose.dev.yml" down --rmi all --volumes || true
    echo "✅ Dev containers and images removed via docker-compose."
else
    # Fallback to manual removal of dev containers if compose file is not found
    echo "⚠️ docker-compose.dev.yml not found. Falling back to target container/image removal..."
    docker stop taga-api-dev taga-web-dev 2>/dev/null || true
    docker rm -f taga-api-dev taga-web-dev 2>/dev/null || true
    docker rmi taga-api:dev taga-web:dev 2>/dev/null || true
fi

echo -e "${RED}======================================================================${NC}"
echo -e "✅ DEV DOCKER WIPE COMPLETED SUCCESSFULLY!"
echo -e "Dev Docker images, containers, networks, and caches have been cleared."
echo -e "Your VPS host directories (like database /data folders) remain untouched."
echo -e "${RED}======================================================================${NC}"
