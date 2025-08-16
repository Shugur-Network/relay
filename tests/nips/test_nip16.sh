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
# RELAY="ws://localhost:8080"
RELAY="wss://shu02.shugur.net"

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

echo -e "${BLUE}Starting Shugur Relay NIP-16 Tests${NC}\n"

# Test NIP-16: Event Treatment
echo -e "\n${YELLOW}Testing NIP-16: Event Treatment${NC}"

# Test 1: Create a replaceable event (kind 0 - metadata)
REPLACEABLE_EVENT=$(nak event -k 0 -c '{"name": "Test User", "about": "Test user description"}' $RELAY)
if [ ! -z "$REPLACEABLE_EVENT" ]; then
    print_result "Create replaceable event (kind 0)" true "16"
else
    print_result "Create replaceable event (kind 0)" false "16"
fi

# Test 2: Replace the replaceable event
if [ ! -z "$REPLACEABLE_EVENT" ]; then
    REPLACED_EVENT=$(nak event -k 0 -c '{"name": "Updated User", "about": "Updated description"}' $RELAY)
    if [ ! -z "$REPLACED_EVENT" ]; then
        print_result "Replace existing replaceable event" true "16"
    else
        print_result "Replace existing replaceable event" false "16"
    fi
fi

# Test 3: Create an ephemeral event (kind 20000)
EPHEMERAL_EVENT=$(nak event -k 20000 -c "This is an ephemeral event" $RELAY)
if [ ! -z "$EPHEMERAL_EVENT" ]; then
    print_result "Create ephemeral event (kind 20000)" true "16"
else
    print_result "Create ephemeral event (kind 20000)" false "16"
fi

# Test 4: Create another ephemeral event (kind 20001)
ANOTHER_EPHEMERAL=$(nak event -k 20001 -c "This is another ephemeral event" $RELAY)
if [ ! -z "$ANOTHER_EPHEMERAL" ]; then
    print_result "Create another ephemeral event (kind 20001)" true "16"
else
    print_result "Create another ephemeral event (kind 20001)" false "16"
fi

# Test 5: Create a non-replaceable event (kind 1)
NON_REPLACEABLE=$(nak event -k 1 -c "This is a non-replaceable event" $RELAY)
if [ ! -z "$NON_REPLACEABLE" ]; then
    print_result "Create non-replaceable event (kind 1)" true "16"
else
    print_result "Create non-replaceable event (kind 1)" false "16"
fi

# Test 6: Create a non-ephemeral event (kind 30000)
NON_EPHEMERAL=$(nak event -k 30000 -c "This is a non-ephemeral event" -t d=test-event $RELAY)
if [ ! -z "$NON_EPHEMERAL" ]; then
    print_result "Create non-ephemeral event (kind 30000)" true "16"
else
    print_result "Create non-ephemeral event (kind 30000)" false "16"
fi

# Test 7: Create a replaceable event with tags (kind 3)
REPLACEABLE_WITH_TAGS=$(nak event -k 3 -c '{"wss://relay1.example.com": {"read": true, "write": true}}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$REPLACEABLE_WITH_TAGS" ]; then
    print_result "Create replaceable event with tags (kind 3)" true "16"
else
    print_result "Create replaceable event with tags (kind 3)" false "16"
fi

# Test 8: Create an ephemeral event with tags
EPHEMERAL_WITH_TAGS=$(nak event -k 20000 -c "This is an ephemeral event with tags" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$EPHEMERAL_WITH_TAGS" ]; then
    print_result "Create ephemeral event with tags" true "16"
else
    print_result "Create ephemeral event with tags" false "16"
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