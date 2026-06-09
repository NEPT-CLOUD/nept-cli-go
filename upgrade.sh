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

# Parse flags
FORCE=false
while [ "$#" -gt 0 ]; do
    case "$1" in
        -f|--force)
            FORCE=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -f, --force    Force upgrade even if local version is already up-to-date"
            echo "  -h, --help     Show this help message"
            exit 0
            ;;
        *)
            error "Unknown argument: $1"
            ;;
    esac
done

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

# Check for curl or wget
if command -v curl >/dev/null 2>&1; then
    DOWNLOADER="curl"
elif command -v wget >/dev/null 2>&1; then
    DOWNLOADER="wget"
else
    error "Either curl or wget is required to run this script."
fi

# Fetch the latest release tag
info "Fetching latest release version from GitHub..."
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

TAG=$(echo "$TAG" | tr -d '[:space:]')

if [ -z "$TAG" ]; then
    error "Failed to retrieve the latest version tag from GitHub."
fi

info "Latest version is $TAG"

# Find local nept path
NEPT_PATH=$(which nept 2>/dev/null || command -v nept 2>/dev/null || true)

# If not in path, check common locations
if [ -z "$NEPT_PATH" ]; then
    for path in "/usr/local/bin/nept" "$HOME/.local/bin/nept" "./nept"; do
        if [ -f "$path" ] && [ -x "$path" ]; then
            NEPT_PATH="$path"
            break
        fi
    done
fi

CURRENT_VERSION=""
if [ -n "$NEPT_PATH" ]; then
    info "Found existing Nept CLI at $NEPT_PATH"
    # Try parsing version as JSON
    CURRENT_VERSION=$("$NEPT_PATH" version -f json 2>/dev/null | grep -o '"version":"[^"]*"' | cut -d'"' -f4 || true)
    
    # Try parsing version as plain text if JSON failed
    if [ -z "$CURRENT_VERSION" ]; then
        CURRENT_VERSION=$("$NEPT_PATH" version 2>/dev/null | grep -i "nept version:" | awk '{print $NF}' || true)
    fi
fi

# Clean current version variable
CURRENT_VERSION=$(echo "$CURRENT_VERSION" | tr -d '[:space:]')

if [ -n "$CURRENT_VERSION" ]; then
    info "Currently installed version is: $CURRENT_VERSION"
else
    warn "Could not determine currently installed version (or Nept CLI is not installed)."
    CURRENT_VERSION="none"
fi

# Compare versions (strip leading 'v' for comparison if present)
LATEST_NORM=$(echo "$TAG" | sed 's/^v//')
CURRENT_NORM=$(echo "$CURRENT_VERSION" | sed 's/^v//')

if [ "$CURRENT_NORM" = "$LATEST_NORM" ] && [ "$FORCE" = false ]; then
    success "Nept CLI is already up-to-date (version $TAG)."
    exit 0
fi

if [ "$FORCE" = true ]; then
    info "Force flag is set. Proceeding with upgrade to $TAG..."
else
    info "Upgrading Nept CLI from $CURRENT_VERSION to $TAG..."
fi

# Determine filename and extension
EXT=""
if [ "$OS" = "windows" ]; then
    EXT=".exe"
fi

BINARY_NAME="nept-${OS}-${ARCH}${EXT}"
DOWNLOAD_URL="https://github.com/NEPT-CLOUD/nept-cli-go/releases/download/${TAG}/${BINARY_NAME}"

# Temporary download directory
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'nept-upgrade')
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

# Determine target install directory and file name
if [ -n "$NEPT_PATH" ]; then
    INSTALL_DIR=$(dirname "$NEPT_PATH")
    TARGET_BIN="$NEPT_PATH"
else
    # Fallback default install directory
    INSTALL_DIR="/usr/local/bin"
    if [ "$OS" = "windows" ]; then
        INSTALL_DIR="/usr/bin"
        if [ ! -w "$INSTALL_DIR" ]; then
            INSTALL_DIR="./"
        fi
    fi
    TARGET_BIN="$INSTALL_DIR/nept${EXT}"
fi

info "Installing upgraded binary to $TARGET_BIN..."

# Create target directory if it doesn't exist (e.g. if target directory was deleted)
TARGET_DIR=$(dirname "$TARGET_BIN")
if [ ! -d "$TARGET_DIR" ]; then
    mkdir -p "$TARGET_DIR"
fi

# Write file logic
if [ -w "$TARGET_DIR" ] && ( [ ! -f "$TARGET_BIN" ] || [ -w "$TARGET_BIN" ] ); then
    mv "$TMP_BIN" "$TARGET_BIN"
else
    if command -v sudo >/dev/null 2>&1; then
        info "Requesting sudo permissions to overwrite $TARGET_BIN"
        sudo mv "$TMP_BIN" "$TARGET_BIN"
    else
        warn "Cannot write to $TARGET_DIR/nept and sudo is not available."
        # Fallback to home directory
        USER_BIN="$HOME/.local/bin"
        mkdir -p "$USER_BIN"
        TARGET_BIN="$USER_BIN/nept${EXT}"
        mv "$TMP_BIN" "$TARGET_BIN"
        warn "Installed to fallback location: $TARGET_BIN. Please ensure this directory is in your PATH."
    fi
fi

# Install the skill folder to the host
SKILL_DIR="$HOME/.nept/skill"
info "Updating skill folder at $SKILL_DIR..."
mkdir -p "$SKILL_DIR"

SKILL_URL="https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/${TAG}/skill/SKILL.md"
if [ "$DOWNLOADER" = "curl" ]; then
    curl -fLo "$SKILL_DIR/SKILL.md" "$SKILL_URL" || {
        warn "Failed to download skill file from $SKILL_URL. Trying fallback to main..."
        curl -fLo "$SKILL_DIR/SKILL.md" "https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/skill/SKILL.md" || warn "Failed to download skill file from fallback URL"
    }
else
    wget -qO "$SKILL_DIR/SKILL.md" "$SKILL_URL" || {
        warn "Failed to download skill file from $SKILL_URL. Trying fallback to main..."
        wget -qO "$SKILL_DIR/SKILL.md" "https://raw.githubusercontent.com/NEPT-CLOUD/nept-cli-go/main/skill/SKILL.md" || warn "Failed to download skill file from fallback URL"
    }
fi

success "Successfully upgraded Nept CLI to version $TAG at $TARGET_BIN"
