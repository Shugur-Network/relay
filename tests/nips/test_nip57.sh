#!/bin/bash
# NIP-57 Lightning Zaps Tests
# Tests for kind 9734 (zap request) and kind 9735 (zap receipt) events

set -e

RELAY_URL="ws://localhost:8081"
TEST_PRIVATE_KEY="0000000000000000000000000000000000000000000000000000000000000001"
TEST_PUBKEY="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
RECIPIENT_PUBKEY="0000000000000000000000000000000000000000000000000000000000000002"
TEST_EVENT_ID="84af8cfb06bb1da8c438d74c59d52ffbdda993f1df6eca37abeb928136d64216"

echo "⚡ Testing NIP-57 Lightning Zaps Implementation"
echo "Relay: $RELAY_URL"
echo "Test pubkey: $TEST_PUBKEY"
echo ""

# Function to test valid zap request
test_valid_zap_request() {
    local description="$1"
    local tags="$2"
    
    echo "Testing VALID ZAP REQUEST: $description"
    
    # Create and publish zap request event
    local result=$(nak event -k 9734 -c "Zap!" $tags --sec $TEST_PRIVATE_KEY $RELAY_URL 2>&1)
    
    if echo "$result" | grep -q "success"; then
        echo "✅ Zap request accepted"
    else
        echo "❌ Zap request rejected: $result"
    fi
    echo ""
}

# Function to test valid zap receipt
test_valid_zap_receipt() {
    local description="$1"
    local tags="$2"
    
    echo "Testing VALID ZAP RECEIPT: $description"
    
    # Create and publish zap receipt event
    local result=$(nak event -k 9735 -c "" $tags --sec $TEST_PRIVATE_KEY $RELAY_URL 2>&1)
    
    if echo "$result" | grep -q "success"; then
        echo "✅ Zap receipt accepted"
    else
        echo "❌ Zap receipt rejected: $result"
    fi
    echo ""
}

# Function to test invalid zap event (should be rejected)
test_invalid_zap_event() {
    local kind="$1"
    local description="$2"
    local tags="$3"
    
    echo "Testing INVALID: $description"
    
    # Create and publish invalid zap event
    local result=$(nak event -k $kind -c "" $tags --sec $TEST_PRIVATE_KEY $RELAY_URL 2>&1)
    
    if echo "$result" | grep -q "msg:" && ! echo "$result" | grep -q "success"; then
        echo "✅ Invalid zap event correctly rejected"
    else
        echo "❌ Invalid zap event incorrectly accepted: $result"
    fi
    echo ""
}

echo "=== Testing Valid Zap Requests (Kind 9734) ==="

# Test basic zap request with required tags
test_valid_zap_request "Basic zap request" "-t relays=wss://relay.example.com -t amount=21000 -t lnurl=lnurl1dp68gurn8ghj7um5v93kketj9ehx2amn9uh8wetvdskkkmn0wahz7mrww4excup0dajx2mrv92x9xp -t p=$RECIPIENT_PUBKEY"

# Test zap request with minimal required tags (just p tag)
test_valid_zap_request "Minimal zap request" "-t p=$RECIPIENT_PUBKEY"

# Test zap request with event zapping (e tag)
test_valid_zap_request "Event zap request" "-t p=$RECIPIENT_PUBKEY -t e=$TEST_EVENT_ID -t k=1"

# Test zap request with addressable event (a tag)
test_valid_zap_request "Addressable event zap" "-t p=$RECIPIENT_PUBKEY -t a=30023:$RECIPIENT_PUBKEY:article1"

# Test zap request with P tag (sender)
test_valid_zap_request "Zap with sender P tag" "-t p=$RECIPIENT_PUBKEY -t P=$TEST_PUBKEY -t amount=5000"

# Test zap request with multiple relays
test_valid_zap_request "Multiple relays" "-t relays=wss://relay1.example.com;wss://relay2.example.com;wss://relay3.example.com -t p=$RECIPIENT_PUBKEY -t amount=10000"

echo "=== Testing Valid Zap Receipts (Kind 9735) ==="

# Create a sample zap request for the description tag
SAMPLE_ZAP_REQUEST="{\"kind\":9734,\"content\":\"Zap!\",\"tags\":[[\"p\",\"$RECIPIENT_PUBKEY\"],[\"amount\",\"21000\"]],\"pubkey\":\"$TEST_PUBKEY\",\"created_at\":1679673265,\"id\":\"test123\",\"sig\":\"testsig\"}"

# Test basic zap receipt with required tags
test_valid_zap_receipt "Basic zap receipt" "-t p=$RECIPIENT_PUBKEY -t bolt11=lnbc210u1p3unwfusp5t9r3yymhpfqculx78u027lxspgxcr2n2987mx2j55nnfs95nxnzqpp5jmrh92pfld78spqs78v9euf2385t83uvpwk9ldrlvf6ch7tpascqhp5zvkrmemgth3tufcvflmzjzfvjt023nazlhljz2n9hattj4f8jq8qxqyjw5qcqpjrzjqtc4fc44feggv7065fqe5m4ytjarg3repr5j9el35xhmtfexc42yczarjuqqfzqqqqqqqqlgqqqqqqgq9q9qxpqysgq079nkq507a5tw7xgttmj4u990j7wfggtrasah5gd4ywfr2pjcn29383tphp4t48gquelz9z78p4cq7ml3nrrphw5w6eckhjwmhezhnqpy6gyf0 -t description=$SAMPLE_ZAP_REQUEST"

