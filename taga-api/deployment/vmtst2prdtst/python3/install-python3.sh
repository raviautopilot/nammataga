#!/bin/bash

# Python 3 Installation Script (Without Snap)
# For Ubuntu/Debian systems

set -e  # Exit on error

LOG_FILE=LOG_FILE="$(pwd)/python-install-$(date +%Y%m%d-%H%M%S).log"
DESIRED_MINOR_VERSION="13"  # Change this to your desired version

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to log messages
log_message() {
    echo -e "$1"
    echo -e "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_message "${RED}Please run as root using sudo${NC}"
    exit 1
fi

log_message "${YELLOW}Starting Python 3 installation without Snap...${NC}"
log_message "Log file: $LOG_FILE"

# Update package list
log_message "${YELLOW}Updating package list...${NC}"
apt update >> "$LOG_FILE" 2>&1

# Check if Python 3 is already installed
if command -v python3 >/dev/null 2>&1; then
    INSTALLED_VERSION=$(python3 --version 2>&1)
    log_message "${GREEN}Python 3 is already installed: $INSTALLED_VERSION${NC}"
else
    # Install Python 3
    log_message "${YELLOW}Installing Python 3...${NC}"
    apt install -y python3 >> "$LOG_FILE" 2>&1
    
    if command -v python3 >/dev/null 2>&1; then
        INSTALLED_VERSION=$(python3 --version 2>&1)
        log_message "${GREEN}Successfully installed: $INSTALLED_VERSION${NC}"
    else
        log_message "${RED}Failed to install Python 3${NC}"
        exit 1
    fi
fi

# Check current Python version
CURRENT_VERSION=$(python3 -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
log_message "${YELLOW}Current Python version: $CURRENT_VERSION${NC}"

# Check if we need to install a specific version
if [[ ! "$CURRENT_VERSION" =~ \.$DESIRED_MINOR_VERSION$ ]]; then
    log_message "${YELLOW}Current Python version ($CURRENT_VERSION) doesn't match desired version (3.$DESIRED_MINOR_VERSION)${NC}"
    
    # Install specific Python version using deadsnakes PPA
    read -p "Do you want to install Python 3.$DESIRED_MINOR_VERSION from deadsnakes PPA? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_message "${YELLOW}Adding deadsnakes PPA...${NC}"
        apt install -y software-properties-common >> "$LOG_FILE" 2>&1
        add-apt-repository -y ppa:deadsnakes/ppa >> "$LOG_FILE" 2>&1
        apt update >> "$LOG_FILE" 2>&1

        # USE THE VARIABLE INSTEAD OF HARDCODED VERSION
        log_message "${YELLOW}Installing Python 3.$DESIRED_MINOR_VERSION...${NC}"
        apt install -y "python3.$DESIRED_MINOR_VERSION" "python3.$DESIRED_MINOR_VERSION-venv" "python3.$DESIRED_MINOR_VERSION-dev" >> "$LOG_FILE" 2>&1

        log_message "${GREEN}Python 3.$DESIRED_MINOR_VERSION installed alongside system Python${NC}"

        # Set up alternatives
        log_message "${YELLOW}Setting up Python alternatives...${NC}"
        update-alternatives --install /usr/bin/python3 python3 "/usr/bin/python3.$DESIRED_MINOR_VERSION" 2 >> "$LOG_FILE" 2>&1
        update-alternatives --install /usr/bin/python3 python3 "/usr/bin/python3" 1 >> "$LOG_FILE" 2>&1
    fi
fi

log_message "${GREEN}Python 3 installation completed successfully!${NC}"
log_message "${YELLOW}Check log file for details: $LOG_FILE${NC}"

# Display installation summary
echo ""
echo "=== INSTALLATION SUMMARY ==="
echo "Python 3: $(command -v python3) - $(python3 --version 2>&1)"
echo "pip3: $(command -v pip3) - $(pip3 --version 2>&1 | cut -d ' ' -f 1-2)"
echo "venv: $(python3 -c 'import venv; print("Available")' 2>/dev/null || echo "Not available")"
echo "============================"
