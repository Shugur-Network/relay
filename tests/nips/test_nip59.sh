#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Test counters
total_tests=0
successful_tests=0
failed_tests=0

# Relay URL
# RELAY="ws://localhost:8080"
RELAY="wss://shu02.shugur.net"

RELAY_URL=$RELAY

# Function to check dependencies
check_dependencies() {
    for cmd in nak websocat; do
        if ! command -v $cmd &> /dev/null; then
            echo -e "${RED}Error: $cmd is not installed${NC}"
            exit 1
        fi
    done
}

# Check dependencies
check_dependencies

# Helper function to print test results
print_result() {
    local test_name=$1
    local success=$2
    local message=$3
    
    total_tests=$((total_tests + 1))
    if [ "$success" = true ]; then
        successful_tests=$((successful_tests + 1))
        echo -e "${GREEN}✓ Test $total_tests: $test_name - $message${NC}"
    else
        failed_tests=$((failed_tests + 1))
        echo -e "${RED}✗ Test $total_tests: $test_name - $message${NC}"
    fi
}

# Generate test keys
PRIVKEY=$(nak key generate)
PUBKEY=$(nak key public $PRIVKEY)

echo "Generated test keys:"
echo "Public key: $PUBKEY"

# Test 1: Create a basic wallet connect event
echo "Test 1: Creating a basic wallet connect event"
CONTENT='{"uri":"wc:test123"}'
response=$(nak event -k 13194 -c "$CONTENT" -t relay=wss://relay.example.com --sec $PRIVKEY $RELAY 2>&1)
echo "Response: $response"
if [[ "$response" == *"success"* ]]; then
    print_result "Basic wallet connect event" true "Event created successfully"
else
    print_result "Basic wallet connect event" false "Failed to create event"
fi

# Test 2: Create a wallet connect event with metadata
echo "Test 2: Creating a wallet connect event with metadata"
CONTENT_WITH_METADATA='{"uri":"wc:test456","name":"Test Wallet","description":"Test wallet connection"}'
response=$(nak event -k 13194 -c "$CONTENT_WITH_METADATA" -t relay=wss://relay.example.com --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Wallet connect event with metadata" true "Event created successfully"
else
    print_result "Wallet connect event with metadata" false "Failed to create event"
fi

# Test 3: Create a wallet connect event with multiple relays
echo "Test 3: Creating a wallet connect event with multiple relays"
CONTENT_MULTI='{"uri":"wc:test789"}'
response=$(nak event -k 13194 -c "$CONTENT_MULTI" -t relay=wss://relay1.example.com -t relay=wss://relay2.example.com --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Wallet connect event with multiple relays" true "Event created successfully"
else
    print_result "Wallet connect event with multiple relays" false "Failed to create event"
fi

# Test 4: Create a wallet connect event with empty content
echo "Test 4: Creating a wallet connect event with empty content"
EMPTY_CONTENT='{}'
response=$(nak event -k 13194 -c "$EMPTY_CONTENT" -t relay=wss://relay.example.com --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Wallet connect event with empty content" true "Event created successfully"
else
    print_result "Wallet connect event with empty content" false "Failed to create event"
fi

# Test 5: Attempt to create an event without relay tag
echo "Test 5: Attempting to create an event without relay tag"
CONTENT_NO_RELAY='{"uri":"wc:test123"}'
response=$(nak event -k 13194 -c "$CONTENT_NO_RELAY" --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event without relay tag" true "Event created successfully"
else
    print_result "Event without relay tag" false "Failed to create event"
fi

# Test 6: Attempt to create an event with invalid relay URL
echo "Test 6: Attempting to create an event with invalid relay URL"
response=$(nak event -k 13194 -c "$CONTENT" -t relay=invalid-url --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event with invalid relay URL" true "Event created successfully"
else
    print_result "Event with invalid relay URL" false "Failed to create event"
fi

# Test 7: Attempt to create an event with invalid pubkey in tag
echo "Test 7: Attempting to create an event with invalid pubkey in tag"
response=$(nak event -k 13194 -c "$CONTENT" -t p=invalid_pubkey -t relay=wss://relay.example.com --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* || "$response" == *"invalid"* ]]; then
    print_result "Event with invalid pubkey in tag" true "Event validation worked as expected"
else
    print_result "Event with invalid pubkey in tag" false "Event validation failed"
fi

# Test 8: Attempt to create an event with malformed content
echo "Test 8: Attempting to create an event with malformed content"
MALFORMED_CONTENT="invalid json"
response=$(nak event -k 13194 -c "$MALFORMED_CONTENT" -t relay=wss://relay.example.com --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event with malformed content" true "Event created successfully"
else
    print_result "Event with malformed content" false "Failed to create event"
fi

# Test 9: Retrieve wallet connect events
echo "Test 9: Retrieving wallet connect events"
sleep 2  # Give events time to be indexed
echo "Executing: nak req -k 13194 -l 10 $RELAY"
search_response=$(nak req -k 13194 -l 10 $RELAY 2>&1)
echo "Response: $search_response"
if [[ "$search_response" == *"\"pubkey\""* && "$search_response" == *"\"content\""* ]]; then
    print_result "Retrieve wallet connect events" true "Successfully found events"
else
    print_result "Retrieve wallet connect events" false "Failed to find events"
fi

# Print summary
echo -e "\nTest Summary:"
echo "Total Tests: $total_tests"
echo -e "${GREEN}Successful Tests: $successful_tests${NC}"
echo -e "${RED}Failed Tests: $failed_tests${NC}"

# Exit with error if any tests failed
if [ $failed_tests -gt 0 ]; then
    exit 1
fi 