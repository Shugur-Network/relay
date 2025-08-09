#!/bin/bash

# Build script for Shugur Relay
# This script builds the relay binary and places it in the bin directory

set -e

# Configuration
BINARY_NAME="relay"
BIN_DIR="./bin"
MAIN_PATH="./cmd"
BUILD_FLAGS="-v"
LDFLAGS="-ldflags \"-w -s\""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
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

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help message"
    echo "  -c, --clean         Clean before building"
    echo "  -r, --race          Build with race detection"
    echo "  -d, --dev           Build development version (with debug info)"
    echo "  -a, --all           Build for all platforms"
    echo "  --linux             Build for Linux only"
    echo "  --darwin            Build for macOS only"
    echo "  --windows           Build for Windows only"
    echo ""
    echo "Examples:"
    echo "  $0                  # Build for current platform"
    echo "  $0 --clean          # Clean and build"
    echo "  $0 --race           # Build with race detection"
    echo "  $0 --all            # Build for all platforms"
}

# Function to clean build artifacts
clean_build() {
    print_info "Cleaning build artifacts..."
    rm -rf "$BIN_DIR"
    go clean
    print_success "Clean completed"
}

# Function to create bin directory
create_bin_dir() {
    mkdir -p "$BIN_DIR"
}

# Function to build for current platform
build_current() {
    local flags="$BUILD_FLAGS"
    local output="$BIN_DIR/$BINARY_NAME"
    
    if [ "$RACE_DETECTION" = "true" ]; then
        flags="$flags -race"
        print_info "Building $BINARY_NAME with race detection..."
    elif [ "$DEV_BUILD" = "true" ]; then
        print_info "Building $BINARY_NAME (development version)..."
    else
        flags="$flags $LDFLAGS"
        print_info "Building $BINARY_NAME..."
    fi
    
    create_bin_dir
    
    if go build $flags -o "$output" "$MAIN_PATH"; then
        print_success "Build completed: $output"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Function to build for specific platform
build_platform() {
    local goos=$1
    local goarch=$2
    local suffix=$3
    
    print_info "Building for $goos/$goarch..."
    create_bin_dir
    
    local output="$BIN_DIR/$BINARY_NAME-$goos-$goarch$suffix"
    
    if GOOS=$goos GOARCH=$goarch go build $BUILD_FLAGS $LDFLAGS -o "$output" "$MAIN_PATH"; then
        print_success "Build completed: $output"
    else
        print_error "Build failed for $goos/$goarch"
        return 1
    fi
}

# Function to build for all platforms
build_all() {
    print_info "Building for all platforms..."
    
    # Linux
    build_platform "linux" "amd64" ""
    
    # macOS
    build_platform "darwin" "amd64" ""
    build_platform "darwin" "arm64" ""
    
    # Windows
    build_platform "windows" "amd64" ".exe"
    
    print_success "All builds completed"
}

# Parse command line arguments
CLEAN_BEFORE_BUILD=false
RACE_DETECTION=false
DEV_BUILD=false
BUILD_ALL=false
BUILD_LINUX=false
BUILD_DARWIN=false
BUILD_WINDOWS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -c|--clean)
            CLEAN_BEFORE_BUILD=true
            shift
            ;;
        -r|--race)
            RACE_DETECTION=true
            shift
            ;;
        -d|--dev)
            DEV_BUILD=true
            shift
            ;;
        -a|--all)
            BUILD_ALL=true
            shift
            ;;
        --linux)
            BUILD_LINUX=true
            shift
            ;;
        --darwin)
            BUILD_DARWIN=true
            shift
            ;;
        --windows)
            BUILD_WINDOWS=true
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
print_info "Starting build process..."

# Clean if requested
if [ "$CLEAN_BEFORE_BUILD" = "true" ]; then
    clean_build
fi

# Check Go installation
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Verify we're in a Go module
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Are you in the project root directory?"
    exit 1
fi

# Execute build based on options
if [ "$BUILD_ALL" = "true" ]; then
    build_all
elif [ "$BUILD_LINUX" = "true" ]; then
    build_platform "linux" "amd64" ""
elif [ "$BUILD_DARWIN" = "true" ]; then
    build_platform "darwin" "amd64" ""
    build_platform "darwin" "arm64" ""
elif [ "$BUILD_WINDOWS" = "true" ]; then
    build_platform "windows" "amd64" ".exe"
else
    build_current
fi

print_success "Build process completed successfully!"
