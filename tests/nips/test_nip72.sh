#!/bin/bash

# Test script for NIP-72 Moderated Communities (Reddit Style)
# Tests community definitions (34550), community posts (1111), and approval events (4550)

# Color definitions for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

RELAY_URL="ws://localhost:8081"
NIP="NIP-72"

# Test counter
test_count=0
failed_count=0
passed_count=0

# Generate test private keys and other test data consistently
TEST_PRIVATE_KEY="0000000000000000000000000000000000000000000000000000000000000001"
TEST_AUTHOR_PUBKEY="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
TEST_MODERATOR_PUBKEY="024d4b6cd1361032ca9bd2aeb9d900aa4d45d9ead80ac9423374c451a7254d0666"
TEST_POST_ID="1111111111111111111111111111111111111111111111111111111111111111"
TEST_COMMUNITY_ID="bitcoin-dev"

# Helper function to print test results
print_result() {
    local test_name=$1
    local success=$2
    local nip=$3
    
    test_count=$((test_count + 1))
    
    if [ "$success" = true ]; then
        echo -e "  ${GREEN}✓${NC} Test $test_count: $test_name ($nip)"
        passed_count=$((passed_count + 1))
    else
        echo -e "  ${RED}✗${NC} Test $test_count: $test_name ($nip)"
        failed_count=$((failed_count + 1))
    fi
}

# Helper function to run nak and capture output
run_nak_test() {
    local event_data="$1"
    local should_succeed="$2"
    
    # Run nak event and capture the full output including stderr
    output=$(echo "$event_data" | timeout 10s nak event --sec "$TEST_PRIVATE_KEY" $RELAY_URL 2>&1)
    exit_code=$?
    
    if [ "$should_succeed" = true ]; then
        # Event should succeed - check if it was published
        if [ $exit_code -eq 0 ] && [[ ! "$output" =~ "failed" ]] && [[ ! "$output" =~ "validation failed" ]] && [[ ! "$output" =~ "error" ]]; then
            return 0
        else
            echo "    Expected success but got failure: $output"
            return 1
        fi
    else
        # Event should fail - check if it was rejected
        if [ $exit_code -ne 0 ] || [[ "$output" =~ "failed" ]] || [[ "$output" =~ "validation failed" ]] || [[ "$output" =~ "error" ]]; then
            return 0
        else
            echo "    Expected failure but got success: $output"
            return 1
        fi
    fi
}

echo -e "${BLUE}Testing $NIP: Moderated Communities${NC}"
echo

# ========================================
# Valid Community Definition Tests (Kind 34550)
# ========================================

echo -e "${YELLOW}Valid Community Definition Tests${NC}"

# Test 1: Basic community definition
community_basic='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["d", "'$TEST_COMMUNITY_ID'"],
        ["name", "Bitcoin Development"],
        ["description", "Community for Bitcoin protocol development discussions"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "wss://relay.example.com", "moderator"],
        ["p", "'$TEST_MODERATOR_PUBKEY'", "wss://relay2.example.com", "moderator"]
    ]
}'

if run_nak_test "$community_basic" true; then
    print_result "Basic community definition" true "$NIP"
else
    print_result "Basic community definition" false "$NIP"
fi

# Test 2: Community with image and multiple relays
community_with_image='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["d", "nostr-dev"],
        ["name", "Nostr Development"],
        ["description", "Community for Nostr protocol development and implementation"],
        ["image", "https://example.com/nostr-logo.png", "512x512"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "wss://relay.example.com", "moderator"],
        ["p", "'$TEST_MODERATOR_PUBKEY'", "", "moderator"],
        ["relay", "wss://relay.example.com", "author"],
        ["relay", "wss://relay2.example.com", "requests"],
        ["relay", "wss://relay3.example.com", "approvals"],
        ["relay", "wss://relay4.example.com"]
    ]
}'

if run_nak_test "$community_with_image" true; then
    print_result "Community with image and multiple relays" true "$NIP"
else
    print_result "Community with image and multiple relays" false "$NIP"
fi

# Test 3: Community with many moderators
community_many_moderators='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["d", "large-community"],
        ["name", "Large Community"],
        ["description", "A community with multiple moderators"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "", "moderator"],
        ["p", "'$TEST_MODERATOR_PUBKEY'", "", "moderator"],
        ["p", "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", "", "moderator"],
        ["p", "123456789012345678901234567890123456789012345678901234567890abcd", "", "moderator"],
        ["p", "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321", "", "moderator"]
    ]
}'

