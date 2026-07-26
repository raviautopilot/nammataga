#!/usr/bin/env bash
set -euo pipefail

# ---------------- Configuration ----------------
MAIN_FILE="main.go"
OUTPUT_BINARY="taga-api"
OS="linux"
ARCH=$(go env GOARCH)

GOBIN="$(go env GOPATH)/bin"

DIRS=(
  .
  ./handler
)

# ---------------- Functions ----------------

check_and_install_swag() {
  # Case 1: swag available in PATH
  if command -v swag >/dev/null 2>&1; then
    echo "✅ swag is already available in PATH"
    return
  fi

  # Case 2: swag exists in GOPATH/bin but not in PATH
  if [ -x "$GOBIN/swag" ]; then
    echo "✅ swag found in GOPATH/bin, adding to PATH..."
    export PATH="$PATH:$GOBIN"
    return
  fi

  # Case 3: swag not installed → install
  echo "⚠️ swag not found. Installing..."

  go install github.com/swaggo/swag/cmd/swag@latest

  export PATH="$PATH:$GOBIN"

  # Final verification
  if ! command -v swag >/dev/null 2>&1; then
    echo "❌ Failed to install swag. Check Go setup."
    exit 1
  fi

  echo "✅ swag installed successfully!"
}

prepare_deps() {
  echo "Ensuring module dependencies are up to date..."
  go mod tidy
  go clean -cache 2>/dev/null || true
}

run_tests() {
  echo -e "\nRunning tests..."
  echo "========================================"
  ./tester.sh
  echo -e "\n✅ All tests passed!\n"
}

generate_docs() {
  echo "Running swag init..."
  swag init \
    --generalInfo "$MAIN_FILE" \
    --dir "$(IFS=,; echo "${DIRS[*]}")" \
    --output ./docs
  echo -e "\n✅ Swagger documentation generated successfully!\n"
}

build_binary() {
  local extension=""
  local target="${OUTPUT_BINARY}${extension}"

  echo "Building for $OS/$ARCH..."
  GOOS="$OS" GOARCH="$ARCH" go build -o "$target" "$MAIN_FILE"

  echo "========================================"
  echo "✅ Build completed successfully!"
  echo "Target: $OS/$ARCH"
  echo "Binary: $target"
  echo "========================================"
}

# ---------------- Main ----------------

check_and_install_swag
# prepare_deps
# run_tests
generate_docs
build_binary