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

echo -e "${BLUE}Starting Shugur Relay NIP-03 Tests${NC}\n"

# Test NIP-03: OpenTimestamps Attestation
echo -e "\n${YELLOW}Testing NIP-03: OpenTimestamps Attestation${NC}"

# First create a regular event to attest
BASE_EVENT=$(nak event -c "This event will be attested with OpenTimestamps" $RELAY)
BASE_EVENT_ID=$(echo "$BASE_EVENT" | jq -r '.id')

if [ -z "$BASE_EVENT_ID" ]; then
    echo -e "${RED}Failed to create base event for testing${NC}"
    exit 1
fi

# Test 1: Valid OpenTimestamps attestation
# Create a base64-encoded OTS file (this is a minimal valid OTS file)
MINIMAL_OTS="AA=="
OTS_ATTESTATION=$(nak event -k 1040 -c "$MINIMAL_OTS" -t "e=$BASE_EVENT_ID" $RELAY)
if [ ! -z "$OTS_ATTESTATION" ]; then
    print_result "Valid OpenTimestamps attestation" true "03"
else
    print_result "Valid OpenTimestamps attestation" false "03"
fi

# Test 2: Missing 'e' tag
MISSING_E_TAG=$(nak event -k 1040 -c "$MINIMAL_OTS" $RELAY 2>&1)
if [[ "$MISSING_E_TAG" == *"missing required 'e' tag"* ]]; then
    print_result "Missing 'e' tag" true "03"
else
    print_result "Missing 'e' tag" false "03"
fi

# Test 3: Valid attestation with 'alt' tag
WITH_ALT_TAG=$(nak event -k 1040 -c "$MINIMAL_OTS" -t "e=$BASE_EVENT_ID" -t "alt=opentimestamps attestation" $RELAY)
if [ ! -z "$WITH_ALT_TAG" ]; then
    print_result "Valid attestation with 'alt' tag" true "03"
else
    print_result "Valid attestation with 'alt' tag" false "03"
fi

# Test 4: Invalid 'alt' tag value
INVALID_ALT_TAG=$(nak event -k 1040 -c "$MINIMAL_OTS" -t "e=$BASE_EVENT_ID" -t "alt=invalid" $RELAY 2>&1)
if [[ "$INVALID_ALT_TAG" == *"invalid"* ]] || [[ "$INVALID_ALT_TAG" == *"❌"* ]]; then
    print_result "Invalid 'alt' tag value" true "03"
else
    print_result "Invalid 'alt' tag value" false "03"
fi

# Test 5: Invalid base64 content
INVALID_BASE64=$(nak event -k 1040 -c "not base64" -t "e=$BASE_EVENT_ID" -t "alt=opentimestamps attestation" $RELAY 2>&1)
if [[ "$INVALID_BASE64" == *"invalid base64 content in OpenTimestamps attestation"* ]]; then
    print_result "Invalid base64 content" true "03"
else
    print_result "Invalid base64 content" false "03"
fi

# Test 6: Multiple 'e' tags
MULTIPLE_E_TAGS=$(nak event -k 1040 -c "$MINIMAL_OTS" -t "e=$BASE_EVENT_ID" -t "e=0000000000000000000000000000000000000000000000000000000000000000" -t "alt=opentimestamps attestation" $RELAY)
if [ ! -z "$MULTIPLE_E_TAGS" ]; then
    print_result "Multiple 'e' tags" true "03"
else
    print_result "Multiple 'e' tags" false "03"
fi

# Test 7: Invalid event ID in 'e' tag
INVALID_EVENT_ID=$(nak event -k 1040 -c "$MINIMAL_OTS" -t "e=invalid" -t "alt=opentimestamps attestation" $RELAY 2>&1)
if [[ "$INVALID_EVENT_ID" == *"invalid event ID in 'e' tag"* ]]; then
    print_result "Invalid event ID in 'e' tag" true "03"
else
    print_result "Invalid event ID in 'e' tag" false "03"
fi

# Test 8: Large OTS file (over 2KB)
# Create a large base64 string (2KB + 1 byte)
LARGE_OTS=$(head -c 2049 /dev/zero | base64)
LARGE_OTS_EVENT=$(nak event -k 1040 -c "$LARGE_OTS" -t "e=$BASE_EVENT_ID" -t "alt=opentimestamps attestation" $RELAY 2>&1)
if [[ "$LARGE_OTS_EVENT" == *"OTS file content too large (max 2KB)"* ]]; then
    print_result "Large OTS file (over 2KB)" true "03"
else
    print_result "Large OTS file (over 2KB)" false "03"
fi

# Test 9: Query attestation by event ID
QUERY_ATTESTATION=$(nak req -k 1040 -t "e=$BASE_EVENT_ID" $RELAY)
if [ ! -z "$QUERY_ATTESTATION" ]; then
    print_result "Query attestation by event ID" true "03"
else
    print_result "Query attestation by event ID" false "03"
fi

# Test 10: Query attestation by kind
QUERY_BY_KIND=$(nak req -k 1040 $RELAY)
if [ ! -z "$QUERY_BY_KIND" ]; then
    print_result "Query attestation by kind" true "03"
else
    print_result "Query attestation by kind" false "03"
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