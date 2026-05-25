#!/usr/bin/env bash
set -euo pipefail

INSTALL_DIR="$HOME/.gon/bin"
BINARY_NAME="gon"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
RESET='\033[0m'

error() {
  echo -e "${RED}error:${RESET} $1" >&2
  exit 1
}

# Resolve script directory (project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Set version from latest commit
COMMIT=$(git -C "$SCRIPT_DIR" rev-parse --short HEAD 2>/dev/null) \
  || error "Failed to get git commit hash. Is this a git repository?"
VERSION="nightly-${COMMIT}"

# Validate environment
if ! command -v go &>/dev/null; then
  error "go is required but not installed."
fi

echo -e "Building ${BOLD}gon ${VERSION}${RESET} from source..."

TMP_FILE=$(mktemp)
trap 'rm -f "$TMP_FILE"' EXIT

echo "  Building binary..."
if ! go build \
  -ldflags="-X gon/internal/version.Version=${VERSION}" \
  -o "$TMP_FILE" \
  "$SCRIPT_DIR/cmd/main.go"; then
  error "Build failed."
fi

# Install binary
echo "  Installing to ${INSTALL_DIR}/${BINARY_NAME}"
mkdir -p "$INSTALL_DIR"
mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Setup PATH
EXPORT_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""

add_to_path() {
  local rc_file="$1"
  if [[ -f "$rc_file" ]]; then
    if ! grep -qF '.gon/bin' "$rc_file"; then
      echo "" >> "$rc_file"
      echo "# gon" >> "$rc_file"
      echo "$EXPORT_LINE" >> "$rc_file"
      echo "  Adding ~/.gon/bin to PATH in $(basename "$rc_file")"
      return 0
    fi
  fi
  return 1
}

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
  UPDATED_RC=0
  if [[ -f "$HOME/.bashrc" ]]; then
    add_to_path "$HOME/.bashrc" && UPDATED_RC=1
  fi
  if [[ -f "$HOME/.zshrc" ]]; then
    add_to_path "$HOME/.zshrc" && UPDATED_RC=1
  fi
  if [[ "$UPDATED_RC" == 0 ]]; then
    echo "  Manually add to your shell config:"
    echo "    $EXPORT_LINE"
  fi
fi

echo ""
echo -e "${GREEN}${BOLD}gon ${VERSION} installed successfully!${RESET}"
echo "Run 'source ~/.bashrc' (or your shell's rc file) or start a new terminal."
