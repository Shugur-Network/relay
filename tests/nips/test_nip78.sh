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

echo -e "${BLUE}Starting Shugur Relay NIP-78 Tests${NC}\n"

# Test NIP-78: Application-specific Data
echo -e "\n${YELLOW}Testing NIP-78: Application-specific Data${NC}"

# Test 1: Create a basic application-specific data event
APP_DATA=$(nak event -k 30078 -c '{"name": "test_app", "data": {"key": "value"}}' -t d=test-app -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$APP_DATA" ]; then
    print_result "Create basic application-specific data event" true "78"
else
    print_result "Create basic application-specific data event" false "78"
fi

# Test 2: Create an application-specific data event with metadata
if [ ! -z "$APP_DATA" ]; then
    METADATA_DATA=$(nak event -k 30078 -c '{"name": "test_app", "data": {"key": "value"}, "metadata": {"version": "1.0", "description": "Test application"}}' -t d=test-app-metadata -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$METADATA_DATA" ]; then
        print_result "Create application-specific data event with metadata" true "78"
    else
        print_result "Create application-specific data event with metadata" false "78"
    fi
fi

# Test 3: Create an application-specific data event with complex data
COMPLEX_DATA=$(nak event -k 30078 -c '{"name": "test_app", "data": {"array": [1, 2, 3], "object": {"nested": "value"}, "boolean": true, "number": 42}}' -t d=test-app-complex -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$COMPLEX_DATA" ]; then
    print_result "Create application-specific data event with complex data" true "78"
else
    print_result "Create application-specific data event with complex data" false "78"
fi

# Test 4: Create an application-specific data event with empty data
EMPTY_DATA=$(nak event -k 30078 -c '{"name": "test_app", "data": {}}' -t d=test-app-empty -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$EMPTY_DATA" ]; then
    print_result "Create application-specific data event with empty data" true "78"
else
    print_result "Create application-specific data event with empty data" false "78"
fi

# Test 5: Attempt to create without name field
NO_NAME=$(nak event -k 30078 -c '{"data": {"key": "value"}}' -t d=test-app-noname -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NO_NAME" == *"invalid"* ]] || [[ "$NO_NAME" == *"❌"* ]] || [[ "$NO_NAME" == *"NIP validation failed"* ]]; then
    print_result "Reject application-specific data event without name" true "78"
else
    print_result "Reject application-specific data event without name" false "78"
fi

# Test 6: Attempt to create without data field
NO_DATA=$(nak event -k 30078 -c '{"name": "test_app"}' -t d=test-app-nodata -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NO_DATA" == *"invalid"* ]] || [[ "$NO_DATA" == *"❌"* ]] || [[ "$NO_DATA" == *"NIP validation failed"* ]]; then
    print_result "Reject application-specific data event without data" true "78"
else
    print_result "Reject application-specific data event without data" false "78"
fi

# Test 7: Attempt to create with invalid data type
INVALID_DATA=$(nak event -k 30078 -c '{"name": "test_app", "data": "invalid"}' -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_DATA" == *"invalid"* ]] || [[ "$INVALID_DATA" == *"❌"* ]]; then
    print_result "Reject application-specific data event with invalid data type" true "78"
else
    print_result "Reject application-specific data event with invalid data type" false "78"
fi

# Test 8: Attempt to create with invalid recipient
INVALID_RECIPIENT=$(nak event -k 30078 -c '{"name": "test_app", "data": {"key": "value"}}' -t p=invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_RECIPIENT" == *"invalid"* ]] || [[ "$INVALID_RECIPIENT" == *"❌"* ]]; then
    print_result "Reject application-specific data event with invalid recipient" true "78"
else
    print_result "Reject application-specific data event with invalid recipient" false "78"
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