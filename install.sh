#!/usr/bin/env sh
set -e

REPO="vmsfigueredo/gflow"
BINARY="gflow"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# detect arch
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# fetch latest version
VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4 | tr -d v)"
if [ -z "$VERSION" ]; then
  echo "Failed to fetch latest version"
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/v${VERSION}/${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"

echo "Installing gflow v${VERSION} (${OS}/${ARCH})..."
curl -fsSL "$URL" | tar xz -C /tmp
install -m755 "/tmp/${BINARY}" "${INSTALL_DIR}/${BINARY}"
rm -f "/tmp/${BINARY}"

echo "Installed: $(${INSTALL_DIR}/${BINARY} version)"