if run_nak_test "$community_many_moderators" true; then
    print_result "Community with many moderators" true "$NIP"
else
    print_result "Community with many moderators" false "$NIP"
fi

# ========================================
# Valid Community Post Tests (Kind 1111)
# ========================================

echo
echo -e "${YELLOW}Valid Community Post Tests${NC}"

# Test 4: Top-level community post
community_post_top_level='{
    "kind": 1111,
    "content": "Welcome to the Bitcoin development community! Looking forward to great discussions.",
    "tags": [
        ["A", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'", "wss://relay.example.com"],
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'", "wss://relay.example.com"],
        ["P", "'$TEST_AUTHOR_PUBKEY'", "wss://relay.example.com"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "wss://relay.example.com"],
        ["K", "34550"],
        ["k", "34550"]
    ]
}'

if run_nak_test "$community_post_top_level" true; then
    print_result "Top-level community post" true "$NIP"
else
    print_result "Top-level community post" false "$NIP"
fi

# Test 5: Nested reply in community
community_post_reply='{
    "kind": 1111,
    "content": "Great point! I agree with the proposal.",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'", "wss://relay.example.com"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "wss://relay.example.com"],
        ["k", "1111"],
        ["e", "'$TEST_POST_ID'", "wss://relay.example.com"],
        ["p", "'$TEST_MODERATOR_PUBKEY'", "wss://relay.example.com"]
    ]
}'

if run_nak_test "$community_post_reply" true; then
    print_result "Nested reply in community" true "$NIP"
else
    print_result "Nested reply in community" false "$NIP"
fi

# Test 6: Community post without relay URLs
community_post_minimal='{
    "kind": 1111,
    "content": "Minimal community post",
    "tags": [
        ["A", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["P", "'$TEST_AUTHOR_PUBKEY'"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"],
        ["K", "34550"],
        ["k", "34550"]
    ]
}'

if run_nak_test "$community_post_minimal" true; then
    print_result "Minimal community post" true "$NIP"
else
    print_result "Minimal community post" false "$NIP"
fi

# ========================================
# Valid Approval Event Tests (Kind 4550)
# ========================================

echo
echo -e "${YELLOW}Valid Approval Event Tests${NC}"

# Test 7: Basic approval event
approval_basic='{
    "kind": 4550,
    "content": "{\"kind\":1111,\"content\":\"This is an approved post\",\"tags\":[[\"A\",\"34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'\"]]}",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'", "wss://relay.example.com"],
        ["e", "'$TEST_POST_ID'", "wss://relay.example.com"],
        ["p", "'$TEST_MODERATOR_PUBKEY'", "wss://relay.example.com"],
        ["k", "1111"]
    ]
}'

if run_nak_test "$approval_basic" true; then
    print_result "Basic approval event" true "$NIP"
else
    print_result "Basic approval event" false "$NIP"
fi

# Test 8: Approval for multiple communities
approval_multi_community='{
    "kind": 4550,
    "content": "{\"kind\":1111,\"content\":\"Cross-posted content\",\"tags\":[[\"A\",\"34550:'$TEST_AUTHOR_PUBKEY':cross-post\"]]}",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':another-community"],
        ["e", "'$TEST_POST_ID'"],
        ["p", "'$TEST_MODERATOR_PUBKEY'"],
        ["k", "1111"]
    ]
}'

if run_nak_test "$approval_multi_community" true; then
    print_result "Approval for multiple communities" true "$NIP"
else
    print_result "Approval for multiple communities" false "$NIP"
fi

# Test 9: Approval with replaceable event reference
approval_replaceable='{
    "kind": 4550,
    "content": "{\"kind\":30023,\"content\":\"Long form content\",\"tags\":[[\"d\",\"article-1\"]]}",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["a", "30023:'$TEST_MODERATOR_PUBKEY':article-1"],
        ["p", "'$TEST_MODERATOR_PUBKEY'"],
        ["k", "30023"]
    ]
}'

if run_nak_test "$approval_replaceable" true; then
    print_result "Approval with replaceable event reference" true "$NIP"
else
    print_result "Approval with replaceable event reference" false "$NIP"
fi

# ========================================
# Invalid Event Tests
# ========================================

echo
echo -e "${YELLOW}Invalid Event Tests${NC}"

# Test 10: Community definition missing d tag (should fail)
community_no_d='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["name", "Invalid Community"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "", "moderator"]
    ]
}'

