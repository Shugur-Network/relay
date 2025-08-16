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

echo -e "${BLUE}Starting Shugur Relay NIP-28 Tests${NC}\n"

# Test NIP-28: Public Chat
echo -e "\n${YELLOW}Testing NIP-28: Public Chat${NC}"

# Test 1: Create a public chat channel
CHANNEL_CREATE=$(nak event -k 40 --content '{"name": "Test Channel", "about": "A test channel for NIP-28", "picture": "https://example.com/channel.jpg"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
CHANNEL_ID=$(echo "$CHANNEL_CREATE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ ! -z "$CHANNEL_CREATE" ]; then
    print_result "Create public chat channel" true "28"
else
    print_result "Create public chat channel" false "28"
fi

# Test 2: Create a channel message
if [ ! -z "$CHANNEL_ID" ]; then
    CHANNEL_MESSAGE=$(nak event -k 41 --content "Hello, this is a test message in the channel" -t e=$CHANNEL_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    MESSAGE_ID=$(echo "$CHANNEL_MESSAGE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    if [ ! -z "$CHANNEL_MESSAGE" ]; then
        print_result "Create channel message" true "28"
    else
        print_result "Create channel message" false "28"
    fi
fi

# Test 3: Create a channel metadata update
if [ ! -z "$CHANNEL_ID" ]; then
    CHANNEL_UPDATE=$(nak event -k 40 --content '{"name": "Updated Channel", "about": "Updated channel description", "picture": "https://example.com/updated.jpg"}' -t e=$CHANNEL_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$CHANNEL_UPDATE" ]; then
        print_result "Update channel metadata" true "28"
    else
        print_result "Update channel metadata" false "28"
    fi
fi

# Test 4: Create a channel hide message
if [ ! -z "$MESSAGE_ID" ]; then
    HIDE_MESSAGE=$(nak event -k 43 --content '{"message_id": "'$MESSAGE_ID'", "reason": "Inappropriate content"}' -t e=$CHANNEL_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$HIDE_MESSAGE" ]; then
        print_result "Hide channel message" true "28"
    else
        print_result "Hide channel message" false "28"
    fi
fi

# Test 5: Create a channel mute user
MUTE_USER=$(nak event -k 44 --content '{"pubkey": "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798", "reason": "Spam"}' -t e=$CHANNEL_ID -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$MUTE_USER" ]; then
    print_result "Mute channel user" true "28"
else
    print_result "Mute channel user" false "28"
fi

# Test 6: Attempt to create channel without name
NO_NAME_CHANNEL=$(nak event -k 40 --content '{"about": "Channel without name"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [ ! -z "$NO_NAME_CHANNEL" ] && [[ "$NO_NAME_CHANNEL" != *"failed"* ]]; then
    print_result "Allow channel creation without name" true "28"
else
    print_result "Allow channel creation without name" false "28"
fi

# Test 7: Attempt to create channel message without channel ID
NO_CHANNEL_MESSAGE=$(nak event -k 41 --content "Attempting to send without channel ID" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NO_CHANNEL_MESSAGE" == *"missing required 'e' tag"* ]] || [[ "$NO_CHANNEL_MESSAGE" == *"failed"* ]]; then
    print_result "Reject channel message without channel ID" true "28"
else
    print_result "Reject channel message without channel ID" false "28"
fi

# Test 8: Attempt to create channel message with invalid channel ID
INVALID_CHANNEL_MESSAGE=$(nak event -k 41 --content "Attempting to send to invalid channel" -t e=invalid_channel_id -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [ ! -z "$INVALID_CHANNEL_MESSAGE" ] && [[ "$INVALID_CHANNEL_MESSAGE" != *"failed"* ]]; then
    print_result "Allow channel message with any channel ID" true "28"
else
    print_result "Allow channel message with any channel ID" false "28"
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