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

echo -e "${BLUE}Starting Shugur Relay NIP-22 Tests${NC}\n"

# Test NIP-22: Event Deletion
echo -e "\n${YELLOW}Testing NIP-22: Event Deletion${NC}"

# Test 1: Create a basic event to delete
ORIGINAL_EVENT=$(nak event -k 1 -c "This is a test event that will be deleted" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
ORIGINAL_EVENT_ID=$(echo "$ORIGINAL_EVENT" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ ! -z "$ORIGINAL_EVENT" ]; then
    print_result "Create basic event for deletion" true "22"
else
    print_result "Create basic event for deletion" false "22"
fi

# Test 2: Create a deletion event with reason
if [ ! -z "$ORIGINAL_EVENT_ID" ]; then
    DELETION_EVENT=$(nak event -k 5 -c "Content removed by user request" -t e=$ORIGINAL_EVENT_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    DELETION_EVENT_ID=$(echo "$DELETION_EVENT" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    if [ ! -z "$DELETION_EVENT" ]; then
        print_result "Create deletion event with reason" true "22"
    else
        print_result "Create deletion event with reason" false "22"
    fi
fi

# Test 3: Create a deletion event without reason
if [ ! -z "$ORIGINAL_EVENT_ID" ]; then
    NO_REASON_DELETION=$(nak event -k 5 -c "" -t e=$ORIGINAL_EVENT_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$NO_REASON_DELETION" ]; then
        print_result "Create deletion event without reason" true "22"
    else
        print_result "Create deletion event without reason" false "22"
    fi
fi

# Test 4: Create a deletion event with multiple event IDs
if [ ! -z "$ORIGINAL_EVENT_ID" ] && [ ! -z "$DELETION_EVENT_ID" ]; then
    MULTI_DELETION=$(nak event -k 5 -c "Deleting multiple events" -t e=$ORIGINAL_EVENT_ID -t e=$DELETION_EVENT_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$MULTI_DELETION" ]; then
        print_result "Create deletion event with multiple event IDs" true "22"
    else
        print_result "Create deletion event with multiple event IDs" false "22"
    fi
fi

# Test 5: Attempt to delete a non-existent event
NONEXISTENT_DELETION=$(nak event -k 5 -c "Attempting to delete non-existent event" -t e=invalid_event_id -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NONEXISTENT_DELETION" == *"invalid"* ]] || [[ "$NONEXISTENT_DELETION" == *"❌"* ]]; then
    print_result "Reject deletion of non-existent event" true "22"
else
    print_result "Reject deletion of non-existent event" false "22"
fi

# Test 6: Attempt to delete without event ID
NO_EVENT_DELETION=$(nak event -k 5 -c "Attempting to delete without event ID" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NO_EVENT_DELETION" == *"missing required"* ]] || [[ "$NO_EVENT_DELETION" == *"failed"* ]]; then
    print_result "Reject deletion without event ID" true "22"
else
    print_result "Reject deletion without event ID" false "22"
fi

# Test 7: Attempt to delete with invalid event ID
INVALID_EVENT_DELETION=$(nak event -k 5 -c "Attempting to delete with invalid event ID" -t e=invalid_event_id -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_EVENT_DELETION" == *"invalid"* ]] || [[ "$INVALID_EVENT_DELETION" == *"❌"* ]]; then
    print_result "Reject deletion with invalid event ID" true "22"
else
    print_result "Reject deletion with invalid event ID" false "22"
fi

# Test 8: Attempt to delete with invalid recipient
INVALID_RECIPIENT_DELETION=$(nak event -k 5 -c "Attempting to delete with invalid recipient" -t e=$ORIGINAL_EVENT -t p=invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_RECIPIENT_DELETION" == *"invalid"* ]] || [[ "$INVALID_RECIPIENT_DELETION" == *"❌"* ]]; then
    print_result "Reject deletion with invalid recipient" true "22"
else
    print_result "Reject deletion with invalid recipient" false "22"
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