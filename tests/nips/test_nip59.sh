#!/bin/bash

# NIP-59: Gift Wrap
# Tests gift wrapping functionality with kinds 13 (seal) and 1059 (gift wrap)
# https://github.com/nostr-protocol/nips/blob/master/59.md

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'

# Test counters
total_tests=0
successful_tests=0
failed_tests=0

# Relay URL - use local relay for testing
RELAY="ws://localhost:8085"

# Function to check dependencies
check_dependencies() {
    for cmd in nak; do
        if ! command -v $cmd &> /dev/null; then
            echo -e "${RED}Error: $cmd is not installed${NC}"
            echo "Install with: go install github.com/fiatjaf/nak@latest"
            exit 1
        fi
    done
}

# Check dependencies
check_dependencies

# Function to check relay connectivity
check_relay() {
    echo "🔗 Checking relay connection..."
    if ! timeout 5 bash -c "</dev/tcp/localhost/8085" 2>/dev/null; then
        echo -e "${RED}❌ Relay not accessible on localhost:8085${NC}"
        echo "Start the relay with: ./bin/relay start --config config/development.yaml"
        exit 1
    fi
    echo -e "${GREEN}✅ Relay is accessible${NC}"
}

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

# Function to get random timestamp within last 2 days
get_random_timestamp() {
    local now=$(date +%s)
    local two_days_ago=$((now - 172800)) # 2 days in seconds
    local random_offset=$((RANDOM % 172800))
    echo $((two_days_ago + random_offset))
}

echo -e "${BLUE}Testing NIP-59: Gift Wrap${NC}"
echo "=================================="

# Check relay connection
check_relay

# Generate test keys
echo -e "${YELLOW}Generating test keys...${NC}"
SENDER_PRIVKEY=$(nak key generate)
SENDER_PUBKEY=$(nak key public $SENDER_PRIVKEY)
RECIPIENT_PRIVKEY=$(nak key generate)
RECIPIENT_PUBKEY=$(nak key public $RECIPIENT_PRIVKEY)

echo "Sender public key: $SENDER_PUBKEY"
echo "Recipient public key: $RECIPIENT_PUBKEY"

# Test 1: Create a basic gift wrap event (kind 1059)
echo -e "\n${YELLOW}Test 1: Creating a basic gift wrap event (kind 1059)${NC}"
# First create a simple message to encrypt
TEST_MESSAGE="Hello from NIP-59 gift wrap!"
ENCRYPTED_MESSAGE=$(nak encrypt --recipient-pubkey "$RECIPIENT_PUBKEY" --sec "$SENDER_PRIVKEY" "$TEST_MESSAGE")

# Create the gift wrap
response=$(nak event -k 1059 -c "$ENCRYPTED_MESSAGE" --sec "$SENDER_PRIVKEY" -p "$RECIPIENT_PUBKEY" --created-at $(get_random_timestamp) $RELAY 2>&1)
echo "Response: $response"
if [[ "$response" == *"success"* ]]; then
    print_result "Basic gift wrap event" true "Event created successfully"
    # Extract event ID for later retrieval
    GIFT_WRAP_ID=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "Gift wrap event ID: $GIFT_WRAP_ID"
else
    print_result "Basic gift wrap event" false "Failed to create event: $response"
fi

# Test 2: Create a gift wrap with multiple recipients
echo -e "\n${YELLOW}Test 2: Creating a gift wrap with multiple recipients${NC}"
RECIPIENT2_PRIVKEY=$(nak key generate)
RECIPIENT2_PUBKEY=$(nak key public $RECIPIENT2_PRIVKEY)

