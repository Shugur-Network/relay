#!/usr/bin/env bash
#
# Build script for Shugur Relay
# - Builds current platform by default (keeps exact original behavior)
# - Adds: arm64 targets, static builds, build tags, checksums, archives, nicer UX
#
# Usage examples:
#   ./build.sh                          # Build for current platform
#   ./build.sh --clean                  # Clean then build current platform
#   ./build.sh --race                   # Race build (dev-ish)
#   ./build.sh --dev                    # Development build (keeps debug info; no -ldflags)
#   ./build.sh --all                    # Build Linux/macOS/Windows (amd64+arm64)
#   ./build.sh --linux                  # Linux (amd64,arm64)
#   ./build.sh --darwin                 # macOS (amd64,arm64)
#   ./build.sh --windows                # Windows (amd64,arm64)
#   ./build.sh --tags "cockroach"       # Pass Go build tags
#   ./build.sh --static                 # Static-ish build (CGO_DISABLED=1, adds -trimpath)
#   ./build.sh --checksum               # Generate SHA256SUMS.txt for produced binaries
#   ./build.sh --archive                # Create .tar.gz (unix) / .zip (windows) artifacts
#   ./build.sh --name relayd            # Override output binary name
#   ./build.sh --main ./cmd/relay       # Override main package path
#
# Notes:
# - Keep ONLY linker flags in LDFLAGS (no '-ldflags' keyword here).
# - For production/minified builds we strip symbols with -w -s.

set -Eeuo pipefail

########################
# Configuration
########################
BINARY_NAME="${BINARY_NAME:-relay}"
BIN_DIR="${BIN_DIR:-./bin}"
MAIN_PATH="${MAIN_PATH:-./cmd}"
BUILD_FLAGS="${BUILD_FLAGS:--v}"

# IMPORTANT: Put ONLY the linker flags here (no '-ldflags' keyword)
LDFLAGS="${LDFLAGS:--w -s}"

# Defaults
DEFAULT_OS_ARCH_MATRIX=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

# Colors
RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
BLUE=$'\033[0;34m'
NC=$'\033[0m'

########################
# Pretty printing
########################
print_info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $*"; }
print_error()   { echo -e "${RED}[ERROR]${NC} $*"; }

trap 'print_error "Failed at line $LINENO"; exit 1' ERR

########################
# Helpers
########################
show_usage() {
  cat <<EOF
Usage: $0 [OPTIONS]

Options:
  -h, --help            Show this help message
  -c, --clean           Clean before building
  -r, --race            Build with race detection (forces CGO on supported OS/ARCH)
  -d, --dev             Development build (keeps debug info; skips -ldflags)
  -a, --all             Build for all platforms (linux/darwin/windows; amd64+arm64)
  --linux               Build for Linux (amd64+arm64)
  --darwin              Build for macOS (amd64+arm64)
  --windows             Build for Windows (amd64+arm64)
  --name NAME           Override output binary base name (default: ${BINARY_NAME})
  --main PATH           Override main package path to build (default: ${MAIN_PATH})
  --tags "TAGS"         Add Go build tags (e.g., "cockroach,netgo")
  --static              Static-ish build: sets CGO_ENABLED=0 and adds -trimpath
  --checksum            Generate SHA256SUMS.txt for produced binaries
  --archive             Create per-target archives (.tar.gz/.zip)

Examples:
  $0                      # Build for current platform
  $0 --clean              # Clean and build
  $0 --race               # Build with race detection
  $0 --all --checksum     # Cross-compile and write checksums
  $0 --darwin --archive   # macOS builds, packaged as archives
EOF
}

abspath() {
  # portable-ish absolute path for files/dirs that may not yet exist
  python3 - <<'PY' "$1"
import os, sys
p = sys.argv[1]
print(os.path.abspath(p))
PY
}

clean_build() {
  print_info "Cleaning build artifacts..."
  rm -rf "$BIN_DIR"
  go clean
  print_success "Clean completed"
}

create_bin_dir() {
  mkdir -p "$BIN_DIR"
}

