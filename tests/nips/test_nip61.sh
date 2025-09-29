#!/bin/bash

# Test script for NIP-61 Nutzaps
# Tests both event types: Nutzap Info Events (10019) and Nutzap Events (9321)

# Color definitions for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

RELAY_URL="ws://localhost:8081"
NIP="NIP-61"

# Test counter
test_count=0
failed_count=0
passed_count=0

# Generate test private keys and other test data consistently
TEST_PRIVATE_KEY="0000000000000000000000000000000000000000000000000000000000000001"
TEST_AUTHOR_PUBKEY="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
TEST_P2PK_PUBKEY="4d4b6cd1361032ca9bd2aeb9d900aa4d45d9ead80ac9423374c451a7254d0766"
TEST_EVENT_ID="1111111111111111111111111111111111111111111111111111111111111111"

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

echo -e "${BLUE}Testing $NIP: Nutzaps${NC}"
echo

# ========================================
# Valid Nutzap Info Event Tests (Kind 10019)
# ========================================

echo -e "${YELLOW}Valid Nutzap Info Event Tests${NC}"

# Test 1: Basic valid nutzap info event
nutzap_info_basic='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "wss://relay1.example.com"],
        ["relay", "wss://relay2.example.com"],
        ["mint", "https://mint1.example.com"],
        ["mint", "https://mint2.example.com", "sat"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_basic" true; then
    print_result "Basic valid nutzap info event" true "$NIP"
else
    print_result "Basic valid nutzap info event" false "$NIP"
fi

# Test 2: Nutzap info with multiple base units
nutzap_info_multi_units='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "wss://relay.example.com"],
        ["mint", "https://stablenut.umint.cash", "usd", "sat"],
        ["mint", "https://minibits.cash", "sat"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_multi_units" true; then
    print_result "Nutzap info with multiple base units" true "$NIP"
else
    print_result "Nutzap info with multiple base units" false "$NIP"
fi

# Test 3: Nutzap info with multiple relays and mints
nutzap_info_multiple='{
    "kind": 10019,
    "content": "Nutzap configuration for payments",
    "tags": [
        ["relay", "wss://nos.lol"],
        ["relay", "wss://relay.damus.io"],
        ["relay", "wss://relay.nostr.band"],
        ["mint", "https://mint.minibits.cash"],
        ["mint", "https://legend.lnbits.com/cashu"],
        ["mint", "https://stablenut.umint.cash", "usd"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_multiple" true; then
    print_result "Nutzap info with multiple relays and mints" true "$NIP"
else
    print_result "Nutzap info with multiple relays and mints" false "$NIP"
fi

# ========================================
# Valid Nutzap Event Tests (Kind 9321)
# ========================================

echo
echo -e "${YELLOW}Valid Nutzap Event Tests${NC}"

# Test 4: Basic nutzap event
nutzap_basic='{
    "kind": 9321,
    "content": "Thanks for this great idea!",
    "tags": [
        ["proof", "{\"amount\":1,\"C\":\"02277c66191736eb72fce9d975d08e3191f8f96afb73ab1eec37e4465683066d3f\",\"id\":\"000a93d6f8a1d2c4\",\"secret\":\"[\\\"P2PK\\\",{\\\"nonce\\\":\\\"b00bdd0467b0090a25bdf2d2f0d45ac4e355c482c1418350f273a04fedaaee83\\\",\\\"data\\\":\\\"02eaee8939e3565e48cc62967e2fde9d8e2a4b3ec0081f29eceff5c64ef10ac1ed\\\"}]\"}"],
        ["u", "https://stablenut.umint.cash"],
        ["e", "'$TEST_EVENT_ID'", "wss://relay.example.com"],
        ["k", "1"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_basic" true; then
    print_result "Basic nutzap event" true "$NIP"
else
    print_result "Basic nutzap event" false "$NIP"
fi

# Test 5: Nutzap with multiple proofs
nutzap_multiple_proofs='{
    "kind": 9321,
    "content": "Multiple small nutzaps!",
    "tags": [
        ["proof", "{\"amount\":1,\"C\":\"02277c66191736eb72fce9d975d08e3191f8f96afb73ab1eec37e4465683066d3f\",\"id\":\"000a93d6f8a1d2c4\",\"secret\":\"[\\\"P2PK\\\",{\\\"nonce\\\":\\\"b00bdd0467b0090a25bdf2d2f0d45ac4e355c482c1418350f273a04fedaaee83\\\",\\\"data\\\":\\\"02eaee8939e3565e48cc62967e2fde9d8e2a4b3ec0081f29eceff5c64ef10ac1ed\\\"}]\"}"],
        ["proof", "{\"amount\":2,\"C\":\"03a45c1b8f234567890abcdef123456789abcdef0123456789abcdef0123456789\",\"id\":\"111b83d6f8a1d2c4\",\"secret\":\"[\\\"P2PK\\\",{\\\"nonce\\\":\\\"c11cee1578c1101b36cef3e3f0e56bd5f466d593d2529461g384b15gefbbff94\\\",\\\"data\\\":\\\"02eaee8939e3565e48cc62967e2fde9d8e2a4b3ec0081f29eceff5c64ef10ac1ed\\\"}]\"}"],
        ["u", "https://mint.example.com"],
        ["e", "'$TEST_EVENT_ID'"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_multiple_proofs" true; then
    print_result "Nutzap with multiple proofs" true "$NIP"
else
    print_result "Nutzap with multiple proofs" false "$NIP"
fi

# Test 6: Nutzap without nutzapped event (just tip)
nutzap_tip_only='{
    "kind": 9321,
    "content": "Just wanted to send you some sats!",
    "tags": [
        ["proof", "{\"amount\":5,\"C\":\"02abc123def456789012345678901234567890123456789012345678901234567\",\"id\":\"222a93d6f8a1d2c4\",\"secret\":\"[\\\"P2PK\\\",{\\\"nonce\\\":\\\"d22dee2689d2212c47def4f4g1f67ce6g577e604e3640572h495c26hgfccgg05\\\",\\\"data\\\":\\\"02eaee8939e3565e48cc62967e2fde9d8e2a4b3ec0081f29eceff5c64ef10ac1ed\\\"}]\"}"],
        ["u", "https://legend.lnbits.com/cashu"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_tip_only" true; then
    print_result "Nutzap tip without specific event" true "$NIP"
else
    print_result "Nutzap tip without specific event" false "$NIP"
fi

# ========================================
# Invalid Event Tests
# ========================================

echo
echo -e "${YELLOW}Invalid Event Tests${NC}"

# Test 7: Nutzap info missing relay tag (should fail)
nutzap_info_no_relay='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["mint", "https://mint.example.com"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_no_relay" false; then
    print_result "Nutzap info missing relay tag (should fail)" true "$NIP"
else
    print_result "Nutzap info missing relay tag (should fail)" false "$NIP"
fi

# Test 8: Nutzap info missing mint tag (should fail)
nutzap_info_no_mint='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "wss://relay.example.com"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_no_mint" false; then
    print_result "Nutzap info missing mint tag (should fail)" true "$NIP"
else
    print_result "Nutzap info missing mint tag (should fail)" false "$NIP"
fi

# Test 9: Nutzap info missing pubkey tag (should fail)
nutzap_info_no_pubkey='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "wss://relay.example.com"],
        ["mint", "https://mint.example.com"]
    ]
}'

if run_nak_test "$nutzap_info_no_pubkey" false; then
    print_result "Nutzap info missing pubkey tag (should fail)" true "$NIP"
else
    print_result "Nutzap info missing pubkey tag (should fail)" false "$NIP"
fi

# Test 10: Nutzap info with invalid relay URL (should fail)
nutzap_info_invalid_relay='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "invalid-relay-url"],
        ["mint", "https://mint.example.com"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_invalid_relay" false; then
    print_result "Nutzap info with invalid relay URL (should fail)" true "$NIP"
else
    print_result "Nutzap info with invalid relay URL (should fail)" false "$NIP"
fi

# Test 11: Nutzap info with invalid mint URL (should fail)
nutzap_info_invalid_mint='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "wss://relay.example.com"],
        ["mint", "ftp://invalid-mint.com"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_invalid_mint" false; then
    print_result "Nutzap info with invalid mint URL (should fail)" true "$NIP"
else
    print_result "Nutzap info with invalid mint URL (should fail)" false "$NIP"
fi

# Test 12: Nutzap info with invalid pubkey format (should fail)
nutzap_info_invalid_pubkey='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "wss://relay.example.com"],
        ["mint", "https://mint.example.com"],
        ["pubkey", "invalid-pubkey"]
    ]
}'

if run_nak_test "$nutzap_info_invalid_pubkey" false; then
    print_result "Nutzap info with invalid pubkey format (should fail)" true "$NIP"
else
    print_result "Nutzap info with invalid pubkey format (should fail)" false "$NIP"
fi

# Test 13: Nutzap event missing proof tag (should fail)
nutzap_no_proof='{
    "kind": 9321,
    "content": "Thanks!",
    "tags": [
        ["u", "https://mint.example.com"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_no_proof" false; then
    print_result "Nutzap event missing proof tag (should fail)" true "$NIP"
else
    print_result "Nutzap event missing proof tag (should fail)" false "$NIP"
fi

# Test 14: Nutzap event missing mint URL (should fail)
nutzap_no_mint='{
    "kind": 9321,
    "content": "Thanks!",
    "tags": [
        ["proof", "{\"amount\":1,\"C\":\"02277c66191736eb72fce9d975d08e3191f8f96afb73ab1eec37e4465683066d3f\",\"id\":\"000a93d6f8a1d2c4\",\"secret\":\"[\\\"P2PK\\\",{\\\"nonce\\\":\\\"b00bdd0467b0090a25bdf2d2f0d45ac4e355c482c1418350f273a04fedaaee83\\\",\\\"data\\\":\\\"02eaee8939e3565e48cc62967e2fde9d8e2a4b3ec0081f29eceff5c64ef10ac1ed\\\"}]\"}"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_no_mint" false; then
    print_result "Nutzap event missing mint URL (should fail)" true "$NIP"
else
    print_result "Nutzap event missing mint URL (should fail)" false "$NIP"
fi

# Test 15: Nutzap event missing recipient (should fail)
nutzap_no_recipient='{
    "kind": 9321,
    "content": "Thanks!",
    "tags": [
        ["proof", "{\"amount\":1,\"C\":\"02277c66191736eb72fce9d975d08e3191f8f96afb73ab1eec37e4465683066d3f\",\"id\":\"000a93d6f8a1d2c4\",\"secret\":\"[\\\"P2PK\\\",{\\\"nonce\\\":\\\"b00bdd0467b0090a25bdf2d2f0d45ac4e355c482c1418350f273a04fedaaee83\\\",\\\"data\\\":\\\"02eaee8939e3565e48cc62967e2fde9d8e2a4b3ec0081f29eceff5c64ef10ac1ed\\\"}]\"}"],
        ["u", "https://mint.example.com"]
    ]
}'

if run_nak_test "$nutzap_no_recipient" false; then
    print_result "Nutzap event missing recipient (should fail)" true "$NIP"
else
    print_result "Nutzap event missing recipient (should fail)" false "$NIP"
fi

# Test 16: Nutzap event with invalid proof JSON (should fail)
nutzap_invalid_proof='{
    "kind": 9321,
    "content": "Thanks!",
    "tags": [
        ["proof", "invalid-json-proof"],
        ["u", "https://mint.example.com"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_invalid_proof" false; then
    print_result "Nutzap event with invalid proof JSON (should fail)" true "$NIP"
else
    print_result "Nutzap event with invalid proof JSON (should fail)" false "$NIP"
fi

# Test 17: Nutzap event with invalid recipient pubkey (should fail)
nutzap_invalid_recipient='{
    "kind": 9321,
    "content": "Thanks!",
    "tags": [
        ["proof", "{\"amount\":1,\"C\":\"02277c66191736eb72fce9d975d08e3191f8f96afb73ab1eec37e4465683066d3f\",\"id\":\"000a93d6f8a1d2c4\",\"secret\":\"[\\\"P2PK\\\",{\\\"nonce\\\":\\\"b00bdd0467b0090a25bdf2d2f0d45ac4e355c482c1418350f273a04fedaaee83\\\",\\\"data\\\":\\\"02eaee8939e3565e48cc62967e2fde9d8e2a4b3ec0081f29eceff5c64ef10ac1ed\\\"}]\"}"],
        ["u", "https://mint.example.com"],
        ["p", "invalid-recipient-pubkey"]
    ]
}'

if run_nak_test "$nutzap_invalid_recipient" false; then
    print_result "Nutzap event with invalid recipient pubkey (should fail)" true "$NIP"
else
    print_result "Nutzap event with invalid recipient pubkey (should fail)" false "$NIP"
fi

# ========================================
# Edge Case Tests
# ========================================

echo
echo -e "${YELLOW}Edge Case Tests${NC}"

# Test 18: Nutzap info with many relays and mints (stress test)
nutzap_info_many='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "wss://relay1.example.com"],
        ["relay", "wss://relay2.example.com"],
        ["relay", "wss://relay3.example.com"],
        ["relay", "wss://relay4.example.com"],
        ["relay", "wss://relay5.example.com"],
        ["mint", "https://mint1.example.com"],
        ["mint", "https://mint2.example.com"],
        ["mint", "https://mint3.example.com"],
        ["mint", "https://mint4.example.com"],
        ["mint", "https://mint5.example.com"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_many" true; then
    print_result "Nutzap info with many relays and mints" true "$NIP"
else
    print_result "Nutzap info with many relays and mints" false "$NIP"
fi

# Test 19: Nutzap with invalid k-tag (non-numeric)
nutzap_invalid_kind='{
    "kind": 9321,
    "content": "Thanks!",
    "tags": [
        ["proof", "{\"amount\":1,\"C\":\"02277c66191736eb72fce9d975d08e3191f8f96afb73ab1eec37e4465683066d3f\",\"id\":\"000a93d6f8a1d2c4\",\"secret\":\"[\\\"P2PK\\\",{\\\"nonce\\\":\\\"b00bdd0467b0090a25bdf2d2f0d45ac4e355c482c1418350f273a04fedaaee83\\\",\\\"data\\\":\\\"02eaee8939e3565e48cc62967e2fde9d8e2a4b3ec0081f29eceff5c64ef10ac1ed\\\"}]\"}"],
        ["u", "https://mint.example.com"],
        ["e", "'$TEST_EVENT_ID'"],
        ["k", "not-a-number"],
        ["p", "'$TEST_AUTHOR_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_invalid_kind" false; then
    print_result "Nutzap with invalid k-tag format (should fail)" true "$NIP"
else
    print_result "Nutzap with invalid k-tag format (should fail)" false "$NIP"
fi

# Test 20: Nutzap info with custom base unit
nutzap_info_custom_unit='{
    "kind": 10019,
    "content": "",
    "tags": [
        ["relay", "wss://relay.example.com"],
        ["mint", "https://mint.example.com", "btc"],
        ["mint", "https://mint2.example.com", "eur"],
        ["pubkey", "'$TEST_P2PK_PUBKEY'"]
    ]
}'

if run_nak_test "$nutzap_info_custom_unit" true; then
    print_result "Nutzap info with custom base units" true "$NIP"
else
    print_result "Nutzap info with custom base units" false "$NIP"
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