#!/bin/bash

# ============================================
# Data Sync Script - VPS to Local (FIXED)
# ============================================

# Configuration
REMOTE_PATH="/apps/taga-api/prd/data/"
LOCAL_PATH="/home/ubuntu/code/github/raviautopilot/nammataga/taga-api/data/"
SSH_ALIAS="sys-taga"
LOG_FILE="/home/ubuntu/code/github/raviautopilot/nammataga/logs/sync-$(date +%Y%m%d-%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Create local directory if it doesn't exist
mkdir -p "$LOCAL_PATH"

# Fix local permissions (make sure we can write)
print_info "Ensuring local directory is writable..."
chmod -R u+w "$LOCAL_PATH" 2>/dev/null || true

# Main sync function
do_sync() {
    print_info "Starting synchronization..."
    print_info "Source: $SSH_ALIAS:$REMOTE_PATH"
    print_info "Destination: $LOCAL_PATH"
    echo "---------------------------------------------------"
    
    # Sync with options that avoid permission errors
    rsync -avz \
        --progress \
        --no-perms \
        --no-owner \
        --no-group \
        --chmod=ugo=rwX \
        --rsync-path="sudo rsync" \
        "$SSH_ALIAS:$REMOTE_PATH" \
        "$LOCAL_PATH" \
        2>&1 | tee "$LOG_FILE"
    
    # Check exit status
    if [ ${PIPESTATUS[0]} -eq 0 ] || [ ${PIPESTATUS[0]} -eq 23 ]; then
        # Exit code 23 is "partial transfer due to errors" - which is fine for permissions
        echo "---------------------------------------------------"
        print_info "Sync completed!"
        print_info "Log saved to: $LOG_FILE"
        
        # Show what was copied
        print_info "Files copied: $(find "$LOCAL_PATH" -type f 2>/dev/null | wc -l)"
        print_info "Total size: $(du -sh "$LOCAL_PATH" 2>/dev/null | cut -f1)"
        return 0
    else
        echo "---------------------------------------------------"
        print_error "Sync failed! Check log: $LOG_FILE"
        return 1
    fi
}

# Show statistics
show_stats() {
    echo "============================================"
    print_info "Local folder stats:"
    echo "  Size: $(du -sh "$LOCAL_PATH" 2>/dev/null || echo 'N/A')"
    echo "  Files: $(find "$LOCAL_PATH" -type f 2>/dev/null | wc -l)"
    echo "  Directories: $(find "$LOCAL_PATH" -type d 2>/dev/null | wc -l)"
    echo "============================================"
}

# Check if we have local write permissions
check_local_permissions() {
    if [ -d "$LOCAL_PATH" ]; then
        if [ -w "$LOCAL_PATH" ]; then
            print_info "Local directory is writable ✓"
            return 0
        else
            print_warning "Local directory is not writable! Attempting to fix..."
            chmod -R u+w "$LOCAL_PATH" 2>/dev/null || true
            if [ -w "$LOCAL_PATH" ]; then
                print_info "Permissions fixed ✓"
                return 0
            else
                print_error "Cannot write to local directory!"
                return 1
            fi
        fi
    else
        print_info "Creating local directory..."
        mkdir -p "$LOCAL_PATH"
        if [ -w "$LOCAL_PATH" ]; then
            print_info "Created and writable ✓"
            return 0
        else
            print_error "Cannot create local directory!"
            return 1
        fi
    fi
}

# Main execution
echo "============================================"
echo "  VPS Data Sync Script"
echo "  $(date)"
echo "============================================"

# Check permissions
if ! check_local_permissions; then
    print_error "Aborting..."
    exit 1
fi

# Perform sync
if do_sync; then
    show_stats
    exit 0
else
    exit 1
fi