# Compose the go build command safely and run it
# Args:
#   $1 -> output path
#   (env GOOS/GOARCH/CGO_ENABLED may be set by caller)
run_build() {
  local output="$1"
  local -a cmd=(go build)

  # Base build flags
  if [[ -n "${BUILD_FLAGS}" ]]; then
    # shellcheck disable=SC2206
    cmd+=(${BUILD_FLAGS})
  fi

  # Build tags
  if [[ -n "${BUILD_TAGS}" ]]; then
    cmd+=(-tags "${BUILD_TAGS}")
  fi

  # Race vs dev vs release
  if [[ "${RACE_DETECTION}" == "true" ]]; then
    cmd+=(-race)
  fi

  # Trimpath (applies to static builds and release by default)
  if [[ "${STATIC_BUILD}" == "true" ]]; then
    cmd+=(-trimpath)
  fi

  # Only pass -ldflags for non-dev builds *and* when LDFLAGS is set
  if [[ "${DEV_BUILD}" != "true" && -n "${LDFLAGS}" ]]; then
    cmd+=(-ldflags "${LDFLAGS}")
  fi

  cmd+=(-o "$output" "$MAIN_PATH")

  print_info "Running: ${cmd[*]}"
  "${cmd[@]}"
}

# Build for current platform
build_current() {
  local output_ext=""
  local goos="$(go env GOOS)"
  local goarch="$(go env GOARCH)"
  [[ "$goos" == "windows" ]] && output_ext=".exe"

  local output="${BIN_DIR}/${BINARY_NAME}${output_ext}"

  if [[ "${RACE_DETECTION}" == "true" ]]; then
    print_info "Building ${BINARY_NAME} with race detection for ${goos}/${goarch}..."
  elif [[ "${DEV_BUILD}" == "true" ]]; then
    print_info "Building ${BINARY_NAME} (development; no -ldflags) for ${goos}/${goarch}..."
  else
    print_info "Building ${BINARY_NAME} (release; -ldflags='${LDFLAGS}') for ${goos}/${goarch}..."
  fi

  create_bin_dir

  local -a envvars=()
  if [[ "${STATIC_BUILD}" == "true" ]]; then
    envvars+=(CGO_ENABLED=0)
  fi
  if [[ "${RACE_DETECTION}" == "true" ]]; then
    envvars+=(CGO_ENABLED=1)
  fi

  if "${envvars[@]:-true}" run_build "$output"; then
    print_success "Build completed: $output"
    ARTIFACTS+=("$output")
  else
    print_error "Build failed"
    exit 1
  fi
}

# Build for a specific platform
# Args: goos goarch
build_platform() {
  local goos="$1"
  local goarch="$2"
  local suffix=""; [[ "$goos" == "windows" ]] && suffix=".exe"
  local output="${BIN_DIR}/${BINARY_NAME}-${goos}-${goarch}${suffix}"

  print_info "Building for ${goos}/${goarch}..."
  create_bin_dir

  local -a envvars=(GOOS="$goos" GOARCH="$goarch")
  if [[ "${STATIC_BUILD}" == "true" ]]; then
    envvars+=(CGO_ENABLED=0)
  fi
  if [[ "${RACE_DETECTION}" == "true" ]]; then
    # race only supported on certain combos; attempt but warn on likely unsupported ones
    case "${goos}/${goarch}" in
      linux/amd64|linux/arm64|darwin/amd64|darwin/arm64|windows/amd64|windows/arm64)
        envvars+=(CGO_ENABLED=1)
        ;;
      *)
        print_warning "Race detector may not be supported on ${goos}/${goarch}; attempting anyway."
        envvars+=(CGO_ENABLED=1)
        ;;
    esac
  fi

  if "${envvars[@]}" run_build "$output"; then
    print_success "Build completed: $output"
    ARTIFACTS+=("$output")
  else
    print_error "Build failed for ${goos}/${goarch}"
    return 1
  fi
}

# Build for multiple platforms
build_matrix() {
  local -a matrix=("$@")
  print_info "Building matrix: ${matrix[*]}"

  for pair in "${matrix[@]}"; do
    IFS='/' read -r goos goarch <<<"$pair"
    build_platform "$goos" "$goarch"
  done

  print_success "Matrix builds completed"
}

