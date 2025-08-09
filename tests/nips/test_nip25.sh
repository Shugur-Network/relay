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

echo -e "${BLUE}Starting Shugur Relay NIP-25 Tests${NC}\n"

# Test NIP-25: Reactions
echo -e "\n${YELLOW}Testing NIP-25: Reactions${NC}"

# Test 1: Create a basic event to react to
ORIGINAL_EVENT=$(nak event -k 1 -c "This is a test event that will receive reactions" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
ORIGINAL_EVENT_ID=$(echo "$ORIGINAL_EVENT" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ ! -z "$ORIGINAL_EVENT" ]; then
    print_result "Create basic event for reactions" true "25"
else
    print_result "Create basic event for reactions" false "25"
fi

# Test 2: Create a positive reaction
if [ ! -z "$ORIGINAL_EVENT_ID" ]; then
    POSITIVE_REACTION=$(nak event -k 7 -c "+" -t e=$ORIGINAL_EVENT_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$POSITIVE_REACTION" ]; then
        print_result "Create positive reaction" true "25"
    else
        print_result "Create positive reaction" false "25"
    fi
fi

# Test 3: Create a negative reaction
if [ ! -z "$ORIGINAL_EVENT_ID" ]; then
    NEGATIVE_REACTION=$(nak event -k 7 -c "-" -t e=$ORIGINAL_EVENT_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$NEGATIVE_REACTION" ]; then
        print_result "Create negative reaction" true "25"
    else
        print_result "Create negative reaction" false "25"
    fi
fi

# Test 4: Create a reaction with emoji
if [ ! -z "$ORIGINAL_EVENT_ID" ]; then
    EMOJI_REACTION=$(nak event -k 7 -c "👍" -t e=$ORIGINAL_EVENT_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$EMOJI_REACTION" ]; then
        print_result "Create reaction with emoji" true "25"
    else
        print_result "Create reaction with emoji" false "25"
    fi
fi

# Test 5: Attempt to react to non-existent event
NONEXISTENT_REACTION=$(nak event -k 7 -c "+" -t e=invalid_event_id -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NONEXISTENT_REACTION" == *"invalid"* ]] || [[ "$NONEXISTENT_REACTION" == *"❌"* ]]; then
    print_result "Reject reaction to non-existent event" true "25"
else
    print_result "Reject reaction to non-existent event" false "25"
fi

# Test 6: Attempt to react without event ID
NO_EVENT_REACTION=$(nak event -k 7 -c "+" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NO_EVENT_REACTION" == *"missing required"* ]] || [[ "$NO_EVENT_REACTION" == *"failed"* ]]; then
    print_result "Reject reaction without event ID" true "25"
else
    print_result "Reject reaction without event ID" false "25"
fi

# Test 7: Attempt to react with invalid content
INVALID_CONTENT_REACTION=$(nak event -k 7 -c "invalid" -t e=$ORIGINAL_EVENT_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_CONTENT_REACTION" == *"invalid"* ]] || [[ "$INVALID_CONTENT_REACTION" == *"❌"* ]]; then
    print_result "Reject reaction with invalid content" true "25"
else
    print_result "Reject reaction with invalid content" false "25"
fi

# Test 8: Attempt to react with invalid recipient
INVALID_RECIPIENT_REACTION=$(nak event -k 7 -c "+" -t e=$ORIGINAL_EVENT_ID -t p=invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_RECIPIENT_REACTION" == *"invalid"* ]] || [[ "$INVALID_RECIPIENT_REACTION" == *"❌"* ]]; then
    print_result "Reject reaction with invalid recipient" true "25"
else
    print_result "Reject reaction with invalid recipient" false "25"
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