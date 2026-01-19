#!/usr/bin/env bash
# build.sh - Cross-compile pdf-cli for multiple platforms

set -e

# Configuration
BINARY_NAME="pdf"
BUILD_DIR="build"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "none")}"
DATE="${DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Platforms to build for
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Print with color
print_status() {
    echo -e "${GREEN}[BUILD]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create build directory
mkdir -p "${BUILD_DIR}"

# Print build info
echo "========================================"
echo "pdf-cli Build Script"
echo "========================================"
echo "Version: ${VERSION}"
echo "Commit:  ${COMMIT}"
echo "Date:    ${DATE}"
echo "========================================"
echo ""

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "${platform}"

    output_name="${BINARY_NAME}-${GOOS}-${GOARCH}"
    if [ "${GOOS}" = "windows" ]; then
        output_name="${output_name}.exe"
    fi

    print_status "Building ${output_name}..."

    if GO111MODULE=on GOOS="${GOOS}" GOARCH="${GOARCH}" go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/${output_name}" ./cmd/pdf; then
        print_status "Successfully built ${output_name}"

        # Calculate and display file size
        size=$(ls -lh "${BUILD_DIR}/${output_name}" | awk '{print $5}')
        echo "  Size: ${size}"
    else
        print_error "Failed to build ${output_name}"
        exit 1
    fi
done

echo ""
echo "========================================"
echo "Build complete!"
echo "========================================"
echo "Binaries are in: ${BUILD_DIR}/"
ls -lh "${BUILD_DIR}/"
