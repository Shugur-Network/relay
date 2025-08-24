#!/bin/bash

# Test script for NIP-65: Relay List Metadata (kind 10002)
# Tests relay list metadata events with 'r' tags and empty content

set -e

RELAY_URL="${RELAY_URL:-ws://localhost:8080}"
TEMP_DIR="/tmp/nip65_test_$$"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing NIP-65: Relay List Metadata${NC}"
echo "Relay URL: $RELAY_URL"

# Create temporary directory for test files
mkdir -p "$TEMP_DIR"
cd "$TEMP_DIR"

# Function to cleanup
cleanup() {
    cd /
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

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

# Check if we have required tools
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}Error: python3 is required for testing${NC}"
    exit 1
fi

# Check if we have required Python modules
python3 -c "import websocket, secp256k1" 2>/dev/null || {
    echo -e "${RED}Error: Required Python modules not found. Install with:${NC}"
    echo "pip3 install websocket-client secp256k1"
    exit 1
}

# Function to generate a test keypair
generate_keypair() {
    local name="$1"
    python3 -c "
import secrets
import hashlib
from secp256k1 import PrivateKey

# Generate private key
privkey_bytes = secrets.token_bytes(32)
privkey_hex = privkey_bytes.hex()

# Generate public key
privkey = PrivateKey(privkey_bytes)
pubkey_bytes = privkey.pubkey.serialize(compressed=True)[1:]  # Remove 0x02 prefix
pubkey_hex = pubkey_bytes.hex()

print(f'PRIVATE_KEY_{name.upper()}={privkey_hex}')
print(f'PUBLIC_KEY_{name.upper()}={pubkey_hex}')
"
}

# Generate test keypairs
echo -e "${YELLOW}Generating test keypairs...${NC}"
eval $(generate_keypair "alice")
eval $(generate_keypair "bob")

echo "Alice pubkey: $PUBLIC_KEY_ALICE"
echo "Bob pubkey: $PUBLIC_KEY_BOB"

# Function to create and sign an event
create_event() {
    local kind="$1"
    local content="$2"
    local tags="$3"
    local private_key="$4"
    local public_key="$5"
    
    python3 -c "
import json
import time
import hashlib
from secp256k1 import PrivateKey

# Event data
kind = $kind
content = '$content'
tags = $tags
created_at = int(time.time())
pubkey = '$public_key'

# Create event object for signing
event_data = [0, pubkey, created_at, kind, tags, content]
event_json = json.dumps(event_data, separators=(',', ':'))
event_hash = hashlib.sha256(event_json.encode()).hexdigest()

# Sign the event
privkey = PrivateKey(bytes.fromhex('$private_key'))
sig = privkey.sign(bytes.fromhex(event_hash), hasher=None)
signature = sig.serialize().hex()

# Create final event
event = {
    'id': event_hash,
    'pubkey': pubkey,
    'created_at': created_at,
    'kind': kind,
    'tags': tags,
    'content': content,
    'sig': signature
}

print(json.dumps(event))
"
}