# Generate checksums
write_checksums() {
  [[ "${#ARTIFACTS[@]}" -eq 0 ]] && { print_warning "No artifacts to checksum"; return 0; }
  local outfile="${BIN_DIR}/SHA256SUMS.txt"
  print_info "Writing checksums to ${outfile}"
  : > "$outfile"
  for f in "${ARTIFACTS[@]}"; do
    if command -v shasum >/dev/null 2>&1; then
      shasum -a 256 "$(abspath "$f")" >> "$outfile"
    elif command -v sha256sum >/dev/null 2>&1; then
      (cd "$(dirname "$f")" && sha256sum "$(basename "$f")") >> "$outfile"
    else
      print_warning "No sha256 tool found; skipped checksums for $f"
    fi
  done
  print_success "Checksums written: ${outfile}"
}

# Archive artifacts per target
archive_artifacts() {
  [[ "${#ARTIFACTS[@]}" -eq 0 ]] && { print_warning "No artifacts to archive"; return 0; }

  print_info "Creating archives for artifacts..."
  for f in "${ARTIFACTS[@]}"; do
    local base="$(basename "$f")"
    local dir="$(dirname "$f")"

    if [[ "$base" == *"-windows-"*".exe" ]]; then
      local zip="${dir}/${base%.exe}.zip"
      if command -v zip >/dev/null 2>&1; then
        (cd "$dir" && zip -q -9 "$(basename "$zip")" "$base")
        print_success "Archive created: $zip"
      else
        print_warning "zip not found; skipping archive for $base"
      fi
    else
      local tgz="${dir}/${base}.tar.gz"
      (cd "$dir" && tar -czf "$(basename "$tgz")" "$base")
      print_success "Archive created: $tgz"
    fi
  done
}

########################
# Parse args
########################
CLEAN_BEFORE_BUILD=false
RACE_DETECTION=false
DEV_BUILD=false
BUILD_ALL=false
BUILD_LINUX=false
BUILD_DARWIN=false
BUILD_WINDOWS=false
STATIC_BUILD=false
DO_CHECKSUM=false
DO_ARCHIVE=false
BUILD_TAGS=""
ARTIFACTS=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) show_usage; exit 0 ;;
    -c|--clean) CLEAN_BEFORE_BUILD=true; shift ;;
    -r|--race) RACE_DETECTION=true; shift ;;
    -d|--dev) DEV_BUILD=true; shift ;;
    -a|--all) BUILD_ALL=true; shift ;;
    --linux) BUILD_LINUX=true; shift ;;
    --darwin) BUILD_DARWIN=true; shift ;;
    --windows) BUILD_WINDOWS=true; shift ;;
    --static) STATIC_BUILD=true; shift ;;
    --checksum) DO_CHECKSUM=true; shift ;;
    --archive) DO_ARCHIVE=true; shift ;;
    --tags) BUILD_TAGS="$2"; shift 2 ;;
    --name) BINARY_NAME="$2"; shift 2 ;;
    --main) MAIN_PATH="$2"; shift 2 ;;
    *)
      print_error "Unknown option: $1"
      show_usage
      exit 1
      ;;
  esac
done

########################
# Main
########################
print_info "Starting build process..."
START_TS=$(date +%s)

# Clean if requested
if [[ "${CLEAN_BEFORE_BUILD}" == "true" ]]; then
  clean_build
fi

# Check Go
if ! command -v go >/dev/null 2>&1; then
  print_error "Go is not installed or not in PATH"
  exit 1
fi

# Verify module
if [[ ! -f "go.mod" ]]; then
  print_error "go.mod not found. Are you in the project root directory?"
  exit 1
fi

# Decide what to build
if [[ "${BUILD_ALL}" == "true" ]]; then
  build_matrix "${DEFAULT_OS_ARCH_MATRIX[@]}"
elif [[ "${BUILD_LINUX}" == "true" ]]; then
  build_matrix "linux/amd64" "linux/arm64"
elif [[ "${BUILD_DARWIN}" == "true" ]]; then
  build_matrix "darwin/amd64" "darwin/arm64"
elif [[ "${BUILD_WINDOWS}" == "true" ]]; then
  build_matrix "windows/amd64" "windows/arm64"
else
  build_current
fi

# Post-processing
[[ "${DO_ARCHIVE}" == "true" ]] && archive_artifacts
[[ "${DO_CHECKSUM}" == "true" ]] && write_checksums

END_TS=$(date +%s)
print_success "Build process completed successfully in $((END_TS-START_TS))s!"
