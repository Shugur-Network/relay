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

echo -e "${BLUE}Starting Shugur Relay NIP-33 Tests${NC}\n"

# Test NIP-33: Addressable Events
echo -e "\n${YELLOW}Testing NIP-33: Addressable Events${NC}"

# Test 1: Create a basic addressable event
PARAM_EVENT=$(nak event -k 30000 -c '{"name": "Test Parameter", "value": "test"}' -t d=test_param -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$PARAM_EVENT" ]; then
    print_result "Create basic addressable event" true "33"
else
    print_result "Create basic addressable event" false "33"
fi

# Test 2: Replace the parameterized event
if [ ! -z "$PARAM_EVENT" ]; then
    REPLACE_EVENT=$(nak event -k 30000 -c '{"name": "Updated Parameter", "value": "updated"}' -t d=test_param -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$REPLACE_EVENT" ]; then
        print_result "Replace parameterized event" true "33"
    else
        print_result "Replace parameterized event" false "33"
    fi
fi

# Test 3: Create a parameterized event with multiple parameters
MULTI_PARAM_EVENT=$(nak event -k 30000 -c '{"name": "Multi Param", "value": "test", "extra": "data"}' -t d=multi_param -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$MULTI_PARAM_EVENT" ]; then
    print_result "Create parameterized event with multiple parameters" true "33"
else
    print_result "Create parameterized event with multiple parameters" false "33"
fi

# Test 4: Create a parameterized event with different kind
DIFF_KIND_EVENT=$(nak event -k 30001 -c '{"name": "Different Kind", "value": "test"}' -t d=diff_kind -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$DIFF_KIND_EVENT" ]; then
    print_result "Create parameterized event with different kind" true "33"
else
    print_result "Create parameterized event with different kind" false "33"
fi

# Test 5: Attempt to create without parameter tag
NO_PARAM_EVENT=$(nak event -k 30000 -c '{"name": "No Parameter", "value": "test"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NO_PARAM_EVENT" == *"missing required 'd' tag"* ]] || [[ "$NO_PARAM_EVENT" == *"failed"* ]]; then
    print_result "Reject event without parameter tag" true "33"
else
    print_result "Reject event without parameter tag" false "33"
fi

# Test 6: Attempt to create with empty parameter tag (should be allowed)
EMPTY_PARAM_EVENT=$(nak event -k 30000 -c '{"name": "Empty Parameter", "value": "test"}' -t d= -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [ ! -z "$EMPTY_PARAM_EVENT" ] && [[ "$EMPTY_PARAM_EVENT" != *"failed"* ]]; then
    print_result "Allow event with empty parameter tag" true "33"
else
    print_result "Allow event with empty parameter tag" false "33"
fi

# Test 7: Attempt to create with invalid kind
INVALID_KIND_EVENT=$(nak event -k 1 -c '{"name": "Invalid Kind", "value": "test"}' -t d=invalid_kind -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_KIND_EVENT" == *"invalid"* ]] || [[ "$INVALID_KIND_EVENT" == *"❌"* ]]; then
    print_result "Reject event with invalid kind" true "33"
else
    print_result "Reject event with invalid kind" false "33"
fi

# Test 8: Attempt to create with invalid recipient
INVALID_RECIPIENT_EVENT=$(nak event -k 30000 -c '{"name": "Invalid Recipient", "value": "test"}' -t d=invalid_recipient -t p=invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_RECIPIENT_EVENT" == *"invalid"* ]] || [[ "$INVALID_RECIPIENT_EVENT" == *"❌"* ]]; then
    print_result "Reject event with invalid recipient" true "33"
else
    print_result "Reject event with invalid recipient" false "33"
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