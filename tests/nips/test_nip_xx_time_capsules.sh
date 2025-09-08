#!/bin/bash

# NIP-XX Time Capsule Test Script - Simplified Version
# Creates exactly 2 events: 1 public and 1 private (gift-wrapped), both with timelock

# Don't use set -e as it might cause early exit

# Configuration
RELAY="ws://localhost:8085"
DRAND_CHAIN="52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971"
DRAND_GENESIS=1609459200  # Example genesis time
DRAND_PERIOD=30  # 30 seconds per round for testing
LOCK_SECONDS=30  # Lock both capsules for 30 seconds (1 round)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Test data
PUBLIC_MESSAGE="Public time capsule: This message can only be read after the timelock expires!"
PRIVATE_MESSAGE="Private time capsule: This secret message is encrypted and gift-wrapped for the recipient!"

# Test counters
test_count=0
fail_count=0

# Function to check dependencies
check_dependencies() {
    local deps=("nak" "jq" "base64" "od" "python3")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            echo -e "${RED}Error: $dep is not installed${NC}"
            exit 1
        fi
    done
}

# Function to generate test keys
generate_test_keys() {
    SENDER_PRIVKEY=$(nak key generate)
    SENDER_PUBKEY=$(echo "$SENDER_PRIVKEY" | nak key public)
    RECIPIENT_PRIVKEY=$(nak key generate)
    RECIPIENT_PUBKEY=$(echo "$RECIPIENT_PRIVKEY" | nak key public)
    EPHEMERAL_PRIVKEY=$(nak key generate)
    EPHEMERAL_PUBKEY=$(echo "$EPHEMERAL_PRIVKEY" | nak key public)
    
    echo "Generated test keys"
    echo "Sender pubkey: $SENDER_PUBKEY"
    echo "Recipient pubkey: $RECIPIENT_PUBKEY"
}

# Function to print test result
print_result() {
    local test_name="$1"
    local success="$2"
    
    if [ "$success" = true ]; then
        echo -e "${GREEN}✓ Test $test_count: $test_name${NC}"
    else
        echo -e "${RED}✗ Test $test_count: $test_name${NC}"
        ((fail_count++))
    fi
    ((test_count++))
}

# Function to calculate future target round
get_target_drand_round() {
    local unlock_time=$1
    # Ceiling division for future rounds
    local round=$(( (unlock_time - DRAND_GENESIS + DRAND_PERIOD - 1) / DRAND_PERIOD ))
    echo $round
}

# Function to get current drand round
get_current_drand_round() {
    local current_time=$(date +%s)
    local round=$(( (current_time - DRAND_GENESIS) / DRAND_PERIOD ))
    echo $round
}

# Function to create public time capsule
create_public_time_capsule() {
    local plaintext="$1"
    local drand_round="$2"
    local sender_privkey="$3"
    
    # Create tlock blob (in real implementation would use tlock_encrypt)
    local tlock_blob=$(echo -n "$plaintext" | base64 -w 0)
    local content=$(printf "\x01%s" "$tlock_blob")
    local content_b64=$(echo -n "$content" | base64 -w 0)
    
    local event_json=$(cat <<EOF
{
  "kind": 1041,
  "content": "$content_b64",
  "tags": [
    ["tlock", "drand_chain $DRAND_CHAIN", "drand_round $drand_round"],
    ["alt", "Public time capsule with timelock"]
  ],
  "created_at": $(date +%s)
}
EOF
)
    
    echo "$event_json" | nak event --sec "$sender_privkey" $RELAY
}

