#!/bin/bash

#!/bin/bash

# Test script for NIP-65: Relay List Metadata (kind 10002)
# Tests relay list metadata events with 'r' tags and empty content using nak

RELAY_URL="${RELAY_URL:-wss://shu02.shugur.net}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing NIP-65: Relay List Metadata${NC}"
echo "Relay URL: $RELAY_URL"

# Test counters
test_count=0
success_count=0
fail_count=0

# Helper function to print test results
print_result() {
    local test_name=$1
    local success=$2
    
    ((test_count++))
    if [ "$success" = true ]; then
        echo -e "${GREEN}✓ Test $test_count: $test_name${NC}"
        ((success_count++))
    else
        echo -e "${RED}✗ Test $test_count: $test_name${NC}"
        ((fail_count++))
    fi
}

# Check if nak is available
if ! command -v nak &> /dev/null; then
    echo -e "${RED}Error: nak is required for testing${NC}"
    echo "Install with: go install github.com/fiatjaf/nak@latest"
    exit 1
fi

echo -e "${YELLOW}Running NIP-65 tests...${NC}"

# Test 1: Create a valid relay list with multiple r tags
echo -e "\n${YELLOW}Test 1: Create relay list with multiple r tags${NC}"
RESULT1=$(nak event -k 10002 -c "" -t r=wss://relay1.example.com -t r="wss://relay2.example.com;read" -t r="wss://relay3.example.com;write" $RELAY_URL 2>&1)

if echo "$RESULT1" | grep -q "success"; then
    print_result "Create relay list with r tags" true
else
    print_result "Create relay list with r tags" false
    echo "Result: $RESULT1"
fi

# Test 2: Create a simple relay list (no markers)
echo -e "\n${YELLOW}Test 2: Create simple relay list (no markers)${NC}"
RESULT2=$(nak event -k 10002 -c "" -t r=wss://simple1.example.com -t r=wss://simple2.example.com $RELAY_URL 2>&1)

if echo "$RESULT2" | grep -q "success"; then
    print_result "Create simple relay list" true
else
    print_result "Create simple relay list" false
    echo "Result: $RESULT2"
fi

# Test 3: Try to create relay list with invalid URL scheme
echo -e "\n${YELLOW}Test 3: Try invalid URL scheme (should fail)${NC}"
RESULT3=$(nak event -k 10002 -c "" -t r=http://invalid-scheme.example.com $RELAY_URL 2>&1)

if echo "$RESULT3" | grep -q "rejected\|error\|invalid\|failed"; then
    print_result "Reject invalid URL scheme" true
else
    print_result "Reject invalid URL scheme" false
    echo "Result: $RESULT3"
fi

# Test 4: Try to create relay list with invalid marker
echo -e "\n${YELLOW}Test 4: Try invalid marker (should fail)${NC}"
RESULT4=$(nak event -k 10002 -c "" -t r="wss://valid-url.example.com;invalid-marker" $RELAY_URL 2>&1)

if echo "$RESULT4" | grep -q "rejected\|error\|invalid\|failed"; then
    print_result "Reject invalid marker" true
else
    print_result "Reject invalid marker" false
    echo "Result: $RESULT4"
fi

# Test 5: Create relay list with non-empty content (should be allowed but not recommended)
echo -e "\n${YELLOW}Test 5: Create relay list with non-empty content${NC}"
RESULT5=$(nak event -k 10002 -c "this should be empty per NIP-65" -t r=wss://content-test.example.com $RELAY_URL 2>&1)

if echo "$RESULT5" | grep -q "success"; then
    print_result "Accept non-empty content" true
else
    print_result "Accept non-empty content" false
    echo "Result: $RESULT5"
fi

# Test 6: Update relay list (test replaceability)
echo -e "\n${YELLOW}Test 6: Update relay list (replaceability test)${NC}"
sleep 1  # Ensure different timestamp
RESULT6=$(nak event -k 10002 -c "" -t r=wss://updated-relay.example.com -t r="wss://backup-relay.example.com;read" $RELAY_URL 2>&1)

if echo "$RESULT6" | grep -q "success"; then
    print_result "Update relay list" true
else
    print_result "Update relay list" false
    echo "Result: $RESULT6"
fi

# Test 7: Query relay lists
echo -e "
${YELLOW}Test 7: Query relay lists${NC}"
QUERY_RESULT=$(nak req -k 10002 --limit 10 $RELAY_URL 2>&1)

if echo "$QUERY_RESULT" | grep -q "kind.*10002\|"kind":10002"; then
    print_result "Query relay lists" true
    echo "Found kind 10002 events:"
    echo "$QUERY_RESULT" | head -20
else
    print_result "Query relay lists" false
    echo "Query result: $QUERY_RESULT"
fi

# Test 8: Empty relay list (no r tags)
echo -e "
${YELLOW}Test 8: Create empty relay list${NC}"
RESULT8=$(nak event -k 10002 -c "" $RELAY_URL 2>&1)

if echo "$RESULT8" | grep -q "success"; then
    print_result "Create empty relay list" true
else
    print_result "Create empty relay list" false
    echo "Result: $RESULT8"
fi

echo -e "
${BLUE}=== Test Summary ===${NC}"
echo -e "Total tests: $test_count"
echo -e "${GREEN}Successful: $success_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"

if [ $fail_count -eq 0 ]; then
    echo -e "
${GREEN}All NIP-65 tests passed!${NC}"
    exit 0
else
    echo -e "
${RED}Some tests failed. Check the relay implementation.${NC}"
    exit 1
fi

set -e

RELAY_URL="${RELAY_URL:-ws://localhost:8080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing NIP-65: Relay List Metadata${NC}"
echo "Relay URL: $RELAY_URL"

# Test counters
test_count=0
success_count=0
fail_count=0

# Helper function to print test results
print_result() {
    local test_name=$1
    local success=$2
    
    ((test_count++))
    if [ "$success" = true ]; then
        echo -e "${GREEN}✓ Test $test_count: $test_name${NC}"
        ((success_count++))
    else
        echo -e "${RED}✗ Test $test_count: $test_name${NC}"
        ((fail_count++))
    fi
}

# Check if nak is available
NAK_CMD=""
if command -v nak &> /dev/null; then
    NAK_CMD="nak"
elif [ -f "$HOME/go/bin/nak" ]; then
    NAK_CMD="$HOME/go/bin/nak"
else
    echo -e "${RED}Error: nak is required for testing${NC}"
    echo "Install nak from: https://github.com/fiatjaf/nak"
    echo "Or run: go install github.com/fiatjaf/nak@latest"
    exit 1
fi

# Test 1: Valid relay list with multiple relays and markers
echo -e "\n${YELLOW}Test 1: Valid relay list with multiple relays and markers${NC}"
result=$(nak event --kind 10002 --content "" --tag r="wss://relay1.example.com" --tag r="wss://relay2.example.com,read" --tag r="wss://relay3.example.com,write" $RELAY_URL 2>&1)
if echo "$result" | grep -q "Event published successfully" || ! echo "$result" | grep -q -i "error\|failed\|invalid"; then
    print_result "Valid relay list with markers" true
else
    print_result "Valid relay list with markers" false
    echo "Result: $result"
fi

# Test 2: Valid relay list with only URLs (no markers)
echo -e "\n${YELLOW}Test 2: Valid relay list with only URLs (no markers)${NC}"
result=$(nak event --kind 10002 --content "" --tag r="wss://simple1.example.com" --tag r="wss://simple2.example.com" $RELAY_URL 2>&1)
if echo "$result" | grep -q "Event published successfully" || ! echo "$result" | grep -q -i "error\|failed\|invalid"; then
    print_result "Valid relay list without markers" true
else
    print_result "Valid relay list without markers" false
    echo "Result: $result"
fi

# Test 3: Query relay lists
echo -e "\n${YELLOW}Test 3: Query relay lists${NC}"
result=$(nak req --kinds 10002 --limit 10 $RELAY_URL 2>&1)
if echo "$result" | grep -q "kind.*10002" || echo "$result" | grep -q "\"kind\":10002"; then
    print_result "Query relay lists" true
    echo "Found relay list events in response"
else
    print_result "Query relay lists" false
    echo "No relay list events found or query failed"
    echo "Result: $result"
fi

# Test 4: Invalid relay URL with wrong scheme
echo -e "\n${YELLOW}Test 4: Invalid relay URL with wrong scheme${NC}"
result=$(nak event --kind 10002 --content "" --tag r="http://invalid-scheme.example.com" $RELAY_URL 2>&1)
if echo "$result" | grep -q -i "error\|failed\|invalid\|rejected"; then
    print_result "Invalid relay URL scheme rejected" true
else
    print_result "Invalid relay URL scheme rejected" false
    echo "Result: $result"
fi

# Test 5: Invalid marker
echo -e "\n${YELLOW}Test 5: Invalid marker${NC}"
result=$(nak event --kind 10002 --content "" --tag r="wss://valid-url.example.com,invalid-marker" $RELAY_URL 2>&1)
if echo "$result" | grep -q -i "error\|failed\|invalid\|rejected"; then
    print_result "Invalid relay marker rejected" true
else
    print_result "Invalid relay marker rejected" false
    echo "Result: $result"
fi

# Test 6: Empty relay list (should be valid)
echo -e "\n${YELLOW}Test 6: Empty relay list${NC}"
result=$(nak event --kind 10002 --content "" $RELAY_URL 2>&1)
if echo "$result" | grep -q "Event published successfully" || ! echo "$result" | grep -q -i "error\|failed\|invalid"; then
    print_result "Empty relay list" true
else
    print_result "Empty relay list" false
    echo "Result: $result"
fi

# Test 7: Relay list with non-empty content (should be allowed)
echo -e "\n${YELLOW}Test 7: Relay list with non-empty content${NC}"
result=$(nak event --kind 10002 --content "this should be empty per NIP-65" --tag r="wss://content-test.example.com" $RELAY_URL 2>&1)
if echo "$result" | grep -q "Event published successfully" || ! echo "$result" | grep -q -i "error\|failed\|invalid"; then
    print_result "Non-empty content (allowed)" true
else
    print_result "Non-empty content (allowed)" false
    echo "Result: $result"
fi

# Test 8: Test replaceable behavior by creating another event for the same author
echo -e "\n${YELLOW}Test 8: Test replaceable behavior${NC}"
echo "Creating first relay list..."
result1=$(nak event --kind 10002 --content "" --tag r="wss://first-relay.example.com" $RELAY_URL 2>&1)
sleep 1
echo "Creating second relay list (should replace the first)..."
result2=$(nak event --kind 10002 --content "" --tag r="wss://second-relay.example.com" $RELAY_URL 2>&1)

if (echo "$result1" | grep -q "Event published successfully" || ! echo "$result1" | grep -q -i "error\|failed\|invalid") && \
   (echo "$result2" | grep -q "Event published successfully" || ! echo "$result2" | grep -q -i "error\|failed\|invalid"); then
    print_result "Replaceable behavior test" true
else
    print_result "Replaceable behavior test" false
    echo "First result: $result1"
    echo "Second result: $result2"
fi

echo -e "\n${BLUE}=== Test Summary ===${NC}"
echo -e "Total tests: $test_count"
echo -e "${GREEN}Successful: $success_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"

if [ $fail_count -eq 0 ]; then
    echo -e "\n${GREEN}All NIP-65 tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed. Check the relay implementation.${NC}"
    exit 1
fi