#!/bin/bash

# Define the base path for the application
BASE_PATH="/apps/taga-api"

# 1. Check if a parameter was provided
if [ -z "$1" ]; then
    echo "Error: No environment specified."
    echo "Usage: $0 [tst|stg|prd]"
    exit 1
fi

# 2. Assign environment and determine target directory
ENV=$1

case "$ENV" in
    tst)
        TARGET_DIR="$BASE_PATH/tst"
        SOURCE_DIR="../dev/"
        ;;
    stg)
        TARGET_DIR="$BASE_PATH/stg"
        SOURCE_DIR="../tst/"
        ;;
    prd)
        TARGET_DIR="$BASE_PATH/prd"
        SOURCE_DIR="../stg/"
        ;;
    *)
        echo "Error: Invalid environment '$ENV'."
        echo "Accepted values: tst, stg, prd"
        exit 1
        ;;
esac

# 3. Check if the source directory exists
if [ ! -d "$SOURCE_DIR" ]; then
    echo "Error: Source directory '$SOURCE_DIR' does not exist."
    echo "Promotion failed instantly."
    exit 1
fi

# 4. Check if the target directory exists
if [ ! -d "$TARGET_DIR" ]; then
    echo "Error: Target directory '$TARGET_DIR' does not exist."
    echo "Promotion failed instantly."
    exit 1
fi

# 5. Define files and directories to copy
# Add or remove items as needed
# Based on the files in the chat, plus common Go project directories
COPY_ITEMS=(
    "config.json"
    "taga-api"
)

# 6. Proceed with promotion logic
echo "Starting promotion to: $ENV"
echo "Source path: $SOURCE_DIR"
echo "Target path: $TARGET_DIR"

# Change to source directory to use relative paths
cd "$SOURCE_DIR" || {
    echo "Error: Cannot change to source directory '$SOURCE_DIR'"
    exit 1
}


# Copy each item
for item in "${COPY_ITEMS[@]}"; do
    # Check if the item exists in the source directory
    if [ -e "$item" ]; then
        echo "Copying $item..."
        # Use -r to copy directories recursively, -u to update only newer files
        cp -ru "$item" "$TARGET_DIR/" 2>/dev/null || cp -r "$item" "$TARGET_DIR/"
    else
        echo "Warning: $item does not exist in source directory. Skipping."
    fi
done

# 7. Update config.json for specific environments
if [ "$ENV" = "tst" ]; then
    echo "Updating config.json for tst environment..."
    CONFIG_FILE="$TARGET_DIR/config.json"

    # Check if jq is available
    if command -v jq >/dev/null 2>&1; then
        # Use jq to update port and reset_password_url
        jq '.port = 1802 | .reset_password_url = "https://tst.nammataga.com"' "$CONFIG_FILE" > "$CONFIG_FILE.tmp" && mv "$CONFIG_FILE.tmp" "$CONFIG_FILE"
        echo "Updated port to 1802 and reset_password_url to https://tst.nammataga.com"
    else
        echo "Error: jq is not installed. Cannot update config.json automatically."
        echo "Please install jq or manually update $CONFIG_FILE"
        exit 1
    fi
fi

echo "Promotion to $ENV completed successfully."
