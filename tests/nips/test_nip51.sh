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
RELAY="ws://localhost:8081"
# RELAY="wss://shu02.shugur.net"

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

# Helper function to check if event was accepted
check_event_accepted() {
    local response=$1
    # Check if the response contains an acceptance message (success)
    if [[ "$response" == *"publishing to ws://localhost:8081... success"* ]]; then
        return 0  # success
    else
        return 1  # failure
    fi
}

# Helper function to check if event was rejected
check_event_rejected() {
    local response=$1
    # Check if the response contains a failure message (rejection)
    if [[ "$response" == *"publishing to ws://localhost:8081... failed"* ]]; then
        return 0  # successfully rejected
    else
        return 1  # not rejected (should have been)
    fi
}

# Helper function to extract event ID from nak output
extract_event_id() {
    local output=$1
    # Try to extract event ID from various possible formats
    echo "$output" | grep -o '[a-f0-9]\{64\}' | head -1
}

# Check if nak is available
if ! command -v nak &> /dev/null; then
    echo -e "${RED}Error: 'nak' command not found. Please install it first.${NC}"
    echo -e "${YELLOW}Install with: go install github.com/fiatjaf/nak@latest${NC}"
    exit 1
fi

echo -e "${BLUE}Starting Shugur Relay NIP-51 (Lists) Tests${NC}\n"
echo -e "${YELLOW}Testing NIP-51: Lists (Standard Lists and Sets)${NC}\n"

# Generate test pubkeys
DUMMY_PUBKEY1="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
DUMMY_PUBKEY2="c6047f9441ed7d6d3045406e95c07cd85c778e4b8cef3ca7abac09b95c709ee5"

# Test 1: Mute List (Kind 10000)
echo -e "Test 1: Mute List Creation (Kind 10000)..."
MUTE_RESPONSE=$(nak event -k 10000 --content "" -t p="$DUMMY_PUBKEY1" -t p="$DUMMY_PUBKEY2" -t t="spam" -t word="badword" $RELAY 2>&1)

if check_event_accepted "$MUTE_RESPONSE"; then
    print_result "Create mute list with p, t, and word tags" true "51"
else
    print_result "Create mute list with p, t, and word tags" false "51"
    echo -e "   Response: $MUTE_RESPONSE"
fi

# Test 2: Pinned Notes (Kind 10001)
echo -e "\nTest 2: Pinned Notes List (Kind 10001)..."
EVENT_ID1="d78ba0d5dce22bfff9db0a9e996c9ef27e2c91051de0c4e1da340e0326b4941e"
EVENT_ID2="f27e2c91051de0c4e1da0d5dce22bfff9db0a9340e0326b4941ed78bae996c9e"

PINNED_RESPONSE=$(nak event -k 10001 --content "" -t e="$EVENT_ID1" -t e="$EVENT_ID2" $RELAY 2>&1)

if check_event_accepted "$PINNED_RESPONSE"; then
    print_result "Create pinned notes list with event references" true "51"
else
    print_result "Create pinned notes list with event references" false "51"
    echo -e "   Response: $PINNED_RESPONSE"
fi

# Test 3: Bookmarks (Kind 10003)
echo -e "\nTest 3: Bookmarks List (Kind 10003)..."
BOOKMARKS_RESPONSE=$(nak event -k 10003 --content "" -t e="$EVENT_ID1" -t a="30023:$DUMMY_PUBKEY1:article1" -t t="nostr" -t r="https://example.com/article" $RELAY 2>&1)

if check_event_accepted "$BOOKMARKS_RESPONSE"; then
    print_result "Create bookmarks with e, a, t, r tags" true "51"
else
    print_result "Create bookmarks with e, a, t, r tags" false "51"
    echo -e "   Response: $BOOKMARKS_RESPONSE"
fi

# Test 4: Communities (Kind 10004)
echo -e "\nTest 4: Communities List (Kind 10004)..."
COMMUNITIES_RESPONSE=$(nak event -k 10004 --content "" -t a="34550:$DUMMY_PUBKEY1:community1" -t a="34550:$DUMMY_PUBKEY2:community2" $RELAY 2>&1)

if check_event_accepted "$COMMUNITIES_RESPONSE"; then
    print_result "Create communities list with community references" true "51"
else
    print_result "Create communities list with community references" false "51"
    echo -e "   Response: $COMMUNITIES_RESPONSE"
fi

# Test 5: Blocked Relays (Kind 10006)
echo -e "\nTest 5: Blocked Relays List (Kind 10006)..."
BLOCKED_RESPONSE=$(nak event -k 10006 --content "" -t relay="wss://spam-relay.com" -t relay="wss://bad-relay.net" $RELAY 2>&1)

if check_event_accepted "$BLOCKED_RESPONSE"; then
    print_result "Create blocked relays list" true "51"
