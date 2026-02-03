#!/bin/sh
set -e

# Shipyard Installation Script
# Downloads and installs the latest (or specified) version of shipyard

# Configuration
REPO="NatoNathan/shipyard"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"
GITHUB_API="https://api.github.com"
GITHUB_DOWNLOAD="https://github.com/${REPO}/releases/download"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Temporary directory for downloads
TMP_DIR=""

# Cleanup function
cleanup() {
  if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
    rm -rf "$TMP_DIR"
  fi
}

# Register cleanup on exit
trap cleanup EXIT INT TERM

# Print functions
print_info() {
  printf "${BLUE}==>${NC} %s\n" "$1" >&2
}

print_success() {
  printf "${GREEN}==>${NC} %s\n" "$1" >&2
}

print_error() {
  printf "${RED}Error:${NC} %s\n" "$1" >&2
}

print_warning() {
  printf "${YELLOW}Warning:${NC} %s\n" "$1" >&2
}

# Check if command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Detect platform
detect_platform() {
  local os arch

  # Detect OS
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  case "$os" in
    darwin) os="darwin" ;;
    linux) os="linux" ;;
    *)
      print_error "Unsupported operating system: $os"
      print_error "Supported: macOS (darwin), Linux"
      exit 1
      ;;
  esac

  # Detect architecture
  arch=$(uname -m)
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)
      print_error "Unsupported architecture: $arch"
      print_error "Supported: x86_64 (amd64), arm64 (aarch64)"
      exit 1
      ;;
  esac

  echo "${os}_${arch}"
}

# Check required tools
check_requirements() {
  local missing=""

  if ! command_exists curl && ! command_exists wget; then
    missing="${missing}curl or wget\n"
  fi

  if ! command_exists tar; then
    missing="${missing}tar\n"
  fi

  if ! command_exists sha256sum && ! command_exists shasum; then
    missing="${missing}sha256sum or shasum\n"
  fi

  if [ -n "$missing" ]; then
    print_error "Missing required tools:"
    printf "%b" "$missing"
    exit 1
  fi
}

# Download file using curl or wget
download_file() {
  local url="$1"
  local output="$2"

  if command_exists curl; then
    curl -fsSL -o "$output" "$url"
  elif command_exists wget; then
    wget -q -O "$output" "$url"
  else
    print_error "Neither curl nor wget found"
    exit 1
  fi
}

