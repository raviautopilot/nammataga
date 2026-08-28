#!/bin/bash
# ==============================================================================
# Script Name: prd-wipe-docker.sh
# Description: HIGHLY DESTRUCTIVE. Cleans up all Docker resources on the VPS.
# Location:    taga-config/docker/prd-wipe-docker.sh
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
echo -e "${RED}⚠️  WARNING: DESTRUCTIVE DANGER WIPE IN PROGRESS${NC}"
echo -e "${RED}======================================================================${NC}"
echo -e "This script will completely wipe the Docker state on this VPS:"
echo -e "  🔥 Stop all running containers"
echo -e "  🔥 Delete all stopped/created containers"
echo -e "  🔥 Delete all loaded Docker images (including frontend and backend)"
echo -e "  🔥 Delete all Docker networks and system volumes"
echo -e "  🔥 Run docker system prune --all --volumes"
echo ""
echo -e "${YELLOW}NOTE:${NC} This will NOT delete your database files located in:"
echo -e "      /apps/taga-api/prd/data"
echo -e "      Your actual database files remain safe on the host VPS hard drive."
echo -e "${RED}======================================================================${NC}"

# Ask for confirmation
read -p "Are you absolutely sure you want to clean out the Docker engine? (y/N): " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "❌ Wipe cancelled. Safe exit."
    exit 0
fi

# Deep verification prompt
echo -e "${YELLOW}Please type 'DANGER_WIPE' in all capitals to proceed:${NC}"
read -p "> " TEXT_VERIFY

if [ "$TEXT_VERIFY" != "DANGER_WIPE" ]; then
    echo "❌ Verification failed. Safe exit."
    exit 0
fi

echo ""
echo "🧹 Wiping docker containers..."
# Stop and remove all containers
CONTAINERS=$(docker ps -a -q)
if [ -n "$CONTAINERS" ]; then
    docker stop $CONTAINERS || true
    docker rm -f $CONTAINERS
    echo "✅ All containers removed."
else
    echo "ℹ️ No containers found to remove."
fi

echo "🧹 Pruning system images, volumes, and caches..."
docker system prune -a --volumes -f

echo "🧹 Pruning unused networks..."
docker network prune -f

echo -e "${RED}======================================================================${NC}"
echo -e "✅ DOCKER WIPE COMPLETED SUCCESSFULLY!"
echo -e "All Docker images, containers, networks, and caches have been cleared."
echo -e "Your VPS host directories (like database /data folders) remain untouched."
echo -e "${RED}======================================================================${NC}"