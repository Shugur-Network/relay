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

echo -e "${BLUE}Starting Shugur Relay NIP-04 Tests${NC}\n"

# Test NIP-04: Encrypted Direct Messages
echo -e "\n${YELLOW}Testing NIP-04: Encrypted Direct Messages${NC}"

# Test 1: Create a basic encrypted direct message
DM_EVENT=$(nak event -k 4 -c "$(echo -n "Hello, this is a test message" | base64)" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$DM_EVENT" == *"success"* ]] && [[ ! "$DM_EVENT" == *"failed"* ]]; then
    print_result "Create basic encrypted direct message" true "04"
else
    print_result "Create basic encrypted direct message" false "04"
fi

# Test 2: Create a direct message with multiple recipients
MULTI_DM_EVENT=$(nak event -k 4 -c "$(echo -n "Hello, this is a test message for multiple recipients" | base64)" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81799 $RELAY 2>&1)
if [[ "$MULTI_DM_EVENT" == *"success"* ]] && [[ ! "$MULTI_DM_EVENT" == *"failed"* ]]; then
    print_result "Create direct message with multiple recipients" true "04"
else
    print_result "Create direct message with multiple recipients" false "04"
fi

# Test 3: Create a direct message with empty content
EMPTY_DM_EVENT=$(nak event -k 4 -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$EMPTY_DM_EVENT" == *"failed"* ]] || [[ "$EMPTY_DM_EVENT" == *"content is required"* ]]; then
    print_result "Reject direct message without content" true "04"
else
    print_result "Reject direct message without content" false "04"
fi

# Test 4: Create a direct message with long content
LONG_CONTENT=$(printf 'a%.0s' {1..32}) # At least 32 bytes for AES-256-CBC
LONG_DM_EVENT=$(nak event -k 4 -c "$(echo -n "$LONG_CONTENT" | base64)" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$LONG_DM_EVENT" == *"success"* ]] && [[ ! "$LONG_DM_EVENT" == *"failed"* ]]; then
    print_result "Create direct message with long content" true "04"
else
    print_result "Create direct message with long content" false "04"
fi

# Test 5: Create a direct message without recipient
NO_RECIPIENT_DM_EVENT=$(nak event -k 4 -c "$(echo -n "Hello, this is a test message without recipient" | base64)" $RELAY 2>&1)
if [[ "$NO_RECIPIENT_DM_EVENT" == *"failed"* ]] || [[ "$NO_RECIPIENT_DM_EVENT" == *"must have at least one recipient"* ]]; then
    print_result "Reject direct message without recipient" true "04"
else
    print_result "Reject direct message without recipient" false "04"
fi

# Test 6: Create a direct message with invalid recipient pubkey
INVALID_RECIPIENT_DM_EVENT=$(nak event -k 4 -c "$(echo -n "Hello, this is a test message" | base64)" -t p=invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_RECIPIENT_DM_EVENT" == *"failed"* ]] || [[ "$INVALID_RECIPIENT_DM_EVENT" == *"invalid pubkey"* ]]; then
    print_result "Reject direct message with invalid recipient pubkey" true "04"
else
    print_result "Reject direct message with invalid recipient pubkey" false "04"
fi

# Test 7: Create a direct message with non-base64 content
NON_BASE64_DM_EVENT=$(nak event -k 4 -c "Hello, this is not base64 encoded" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NON_BASE64_DM_EVENT" == *"failed"* ]] || [[ "$NON_BASE64_DM_EVENT" == *"must be base64 encoded"* ]]; then
    print_result "Reject direct message with non-base64 content" true "04"
else
    print_result "Reject direct message with non-base64 content" false "04"
fi

# Test 8: Create a direct message with multiple p tags but same recipient
DUPLICATE_RECIPIENT_DM_EVENT=$(nak event -k 4 -c "$(echo -n "Hello, this is a test message" | base64)" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$DUPLICATE_RECIPIENT_DM_EVENT" == *"success"* ]] && [[ ! "$DUPLICATE_RECIPIENT_DM_EVENT" == *"failed"* ]]; then
    print_result "Create direct message with duplicate recipient" true "04"
else
    print_result "Create direct message with duplicate recipient" false "04"
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