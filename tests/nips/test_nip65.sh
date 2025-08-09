#!/bin/bash

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
RELAY="ws://localhost:8080"

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
    ((test_count++))
}

echo -e "${BLUE}Starting Shugur Relay NIP-65 Tests${NC}\n"

# Test NIP-65: Relay List Metadata
echo -e "\n${YELLOW}Testing NIP-65: Relay List Metadata${NC}"

# Test 1: Create a basic relay list
RELAY_LIST=$(nak event -k 10002 -c '{"relays": {"wss://relay1.example.com": {"read": true, "write": true}, "wss://relay2.example.com": {"read": true, "write": false}}}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$RELAY_LIST" ]; then
    print_result "Create basic relay list" true "65"
else
    print_result "Create basic relay list" false "65"
fi

# Test 2: Create a relay list with metadata
if [ ! -z "$RELAY_LIST" ]; then
    METADATA_LIST=$(nak event -k 10002 -c '{"relays": {"wss://relay3.example.com": {"read": true, "write": true, "metadata": {"name": "Test Relay", "description": "A test relay"}}}}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$METADATA_LIST" ]; then
        print_result "Create relay list with metadata" true "65"
    else
        print_result "Create relay list with metadata" false "65"
    fi
fi

# Test 3: Create a relay list with multiple relays
MULTI_RELAY_LIST=$(nak event -k 10002 -c '{"relays": {"wss://relay4.example.com": {"read": true, "write": true}, "wss://relay5.example.com": {"read": true, "write": true}, "wss://relay6.example.com": {"read": true, "write": false}}}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$MULTI_RELAY_LIST" ]; then
    print_result "Create relay list with multiple relays" true "65"
else
    print_result "Create relay list with multiple relays" false "65"
fi

# Test 4: Create a relay list with empty relays
EMPTY_LIST=$(nak event -k 10002 -c '{"relays": {}}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$EMPTY_LIST" ]; then
    print_result "Create relay list with empty relays" true "65"
else
    print_result "Create relay list with empty relays" false "65"
fi

# Test 5: Attempt to create without relays field
NO_RELAYS=$(nak event -k 10002 -c '{"invalid": "data"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NO_RELAYS" == *"invalid"* ]] || [[ "$NO_RELAYS" == *"❌"* ]]; then
    print_result "Reject relay list without relays field" true "65"
else
    print_result "Reject relay list without relays field" false "65"
fi

# Test 6: Attempt to create with invalid relay URL
INVALID_URL=$(nak event -k 10002 -c '{"relays": {"invalid_url": {"read": true, "write": true}}}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_URL" == *"invalid"* ]] || [[ "$INVALID_URL" == *"❌"* ]]; then
    print_result "Reject relay list with invalid relay URL" true "65"
else
    print_result "Reject relay list with invalid relay URL" false "65"
fi

# Test 7: Attempt to create with invalid permissions
INVALID_PERMS=$(nak event -k 10002 -c '{"relays": {"wss://relay7.example.com": {"read": "invalid", "write": true}}}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_PERMS" == *"invalid"* ]] || [[ "$INVALID_PERMS" == *"❌"* ]]; then
    print_result "Reject relay list with invalid permissions" true "65"
else
    print_result "Reject relay list with invalid permissions" false "65"
fi

# Test 8: Attempt to create with invalid recipient
INVALID_RECIPIENT=$(nak event -k 10002 -c '{"relays": {"wss://relay8.example.com": {"read": true, "write": true}}}' -t p=invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_RECIPIENT" == *"invalid"* ]] || [[ "$INVALID_RECIPIENT" == *"❌"* ]]; then
    print_result "Reject relay list with invalid recipient" true "65"
else
    print_result "Reject relay list with invalid recipient" false "65"
fi

# Print summary
echo -e "\n${BLUE}Test Summary:${NC}"
echo -e "Total tests: $test_count"
echo -e "${GREEN}Successful: $success_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"

# Exit with error if any tests failed
if [ $fail_count -gt 0 ]; then
    exit 1
else
    exit 0
fi 