else
    print_result "Create blocked relays list" false "51"
    echo -e "   Response: $BLOCKED_RESPONSE"
fi

# Test 6: Interests (Kind 10015)
echo -e "\nTest 6: Interests List (Kind 10015)..."
INTERESTS_RESPONSE=$(nak event -k 10015 --content "" -t t="bitcoin" -t t="nostr" -t t="lightning" -t a="30015:$DUMMY_PUBKEY1:crypto" $RELAY 2>&1)

if check_event_accepted "$INTERESTS_RESPONSE"; then
    print_result "Create interests list with hashtags and sets" true "51"
else
    print_result "Create interests list with hashtags and sets" false "51"
    echo -e "   Response: $INTERESTS_RESPONSE"
fi

# Test 7: Follow Set (Kind 30000) - Requires 'd' tag
echo -e "\nTest 7: Follow Set (Kind 30000)..."
FOLLOW_SET_RESPONSE=$(nak event -k 30000 --content "" -t d="crypto-follows" -t title="Crypto Enthusiasts" -t p="$DUMMY_PUBKEY1" -t p="$DUMMY_PUBKEY2" $RELAY 2>&1)

if check_event_accepted "$FOLLOW_SET_RESPONSE"; then
    print_result "Create follow set with d tag and metadata" true "51"
else
    print_result "Create follow set with d tag and metadata" false "51"
    echo -e "   Response: $FOLLOW_SET_RESPONSE"
fi

# Test 8: Relay Set (Kind 30002) - Requires 'd' tag
echo -e "\nTest 8: Relay Set (Kind 30002)..."
RELAY_SET_RESPONSE=$(nak event -k 30002 --content "" -t d="my-relays" -t title="My Relay Set" -t relay="wss://relay1.com" -t relay="wss://relay2.com" $RELAY 2>&1)

if check_event_accepted "$RELAY_SET_RESPONSE"; then
    print_result "Create relay set with relays and metadata" true "51"
else
    print_result "Create relay set with relays and metadata" false "51"
    echo -e "   Response: $RELAY_SET_RESPONSE"
fi

# Test 9: Bookmark Set (Kind 30003) - Requires 'd' tag
echo -e "\nTest 9: Bookmark Set (Kind 30003)..."
BOOKMARK_SET_RESPONSE=$(nak event -k 30003 --content "" -t d="tech-bookmarks" -t title="Tech Articles" -t description="Collection of tech articles" -t e="$EVENT_ID1" -t a="30023:$DUMMY_PUBKEY1:article1" -t r="https://example.com" $RELAY 2>&1)

if check_event_accepted "$BOOKMARK_SET_RESPONSE"; then
    print_result "Create bookmark set with mixed content types" true "51"
else
    print_result "Create bookmark set with mixed content types" false "51"
    echo -e "   Response: $BOOKMARK_SET_RESPONSE"
fi

# Test 10: Curation Set (Kind 30004) - Requires 'd' tag
echo -e "\nTest 10: Curation Set (Kind 30004)..."
CURATION_SET_RESPONSE=$(nak event -k 30004 --content "" -t d="best-articles" -t title="Best Articles" -t description="Curated collection of great articles" -t a="30023:$DUMMY_PUBKEY1:article1" -t e="$EVENT_ID1" $RELAY 2>&1)

if check_event_accepted "$CURATION_SET_RESPONSE"; then
    print_result "Create curation set with articles and notes" true "51"
else
    print_result "Create curation set with articles and notes" false "51"
    echo -e "   Response: $CURATION_SET_RESPONSE"
fi

# Test 11: Kind Mute Set (Kind 30007) - 'd' tag must be kind string
echo -e "\nTest 11: Kind Mute Set (Kind 30007)..."
KIND_MUTE_RESPONSE=$(nak event -k 30007 --content "" -t d="1" -t p="$DUMMY_PUBKEY1" -t p="$DUMMY_PUBKEY2" $RELAY 2>&1)

if check_event_accepted "$KIND_MUTE_RESPONSE"; then
    print_result "Create kind mute set with kind string as d tag" true "51"
else
    print_result "Create kind mute set with kind string as d tag" false "51"
    echo -e "   Response: $KIND_MUTE_RESPONSE"
fi

# Test 12: Interest Set (Kind 30015) - Requires 'd' tag
echo -e "\nTest 12: Interest Set (Kind 30015)..."
INTEREST_SET_RESPONSE=$(nak event -k 30015 --content "" -t d="crypto" -t title="Crypto Topics" -t t="bitcoin" -t t="nostr" -t t="lightning" $RELAY 2>&1)

if check_event_accepted "$INTEREST_SET_RESPONSE"; then
    print_result "Create interest set with hashtags" true "51"
else
    print_result "Create interest set with hashtags" false "51"
    echo -e "   Response: $INTEREST_SET_RESPONSE"
