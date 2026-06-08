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
for arg in "$@"; do
    case $arg in
        -y|--yes) FORCE=true ;;
    esac
done

# Detect OS
OS_UNAME=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS_UNAME" in
    darwin*)  OS="darwin" ;;
    linux*)   OS="linux" ;;
    msys*|mingw*|cygwin*) OS="windows" ;;
    *)        OS="unknown" ;;
esac

# Find all possible locations of the binary
BINARY_PATHS=""

# Check path resolved by shell
RESOLVED_BIN=$(command -v nept 2>/dev/null || true)
if [ -n "$RESOLVED_BIN" ]; then
    BINARY_PATHS="$RESOLVED_BIN"
fi

# Standard path checks
EXT=""
if [ "$OS" = "windows" ]; then
    EXT=".exe"
fi

STANDARDS="/usr/local/bin/nept${EXT} $HOME/.local/bin/nept${EXT} $HOME/bin/nept${EXT}"
if [ "$OS" = "windows" ]; then
    STANDARDS="$STANDARDS $HOME/.nept/bin/nept.exe"
fi

for path in $STANDARDS; do
    if [ -f "$path" ]; then
        # Check if already added to BINARY_PATHS
        case "$BINARY_PATHS" in
            *"$path"*) ;;
            *) BINARY_PATHS="$BINARY_PATHS $path" ;;
        esac
    fi
done

# Remove duplicates/leading space
BINARY_PATHS=$(echo "$BINARY_PATHS" | tr ' ' '\n' | sort -u | tr '\n' ' ')

REMOVED_ANY=false

for path in $BINARY_PATHS; do
    if [ -f "$path" ]; then
        info "Found binary at: $path"
        # In Unix, deleting a file requires write permissions on its parent directory
        if [ -w "$(dirname "$path")" ]; then
            rm -f "$path"
            REMOVED_ANY=true
            info "Removed binary."
        else
            if command -v sudo >/dev/null 2>&1; then
                info "Requesting sudo permissions to remove $path"
                sudo rm -f "$path"
                REMOVED_ANY=true
                info "Removed binary."
            else
                warn "Cannot remove $path: parent directory is not writable and sudo is not available."
            fi
        fi
    fi
done

# Remove Windows local installation folder if it exists
if [ "$OS" = "windows" ]; then
    WIN_DIR="$HOME/.nept"
    if [ -d "$WIN_DIR" ]; then
        info "Cleaning up directory: $WIN_DIR"
        rm -rf "$WIN_DIR"
        REMOVED_ANY=true
    fi
fi

# Prompt or remove configuration file
CONFIG_FILE="$HOME/.nept.yaml"
if [ -f "$CONFIG_FILE" ]; then
    DELETE_CONFIG=false
    if [ "$FORCE" = true ]; then
        DELETE_CONFIG=true
    elif [ -t 0 ]; then
        printf "Do you want to delete the global configuration file ($CONFIG_FILE)? [y/N]: "
        read -r response
        case "$response" in
            [yY][eE][sS]|[yY]) DELETE_CONFIG=true ;;
        esac
    fi

    if [ "$DELETE_CONFIG" = true ]; then
        rm -f "$CONFIG_FILE"
        success "Removed configuration file: $CONFIG_FILE"
    else
        info "Kept configuration file: $CONFIG_FILE"
    fi
fi

if [ "$REMOVED_ANY" = true ]; then
    success "Successfully uninstalled nept."
else
    warn "No nept installations were found."
fi
