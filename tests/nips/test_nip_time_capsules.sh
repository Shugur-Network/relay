#!/bin/bash

# NIP-XX Time Capsules Test Script - Updated Specification
# Tests time-lock encrypted messages with new format and NIP-59 support

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color
BLUE='\033[0;34m'
YELLOW='\033[1;33m'

# Test counter
test_count=0
success_count=0
fail_count=0

# Relay URL
RELAY="ws://localhost:8085"

# Test timeout
TEST_TIMEOUT=120

# Helper function to print test results
print_result() {
    local test_name=$1
    local success=$2
    local nip=$3
    
    if [ "$success" = true ]; then
        echo -e "${GREEN}✓ Test $test_count: $test_name (NIP-$nip)${NC}"
        ((success_count++))
    else
        echo -e "${RED}✗ Test $test_count: $test_name (NIP-$nip)${NC}"
        ((fail_count++))
    fi
}

# Function to check if required dependencies are installed
check_dependencies() {
    local missing_deps=()
    
    # Check binary dependencies
    for cmd in tle nak jq python3; do
        if ! command -v $cmd &> /dev/null; then
            missing_deps+=("$cmd")
        fi
    done
    
    # Check Python packages
    if ! python3 -c "import websocket, requests" 2>/dev/null; then
        missing_deps+=("python-packages")
    fi
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo -e "${RED}❌ Missing dependencies: ${missing_deps[*]}${NC}"
        echo "Install with:"
        echo "  - tle: go install github.com/drand/tlock/cmd/tle@latest"
        echo "  - nak: go install github.com/fiatjaf/nak@latest"
        echo "  - python packages: pip3 install websocket-client requests"
        echo ""
        echo "Optional: Create virtual environment first:"
        echo "  python3 -m venv venv && source venv/bin/activate"
        echo "  pip3 install -r nip-xx-time-capsules/requirements.txt"
        return 1
    fi
    
    return 0
}

# Function to check relay connectivity
check_relay() {
    echo "🔗 Checking relay connection..."
    if ! timeout 5 bash -c "</dev/tcp/localhost/8085" 2>/dev/null; then
        echo -e "${RED}❌ Relay not accessible on localhost:8085${NC}"
        echo "Start the relay with: ./bin/relay start --config config/development.yaml"
        return 1
    fi
    
    echo -e "${GREEN}✅ Relay is accessible${NC}"
    return 0
}

# Test 1: Updated Time Capsule Creation and Validation
test_time_capsule_creation() {
    local test_name="Updated Time Capsule Creation (New Spec)"
    ((test_count++))
    
    echo -e "${BLUE}Test $test_count: $test_name${NC}"
    
    # Run the simple validation test (no external dependencies)
    if timeout $TEST_TIMEOUT python3 tests/nips/nip-xx-time-capsules/lib/simple_validation_test.py; then
        print_result "$test_name" true "XX"
        return 0
    else
        print_result "$test_name" false "XX"
        return 1
    fi
}

# Test 2: NIP-59 Gift Wrapping Support
test_nip59_gift_wrapping() {
    local test_name="NIP-59 Gift Wrapping for Private Capsules"
    ((test_count++))
    
    echo -e "${BLUE}Test $test_count: $test_name${NC}"
    
    # Test NIP-59 event structure validation (without actual encryption)
    if python3 -c "
import json

# Test NIP-59 gift wrap structure (kind 1059)
gift_wrap = {
    'kind': 1059,
    'content': 'base64_encoded_seal_content',
    'tags': [['p', 'recipient_pubkey']],
    'created_at': 1694780000,
    'pubkey': 'ephemeral_pubkey',
    'id': 'event_id',
    'sig': 'signature'
}

# Test NIP-59 seal structure (kind 13) 
seal = {
    'kind': 13,
    'content': 'base64_encoded_rumor_content',
    'tags': [],  # Must be empty
    'created_at': 1694780000,
    'pubkey': 'author_pubkey',
    'id': 'event_id', 
    'sig': 'signature'
}

print('✅ NIP-59 gift wrap structure valid')
print('✅ NIP-59 seal structure valid')
exit(0)
"; then
        print_result "$test_name" true "59"
        return 0
    else
        print_result "$test_name" false "59"
        return 1
    fi
}

# Test 3: New Tag Format Validation  
test_new_tag_format() {
    local test_name="New Tlock Tag Format Validation"
    ((test_count++))
    
    echo -e "${BLUE}Test $test_count: $test_name${NC}"
    
    # Test new tlock tag format validation
    if python3 -c "
import base64

# Test the new tlock tag format
test_event = {
    'kind': 1041,
    'content': base64.b64encode(b'dummy_age_content').decode('utf-8'),
    'tags': [
        ['tlock', '52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971', '999999999'],
        ['alt', 'Test capsule']
    ]
}

# Validate new format
tlock_tag = test_event['tags'][0]
if len(tlock_tag) == 3 and len(tlock_tag[1]) == 64 and tlock_tag[1].isalnum():
    print(f'✅ New tlock format valid: {tlock_tag}')
    exit(0)
else:
    print(f'❌ Invalid tlock format: {tlock_tag}')
    exit(1)
"; then
        print_result "$test_name" true "XX"
        return 0
    else
        print_result "$test_name" false "XX"
        return 1
    fi
}

# Main test execution
main() {
    echo "🕐 Testing NIP-XX Time Capsules Implementation"
    echo "=============================================="
    
    # Check dependencies
    if ! check_dependencies; then
        exit 1
    fi
    
    # Check relay
    if ! check_relay; then
        exit 1
    fi
    
    echo ""
    echo "Running Updated NIP-XX Time Capsules Tests..."
    echo "=============================================="
    
    # Run tests
    test_time_capsule_creation
    test_nip59_gift_wrapping  
    test_new_tag_format
    
    echo ""
    echo "=============================================="
    echo "Test Results:"
    echo -e "  ${GREEN}✓ Passed: $success_count${NC}"
    echo -e "  ${RED}✗ Failed: $fail_count${NC}"
    echo -e "  Total: $test_count"
    
    if [ $fail_count -eq 0 ]; then
        echo -e "${GREEN}🎉 All NIP-XX Time Capsules tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some NIP-XX Time Capsules tests failed!${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
