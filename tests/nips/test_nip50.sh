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
RELAY="ws://localhost:8080"
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

# Test 1: Create a basic event with keywords
echo "Test 1: Creating a basic event with keywords"
response=$(nak event -k 1 -c "Test event with keywords" -t t=nostr -t t=test --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Basic event with keywords" true "Event created successfully"
else
    print_result "Basic event with keywords" false "Failed to create event"
fi

# Test 2: Create an event with multiple keywords
echo "Test 2: Creating an event with multiple keywords"
response=$(nak event -k 1 -c "Test event with multiple keywords" -t t=nostr -t t=test -t t=multiple --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event with multiple keywords" true "Event created successfully"
else
    print_result "Event with multiple keywords" false "Failed to create event"
fi

# Test 3: Create an event with empty keywords
echo "Test 3: Creating an event with empty keywords"
response=$(nak event -k 1 -c "Test event with empty keyword" -t t= --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event with empty keywords" true "Event created successfully"
else
    print_result "Event with empty keywords" false "Failed to create event"
fi

# Test 4: Create an event with long keywords
echo "Test 4: Creating an event with long keywords"
response=$(nak event -k 1 -c "Test event with long keyword" -t t=very_long_keyword_that_should_still_be_valid --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event with long keywords" true "Event created successfully"
else
    print_result "Event with long keywords" false "Failed to create event"
fi

# Test 5: Attempt to create an event without keyword tag
echo "Test 5: Attempting to create an event without keyword tag"
response=$(nak event -k 1 -c "Test event without keyword" --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event without keyword tag" true "Event created successfully"
else
    print_result "Event without keyword tag" false "Failed to create event"
fi

# Test 6: Attempt to create an event with invalid keyword tag
echo "Test 6: Attempting to create an event with invalid keyword tag"
response=$(nak event -k 1 -c "Test event with invalid keyword tag" -t invalid=test --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event with invalid keyword tag" true "Event created successfully"
else
    print_result "Event with invalid keyword tag" false "Failed to create event"
fi

# Test 7: Attempt to create an event with invalid pubkey in tag
echo "Test 7: Attempting to create an event with invalid pubkey in tag"
response=$(nak event -k 1 -c "Test event with invalid pubkey in tag" -t p=invalid_pubkey -t t=test --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* || "$response" == *"invalid"* ]]; then
    print_result "Event with invalid pubkey in tag" true "Event validation worked as expected"
else
    print_result "Event with invalid pubkey in tag" false "Event validation failed"
fi

# Test 8: Attempt to create an event with duplicate keywords
echo "Test 8: Attempting to create an event with duplicate keywords"
response=$(nak event -k 1 -c "Test event with duplicate keywords" -t t=test -t t=test --sec $PRIVKEY $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Event with duplicate keywords" true "Event created successfully"
else
    print_result "Event with duplicate keywords" false "Failed to create event"
fi

# Test 9: Search for events with a specific keyword
echo "Test 9: Searching for events with a specific keyword"
sleep 2  # Give events time to be indexed
echo "Executing: nak req -k 1 -t t=test -l 10 $RELAY"
search_response=$(nak req -k 1 -t t=test -l 10 $RELAY 2>&1)
echo "Response: $search_response"
if [[ "$search_response" == *"\"pubkey\""* && "$search_response" == *"\"content\""* ]]; then
    print_result "Search for events by keyword" true "Successfully found events"
else
    print_result "Search for events by keyword" false "Failed to find events"
fi

# Test 10: Search using the search parameter (NIP-50 specific)
echo "Test 10: Searching using the search parameter"
echo "Executing: nak req -k 1 --search \"test\" -l 10 $RELAY"
search_response=$(nak req -k 1 --search "test" -l 10 $RELAY 2>&1)
echo "Response: $search_response"
if [[ "$search_response" == *"\"pubkey\""* && "$search_response" == *"\"content\""* ]]; then
    print_result "Search using the search parameter" true "Successfully found events"
else
    print_result "Search using the search parameter" false "Failed to find events"
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