# Function to send event to relay
send_event() {
    local event="$1"
    local expected_ok="$2"
    local test_name="$3"
    
    echo "Sending event for: $test_name"
    
    response=$(timeout 10 python3 -c "
import json
import websocket
import sys

try:
    event = json.loads('$event')
    ws = websocket.create_connection('$RELAY_URL', timeout=5)
    
    # Send EVENT message
    message = json.dumps(['EVENT', event])
    ws.send(message)
    
    # Wait for OK response
    response = ws.recv()
    ws.close()
    
    print(response)
except Exception as e:
    print(f'ERROR: {e}')
    sys.exit(1)
" 2>/dev/null)
    
    echo "Response: $response"
    
    # Parse OK response
    if echo "$response" | grep -q '"OK"'; then
        if echo "$response" | grep -q "$expected_ok"; then
            print_result "$test_name" true
            return 0
        else
            print_result "$test_name" false
            return 1
        fi
    else
        print_result "$test_name" false
        return 1
    fi
}

# Function to query events
query_events() {
    local filter="$1"
    local description="$2"
    
    echo "Querying: $description"
    
    response=$(timeout 10 python3 -c "
import json
import websocket
import time

try:
    filter_obj = json.loads('$filter')
    ws = websocket.create_connection('$RELAY_URL', timeout=5)
    
    # Send REQ message
    sub_id = 'test_sub_' + str(int(time.time()))
    message = json.dumps(['REQ', sub_id, filter_obj])
    ws.send(message)
    
    events = []
    timeout_count = 0
    while timeout_count < 30:  # 3 second timeout
        try:
            response = ws.recv()
            msg = json.loads(response)
            
            if msg[0] == 'EVENT':
                events.append(msg[2])
            elif msg[0] == 'EOSE':
                break
        except:
            timeout_count += 1
            time.sleep(0.1)
    
    ws.close()
    print(json.dumps(events))
except Exception as e:
    print('[]')  # Return empty array on error
" 2>/dev/null)
    
    event_count=$(echo "$response" | python3 -c "import json, sys; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
    echo "Found $event_count events"
    
    return $event_count
}

# Test 1: Valid relay list with multiple relays and markers
echo -e "\n${YELLOW}Test 1: Valid relay list with multiple relays and markers${NC}"
relay_tags='[["r", "wss://relay1.example.com"], ["r", "wss://relay2.example.com", "read"], ["r", "wss://relay3.example.com", "write"]]'
event1=$(create_event 10002 "" "$relay_tags" "$PRIVATE_KEY_ALICE" "$PUBLIC_KEY_ALICE")
send_event "$event1" "true" "Valid relay list with markers"

# Test 2: Valid relay list with only URLs (no markers)
echo -e "\n${YELLOW}Test 2: Valid relay list with only URLs (no markers)${NC}"
simple_tags='[["r", "wss://simple1.example.com"], ["r", "wss://simple2.example.com"]]'
event2=$(create_event 10002 "" "$simple_tags" "$PRIVATE_KEY_BOB" "$PUBLIC_KEY_BOB")
send_event "$event2" "true" "Valid relay list without markers"

# Test 3: Query Alice's relay list
echo -e "\n${YELLOW}Test 3: Query Alice's relay list${NC}"
filter='{"kinds": [10002], "authors": ["'$PUBLIC_KEY_ALICE'"]}'
query_events "$filter" "Alice's relay list"
alice_events=$?

if [ $alice_events -eq 1 ]; then
    print_result "Query Alice's relay list" true
else
    print_result "Query Alice's relay list" false
fi

# Test 4: Replaceable event - update Alice's relay list
echo -e "\n${YELLOW}Test 4: Update Alice's relay list (replaceable behavior)${NC}"
sleep 2  # Ensure different timestamp
new_relay_tags='[["r", "wss://new-relay.example.com"], ["r", "wss://backup-relay.example.com", "read"]]'
event3=$(create_event 10002 "" "$new_relay_tags" "$PRIVATE_KEY_ALICE" "$PUBLIC_KEY_ALICE")
send_event "$event3" "true" "Update Alice's relay list"

# Verify replacement worked
echo "Verifying relay list replacement..."
sleep 1
query_events "$filter" "Alice's updated relay list"
alice_events_after=$?

if [ $alice_events_after -eq 1 ]; then
    print_result "Replaceable event behavior" true
else
    print_result "Replaceable event behavior" false
fi

# Test 5: Query all relay lists
echo -e "\n${YELLOW}Test 5: Query all relay lists${NC}"
all_filter='{"kinds": [10002]}'
query_events "$all_filter" "All relay lists"
total_events=$?

if [ $total_events -eq 2 ]; then
    print_result "Query all relay lists" true
else
    print_result "Query all relay lists" false
fi

# Test 6: Invalid relay URL with wrong scheme
echo -e "\n${YELLOW}Test 6: Invalid relay URL with wrong scheme${NC}"
invalid_scheme_tags='[["r", "http://invalid-scheme.example.com"]]'
event4=$(create_event 10002 "" "$invalid_scheme_tags" "$PRIVATE_KEY_ALICE" "$PUBLIC_KEY_ALICE")
send_event "$event4" "false" "Invalid relay URL scheme"

# Test 7: Invalid marker
echo -e "\n${YELLOW}Test 7: Invalid marker${NC}"
invalid_marker_tags='[["r", "wss://valid-url.example.com", "invalid-marker"]]'
event5=$(create_event 10002 "" "$invalid_marker_tags" "$PRIVATE_KEY_ALICE" "$PUBLIC_KEY_ALICE")
send_event "$event5" "false" "Invalid relay marker"

# Test 8: Malformed r tag (missing URL)
echo -e "\n${YELLOW}Test 8: Malformed r tag (missing URL)${NC}"
malformed_tags='[["r"]]'
event6=$(create_event 10002 "" "$malformed_tags" "$PRIVATE_KEY_ALICE" "$PUBLIC_KEY_ALICE")
send_event "$event6" "false" "Malformed r tag"

# Test 9: Valid event with non-empty content (allowed but not recommended)
echo -e "\n${YELLOW}Test 9: Valid event with non-empty content${NC}"
content_tags='[["r", "wss://content-test.example.com"]]'
event7=$(create_event 10002 "this should be empty per NIP-65" "$content_tags" "$PRIVATE_KEY_BOB" "$PUBLIC_KEY_BOB")
send_event "$event7" "true" "Non-empty content (allowed)"

# Test 10: Empty event (no r tags)
echo -e "\n${YELLOW}Test 10: Empty event (no r tags)${NC}"
empty_tags='[]'
event8=$(create_event 10002 "" "$empty_tags" "$PRIVATE_KEY_ALICE" "$PUBLIC_KEY_ALICE")
send_event "$event8" "true" "Empty relay list"

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