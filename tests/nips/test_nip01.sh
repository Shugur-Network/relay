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

echo -e "${BLUE}Starting Shugur Relay NIPs Tests${NC}\n"

# Test NIP-01: Basic Protocol Flow Semantics
echo -e "\n${YELLOW}Testing NIP-01: Basic Protocol Flow Semantics${NC}"

# Test 1: Basic event creation and publishing
EVENT=$(nak event -c "Testing NIP-01 basic event" $RELAY)
if [ ! -z "$EVENT" ]; then
    print_result "Basic event creation and publishing" true "01"
else
    print_result "Basic event creation and publishing" false "01"
fi

# Test 2: Event with tags
TAGGED_EVENT=$(nak event -c "Testing NIP-01 event with tags" -t t=test $RELAY)
if [ ! -z "$TAGGED_EVENT" ]; then
    print_result "Event with tags" true "01"
else
    print_result "Event with tags" false "01"
fi

# Test 3: Event with custom kind (using a supported kind)
CUSTOM_EVENT=$(nak event -k 1 -c "Testing NIP-01 custom kind event" $RELAY)
if [ ! -z "$CUSTOM_EVENT" ]; then
    print_result "Custom kind event" true "01"
else
    print_result "Custom kind event" false "01"
fi

# Test 4: Event with created_at timestamp
TIMESTAMP_EVENT=$(nak event -c "Testing NIP-01 event with timestamp" --created-at $(date +%s) $RELAY)
if [ ! -z "$TIMESTAMP_EVENT" ]; then
    print_result "Event with timestamp" true "01"
else
    print_result "Event with timestamp" false "01"
fi

# Test 5: Event with pubkey
PUBKEY_EVENT=$(nak event -c "Testing NIP-01 event with pubkey" -p 79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$PUBKEY_EVENT" ]; then
    print_result "Event with pubkey" true "01"
else
    print_result "Event with pubkey" false "01"
fi

# Test 6: Event with all fields (without signature)
COMPLETE_EVENT=$(nak event -k 1 -c "Testing NIP-01 complete event" -t t=test --created-at $(date +%s) -p 79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$COMPLETE_EVENT" ]; then
    print_result "Event with all fields" true "01"
else
    print_result "Event with all fields" false "01"
fi

# Test 7: Event with invalid pubkey
INVALID_PUBKEY_EVENT=$(nak event -c "Testing NIP-01 invalid pubkey" -p invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_PUBKEY_EVENT" == *"invalid"* ]] || [[ "$INVALID_PUBKEY_EVENT" == *"❌"* ]]; then
    print_result "Event with invalid pubkey" true "01"
else
    print_result "Event with invalid pubkey" false "01"
fi

# Test 8: Event with invalid timestamp
INVALID_TIMESTAMP_EVENT=$(nak event -c "Testing NIP-01 invalid timestamp" --created-at -1 $RELAY 2>&1)
if [[ "$INVALID_TIMESTAMP_EVENT" == *"invalid"* ]] || [[ "$INVALID_TIMESTAMP_EVENT" == *"❌"* ]]; then
    print_result "Event with invalid timestamp" true "01"
else
    print_result "Event with invalid timestamp" false "01"
fi

# Test 9: Event with multiple tags
MULTI_TAG_EVENT=$(nak event -c "Testing NIP-01 event with multiple tags" -t t=test -t t=example $RELAY)
if [ ! -z "$MULTI_TAG_EVENT" ]; then
    print_result "Event with multiple tags" true "01"
else
    print_result "Event with multiple tags" false "01"
fi

# Test 10: Event with empty content
EMPTY_CONTENT_EVENT=$(nak event -c "" $RELAY)
if [ ! -z "$EMPTY_CONTENT_EVENT" ]; then
    print_result "Event with empty content" true "01"
else
    print_result "Event with empty content" false "01"
fi

# Test NIP-01: Filter Tests
echo -e "\n${YELLOW}Testing NIP-01: Filter Semantics${NC}"

# Test 11: Basic filter by kind
KIND_FILTER=$(nak req -k 1 $RELAY)
if [ ! -z "$KIND_FILTER" ]; then
    print_result "Filter by kind" true "01"
else
    print_result "Filter by kind" false "01"
fi

# Test 12: Filter by author
AUTHOR_FILTER=$(nak req -k 1 --author 79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$AUTHOR_FILTER" ]; then
    print_result "Filter by author" true "01"
else
    print_result "Filter by author" false "01"
fi

# Test 13: Filter by time range
TIME_FILTER=$(nak req -k 1 --since $(($(date +%s) - 3600)) --until $(date +%s) $RELAY)
if [ ! -z "$TIME_FILTER" ]; then
    print_result "Filter by time range" true "01"
else
    print_result "Filter by time range" false "01"
fi

# Test 14: Filter by tag
TAG_FILTER=$(nak req -k 1 -t t=test $RELAY)
if [ ! -z "$TAG_FILTER" ]; then
    print_result "Filter by tag" true "01"
else
    print_result "Filter by tag" false "01"
fi

# Test 15: Filter with limit
LIMIT_FILTER=$(nak req -k 1 -l 1 $RELAY)
if [ ! -z "$LIMIT_FILTER" ]; then
    print_result "Filter with limit" true "01"
else
    print_result "Filter with limit" false "01"
fi

# Test 16: Complex filter (multiple conditions)
COMPLEX_FILTER=$(nak req -k 1 --author 79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 -t t=test --since $(($(date +%s) - 3600)) --until $(date +%s) -l 1 $RELAY)
if [ ! -z "$COMPLEX_FILTER" ]; then
    print_result "Complex filter" true "01"
else
    print_result "Complex filter" false "01"
fi

# Test 17: Filter with invalid author
INVALID_AUTHOR_FILTER=$(nak req -k 1 --author invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_AUTHOR_FILTER" == *"invalid"* ]] || [[ "$INVALID_AUTHOR_FILTER" == *"❌"* ]]; then
    print_result "Filter with invalid author" true "01"
else
    print_result "Filter with invalid author" false "01"
fi

# Test 18: Filter with invalid time range (since > until)
INVALID_TIME_FILTER=$(nak req -k 1 --since $(($(date +%s) + 3600)) --until $(date +%s) $RELAY)
if [ -z "$INVALID_TIME_FILTER" ]; then
    print_result "Filter with invalid time range (since > until) returns empty set" true "01"
else
    print_result "Filter with invalid time range (since > until) returns empty set" false "01"
fi

# Test 19: Filter with invalid limit
INVALID_LIMIT_FILTER=$(nak req -k 1 -l -1 $RELAY 2>&1)
if [[ "$INVALID_LIMIT_FILTER" == *"invalid"* ]] || [[ "$INVALID_LIMIT_FILTER" == *"❌"* ]]; then
    print_result "Filter with invalid limit" true "01"
else
    print_result "Filter with invalid limit" false "01"
fi

# Test 20: Filter with multiple authors
MULTI_AUTHOR_FILTER=$(nak req -k 1 --author 79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 --author 82341f882b6eabcd2ba7f1ef90aad961cf074af15b9ef44a09f9d2a8fbfbe6a2 $RELAY)
if [ ! -z "$MULTI_AUTHOR_FILTER" ]; then
    print_result "Filter with multiple authors" true "01"
else
    print_result "Filter with multiple authors" false "01"
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