fi

# Test 13: Emoji Set (Kind 30030) - Requires 'd' tag
echo -e "\nTest 13: Emoji Set (Kind 30030)..."
EMOJI_SET_RESPONSE=$(nak event -k 30030 --content "" -t d="reactions" -t title="Reaction Emojis" -t emoji="🚀" -t emoji="⚡" -t emoji="💜" $RELAY 2>&1)

if check_event_accepted "$EMOJI_SET_RESPONSE"; then
    print_result "Create emoji set with emojis" true "51"
else
    print_result "Create emoji set with emojis" false "51"
    echo -e "   Response: $EMOJI_SET_RESPONSE"
fi

# Test 14: Starter Pack (Kind 39089) - Requires 'd' tag
echo -e "\nTest 14: Starter Pack (Kind 39089)..."
STARTER_PACK_RESPONSE=$(nak event -k 39089 --content "" -t d="nostr-beginners" -t title="Nostr for Beginners" -t description="Great accounts to follow when starting with Nostr" -t p="$DUMMY_PUBKEY1" -t p="$DUMMY_PUBKEY2" $RELAY 2>&1)

if check_event_accepted "$STARTER_PACK_RESPONSE"; then
    print_result "Create starter pack with recommended pubkeys" true "51"
else
    print_result "Create starter pack with recommended pubkeys" false "51"
    echo -e "   Response: $STARTER_PACK_RESPONSE"
fi

# Error Tests - These should fail

# Test 15: Invalid Set without 'd' tag (should fail)
echo -e "\nTest 15: Invalid Follow Set without 'd' tag (should fail)..."
INVALID_SET_RESPONSE=$(nak event -k 30000 --content "" -t title="Invalid Set" -t p="$DUMMY_PUBKEY1" $RELAY 2>&1)

if check_event_rejected "$INVALID_SET_RESPONSE"; then
    print_result "Invalid set without d tag correctly rejected" true "51"
else
    print_result "Invalid set without d tag should have been rejected" false "51"
    echo -e "   Response: $INVALID_SET_RESPONSE"
fi

# Test 16: Invalid mute word (uppercase, should fail)
echo -e "\nTest 16: Invalid mute word (uppercase, should fail)..."
INVALID_WORD_RESPONSE=$(nak event -k 10000 --content "" -t word="BADWORD" $RELAY 2>&1)

if check_event_rejected "$INVALID_WORD_RESPONSE"; then
    print_result "Invalid uppercase word correctly rejected" true "51"
else
    print_result "Invalid uppercase word should have been rejected" false "51"
    echo -e "   Response: $INVALID_WORD_RESPONSE"
fi

# Test 17: Invalid relay URL (should fail)
echo -e "\nTest 17: Invalid relay URL (should fail)..."
INVALID_RELAY_RESPONSE=$(nak event -k 10006 --content "" -t relay="not-a-websocket-url" $RELAY 2>&1)

if check_event_rejected "$INVALID_RELAY_RESPONSE"; then
    print_result "Invalid relay URL correctly rejected" true "51"
else
    print_result "Invalid relay URL should have been rejected" false "51"
    echo -e "   Response: $INVALID_RELAY_RESPONSE"
fi

# Test 18: Encrypted Content Support
echo -e "\nTest 18: Mute List with Encrypted Content..."
ENCRYPTED_CONTENT="TJob1dQrf2ndsmdbeGU+05HT5GMnBSx3fx8QdDY/g3NvCa7klfzgaQCmRZuo1d3WQjHDOjzSY1+MgTK5WjewFFumCcOZniWtOMSga9tJk1ky00tLoUUzyLnb1v9x95h/iT/KpkICJyAwUZ+LoJBUzLrK52wNTMt8M5jSLvCkRx8C0BmEwA/00pjOp4eRndy19H4WUUehhjfV2/VV/k4hMAjJ7Bb5Hp9xdmzmCLX9+64+MyeIQQjQAHPj8dkSsRahP7KS3MgMpjaF8nL48Bg5"

ENCRYPTED_RESPONSE=$(nak event -k 10000 --content "$ENCRYPTED_CONTENT" -t p="$DUMMY_PUBKEY1" $RELAY 2>&1)

if check_event_accepted "$ENCRYPTED_RESPONSE"; then
    print_result "Mute list with encrypted content accepted" true "51"
else
    print_result "Mute list with encrypted content accepted" false "51"
    echo -e "   Response: $ENCRYPTED_RESPONSE"
fi

# Print Summary
echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}           Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total Tests: ${YELLOW}$test_count${NC}"
echo -e "Passed: ${GREEN}$success_count${NC}"
echo -e "Failed: ${RED}$fail_count${NC}"

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed! NIP-51 implementation is working correctly.${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please review the NIP-51 implementation.${NC}"
    exit 1
fi