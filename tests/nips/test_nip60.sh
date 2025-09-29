#!/bin/bash

# Test script for NIP-60 Cashu Wallets
# Tests all event types: Wallet Events (17375), Token Events (7375), Spending History (7376), Quote Events (7374)

# Color definitions for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

RELAY_URL="ws://localhost:8081"
NIP="NIP-60"

# Test counter
test_count=0
failed_count=0
passed_count=0

# Generate test private keys and other test data consistently
TEST_PRIVATE_KEY="0000000000000000000000000000000000000000000000000000000000000001"
TEST_AUTHOR_PUBKEY="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
TEST_EVENT_ID_1="1111111111111111111111111111111111111111111111111111111111111111"
TEST_EVENT_ID_2="2222222222222222222222222222222222222222222222222222222222222222"

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

echo -e "${BLUE}Testing $NIP: Cashu Wallets${NC}"
echo

# ========================================
# Valid Wallet Event Tests (Kind 17375)
# ========================================

echo -e "${YELLOW}Valid Wallet Event Tests${NC}"

# Test 1: Basic valid wallet event with encrypted content
wallet_event='{
    "kind": 17375,
    "content": "AgABC123def456ghi789jklmno012pqr345stu678vwx901yzABCDEF234567890",
    "tags": []
}'

if run_nak_test "$wallet_event" true; then
    print_result "Basic valid wallet event" true "$NIP"
else
    print_result "Basic valid wallet event" false "$NIP"
fi

# Test 2: Wallet event with metadata tags
wallet_with_metadata='{
    "kind": 17375,
    "content": "AgABC123def456ghi789jklmno012pqr345stu678vwx901yzABCDEF234567890",
    "tags": [
        ["client", "cashu-wallet-app"],
        ["alt", "Cashu wallet backup"]
    ]
}'

if run_nak_test "$wallet_with_metadata" true; then
    print_result "Wallet event with metadata tags" true "$NIP"
else
    print_result "Wallet event with metadata tags" false "$NIP"
fi

# ========================================
# Valid Token Event Tests (Kind 7375)
# ========================================

echo
echo -e "${YELLOW}Valid Token Event Tests${NC}"

# Test 3: Basic token event
token_event='{
    "kind": 7375,
    "content": "AgABC789xyz123abc456def789ghi012jkl345mno678pqr901stu234vwx567yza",
    "tags": []
}'

if run_nak_test "$token_event" true; then
    print_result "Basic token event" true "$NIP"
else
    print_result "Basic token event" false "$NIP"
fi

# Test 4: Token event with client metadata
token_with_client='{
    "kind": 7375,
    "content": "AgABC789xyz123abc456def789ghi012jkl345mno678pqr901stu234vwx567yza",
    "tags": [
        ["client", "cashu-mobile-wallet"]
    ]
}'

if run_nak_test "$token_with_client" true; then
    print_result "Token event with client metadata" true "$NIP"
else
    print_result "Token event with client metadata" false "$NIP"
fi

# ========================================
# Valid Spending History Event Tests (Kind 7376)
# ========================================

echo
echo -e "${YELLOW}Valid Spending History Event Tests${NC}"

# Test 5: Basic spending history event with public e-tag (redeemed)
spending_history='{
    "kind": 7376,
    "content": "AgABC456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
    "tags": [
        ["e", "'$TEST_EVENT_ID_1'", "", "redeemed"]
    ]
}'

if run_nak_test "$spending_history" true; then
    print_result "Basic spending history with redeemed e-tag" true "$NIP"
else
    print_result "Basic spending history with redeemed e-tag" false "$NIP"
fi

# Test 6: Spending history with multiple e-tags
spending_multiple_tags='{
    "kind": 7376,
    "content": "AgABC456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
    "tags": [
        ["e", "'$TEST_EVENT_ID_1'", "", "destroyed"],
        ["e", "'$TEST_EVENT_ID_2'", "", "created"]
    ]
}'

if run_nak_test "$spending_multiple_tags" true; then
    print_result "Spending history with multiple e-tags" true "$NIP"
else
    print_result "Spending history with multiple e-tags" false "$NIP"
fi

# Test 7: Spending history with redeemed marker (should be unencrypted)
spending_redeemed='{
    "kind": 7376,
    "content": "AgABC456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
    "tags": [
        ["e", "'$TEST_EVENT_ID_1'", "", "redeemed"],
        ["e", "'$TEST_EVENT_ID_2'", "", "created"]
    ]
}'

if run_nak_test "$spending_redeemed" true; then
    print_result "Spending history with redeemed marker" true "$NIP"
else
    print_result "Spending history with redeemed marker" false "$NIP"
fi

