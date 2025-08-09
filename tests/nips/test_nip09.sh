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

# Relay URL - using a known working relay
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

echo -e "${BLUE}Starting Shugur Relay NIP-09 Tests${NC}\n"

# Test NIP-09: Event Deletion
echo -e "\n${YELLOW}Testing NIP-09: Event Deletion${NC}"

# Test 1: Create a basic event to delete
ORIGINAL_EVENT=$(nak event -k 1 -c "This is a test event that will be deleted" $RELAY)
EVENT_ID=$(echo "$ORIGINAL_EVENT" | jq -r .id)
if [ ! -z "$EVENT_ID" ]; then
    print_result "Create original event for deletion" true "09"
else
    print_result "Create original event for deletion" false "09"
fi

# Test 2: Create a deletion event
if [ ! -z "$EVENT_ID" ]; then
    DELETION_EVENT=$(nak event -k 5 -c "Deleting the original event" -t e="$EVENT_ID" $RELAY)
    if [ ! -z "$DELETION_EVENT" ]; then
        print_result "Create deletion event" true "09"
    else
        print_result "Create deletion event" false "09"
    fi
fi

# Test 3: Create a deletion event without reason
if [ ! -z "$EVENT_ID" ]; then
    NO_REASON_DELETION=$(nak event -k 5 -c "" -t e="$EVENT_ID" $RELAY)
    if [ ! -z "$NO_REASON_DELETION" ]; then
        print_result "Create deletion event without reason" true "09"
    else
        print_result "Create deletion event without reason" false "09"
    fi
fi

# Test 4: Create a deletion event with multiple event IDs
MULTI_DELETION_EVENT=$(nak event -k 5 -c "Deleting multiple events" -t e="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798" -t e:"79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81799" $RELAY)
if [ ! -z "$MULTI_DELETION_EVENT" ]; then
    print_result "Create deletion event with multiple event IDs" true "09"
else
    print_result "Create deletion event with multiple event IDs" false "09"
fi

# Test 5: Create a deletion event without event ID
NO_EVENT_DELETION=$(nak event -k 5 -c "Attempting to delete without event ID" $RELAY 2>&1)
if [[ "$NO_EVENT_DELETION" == *"invalid"* ]] || [[ "$NO_EVENT_DELETION" == *"❌"* ]] || [[ "$NO_EVENT_DELETION" == *"failed"* ]]; then
    print_result "Reject deletion event without event ID" true "09"
else
    print_result "Reject deletion event without event ID" false "09"
fi

# Test 6: Create a deletion event with invalid event ID
INVALID_EVENT_DELETION=$(nak event -k 5 -c "Attempting to delete with invalid event ID" -t e="invalid_event_id" $RELAY 2>&1)
if [[ "$INVALID_EVENT_DELETION" == *"invalid"* ]] || [[ "$INVALID_EVENT_DELETION" == *"❌"* ]] || [[ "$INVALID_EVENT_DELETION" == *"failed"* ]]; then
    print_result "Reject deletion event with invalid event ID" true "09"
else
    print_result "Reject deletion event with invalid event ID" false "09"
fi

# Test 7: Create a deletion event for a non-existent event
NON_EXISTENT_DELETION=$(nak event -k 5 -c "Attempting to delete non-existent event" -t e="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798" $RELAY)
if [ ! -z "$NON_EXISTENT_DELETION" ]; then
    print_result "Allow deletion event for non-existent event" true "09"
else
    print_result "Allow deletion event for non-existent event" false "09"
fi

# Test 8: Create a deletion event with duplicate event IDs
DUPLICATE_DELETION=$(nak event -k 5 -c "Attempting to delete with duplicate event IDs" -t e="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798" -t e:"79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798" $RELAY)
if [ ! -z "$DUPLICATE_DELETION" ]; then
    print_result "Allow deletion event with duplicate event IDs" true "09"
else
    print_result "Allow deletion event with duplicate event IDs" false "09"
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