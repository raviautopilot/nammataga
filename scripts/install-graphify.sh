#!/bin/bash

# ==============================================================================
# Script Name: install-graphify.sh
# Description: Installs Python 3, Pip, and graphify globally. Includes
#              post-installation tests and detailed logging.
# ==============================================================================

set -euo pipefail

LOG_FILE="/var/log/graphify_install.log"

# Ensure user is root
if [ "$EUID" -ne 0 ]; then 
    echo "Error: This script must be run with sudo privileges to install globally." >&2
    exit 1
fi

# Initialize log file
touch "$LOG_FILE"
chmod 644 "$LOG_FILE"

# Logging function
log() {
    local msg="$1"
    local timestamp
    timestamp=$(date "+%Y-%m-%d %H:%M:%S")
    echo "[$timestamp] $msg" | tee -a "$LOG_FILE"
}

# Error handling function
handle_error() {
    local step="$1"
    local exit_code="$2"
    local advice="$3"
    
    echo ""
    echo "==================================================" | tee -a "$LOG_FILE"
    echo "❌ INSTALLATION FAILED" | tee -a "$LOG_FILE"
    echo "Step: $step" | tee -a "$LOG_FILE"
    echo "Exit Code: $exit_code" | tee -a "$LOG_FILE"
    echo "Explanation: The command failed to execute properly and returned a non-zero status." | tee -a "$LOG_FILE"
    echo "Suggested Action: $advice" | tee -a "$LOG_FILE"
    echo "More Details: Please check the detailed logs in $LOG_FILE" | tee -a "$LOG_FILE"
    echo "==================================================" | tee -a "$LOG_FILE"
    
    exit "$exit_code"
}

log "🚀 Starting global installation of graphify..."

# Step 1: Update package lists
log "Step 1/5: Updating system package repositories..."
exit_code=0
apt-get update -y >> "$LOG_FILE" 2>&1 || exit_code=$?
if [ $exit_code -ne 0 ]; then
    handle_error "Updating package repositories (apt-get update)" $exit_code \
        "Check your internet connection, verify your DNS settings, or ensure the configured apt mirrors are reachable."
fi

# Step 2: Install prerequisites
log "Step 2/5: Installing Python3 and Pip prerequisites..."
exit_code=0
apt-get install -y python3 python3-pip >> "$LOG_FILE" 2>&1 || exit_code=$?
if [ $exit_code -ne 0 ]; then
    handle_error "Installing python3 and python3-pip" $exit_code \
        "Run 'sudo apt-get --fix-broken install' to resolve system package conflicts, or manually run the install command to see specific missing dependencies."
fi

# Step 3: Handle PIP flags for newer systems
log "Step 3/5: Checking pip environment compatibility..."
PIP_FLAGS="--ignore-installed"
if pip3 install --help 2>&1 | grep -q "break-system-packages"; then
    log "Notice: Newer Debian/Ubuntu environment detected. Adding --break-system-packages to pip."
    PIP_FLAGS="--ignore-installed --break-system-packages"
fi

# Step 4: Install graphify
log "Step 4/5: Installing graphify globally via pip..."
exit_code=0
pip3 install $PIP_FLAGS "graphifyy[gemini]" >> "$LOG_FILE" 2>&1 || exit_code=$?
if [ $exit_code -ne 0 ]; then
    handle_error "Installing graphify via pip3" $exit_code \
        "Check if 'pip3' has access to PyPI (network/firewall issue). Alternatively, run 'sudo pip3 install \"graphifyy[gemini]\"' manually to view the Python traceback."
fi

# Step 5: Post-installation tests
log "Step 5/5: Running post-installation validation tests..."
if ! command -v graphify >> "$LOG_FILE" 2>&1; then
    handle_error "Locating the graphify executable" 1 \
        "The 'graphify' binary was not found in the system PATH. Ensure the pip binary installation path (usually /usr/local/bin) is included in your global PATH."
fi

exit_code=0
graphify --help >> "$LOG_FILE" 2>&1 || exit_code=$?
if [ $exit_code -ne 0 ]; then
    handle_error "Executing graphify binary" $exit_code \
        "The installation completed, but 'graphify' crashed when executed. You may have conflicting Python libraries. Run 'graphify' manually to inspect the error."
fi

log "✅ Success! graphify is successfully installed globally and passed all tests."