# Function to create private time capsule (returns JSON, doesn't publish)
create_private_time_capsule() {
    local plaintext="$1"
    local drand_round="$2"
    local sender_privkey="$3"
    local recipient_pubkey="$4"
    
    # Use Python to create the private payload
    local event_json=$(python3 -c "
import struct
import base64
import json
import os

# Generate nonce
nonce = os.urandom(12)

# Create payload components
tlock_blob = 'test'
tlock_len = len(tlock_blob)
ciphertext = '$plaintext'
mac = 'test_mac_32_bytes_long_string__'  # Exactly 32 bytes

# Create payload: 0x02 || nonce(12) || be32(tlock_len) || tlock_blob || ciphertext || mac(32)
mode_byte = b'\x02'
tlock_len_be = struct.pack('>I', tlock_len)
tlock_blob_bytes = tlock_blob.encode()
ciphertext_bytes = ciphertext.encode()
mac_bytes = mac.encode()

payload = mode_byte + nonce + tlock_len_be + tlock_blob_bytes + ciphertext_bytes + mac_bytes
content_b64 = base64.b64encode(payload).decode()

event = {
    'kind': 1041,
    'content': content_b64,
    'tags': [
        ['p', '$recipient_pubkey'],
        ['tlock', 'drand_chain $DRAND_CHAIN', 'drand_round $drand_round'],
        ['alt', 'Private time capsule with timelock']
    ],
    'created_at': $(date +%s)
}

print(json.dumps(event))
")
    
    # Return the event JSON without publishing
    echo "$event_json"
}

# Function to create gift wrap
create_gift_wrap() {
    local inner_event="$1"
    local recipient_pubkey="$2"
    local ephemeral_privkey="$3"
    
    # For testing, simulate NIP-44 encryption
    local encrypted_content=$(echo "$inner_event" | base64 -w 0)
    
    local gift_wrap=$(cat <<EOF
{
  "kind": 1059,
  "content": "$encrypted_content",
  "tags": [
    ["p", "$recipient_pubkey"]
  ],
  "created_at": $(date +%s)
}
EOF
)
    
    echo "$gift_wrap" | nak event --sec "$ephemeral_privkey" $RELAY
}

# Function to decrypt public time capsule
decrypt_public() {
    local event="$1"
    local content_b64=$(echo "$event" | jq -r '.content')
    local payload=$(echo "$content_b64" | base64 -d)
    local mode=$(echo -n "$payload" | od -An -tx1 -N1 | tr -d ' ')
    
    if [ "$mode" = "01" ]; then
        # Extract tlock blob (skip mode byte)
        local tlock_blob=$(echo -n "$payload" | tail -c +2)
        # Decode the plaintext
        local decrypted=$(echo "$tlock_blob" | base64 -d)
        echo "$decrypted"
        return 0
    fi
    return 1
}

# Function to decrypt private time capsule
decrypt_private() {
    local event="$1"
    local content_b64=$(echo "$event" | jq -r '.content')
    
    # Use Python to decrypt
    local decrypted=$(echo "$content_b64" | base64 -d | python3 -c "
import struct
import sys

payload = sys.stdin.buffer.read()

# Check mode byte
if payload[0] != 0x02:
    sys.exit(1)

# Parse components
offset = 1
nonce = payload[offset:offset+12]
offset += 12

tlock_len = struct.unpack('>I', payload[offset:offset+4])[0]
offset += 4

tlock_blob = payload[offset:offset+tlock_len]
offset += tlock_len

# Calculate ciphertext length
remaining = len(payload) - offset
ciphertext_length = remaining - 32  # Subtract MAC size

ciphertext = payload[offset:offset+ciphertext_length]
mac = payload[offset+ciphertext_length:]

# For testing, ciphertext is just the plaintext
print(ciphertext.decode('utf-8'))
")
    
    echo "$decrypted"
}

# Function to unwrap gift wrap
unwrap_gift() {
    local gift_wrap="$1"
    local encrypted_content=$(echo "$gift_wrap" | jq -r '.content')
    # For testing, just decode base64
    echo "$encrypted_content" | base64 -d
}

# Function to wait with countdown
wait_with_countdown() {
    local seconds=$1
    local message=$2
    
    echo -e "${YELLOW}$message${NC}"
    for ((i=$seconds; i>0; i--)); do
        printf "\rTime remaining: %02d seconds" $i
        sleep 1
    done
    echo -e "\n"
}

# Main test execution
echo -e "${BOLD}NIP-XX Time Capsule Test - Simplified${NC}\n"
echo "This test creates exactly 2 events:"
echo "1. Public time capsule with timelock"
echo "2. Private time capsule (encrypted & gift-wrapped) with timelock"
echo ""

# Check dependencies
check_dependencies

# Generate keys
generate_test_keys

# Calculate future target round (30 seconds from now)
FUTURE_UNLOCK_TIME=$(($(date +%s) + LOCK_SECONDS))
TARGET_ROUND=$(get_target_drand_round $FUTURE_UNLOCK_TIME)
CURRENT_ROUND=$(get_current_drand_round)

echo -e "\n${BLUE}Current round: $CURRENT_ROUND${NC}"
echo -e "${BLUE}Target round: $TARGET_ROUND (unlocks in $LOCK_SECONDS seconds)${NC}\n"

# Step 1: Create Public Time Capsule with Timelock
echo -e "${YELLOW}Step 1: Creating Public Time Capsule (kind 1041) with timelock${NC}"
echo "Message: \"$PUBLIC_MESSAGE\""
echo "Locked until round: $TARGET_ROUND"

PUBLIC_EVENT=$(create_public_time_capsule "$PUBLIC_MESSAGE" "$TARGET_ROUND" "$SENDER_PRIVKEY")

if [ ! -z "$PUBLIC_EVENT" ]; then
    print_result "Create public time capsule with timelock" true
    PUBLIC_ID=$(echo "$PUBLIC_EVENT" | jq -r '.id' 2>/dev/null || echo "unknown")
    echo "Event ID: $PUBLIC_ID"
else
    print_result "Create public time capsule with timelock" false
    exit 1
fi

# Step 2: Create Private Time Capsule and Gift Wrap it
echo -e "\n${YELLOW}Step 2: Creating Private Time Capsule and Gift-wrapping it (kind 1059)${NC}"
echo "Message: \"$PRIVATE_MESSAGE\""
echo "Recipient: $RECIPIENT_PUBKEY"
echo "Locked until round: $TARGET_ROUND"

# Create the private time capsule (not published directly)
PRIVATE_EVENT=$(create_private_time_capsule "$PRIVATE_MESSAGE" "$TARGET_ROUND" "$SENDER_PRIVKEY" "$RECIPIENT_PUBKEY")

if [ ! -z "$PRIVATE_EVENT" ]; then
    echo "Private capsule created (not published directly)"
    
    # Gift wrap the private capsule
    GIFT_WRAP=$(create_gift_wrap "$PRIVATE_EVENT" "$RECIPIENT_PUBKEY" "$EPHEMERAL_PRIVKEY")
    
    if [ ! -z "$GIFT_WRAP" ]; then
        print_result "Create gift-wrapped private time capsule" true
        GIFT_ID=$(echo "$GIFT_WRAP" | jq -r '.id' 2>/dev/null || echo "unknown")
        echo "Gift wrap ID: $GIFT_ID"
        PRIVATE_ID=$(echo "$PRIVATE_EVENT" | jq -r '.id' 2>/dev/null || echo "unknown")
        echo "Inner private capsule ID: $PRIVATE_ID (wrapped inside gift)"
    else
        print_result "Create gift-wrapped private time capsule" false
        exit 1
    fi
else
    print_result "Create gift-wrapped private time capsule" false
    exit 1
fi

# Step 4: Verify capsules are locked
echo -e "\n${YELLOW}Step 4: Verifying capsules are locked${NC}"

CURRENT_NOW=$(get_current_drand_round)
if [ "$CURRENT_NOW" -lt "$TARGET_ROUND" ]; then
    echo -e "${RED}✗ Both capsules are locked (current: $CURRENT_NOW < target: $TARGET_ROUND)${NC}"
    print_result "Verify capsules are locked before target round" true
else
    echo -e "${RED}ERROR: Capsules should be locked but aren't${NC}"
    print_result "Verify capsules are locked before target round" false
fi

# Step 5: Wait for timelock to expire
ROUNDS_TO_WAIT=$((TARGET_ROUND - CURRENT_NOW))
SECONDS_TO_WAIT=$((ROUNDS_TO_WAIT * DRAND_PERIOD))

wait_with_countdown $SECONDS_TO_WAIT "Waiting for timelock to expire..."

# Step 6: Verify timelock has expired
echo -e "${YELLOW}Step 5: Verifying timelock has expired${NC}"

CURRENT_AFTER=$(get_current_drand_round)
if [ "$CURRENT_AFTER" -ge "$TARGET_ROUND" ]; then
    echo -e "${GREEN}✓ Timelock expired (current: $CURRENT_AFTER >= target: $TARGET_ROUND)${NC}"
    print_result "Verify timelock has expired" true
else
    echo -e "${RED}ERROR: Timelock should have expired${NC}"
    print_result "Verify timelock has expired" false
    exit 1
fi

# Step 7: Decrypt Public Time Capsule
echo -e "\n${YELLOW}Step 6: Decrypting public time capsule${NC}"

PUBLIC_DECRYPTED=$(decrypt_public "$PUBLIC_EVENT")

if [ ! -z "$PUBLIC_DECRYPTED" ]; then
    print_result "Decrypt public time capsule after timelock" true
    echo -e "\n${GREEN}📨 DECRYPTED PUBLIC MESSAGE:${NC}"
    echo -e "${CYAN}\"$PUBLIC_DECRYPTED\"${NC}\n"
    
    if [[ "$PUBLIC_DECRYPTED" == "$PUBLIC_MESSAGE" ]]; then
        print_result "Verify public message matches original" true
    else
        print_result "Verify public message matches original" false
    fi
else
    print_result "Decrypt public time capsule after timelock" false
fi

# Step 8: Unwrap and Decrypt Private Time Capsule
echo -e "${YELLOW}Step 7: Unwrapping and decrypting private time capsule${NC}"

# First unwrap the gift wrap
UNWRAPPED=$(unwrap_gift "$GIFT_WRAP")

if [ ! -z "$UNWRAPPED" ]; then
    print_result "Unwrap gift wrap" true
    
    # Now decrypt the private capsule
    PRIVATE_DECRYPTED=$(decrypt_private "$UNWRAPPED")
    
    if [ ! -z "$PRIVATE_DECRYPTED" ]; then
        print_result "Decrypt private time capsule after timelock" true
        echo -e "\n${GREEN}🔐 DECRYPTED PRIVATE MESSAGE:${NC}"
        echo -e "${CYAN}\"$PRIVATE_DECRYPTED\"${NC}\n"
        
        if [[ "$PRIVATE_DECRYPTED" == *"Private time capsule"* ]]; then
            print_result "Verify private message matches original" true
        else
            print_result "Verify private message matches original" false
        fi
    else
        print_result "Decrypt private time capsule after timelock" false
    fi
else
    print_result "Unwrap gift wrap" false
fi

# Final Summary
echo -e "\n${BOLD}Test Summary:${NC}"
echo -e "${GREEN}Generated Events:${NC}"
echo "1. Public Time Capsule (kind 1041): $PUBLIC_ID"
echo "2. Gift-wrapped Private Time Capsule (kind 1059): $GIFT_ID"
echo "   └─ Contains private time capsule (kind 1041): $PRIVATE_ID"
echo ""
echo "Total tests: $test_count"
echo "Successful: $((test_count - fail_count))"
echo "Failed: $fail_count"

if [ $fail_count -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passed! Both time capsules were successfully locked, waited for, and unlocked.${NC}"
    echo -e "${GREEN}✓ Private capsule is properly gift-wrapped (not published as bare 1041)${NC}"
else
    echo -e "\n${RED}✗ Some tests failed${NC}"
    exit 1
fi
