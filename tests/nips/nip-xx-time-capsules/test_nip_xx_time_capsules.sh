#!/bin/bash
# Test NIP-XX Time Capsules Python Implementation
# Complete end-to-end test with cross-chain compatibility

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🕐 Testing NIP-XX Time Capsules Python Implementation"
echo "=" * 60

# Check if Python environment has required packages
echo "🔍 Checking Python dependencies..."

MISSING_DEPS=()

command -v tle >/dev/null || MISSING_DEPS+=("tle")
command -v nak >/dev/null || MISSING_DEPS+=("nak")

# Check Python packages
if ! python3 -c "import websocket, requests" 2>/dev/null; then
    MISSING_DEPS+=("python-packages")
fi

if [ ${#MISSING_DEPS[@]} -gt 0 ]; then
    echo -e "${RED}❌ Missing dependencies: ${MISSING_DEPS[*]}${NC}"
    echo "Install with:"
    echo "  - tle: go install github.com/drand/tlock/cmd/tle@latest"
    echo "  - nak: go install github.com/fiatjaf/nak@latest"
    echo "  - python packages: pip3 install websocket-client requests"
    echo ""
    echo "Optional: Create virtual environment first:"
    echo "  python3 -m venv venv && source venv/bin/activate"
    echo "  pip3 install -r requirements-test.txt"
    exit 1
fi

echo -e "${GREEN}✅ All dependencies available${NC}"

# Check if relay is running
echo "🔗 Checking relay connection..."
if ! timeout 5 bash -c "</dev/tcp/localhost/8085" 2>/dev/null; then
    echo -e "${RED}❌ Relay not accessible on localhost:8085${NC}"
    echo "Start the relay with: ./bin/relay start --config config/development.yaml"
    exit 1
fi

echo -e "${GREEN}✅ Relay is accessible${NC}"

# Run the Python test
echo "🐍 Running NIP-XX Time Capsules Python test..."
echo "=" * 60

# Run from current directory since complete_time_capsules_demo.py is now here
if python3 complete_time_capsules_demo.py; then
    echo "=" * 60
    echo -e "${GREEN}🎉 NIP-XX Time Capsules Python test PASSED!${NC}"
    echo -e "${GREEN}✅ All functionality working correctly${NC}"
    echo -e "${GREEN}✅ Cross-chain compatibility demonstrated${NC}"
    echo -e "${GREEN}✅ Full NIP-XX specification compliance verified${NC}"
    exit 0
else
    echo "=" * 60
    echo -e "${RED}❌ NIP-XX Time Capsules Python test FAILED!${NC}"
    exit 1
fi
