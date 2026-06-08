#!/bin/sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    printf "${BLUE}[info]${NC} %s\n" "$1"
}

success() {
    printf "${GREEN}[success]${NC} %s\n" "$1"
}

warn() {
    printf "${YELLOW}[warning]${NC} %s\n" "$1"
}

error() {
    printf "${RED}[error]${NC} %s\n" "$1" >&2
    exit 1
}

# Detect OS
OS_UNAME=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS_UNAME" in
    darwin*)  OS="darwin" ;;
    linux*)   OS="linux" ;;
    msys*|mingw*|cygwin*) OS="windows" ;;
    *)        error "Unsupported operating system: $OS_UNAME" ;;
esac

# Detect Architecture
ARCH_UNAME=$(uname -m | tr '[:upper:]' '[:lower:]')
case "$ARCH_UNAME" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)            error "Unsupported architecture: $ARCH_UNAME" ;;
esac

info "Detected platform: ${OS}-${ARCH}"

# Check for curl or wget
if command -v curl >/dev/null 2>&1; then
    DOWNLOADER="curl"
elif command -v wget >/dev/null 2>&1; then
    DOWNLOADER="wget"
else
    error "Either curl or wget is required to run this script."
fi

# Fetch the latest release tag
info "Fetching latest release version..."
TAG=""
if [ "$DOWNLOADER" = "curl" ]; then
    TAG=$(curl -sI https://github.com/NEPT-CLOUD/nept-cli-go/releases/latest | grep -i location | sed -n 's|.*/tag/\(.*\)|\1|p' | tr -d '\r')
    if [ -z "$TAG" ]; then
        TAG=$(curl -s https://api.github.com/repos/NEPT-CLOUD/nept-cli-go/releases/latest | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    fi
else
    # wget fallback
    TAG=$(wget --max-redirect=0 https://github.com/NEPT-CLOUD/nept-cli-go/releases/latest 2>&1 | grep -i location | sed -n 's|.*/tag/\(.*\)|\1|p' | tr -d '\r')
    if [ -z "$TAG" ]; then
        TAG=$(wget -qO- https://api.github.com/repos/NEPT-CLOUD/nept-cli-go/releases/latest | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    fi
fi

# Clean tag variable
TAG=$(echo "$TAG" | tr -d '[:space:]')

if [ -z "$TAG" ]; then
    error "Failed to retrieve the latest version tag from GitHub."
fi

info "Latest version is $TAG"

# Determine filename and extension
EXT=""
if [ "$OS" = "windows" ]; then
    EXT=".exe"
fi

BINARY_NAME="nept-${OS}-${ARCH}${EXT}"
DOWNLOAD_URL="https://github.com/NEPT-CLOUD/nept-cli-go/releases/download/${TAG}/${BINARY_NAME}"

# Temporary download directory
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'nept-install')
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

TMP_BIN="$TMP_DIR/nept${EXT}"

info "Downloading binary from: $DOWNLOAD_URL"
if [ "$DOWNLOADER" = "curl" ]; then
    curl -fLo "$TMP_BIN" "$DOWNLOAD_URL" || error "Failed to download $DOWNLOAD_URL"
else
    wget -qO "$TMP_BIN" "$DOWNLOAD_URL" || error "Failed to download $DOWNLOAD_URL"
fi

chmod +x "$TMP_BIN"

# Determine where to install
INSTALL_DIR="/usr/local/bin"
# If we don't have write permission to /usr/local/bin, try sudo or fallback to user local bin
if [ "$OS" = "windows" ]; then
    INSTALL_DIR="/usr/bin"
    if [ ! -w "$INSTALL_DIR" ]; then
        INSTALL_DIR="./"
    fi
fi

info "Installing to $INSTALL_DIR/nept..."

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_BIN" "$INSTALL_DIR/nept${EXT}"
else
    if command -v sudo >/dev/null 2>&1; then
        info "Requesting sudo permissions to install to $INSTALL_DIR"
        sudo mv "$TMP_BIN" "$INSTALL_DIR/nept${EXT}"
    else
        warn "Cannot write to $INSTALL_DIR and sudo is not available."
        # Fallback to home directory
        USER_BIN="$HOME/.local/bin"
        mkdir -p "$USER_BIN"
        mv "$TMP_BIN" "$USER_BIN/nept${EXT}"
        INSTALL_DIR="$USER_BIN"
        warn "Installed to $INSTALL_DIR. Please ensure this directory is in your PATH."
    fi
fi

success "Successfully installed nept to $INSTALL_DIR/nept${EXT}"
