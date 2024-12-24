#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Default values
GITHUB_REPO="oceanbase/ob-operator"
GITHUB_HOST="https://github.com"
VERSION="0.1.0"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
BINARY_NAME="okctl"
USE_PROXY=false

# Map architecture names
case ${ARCH} in
    x86_64|amd64)
        ARCH=amd64
        ;;
    aarch64|arm64)
        ARCH=arm64
        ;;
    *)
        echo "Unsupported architecture: ${ARCH}"
        exit 1
        ;;
esac

# Help message
usage() {
    cat <<EOF
Usage: $0 [options]

Options:
    -r, --repo        GitHub repository (default: ${GITHUB_REPO})
    -v, --version     Version to install (default: ${VERSION})
    -h, --help        Show this help message
EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case ${key} in
        -r|--repo)
            GITHUB_REPO="$2"
            shift
            shift
            ;;
        -v|--version)
            VERSION="$2"
            shift
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        -p|--proxy)
            USE_PROXY=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Check if required tools are installed
check_requirements() {
    local missing_tools=()
    
    if ! command -v curl >/dev/null 2>&1; then
        missing_tools+=("curl")
    fi
    if ! command -v tar >/dev/null 2>&1; then
        missing_tools+=("tar")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        echo "Error: Required tools are missing: ${missing_tools[*]}"
        echo "Please install them and try again"
        exit 1
    fi
}

# Download and install the binary
install_binary() {
    local version=$1
    local BINARY_NAME="okctl"
    if [ $USE_PROXY = true ]; then
        GITHUB_HOST="https://gh.wewell.org/https://github.com"
    fi
    local download_url="${GITHUB_HOST}/${GITHUB_REPO}/releases/download/cli-${version}/${BINARY_NAME}_${version}_${OS}_${ARCH}.tar.gz"
    
    echo "Downloading ${BINARY_NAME} ${version} from ${download_url} ..."
    if ! curl -L -o "${BINARY_NAME}_${version}.tar.gz" "${download_url}"; then
        echo "Error: Failed to download ${download_url}"
        exit 1
    fi
    
    echo "Extracting to current directory..."
    if ! tar -xzf "${BINARY_NAME}_${version}.tar.gz"; then
        echo "Error: Failed to extract archive"
        exit 1
    fi
    
    # Clean up downloaded archive
    rm -f "${BINARY_NAME}_${version}.tar.gz"
    
    echo "Successfully extracted ${BINARY_NAME}-${version} (./${BINARY_NAME}) to current directory"
    echo "Execute ./${BINARY_NAME} --help to get started"

    echo "Recommend to move the binary to a directory in your PATH, e.g. /usr/local/bin/ like so:"
    echo "sudo mv ./${BINARY_NAME} /usr/local/bin/"
}

main() {
    # Check requirements
    check_requirements
    
    # Install binary
    install_binary "${VERSION}"
}

main