# ========================================
# Valid Quote Event Tests (Kind 7374)
# ========================================

echo
echo -e "${YELLOW}Valid Quote Event Tests${NC}"

# Test 8: Basic quote event with required tags
future_expiration=$(($(date +%s) + 86400))  # 1 day from now
quote_event="{
    \"kind\": 7374,
    \"content\": \"AgABC123quoteid456def789ghi012encrypted890\",
    \"tags\": [
        [\"expiration\", \"$future_expiration\"],
        [\"mint\", \"https://mint.example.com\"]
    ]
}"

if run_nak_test "$quote_event" true; then
    print_result "Basic quote event with required tags" true "$NIP"
else
    print_result "Basic quote event with required tags" false "$NIP"
fi

# Test 9: Quote event with HTTPS mint URL
quote_https_mint="{
    \"kind\": 7374,
    \"content\": \"AgABC123quoteid456def789ghi012encrypted890\",
    \"tags\": [
        [\"expiration\", \"$future_expiration\"],
        [\"mint\", \"https://stablenut.umint.cash\"]
    ]
}"

if run_nak_test "$quote_https_mint" true; then
    print_result "Quote event with HTTPS mint URL" true "$NIP"
else
    print_result "Quote event with HTTPS mint URL" false "$NIP"
fi

# ========================================
# Invalid Event Tests
# ========================================

echo
echo -e "${YELLOW}Invalid Event Tests${NC}"

# Test 10: Wallet event with empty content (should fail)
wallet_empty_content='{
    "kind": 17375,
    "content": "",
    "tags": []
}'

if run_nak_test "$wallet_empty_content" false; then
    print_result "Wallet event with empty content (should fail)" true "$NIP"
else
    print_result "Wallet event with empty content (should fail)" false "$NIP"
fi

# Test 11: Wallet event with content too short (should fail)
wallet_short_content='{
    "kind": 17375,
    "content": "short",
    "tags": []
}'

if run_nak_test "$wallet_short_content" false; then
    print_result "Wallet event with short content (should fail)" true "$NIP"
else
    print_result "Wallet event with short content (should fail)" false "$NIP"
fi

# Test 12: Token event with empty content (should fail)
token_empty_content='{
    "kind": 7375,
    "content": "",
    "tags": []
}'

if run_nak_test "$token_empty_content" false; then
    print_result "Token event with empty content (should fail)" true "$NIP"
else
    print_result "Token event with empty content (should fail)" false "$NIP"
fi

# Test 13: Spending history with invalid e-tag format (should fail)
spending_invalid_etag='{
    "kind": 7376,
    "content": "AgABC456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
    "tags": [
        ["e", "invalid-event-id", "", "redeemed"]
    ]
}'

if run_nak_test "$spending_invalid_etag" false; then
    print_result "Spending history with invalid e-tag format (should fail)" true "$NIP"
else
    print_result "Spending history with invalid e-tag format (should fail)" false "$NIP"
fi

# Test 14: Spending history with invalid marker (should fail)
spending_invalid_marker='{
    "kind": 7376,
    "content": "AgABC456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
    "tags": [
        ["e", "'$TEST_EVENT_ID_1'", "", "invalid-marker"]
    ]
}'

if run_nak_test "$spending_invalid_marker" false; then
    print_result "Spending history with invalid marker (should fail)" true "$NIP"
else
    print_result "Spending history with invalid marker (should fail)" false "$NIP"
fi

# Test 15: Spending history with incomplete e-tag (should fail)
spending_incomplete_etag='{
    "kind": 7376,
    "content": "AgABC456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
    "tags": [
        ["e", "'$TEST_EVENT_ID_1'"]
    ]
}'

if run_nak_test "$spending_incomplete_etag" false; then
    print_result "Spending history with incomplete e-tag (should fail)" true "$NIP"
else
    print_result "Spending history with incomplete e-tag (should fail)" false "$NIP"
fi

# Test 16: Quote event missing expiration tag (should fail)
quote_no_expiration='{
    "kind": 7374,
    "content": "AgABC123quoteid456def789ghi012encrypted890",
    "tags": [
        ["mint", "https://mint.example.com"]
    ]
}'

if run_nak_test "$quote_no_expiration" false; then
    print_result "Quote event missing expiration (should fail)" true "$NIP"
else
    print_result "Quote event missing expiration (should fail)" false "$NIP"
fi

# Test 17: Quote event missing mint tag (should fail)
quote_no_mint='{
    "kind": 7374,
    "content": "AgABC123quoteid456def789ghi012encrypted890",
    "tags": [
        ["expiration", "1735518000"]
    ]
}'

if run_nak_test "$quote_no_mint" false; then
    print_result "Quote event missing mint tag (should fail)" true "$NIP"