# Get latest version from GitHub API
get_latest_version() {
  local response version

  print_info "Fetching latest version from GitHub..."

  response=$(download_file "${GITHUB_API}/repos/${REPO}/releases/latest" -)

  if [ $? -ne 0 ]; then
    print_error "Failed to fetch latest release information"
    exit 1
  fi

  # Extract tag_name from JSON response
  version=$(echo "$response" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')

  if [ -z "$version" ]; then
    print_error "Could not determine latest version"
    exit 1
  fi

  echo "$version"
}

# Verify checksum
verify_checksum() {
  local file="$1"
  local checksums_file="$2"
  local filename=$(basename "$file")

  print_info "Verifying checksum..."

  # Extract expected checksum for our file
  local expected_sum=$(grep "$filename" "$checksums_file" | awk '{print $1}')

  if [ -z "$expected_sum" ]; then
    print_error "Could not find checksum for $filename"
    return 1
  fi

  # Calculate actual checksum
  local actual_sum
  if command_exists sha256sum; then
    actual_sum=$(sha256sum "$file" | awk '{print $1}')
  elif command_exists shasum; then
    actual_sum=$(shasum -a 256 "$file" | awk '{print $1}')
  else
    print_error "No checksum tool available"
    return 1
  fi

  if [ "$expected_sum" != "$actual_sum" ]; then
    print_error "Checksum verification failed!"
    print_error "Expected: $expected_sum"
    print_error "Got:      $actual_sum"
    return 1
  fi

  print_success "Checksum verified"
  return 0
}

# Download and install
install_shipyard() {
  local version="$1"
  local platform="$2"
  local archive_name="shipyard_${version}_${platform}.tar.gz"
  local download_url="${GITHUB_DOWNLOAD}/${version}/${archive_name}"
  local checksums_url="${GITHUB_DOWNLOAD}/${version}/checksums.txt"

  # Create temporary directory
  TMP_DIR=$(mktemp -d)

  print_info "Downloading shipyard ${version} for ${platform}..."

  # Download archive
  if ! download_file "$download_url" "${TMP_DIR}/${archive_name}"; then
    print_error "Failed to download ${archive_name}"
    print_error "URL: ${download_url}"
    exit 1
  fi

  # Download checksums
  print_info "Downloading checksums..."
  if ! download_file "$checksums_url" "${TMP_DIR}/checksums.txt"; then
    print_error "Failed to download checksums"
    exit 1
  fi

  # Verify checksum
  if ! verify_checksum "${TMP_DIR}/${archive_name}" "${TMP_DIR}/checksums.txt"; then
    exit 1
  fi

  # Extract archive
  print_info "Extracting archive..."
  tar -xzf "${TMP_DIR}/${archive_name}" -C "$TMP_DIR"

  if [ ! -f "${TMP_DIR}/shipyard" ]; then
    print_error "Binary not found in archive"
    exit 1
  fi

  # Install binary
  print_info "Installing to ${INSTALL_DIR}/shipyard..."

  # Check if we need sudo
  if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP_DIR}/shipyard" "${INSTALL_DIR}/shipyard"
    chmod +x "${INSTALL_DIR}/shipyard"
  else
    print_warning "Installation directory requires elevated permissions"
    if command_exists sudo; then
      sudo mv "${TMP_DIR}/shipyard" "${INSTALL_DIR}/shipyard"
      sudo chmod +x "${INSTALL_DIR}/shipyard"
    else
      print_error "Cannot write to ${INSTALL_DIR} and sudo not available"
      exit 1
    fi
  fi

  print_success "Shipyard ${version} installed successfully!"

  # Verify installation
  if command_exists shipyard; then
    print_info "Installed version:"
    shipyard --version
  else
    print_warning "shipyard command not found in PATH"
    print_warning "You may need to add ${INSTALL_DIR} to your PATH"
    print_warning "Or run: export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
}

# Show help
show_help() {
  cat <<EOF
Shipyard Installation Script

Usage: $0 [OPTIONS]

OPTIONS:
  --version VERSION    Install specific version (e.g., v1.2.3)
  --prefix PATH        Install to custom directory (default: /usr/local/bin)
  --help               Show this help message

ENVIRONMENT VARIABLES:
  VERSION              Version to install (default: latest)
  INSTALL_DIR          Installation directory (default: /usr/local/bin)

EXAMPLES:
  # Install latest version
  $0

  # Install specific version
  $0 --version v1.2.3
  VERSION=v1.2.3 $0

  # Install to custom directory
  $0 --prefix ~/.local/bin
  INSTALL_DIR=~/.local/bin $0

  # Quick install via curl
  curl -sSL https://raw.githubusercontent.com/${REPO}/main/install.sh | sh

REQUIREMENTS:
  - curl or wget
  - tar
  - sha256sum or shasum

SUPPORTED PLATFORMS:
  - macOS (Intel, Apple Silicon)
  - Linux (x86_64, ARM64)

For more information, visit: https://github.com/${REPO}
EOF
}

# Parse arguments
parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --version)
        VERSION="$2"
        shift 2
        ;;
      --prefix)
        INSTALL_DIR="$2"
        shift 2
        ;;
      --help|-h)
        show_help
        exit 0
        ;;
      *)
        print_error "Unknown option: $1"
        show_help
        exit 1
        ;;
    esac
  done
}

# Main function
main() {
  # Parse command line arguments
  parse_args "$@"

  # Check requirements
  check_requirements

  # Detect platform
  local platform
  platform=$(detect_platform)
  print_info "Detected platform: ${platform}"

  # Determine version
  local version="$VERSION"
  if [ "$version" = "latest" ]; then
    version=$(get_latest_version)
  fi

  # Ensure version starts with 'v'
  case "$version" in
    v*) ;;
    *) version="v${version}" ;;
  esac

  print_info "Installing version: ${version}"

  # Install
  install_shipyard "$version" "$platform"

  print_success "Installation complete!"
  print_info ""
  print_info "Get started with: shipyard --help"
}

# Run main function
main "$@"
