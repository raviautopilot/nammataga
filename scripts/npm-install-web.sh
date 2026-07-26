#!/bin/bash
# ==============================================================================
# Script Name: npm-install-web.sh
# Description: Installs npm dependencies for the frontend from the root directory.
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "📦 Running npm install for taga-web..."
cd "$PROJECT_ROOT/taga-web"
./npm-install.sh
