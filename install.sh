#!/bin/bash

# ShareSeer MCP Server Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/shareseer/mcp-server/main/install.sh | sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# GitHub repository
REPO="shareseer/mcp-server"
BINARY_NAME="shareseer-mcp"

# Default installation directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect platform and architecture
detect_platform() {
    local os arch
    
    # Detect OS
    case "$(uname -s)" in
        Darwin*)    os="darwin" ;;
        Linux*)     os="linux" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          
            print_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    if [ "$os" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
        PLATFORM_BINARY="${BINARY_NAME%.*}-${os}-${arch}.exe"
    else
        PLATFORM_BINARY="${BINARY_NAME}-${os}-${arch}"
    fi
    
    print_status "Detected platform: ${os}-${arch}"
}

# Get the latest release version
get_latest_version() {
    print_status "Fetching latest release information..."
    
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "curl or wget is required to download the binary"
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        print_error "Failed to get latest version"
        exit 1
    fi
    
    print_status "Latest version: ${VERSION}"
}

# Download the binary
download_binary() {
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${PLATFORM_BINARY}"
    local temp_file="/tmp/${PLATFORM_BINARY}"
    
    print_status "Downloading ${PLATFORM_BINARY}..."
    print_status "URL: ${download_url}"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "${temp_file}" "${download_url}"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "${temp_file}" "${download_url}"
    else
        print_error "curl or wget is required to download the binary"
        exit 1
    fi
    
    if [ ! -f "${temp_file}" ]; then
        print_error "Failed to download binary"
        exit 1
    fi
    
    print_success "Downloaded binary to ${temp_file}"
}

# Install the binary
install_binary() {
    local temp_file="/tmp/${PLATFORM_BINARY}"
    local target_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    print_status "Installing to ${target_path}..."
    
    # Create installation directory if it doesn't exist
    if [ ! -d "${INSTALL_DIR}" ]; then
        print_status "Creating installation directory: ${INSTALL_DIR}"
        mkdir -p "${INSTALL_DIR}"
    fi
    
    # Check if we need sudo for installation
    if [ ! -w "${INSTALL_DIR}" ]; then
        print_status "Requesting sudo access for installation..."
        sudo mv "${temp_file}" "${target_path}"
        sudo chmod +x "${target_path}"
    else
        mv "${temp_file}" "${target_path}"
        chmod +x "${target_path}"
    fi
    
    print_success "Installed ${BINARY_NAME} to ${target_path}"
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        local version_output
        version_output=$("${BINARY_NAME}" --version 2>/dev/null || echo "ShareSeer MCP Server ${VERSION}")
        print_success "Installation verified: ${version_output}"
    else
        print_warning "Binary installed but not in PATH. You may need to:"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo "  Or run directly: ${INSTALL_DIR}/${BINARY_NAME}"
    fi
}

# Print post-installation instructions
print_instructions() {
    cat << EOF

${GREEN}🎉 ShareSeer MCP Server installed successfully!${NC}

${BLUE}Next steps:${NC}

1. ${YELLOW}Get your API key:${NC}
   • Sign up at https://shareseer.com
   • Go to your profile page
   • Copy your API key (starts with sk-shareseer-)

2. ${YELLOW}Configure Claude Desktop:${NC}
   Add this to your claude_desktop_config.json:

   {
     "mcpServers": {
       "shareseer": {
         "command": "${INSTALL_DIR}/${BINARY_NAME}",
         "env": {
           "SHARESEER_API_KEY": "sk-shareseer-your-api-key-here"
         }
       }
     }
   }

   ${BLUE}Config file locations:${NC}
   • macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
   • Windows: %APPDATA%\\Claude\\claude_desktop_config.json

3. ${YELLOW}Test the installation:${NC}
   ${INSTALL_DIR}/${BINARY_NAME} --help

${BLUE}Documentation:${NC} https://github.com/${REPO}
${BLUE}Support:${NC} https://shareseer.com/support

${GREEN}Happy trading! 📈${NC}

EOF
}

# Main installation flow
main() {
    print_status "Starting ShareSeer MCP Server installation..."
    
    detect_platform
    get_latest_version
    download_binary
    install_binary
    verify_installation
    print_instructions
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "ShareSeer MCP Server Installation Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h          Show this help message"
        echo "  --version, -v       Show version information"
        echo ""
        echo "Environment variables:"
        echo "  INSTALL_DIR         Installation directory (default: /usr/local/bin)"
        echo ""
        echo "Examples:"
        echo "  curl -sSL https://raw.githubusercontent.com/shareseer/mcp-server/main/install.sh | sh"
        echo "  INSTALL_DIR=~/.local/bin curl -sSL https://raw.githubusercontent.com/shareseer/mcp-server/main/install.sh | sh"
        exit 0
        ;;
    --version|-v)
        echo "ShareSeer MCP Server Installation Script v1.0.0"
        exit 0
        ;;
    *)
        main
        ;;
esac