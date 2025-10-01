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
RELAY="ws://localhost:8081"
# RELAY="wss://shu02.shugur.net"

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

echo -e "${BLUE}Starting Shugur Relay NIP-02 Tests${NC}\n"

# Test NIP-02: Contact List
echo -e "\n${YELLOW}Testing NIP-02: Contact List${NC}"

# Test 1: Create a contact list event with basic contacts
CONTACT_LIST=$(nak event -k 3 -p "02f7f44f4d3d0c0a0d0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f" -p "03f7f44f4d3d0c0a0d0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f" $RELAY)
if [ ! -z "$CONTACT_LIST" ]; then
    print_result "Create contact list event with basic contacts" true "02"
else
    print_result "Create contact list event with basic contacts" false "02"
fi

# Test 2: Create contact list with multiple contacts
MULTI_CONTACT_LIST=$(nak event -k 3 -p "02f7f44f4d3d0c0a0d0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f" -p "03f7f44f4d3d0c0a0d0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f" -p "04f7f44f4d3d0c0a0d0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f0a0b0c0d0e0f" $RELAY)
if [ ! -z "$MULTI_CONTACT_LIST" ]; then
    print_result "Create contact list with multiple contacts" true "02"
else
    print_result "Create contact list with multiple contacts" false "02"
fi

# Test 3: Create contact list with empty content
EMPTY_CONTACT_LIST=$(nak event -k 3 $RELAY)
if [ ! -z "$EMPTY_CONTACT_LIST" ]; then
    print_result "Create empty contact list" true "02"
else
    print_result "Create empty contact list" false "02"
fi

# Test 4: Create contact list with invalid pubkey
INVALID_PUBKEY_CONTACT_LIST=$(nak event -k 3 -p "invalid-pubkey" $RELAY 2>&1)
if [[ "$INVALID_PUBKEY_CONTACT_LIST" == *"invalid"* ]] || [[ "$INVALID_PUBKEY_CONTACT_LIST" == *"❌"* ]]; then
    print_result "Reject contact list with invalid pubkey" true "02"
else
    print_result "Reject contact list with invalid pubkey" false "02"
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