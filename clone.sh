#!/bin/bash

# Exit on any error
set -e

# Check if target directory is provided
if [ -z "$1" ]; then
    echo "Error: Target directory not specified."
    echo "Usage: $0 <target_directory>"
    exit 1
fi

TARGET_DIR="$1"
SOURCE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Ensure source is a git repository
if ! git -C "$SOURCE_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "Error: Source directory is not a git repository."
    exit 1
fi

# Create target directory if it doesn't exist
mkdir -p "$TARGET_DIR"
# Resolve absolute path of target directory
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"

# Check if target directory is same as source directory
if [ "$SOURCE_DIR" = "$TARGET_DIR" ]; then
    echo "Error: Target directory cannot be the same as the source directory."
    exit 1
fi

echo "Cloning project from: $SOURCE_DIR"
echo "To target location:   $TARGET_DIR"

# 1. Copy the .git directory to preserve git history and repository configuration
if [ -d "$SOURCE_DIR/.git" ]; then
    echo "Copying git repository metadata (.git)..."
    # Ensure any existing .git in target is cleared
    rm -rf "$TARGET_DIR/.git"
    cp -rp "$SOURCE_DIR/.git" "$TARGET_DIR/"
fi

# 2. Check if target directory is inside the source directory to prevent recursion
EXCLUDE_PATTERN=""
if [[ "$TARGET_DIR" == "$SOURCE_DIR"/* ]]; then
    REL_TARGET="${TARGET_DIR#$SOURCE_DIR/}"
    # Prepare grep pattern to filter out files inside the target directory
    EXCLUDE_PATTERN="^${REL_TARGET}/"
fi

get_files() {
    if [ -n "$EXCLUDE_PATTERN" ]; then
        git ls-files -z -c -o --exclude-standard | grep -z -v "$EXCLUDE_PATTERN"
    else
        git ls-files -z -c -o --exclude-standard
    fi
}

# 3. Copy all tracked files and untracked (but non-ignored) files.
# This ensures that all .gitignored files/folders (such as evidence/,
# graphify-out/, logs, node_modules etc.) are excluded.
echo "Copying non-ignored project files..."
if command -v rsync >/dev/null 2>&1; then
    # Use rsync for copying if available (preserves file times/permissions and shows progress)
    get_files | rsync -0 -av --files-from=- "$SOURCE_DIR/" "$TARGET_DIR/"
else
    # Fallback to tar if rsync is not available
    get_files | tar --null -cf - -C "$SOURCE_DIR" -T - | tar -xf - -C "$TARGET_DIR"
fi

echo "Successfully cloned project to $TARGET_DIR (excluding gitignored files)."