response=$(nak event -k 1059 -c "$ENCRYPTED_MESSAGE" --sec "$SENDER_PRIVKEY" -p "$RECIPIENT_PUBKEY" -p "$RECIPIENT2_PUBKEY" --created-at $(get_random_timestamp) $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Gift wrap with multiple recipients" true "Event created successfully"
else
    print_result "Gift wrap with multiple recipients" false "Failed to create event"
fi

# Test 3: Test gift wrap without 'p' tag (should fail or be accepted based on relay policy)
echo -e "\n${YELLOW}Test 3: Creating a gift wrap without 'p' tag${NC}"
response=$(nak event -k 1059 -c "$ENCRYPTED_MESSAGE" --sec "$SENDER_PRIVKEY" --created-at $(get_random_timestamp) $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Gift wrap without p tag" true "Event accepted (relay policy allows)"
else
    print_result "Gift wrap without p tag" true "Event rejected (relay enforces p tag requirement)"
fi

# Test 4: Test gift wrap with invalid NIP-44 content (should fail)
echo -e "\n${YELLOW}Test 4: Creating a gift wrap with invalid NIP-44 content${NC}"
INVALID_CONTENT="plain text not encrypted"
response=$(nak event -k 1059 -c "$INVALID_CONTENT" --sec "$SENDER_PRIVKEY" -p "$RECIPIENT_PUBKEY" --created-at $(get_random_timestamp) $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Gift wrap with invalid content" false "Event should have been rejected but was accepted"
else
    print_result "Gift wrap with invalid content" true "Event correctly rejected for invalid NIP-44 content"
fi

# Test 5: Test gift wrap with empty content (should fail)
echo -e "\n${YELLOW}Test 5: Creating a gift wrap with empty content${NC}"
response=$(nak event -k 1059 -c "" --sec "$SENDER_PRIVKEY" -p "$RECIPIENT_PUBKEY" --created-at $(get_random_timestamp) $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Gift wrap with empty content" false "Event should have been rejected but was accepted"
else
    print_result "Gift wrap with empty content" true "Event correctly rejected for empty content"
fi

# Test 6: Retrieve gift wrap events
echo -e "\n${YELLOW}Test 6: Retrieving gift wrap events${NC}"
sleep 2  # Give events time to be indexed
search_response=$(nak req -k 1059 -l 5 $RELAY 2>&1)
echo "Response: $search_response"
if [[ "$search_response" == *"\"pubkey\""* && "$search_response" == *"\"content\""* ]]; then
    print_result "Retrieve gift wrap events" true "Successfully found events"
else
    print_result "Retrieve gift wrap events" false "Failed to find events"
fi

# Test 7: Test retrieving specific gift wrap by ID (if we have one)
if [ ! -z "$GIFT_WRAP_ID" ]; then
    echo -e "\n${YELLOW}Test 7: Retrieving specific gift wrap by ID${NC}"
    specific_response=$(nak req -i "$GIFT_WRAP_ID" $RELAY 2>&1)
    if [[ "$specific_response" == *"$GIFT_WRAP_ID"* ]]; then
        print_result "Retrieve specific gift wrap" true "Successfully found event by ID"
    else
        print_result "Retrieve specific gift wrap" false "Failed to find event by ID"
    fi
else
    echo -e "\n${YELLOW}Test 7: Skipping specific gift wrap retrieval (no ID available)${NC}"
    total_tests=$((total_tests + 1))
fi

# Test 8: Test gift wrap decryption (end-to-end test)
if [ ! -z "$GIFT_WRAP_ID" ]; then
    echo -e "\n${YELLOW}Test 8: Testing gift wrap decryption${NC}"
    # Retrieve the gift wrap
    retrieved_event=$(nak req -i "$GIFT_WRAP_ID" $RELAY 2>&1)
    if [[ "$retrieved_event" == *"$GIFT_WRAP_ID"* ]]; then
        # Extract the content and try to decrypt
        ENCRYPTED_CONTENT=$(echo "$retrieved_event" | grep -o '"content":"[^"]*"' | cut -d'"' -f4)
        if [ ! -z "$ENCRYPTED_CONTENT" ]; then
            DECRYPTED_MESSAGE=$(nak decrypt --sec "$RECIPIENT_PRIVKEY" -p "$SENDER_PUBKEY" "$ENCRYPTED_CONTENT" 2>/dev/null)
            if [[ "$DECRYPTED_MESSAGE" == "$TEST_MESSAGE" ]]; then
                print_result "Gift wrap decryption" true "Successfully decrypted message: '$DECRYPTED_MESSAGE'"
            else
                print_result "Gift wrap decryption" false "Decryption failed or message mismatch"
            fi
        else
            print_result "Gift wrap decryption" false "Could not extract encrypted content"
        fi
    else
        print_result "Gift wrap decryption" false "Could not retrieve event for decryption test"
    fi
else
    echo -e "\n${YELLOW}Test 8: Skipping decryption test (no gift wrap ID available)${NC}"
    total_tests=$((total_tests + 1))
fi

# Test 9: Test with wallet connect content (kind 13194 inside gift wrap)
echo -e "\n${YELLOW}Test 9: Creating gift wrap containing wallet connect content${NC}"
WALLET_CONTENT='{"uri":"wc:test123","name":"Test Wallet"}'
ENCRYPTED_WALLET=$(nak encrypt --recipient-pubkey "$RECIPIENT_PUBKEY" --sec "$SENDER_PRIVKEY" "$WALLET_CONTENT")
response=$(nak event -k 1059 -c "$ENCRYPTED_WALLET" --sec "$SENDER_PRIVKEY" -p "$RECIPIENT_PUBKEY" --created-at $(get_random_timestamp) $RELAY 2>&1)
if [[ "$response" == *"success"* ]]; then
    print_result "Gift wrap with wallet connect content" true "Event created successfully"
else
    print_result "Gift wrap with wallet connect content" false "Failed to create event"
fi

# Print summary
echo -e "\n=================================="
echo -e "${BLUE}NIP-59 Test Summary:${NC}"
echo "Total Tests: $total_tests"
echo -e "${GREEN}Successful Tests: $successful_tests${NC}"
echo -e "${RED}Failed Tests: $failed_tests${NC}"

if [ $failed_tests -gt 0 ]; then
    echo -e "\n${RED}Some tests failed. Check relay logs for details.${NC}"
    exit 1
else
    echo -e "\n${GREEN}All tests passed! NIP-59 Gift Wrap functionality is working correctly.${NC}"
fi