if run_nak_test "$community_no_d" false; then
    print_result "Community definition missing d tag (should fail)" true "$NIP"
else
    print_result "Community definition missing d tag (should fail)" false "$NIP"
fi

# Test 11: Community definition with no moderators (should fail)
community_no_moderators='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["d", "no-mods"],
        ["name", "No Moderators Community"]
    ]
}'

if run_nak_test "$community_no_moderators" false; then
    print_result "Community definition with no moderators (should fail)" true "$NIP"
else
    print_result "Community definition with no moderators (should fail)" false "$NIP"
fi

# Test 12: Community definition with invalid moderator pubkey (should fail)
community_invalid_moderator='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["d", "invalid-mod"],
        ["name", "Invalid Moderator Community"],
        ["p", "invalid-pubkey", "", "moderator"]
    ]
}'

if run_nak_test "$community_invalid_moderator" false; then
    print_result "Community definition with invalid moderator pubkey (should fail)" true "$NIP"
else
    print_result "Community definition with invalid moderator pubkey (should fail)" false "$NIP"
fi

# Test 13: Community definition with invalid relay marker (should fail)
community_invalid_relay='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["d", "invalid-relay"],
        ["name", "Invalid Relay Community"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "", "moderator"],
        ["relay", "wss://relay.example.com", "invalid-marker"]
    ]
}'

if run_nak_test "$community_invalid_relay" false; then
    print_result "Community definition with invalid relay marker (should fail)" true "$NIP"
else
    print_result "Community definition with invalid relay marker (should fail)" false "$NIP"
fi

# Test 14: Community post missing A tag (should fail)
community_post_no_A='{
    "kind": 1111,
    "content": "Missing A tag",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["P", "'$TEST_AUTHOR_PUBKEY'"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"],
        ["K", "34550"],
        ["k", "34550"]
    ]
}'

if run_nak_test "$community_post_no_A" false; then
    print_result "Community post missing A tag (should fail)" true "$NIP"
else
    print_result "Community post missing A tag (should fail)" false "$NIP"
fi

# Test 15: Community post missing lowercase tags (should fail)
community_post_no_lowercase='{
    "kind": 1111,
    "content": "Missing lowercase tags",
    "tags": [
        ["A", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["P", "'$TEST_AUTHOR_PUBKEY'"],
        ["K", "34550"]
    ]
}'

if run_nak_test "$community_post_no_lowercase" true; then
    print_result "Community post with uppercase tags (should pass)" true "$NIP"
else
    print_result "Community post with uppercase tags (should pass)" false "$NIP"
fi

# Test 16: Community post with invalid community reference (should fail)
community_post_invalid_ref='{
    "kind": 1111,
    "content": "Invalid community reference",
    "tags": [
        ["A", "invalid:reference:format"],
        ["a", "invalid:reference:format"],
        ["P", "'$TEST_AUTHOR_PUBKEY'"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"],
        ["K", "34550"],
        ["k", "34550"]
    ]
}'

if run_nak_test "$community_post_invalid_ref" false; then
    print_result "Community post with invalid community reference (should fail)" true "$NIP"
else
    print_result "Community post with invalid community reference (should fail)" false "$NIP"
fi

# Test 17: Approval event missing community a tag (should fail)
approval_no_community='{
    "kind": 4550,
    "content": "{\"kind\":1111,\"content\":\"Content\"}",
    "tags": [
        ["e", "'$TEST_POST_ID'"],
        ["p", "'$TEST_MODERATOR_PUBKEY'"],
        ["k", "1111"]
    ]
}'

if run_nak_test "$approval_no_community" false; then
    print_result "Approval event missing community a tag (should fail)" true "$NIP"
else
    print_result "Approval event missing community a tag (should fail)" false "$NIP"
fi

# Test 18: Approval event missing post reference (should fail)
approval_no_post_ref='{
    "kind": 4550,
    "content": "{\"kind\":1111,\"content\":\"Content\"}",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["p", "'$TEST_MODERATOR_PUBKEY'"],
        ["k", "1111"]
    ]
}'

if run_nak_test "$approval_no_post_ref" false; then
    print_result "Approval event missing post reference (should fail)" true "$NIP"
else
    print_result "Approval event missing post reference (should fail)" false "$NIP"
fi

