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

echo -e "${BLUE}Starting Shugur Relay NIP-20 Tests${NC}\n"

# Test NIP-20: Command Results
echo -e "\n${YELLOW}Testing NIP-20: Command Results${NC}"

# Test 1: Create a basic command result
COMMAND_RESULT=$(nak event -k 24133 -c '{"result": "success", "command": "test", "message": "Command executed successfully"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
COMMAND_EVENT_ID=$(echo "$COMMAND_RESULT" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ ! -z "$COMMAND_RESULT" ]; then
    print_result "Create basic command result" true "20"
else
    print_result "Create basic command result" false "20"
fi

# Test 2: Create a command result with error
ERROR_RESULT=$(nak event -k 24133 -c '{"result": "error", "command": "test", "message": "Command failed", "error": "Invalid input"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$ERROR_RESULT" ]; then
    print_result "Create command result with error" true "20"
else
    print_result "Create command result with error" false "20"
fi

# Test 3: Create a command result with metadata
METADATA_JSON='{"result":"success","command":"test","message":"Command executed","metadata":{"duration":100,"timestamp":1234567890}}'
METADATA_RESULT=$(nak event -k 24133 -c "$METADATA_JSON" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$METADATA_RESULT" ]; then
    print_result "Create command result with metadata" true "20"
else
    print_result "Create command result with metadata" false "20"
fi

# Test 4: Create a command result with reply
if [ ! -z "$COMMAND_EVENT_ID" ]; then
    REPLY_RESULT=$(nak event -k 24133 -c '{"result": "success", "command": "test", "message": "Reply to command"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 -t e=$COMMAND_EVENT_ID $RELAY)
    if [ ! -z "$REPLY_RESULT" ]; then
        print_result "Create command result with reply" true "20"
    else
        print_result "Create command result with reply" false "20"
    fi
fi

# Test 5: Create a command result without required fields
INVALID_RESULT=$(nak event -k 24133 -c '{"invalid": "data"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_RESULT" == *"invalid"* ]] || [[ "$INVALID_RESULT" == *"❌"* ]]; then
    print_result "Reject command result without required fields" true "20"
else
    print_result "Reject command result without required fields" false "20"
fi

# Test 6: Create a command result with invalid result type
INVALID_TYPE_RESULT=$(nak event -k 24133 -c '{"result": "invalid", "command": "test", "message": "Invalid result type"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_TYPE_RESULT" == *"invalid"* ]] || [[ "$INVALID_TYPE_RESULT" == *"❌"* ]]; then
    print_result "Reject command result with invalid result type" true "20"
else
    print_result "Reject command result with invalid result type" false "20"
fi

# Test 7: Create a command result with invalid recipient
INVALID_RECIPIENT_RESULT=$(nak event -k 24133 -c '{"result": "success", "command": "test", "message": "Invalid recipient"}' -t p=invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_RECIPIENT_RESULT" == *"invalid"* ]] || [[ "$INVALID_RECIPIENT_RESULT" == *"❌"* ]]; then
    print_result "Reject command result with invalid recipient" true "20"
else
    print_result "Reject command result with invalid recipient" false "20"
fi

# Test 8: Create a command result with malformed JSON
MALFORMED_RESULT=$(nak event -k 24133 -c '{invalid json}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$MALFORMED_RESULT" == *"invalid"* ]] || [[ "$MALFORMED_RESULT" == *"❌"* ]]; then
    print_result "Reject command result with malformed JSON" true "20"
else
    print_result "Reject command result with malformed JSON" false "20"
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