else
    print_result "Quote event missing mint tag (should fail)" false "$NIP"
fi

# Test 18: Quote event with invalid expiration timestamp (should fail)
quote_invalid_expiration='{
    "kind": 7374,
    "content": "AgABC123quoteid456def789ghi012encrypted890",
    "tags": [
        ["expiration", "invalid-timestamp"],
        ["mint", "https://mint.example.com"]
    ]
}'

if run_nak_test "$quote_invalid_expiration" false; then
    print_result "Quote event with invalid expiration (should fail)" true "$NIP"
else
    print_result "Quote event with invalid expiration (should fail)" false "$NIP"
fi

# Test 19: Quote event with invalid mint URL (should fail)
quote_invalid_mint='{
    "kind": 7374,
    "content": "AgABC123quoteid456def789ghi012encrypted890",
    "tags": [
        ["expiration", "1735518000"],
        ["mint", "not-a-valid-url"]
    ]
}'

if run_nak_test "$quote_invalid_mint" false; then
    print_result "Quote event with invalid mint URL (should fail)" true "$NIP"
else
    print_result "Quote event with invalid mint URL (should fail)" false "$NIP"
fi

# Test 20: Quote event with non-HTTP(S) mint URL (should fail)
quote_ftp_mint='{
    "kind": 7374,
    "content": "AgABC123quoteid456def789ghi012encrypted890",
    "tags": [
        ["expiration", "1735518000"],
        ["mint", "ftp://mint.example.com"]
    ]
}'

if run_nak_test "$quote_ftp_mint" false; then
    print_result "Quote event with FTP mint URL (should fail)" true "$NIP"
else
    print_result "Quote event with FTP mint URL (should fail)" false "$NIP"
fi

# ========================================
# Edge Case Tests
# ========================================

echo
echo -e "${YELLOW}Edge Case Tests${NC}"

# Test 21: Wallet event with maximum content length
max_content=$(printf 'AgA%.0s' {1..200})  # Create long base64-like content
wallet_max_content="{
    \"kind\": 17375,
    \"content\": \"$max_content\",
    \"tags\": []
}"

if run_nak_test "$wallet_max_content" true; then
    print_result "Wallet event with maximum content length" true "$NIP"
else
    print_result "Wallet event with maximum content length" false "$NIP"
fi

# Test 22: Quote event with HTTP (non-HTTPS) mint
quote_http_mint="{
    \"kind\": 7374,
    \"content\": \"AgABC123quoteid456def789ghi012encrypted890\",
    \"tags\": [
        [\"expiration\", \"$future_expiration\"],
        [\"mint\", \"http://localhost:3338\"]
    ]
}"

if run_nak_test "$quote_http_mint" true; then
    print_result "Quote event with HTTP mint URL" true "$NIP"
else
    print_result "Quote event with HTTP mint URL" false "$NIP"
fi

# Test 23: Spending history with all valid marker types
spending_all_markers='{
    "kind": 7376,
    "content": "AgABC456def789ghi012jkl345mno678pqr901stu234vwx567yza890bcd123efg",
    "tags": [
        ["e", "'$TEST_EVENT_ID_1'", "", "created"],
        ["e", "'$TEST_EVENT_ID_2'", "", "destroyed"],
        ["e", "3333333333333333333333333333333333333333333333333333333333333333", "", "redeemed"]
    ]
}'

if run_nak_test "$spending_all_markers" true; then
    print_result "Spending history with all marker types" true "$NIP"
else
    print_result "Spending history with all marker types" false "$NIP"
fi

# Test 24: Quote event with future expiration (2 weeks from now)
future_expiration=$(($(date +%s) + 1209600))  # 2 weeks in seconds
quote_future_expiration="{
    \"kind\": 7374,
    \"content\": \"AgABC123quoteid456def789ghi012encrypted890\",
    \"tags\": [
        [\"expiration\", \"$future_expiration\"],
        [\"mint\", \"https://mint.example.com\"]
    ]
}"

if run_nak_test "$quote_future_expiration" true; then
    print_result "Quote event with future expiration" true "$NIP"
else
    print_result "Quote event with future expiration" false "$NIP"
fi

# Test 25: Token event with expiration tag (optional but allowed)
token_with_expiration="{
    \"kind\": 7375,
    \"content\": \"AgABC789xyz123abc456def789ghi012jkl345mno678pqr901stu234vwx567yza\",
    \"tags\": [
        [\"expiration\", \"$future_expiration\"]
    ]
}"

if run_nak_test "$token_with_expiration" true; then
    print_result "Token event with expiration tag" true "$NIP"
else
    print_result "Token event with expiration tag" false "$NIP"
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