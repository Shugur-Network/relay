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
RELAY_INFO_URL="http://localhost:8080"

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

echo -e "${BLUE}Starting Shugur Relay NIP-11 Tests${NC}\n"

# Test NIP-11: Relay Information Document
echo -e "\n${YELLOW}Testing NIP-11: Relay Information Document${NC}"

# Test 1: Basic relay information retrieval with correct Accept header
RELAY_INFO=$(curl -s -H "Accept: application/nostr+json" $RELAY_INFO_URL)
if [ ! -z "$RELAY_INFO" ]; then
    print_result "Basic relay information retrieval" true "11"
else
    print_result "Basic relay information retrieval" false "11"
fi

# Test 2: Check for required fields
if [ ! -z "$RELAY_INFO" ]; then
    # Check for name field
    if echo "$RELAY_INFO" | jq -e '.name' >/dev/null 2>&1; then
        print_result "Relay information contains name field" true "11"
    else
        print_result "Relay information contains name field" false "11"
    fi

    # Check for description field
    if echo "$RELAY_INFO" | jq -e '.description' >/dev/null 2>&1; then
        print_result "Relay information contains description field" true "11"
    else
        print_result "Relay information contains description field" false "11"
    fi

    # Check for pubkey field
    if echo "$RELAY_INFO" | jq -e '.pubkey' >/dev/null 2>&1; then
        print_result "Relay information contains pubkey field" true "11"
    else
        print_result "Relay information contains pubkey field" false "11"
    fi

    # Check for contact field
    if echo "$RELAY_INFO" | jq -e '.contact' >/dev/null 2>&1; then
        print_result "Relay information contains contact field" true "11"
    else
        print_result "Relay information contains contact field" false "11"
    fi

    # Check for supported_nips field
    if echo "$RELAY_INFO" | jq -e '.supported_nips' >/dev/null 2>&1; then
        print_result "Relay information contains supported_nips field" true "11"
    else
        print_result "Relay information contains supported_nips field" false "11"
    fi

    # Check for software field
    if echo "$RELAY_INFO" | jq -e '.software' >/dev/null 2>&1; then
        print_result "Relay information contains software field" true "11"
    else
        print_result "Relay information contains software field" false "11"
    fi

    # Check for version field
    if echo "$RELAY_INFO" | jq -e '.version' >/dev/null 2>&1; then
        print_result "Relay information contains version field" true "11"
    else
        print_result "Relay information contains version field" false "11"
    fi
fi

# Test 3: Check for valid JSON format
if echo "$RELAY_INFO" | jq '.' >/dev/null 2>&1; then
    print_result "Relay information is valid JSON" true "11"
else
    print_result "Relay information is valid JSON" false "11"
fi

# Test 4: Check for supported NIPs
if [ ! -z "$RELAY_INFO" ]; then
    # Check if NIP-01 is supported
    if echo "$RELAY_INFO" | jq -e '.supported_nips[] | select(. == 1)' >/dev/null 2>&1; then
        print_result "Relay supports NIP-01" true "11"
    else
        print_result "Relay supports NIP-01" false "11"
    fi

    # Check if NIP-02 is supported
    if echo "$RELAY_INFO" | jq -e '.supported_nips[] | select(. == 2)' >/dev/null 2>&1; then
        print_result "Relay supports NIP-02" true "11"
    else
        print_result "Relay supports NIP-02" false "11"
    fi

    # Check if NIP-04 is supported
    if echo "$RELAY_INFO" | jq -e '.supported_nips[] | select(. == 4)' >/dev/null 2>&1; then
        print_result "Relay supports NIP-04" true "11"
    else
        print_result "Relay supports NIP-04" false "11"
    fi

    # Check if NIP-09 is supported
    if echo "$RELAY_INFO" | jq -e '.supported_nips[] | select(. == 9)' >/dev/null 2>&1; then
        print_result "Relay supports NIP-09" true "11"
    else
        print_result "Relay supports NIP-09" false "11"
    fi
fi

# Test 5: Check for rate limiting information
if [ ! -z "$RELAY_INFO" ]; then
    if echo "$RELAY_INFO" | jq -e '.limitation' >/dev/null 2>&1; then
        print_result "Relay provides rate limiting information" true "11"
    else
        print_result "Relay provides rate limiting information" false "11"
    fi
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