# Test 19: Approval event with empty content (should fail)
approval_empty_content='{
    "kind": 4550,
    "content": "",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["e", "'$TEST_POST_ID'"],
        ["p", "'$TEST_MODERATOR_PUBKEY'"],
        ["k", "1111"]
    ]
}'

if run_nak_test "$approval_empty_content" false; then
    print_result "Approval event with empty content (should fail)" true "$NIP"
else
    print_result "Approval event with empty content (should fail)" false "$NIP"
fi

# Test 20: Approval event with invalid JSON content (should fail)
approval_invalid_json='{
    "kind": 4550,
    "content": "invalid-json-content",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["e", "'$TEST_POST_ID'"],
        ["p", "'$TEST_MODERATOR_PUBKEY'"],
        ["k", "1111"]
    ]
}'

if run_nak_test "$approval_invalid_json" false; then
    print_result "Approval event with invalid JSON content (should fail)" true "$NIP"
else
    print_result "Approval event with invalid JSON content (should fail)" false "$NIP"
fi

# ========================================
# Edge Case Tests
# ========================================

echo
echo -e "${YELLOW}Edge Case Tests${NC}"

# Test 21: Community with very long identifier (should fail)
community_long_id='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["d", "this-is-a-very-long-community-identifier-that-exceeds-the-maximum-allowed-length-of-255-characters-and-should-be-rejected-by-the-validation-system-because-it-is-too-long-for-practical-use-in-a-distributed-system-where-bandwidth-matters"],
        ["name", "Long ID Community"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "", "moderator"]
    ]
}'

if run_nak_test "$community_long_id" false; then
    print_result "Community with very long identifier (should fail)" true "$NIP"
else
    print_result "Community with very long identifier (should fail)" false "$NIP"
fi

# Test 22: Community with invalid image dimensions (should fail)
community_invalid_image='{
    "kind": 34550,
    "content": "",
    "tags": [
        ["d", "invalid-image"],
        ["name", "Invalid Image Community"],
        ["image", "https://example.com/image.png", "not-dimensions"],
        ["p", "'$TEST_AUTHOR_PUBKEY'", "", "moderator"]
    ]
}'

if run_nak_test "$community_invalid_image" false; then
    print_result "Community with invalid image dimensions (should fail)" true "$NIP"
else
    print_result "Community with invalid image dimensions (should fail)" false "$NIP"
fi

# Test 23: Community post with wrong K tag value (should fail)
community_post_wrong_K='{
    "kind": 1111,
    "content": "Wrong K tag",
    "tags": [
        ["A", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["P", "'$TEST_AUTHOR_PUBKEY'"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"],
        ["K", "1111"],
        ["k", "34550"]
    ]
}'

if run_nak_test "$community_post_wrong_K" false; then
    print_result "Community post with wrong K tag value (should fail)" true "$NIP"
else
    print_result "Community post with wrong K tag value (should fail)" false "$NIP"
fi

# Test 24: Valid cross-post (kind 6)
cross_post_basic='{
    "kind": 6,
    "content": "{\"kind\":1111,\"content\":\"Original post content\",\"tags\":[[\"A\",\"34550:'$TEST_AUTHOR_PUBKEY':source-community\"]]}",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["e", "'$TEST_POST_ID'"],
        ["p", "'$TEST_MODERATOR_PUBKEY'"]
    ]
}'

if run_nak_test "$cross_post_basic" true; then
    print_result "Valid cross-post (kind 6)" true "$NIP"
else
    print_result "Valid cross-post (kind 6)" false "$NIP"
fi

# Test 25: Backwards compatibility kind 1 post
backwards_compat_post='{
    "kind": 1,
    "content": "Legacy kind 1 post to community",
    "tags": [
        ["a", "34550:'$TEST_AUTHOR_PUBKEY':'$TEST_COMMUNITY_ID'"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"]
    ]
}'

if run_nak_test "$backwards_compat_post" true; then
    print_result "Backwards compatibility kind 1 post" true "$NIP"
else
    print_result "Backwards compatibility kind 1 post" false "$NIP"
fi

# ========================================
# Summary
# ========================================

echo
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "Total tests: $test_count"
echo -e "Passed: ${GREEN}$passed_count${NC}"
echo -e "Failed: ${RED}$failed_count${NC}"

if [ $failed_count -eq 0 ]; then
    echo -e "${GREEN}All $NIP tests passed! 🎉${NC}"
    exit 0
else
    echo -e "${RED}$failed_count test(s) failed${NC}"
    exit 1
fi