# Test zap receipt with sender P tag
test_valid_zap_receipt "Zap receipt with sender" "-t p=$RECIPIENT_PUBKEY -t P=$TEST_PUBKEY -t bolt11=lnbc100u1p3unwfusp5test -t description=$SAMPLE_ZAP_REQUEST"

# Test zap receipt with event tags
test_valid_zap_receipt "Event zap receipt" "-t p=$RECIPIENT_PUBKEY -t e=$TEST_EVENT_ID -t k=1 -t bolt11=lnbc50u1p3unwfusp5test -t description=$SAMPLE_ZAP_REQUEST"

# Test zap receipt with preimage
test_valid_zap_receipt "Zap receipt with preimage" "-t p=$RECIPIENT_PUBKEY -t bolt11=lnbc25u1p3unwfusp5test -t description=$SAMPLE_ZAP_REQUEST -t preimage=5d006d2cf1e73c7148e7519a4c68adc81642ce0e25a432b2434c99f97344c15f"

echo "=== Testing Invalid Zap Requests ==="

# Test zap request without p tag
test_invalid_zap_event 9734 "Zap request without p tag" "-t amount=21000"

# Test zap request with multiple p tags
test_invalid_zap_event 9734 "Zap request with multiple p tags" "-t p=$RECIPIENT_PUBKEY -t p=$TEST_PUBKEY"

# Test zap request with multiple e tags
test_invalid_zap_event 9734 "Zap request with multiple e tags" "-t p=$RECIPIENT_PUBKEY -t e=$TEST_EVENT_ID -t e=1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

# Test zap request with invalid amount
test_invalid_zap_event 9734 "Zap request with invalid amount" "-t p=$RECIPIENT_PUBKEY -t amount=invalid"

# Test zap request with zero amount
test_invalid_zap_event 9734 "Zap request with zero amount" "-t p=$RECIPIENT_PUBKEY -t amount=0"

# Test zap request with invalid pubkey
test_invalid_zap_event 9734 "Zap request with invalid pubkey" "-t p=invalidpubkey"

# Test zap request with invalid event ID
test_invalid_zap_event 9734 "Zap request with invalid event ID" "-t p=$RECIPIENT_PUBKEY -t e=invalid"

# Test zap request with invalid lnurl
test_invalid_zap_event 9734 "Zap request with invalid lnurl" "-t p=$RECIPIENT_PUBKEY -t lnurl=invalid"

# Test zap request with invalid relay URL
test_invalid_zap_event 9734 "Zap request with invalid relay" "-t p=$RECIPIENT_PUBKEY -t relays=invalid_url"

# Test zap request with invalid event coordinate
test_invalid_zap_event 9734 "Zap request with invalid a tag" "-t p=$RECIPIENT_PUBKEY -t a=invalid:coordinate"

echo "=== Testing Invalid Zap Receipts ==="

# Test zap receipt without p tag
test_invalid_zap_event 9735 "Zap receipt without p tag" "-t bolt11=lnbc10u1p3test -t description='$SAMPLE_ZAP_REQUEST'"

# Test zap receipt without bolt11 tag
test_invalid_zap_event 9735 "Zap receipt without bolt11 tag" "-t p=$RECIPIENT_PUBKEY -t description='$SAMPLE_ZAP_REQUEST'"

# Test zap receipt without description tag
test_invalid_zap_event 9735 "Zap receipt without description tag" "-t p=$RECIPIENT_PUBKEY -t bolt11=lnbc10u1p3test"

# Test zap receipt with invalid bolt11
test_invalid_zap_event 9735 "Zap receipt with invalid bolt11" "-t p=$RECIPIENT_PUBKEY -t bolt11=invalid -t description='$SAMPLE_ZAP_REQUEST'"

# Test zap receipt with invalid description JSON
test_invalid_zap_event 9735 "Zap receipt with invalid description" "-t p=$RECIPIENT_PUBKEY -t bolt11=lnbc10u1p3test -t description=invalid_json"

# Test zap receipt with invalid preimage
test_invalid_zap_event 9735 "Zap receipt with invalid preimage" "-t p=$RECIPIENT_PUBKEY -t bolt11=lnbc10u1p3test -t description='$SAMPLE_ZAP_REQUEST' -t preimage=invalid"

echo "=== Testing Edge Cases ==="

# Test very large amount
test_valid_zap_request "Large amount zap" "-t p=$RECIPIENT_PUBKEY -t amount=1000000000"

# Test zap request with content
test_valid_zap_request "Zap with message" "-t p=$RECIPIENT_PUBKEY -t amount=21000"

# Test empty relay list
test_invalid_zap_event 9734 "Empty relay list" "-t p=$RECIPIENT_PUBKEY -t relays="

# Test maximum amount
test_invalid_zap_event 9734 "Amount too large" "-t p=$RECIPIENT_PUBKEY -t amount=999999999999"

echo "=== Testing Zap Query ==="

# Test querying zap requests and receipts
echo "Querying recent zap events..."
nak req -k 9734,9735 --limit 5 $RELAY_URL 2>/dev/null | head -5

echo ""
echo "⚡ NIP-57 Lightning Zaps tests completed!"
echo "Review the results above to verify all valid zap events were accepted"
echo "and all invalid zap events were properly rejected by the relay."