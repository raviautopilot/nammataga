#!/usr/bin/env bash

# Exit immediately if any critical command fails, treat unset variables as errors
set -euo pipefail

# --- CONFIGURATION ---
# Change this path if running on VPS (/apps/taga-api/prd/data/image) 
# or keep your local test path as needed.
DATA_DIR="/home/sudhan_dev/Downloads/code/nammataga/taga-api/data"
TARGET_DIR="${DATA_DIR}/image"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="${DATA_DIR}/image_backup_${TIMESTAMP}.tar.gz"

# Colors for clear console logging
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Starting Image Optimization ===${NC}"

# 1. Verify target directory exists
if [ ! -d "$TARGET_DIR" ]; then
    echo -e "${RED}Error: Target directory $TARGET_DIR does not exist!${NC}"
    exit 1
fi

cd "$TARGET_DIR"
echo -e "${GREEN}[+] Working inside directory: $(pwd)${NC}"

# 2. Check for required optimization tools
MISSING_TOOLS=()
command -v jpegoptim >/dev/null 2>&1 || MISSING_TOOLS+=("jpegoptim")
command -v optipng >/dev/null 2>&1 || MISSING_TOOLS+=("optipng")

if [ ${#MISSING_TOOLS[@]} -gt 0 ]; then
    echo -e "${RED}Error: Missing required tools: ${MISSING_TOOLS[*]}${NC}"
    echo -e "${YELLOW}Please install them first (e.g., sudo apt install jpegoptim optipng)${NC}"
    exit 1
fi

# 3. Check initial folder size
INITIAL_SIZE=$(du -sh . | awk '{print $1}')
echo -e "${GREEN}[+] Current image folder size: $INITIAL_SIZE${NC}"

# 4. Create secure backup archive in the parent data directory
echo -e "${YELLOW}[+] Creating safety backup archive at: $BACKUP_FILE ...${NC}"
tar -czf "$BACKUP_FILE" -C "$TARGET_DIR" .
echo -e "${GREEN}[+] Backup successfully created.${NC}"

# 5. Optimize JPEG / JPG files robustly (silencing minor header warnings without crashing)
echo -e "${YELLOW}[+] Optimizing all JPEG/JPG images (stripping metadata, max quality 85%)...${NC}"
find . -type f \( -iname "*.jpg" -o -iname "*.jpeg" \) -print0 | while IFS= read -r -d '' img; do
    # Ensure file is writable before optimizing
    chmod u+w "$img" 2>/dev/null || true
    
    # Run jpegoptim and suppress non-critical header warnings
    jpegoptim --max=85 --strip-all --quiet "$img" 2>/dev/null || true
done
echo -e "${GREEN}[+] JPEG optimization completed.${NC}"

# 6. Optimize PNG files robustly (lossless)
echo -e "${YELLOW}[+] Optimizing all PNG images (lossless compression & stripping metadata)...${NC}"
find . -type f -iname "*.png" -print0 | while IFS= read -r -d '' img; do
    # Ensure file is writable before optimizing
    chmod u+w "$img" 2>/dev/null || true
    
    optipng -o5 -strip all -quiet "$img" 2>/dev/null || true
done
echo -e "${GREEN}[+] PNG optimization completed.${NC}"

# 7. Check final folder size
FINAL_SIZE=$(du -sh . | awk '{print $1}')

echo -e "\n--------------------------------------------------"
echo -e "${GREEN}=== Image Optimization Finished Successfully! ===${NC}"
echo -e "Initial Size : ${INITIAL_SIZE}"
echo -e "Final Size   : ${FINAL_SIZE}"
echo -e "Backup File  : ${BACKUP_FILE}"
echo -e "--------------------------------------------------"