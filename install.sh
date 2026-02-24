#!/bin/sh
# qcli installer — downloads the latest release binary for your platform.
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/JValdivia23/quota-cli/main/install.sh | sh
#
# Installs to /usr/local/bin by default. Set INSTALL_DIR to override:
#   INSTALL_DIR=~/.local/bin curl -fsSL ... | sh

set -e

REPO="JValdivia23/quota-cli"
BINARY="qcli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  OS_TAG="Linux" ;;
  Darwin) OS_TAG="Darwin" ;;
  *)
    echo "Unsupported OS: $OS"
    echo "Please download manually from: https://github.com/$REPO/releases"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)  ARCH_TAG="x86_64" ;;
  arm64|aarch64) ARCH_TAG="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    echo "Please download manually from: https://github.com/$REPO/releases"
    exit 1
    ;;
esac

# Fetch latest release tag
echo "Fetching latest qcli release..."
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' | head -1 | sed 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Could not determine latest release. Check https://github.com/$REPO/releases"
  exit 1
fi

ARCHIVE="${BINARY}_${OS_TAG}_${ARCH_TAG}.tar.gz"
URL="https://github.com/$REPO/releases/download/$LATEST/$ARCHIVE"

echo "Downloading qcli $LATEST for $OS_TAG/$ARCH_TAG..."
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "$URL" -o "$TMP_DIR/$ARCHIVE"
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

# Install
if [ ! -d "$INSTALL_DIR" ]; then
  mkdir -p "$INSTALL_DIR"
fi

if [ -w "$INSTALL_DIR" ]; then
  cp "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo cp "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "✓ qcli $LATEST installed to $INSTALL_DIR/$BINARY"
echo ""

# Verify install
if command -v qcli >/dev/null 2>&1; then
  echo "Run: qcli status"
else
  echo "Note: Make sure $INSTALL_DIR is in your PATH."
  echo "  Add to your shell profile:  export PATH=\"\$PATH:$INSTALL_DIR\""
fi
