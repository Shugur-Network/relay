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

# Relay URL - adjust as needed
RELAY="ws://localhost:8080"
# RELAY="wss://shu02.shugur.net"

# Generate test keys
PRIVKEY=$(openssl rand -hex 32)
PUBKEY=$(echo $PRIVKEY | nak key-public)

# Helper function to print test results
print_result() {
    local test_name=$1
    local success=$2
    local details=$3
    
    if [ "$success" = true ]; then
        echo -e "${GREEN}✓ Test $test_count: $test_name${NC}"
        ((success_count++))
    else
        echo -e "${RED}✗ Test $test_count: $test_name${NC}"
        if [ ! -z "$details" ]; then
            echo -e "   ${RED}Details: $details${NC}"
        fi
        ((fail_count++))
    fi
    ((test_count++))
}

# Helper function to check if event was accepted
check_event_accepted() {
    local response=$1
    # Check if the response contains acceptance indicators
    if [[ "$response" == *"\"true\""* ]] || [[ "$response" == *"published"* ]] || [[ "$response" == *"OK"* ]]; then
        return 0  # success
    else
        return 1  # failure
    fi
}

# Helper function to extract event ID from nak output
extract_event_id() {
    local output=$1
    echo "$output" | grep -o '[a-f0-9]\{64\}' | head -1
}

# Generate future timestamp (1 hour from now)
FUTURE_TIME=$(($(date +%s) + 3600))

# Generate witness keys
WITNESS1_PRIVKEY=$(openssl rand -hex 32)
WITNESS1_PUBKEY=$(echo $WITNESS1_PRIVKEY | nak key-public)
WITNESS2_PRIVKEY=$(openssl rand -hex 32)
WITNESS2_PUBKEY=$(echo $WITNESS2_PRIVKEY | nak key-public)
WITNESS3_PRIVKEY=$(openssl rand -hex 32)
WITNESS3_PUBKEY=$(echo $WITNESS3_PRIVKEY | nak key-public)

echo -e "${BLUE}Starting Time Capsules Tests${NC}\n"
echo -e "${YELLOW}Testing Time Capsules: Kinds 11990, 31990, 11991${NC}\n"
echo -e "Test Keys Generated:"
echo -e "  Author PubKey: $PUBKEY"
echo -e "  Witness 1: $WITNESS1_PUBKEY"
echo -e "  Witness 2: $WITNESS2_PUBKEY"
echo -e "  Witness 3: $WITNESS3_PUBKEY"
echo -e "  Unlock Time: $FUTURE_TIME ($(date -d @$FUTURE_TIME))\n"

# Test 1: Valid Time Capsule Creation (Kind 11990)
echo -e "Test 1: Valid Time Capsule Creation (Kind 11990)..."
CAPSULE_RESPONSE=$(nak event -k 11990 \
    --content "This is encrypted time capsule content - base64encoded_ciphertext_here" \
    -t x-cap=v1 \
    -t u=threshold,t,3,n,3,T,$FUTURE_TIME \
    -t w="$WITNESS1_PUBKEY","$WITNESS2_PUBKEY","$WITNESS3_PUBKEY" \
    -t w-commit=merkle_root_placeholder \
    -t enc=nip44:v1,alg,xchacha20poly1305 \
    -t loc=inline \
    --sec $PRIVKEY \
    $RELAY 2>&1)

CAPSULE_ID=$(extract_event_id "$CAPSULE_RESPONSE")
if [ ! -z "$CAPSULE_ID" ] && check_event_accepted "$CAPSULE_RESPONSE"; then
    print_result "Create valid time capsule (kind 11990)" true
    echo -e "   Capsule ID: $CAPSULE_ID"
else
    print_result "Create valid time capsule (kind 11990)" false "$CAPSULE_RESPONSE"
fi

# Test 2: Parameterized Replaceable Time Capsule (Kind 31990)
echo -e "\nTest 2: Parameterized Replaceable Time Capsule (Kind 31990)..."
REPLACEABLE_RESPONSE=$(nak event -k 31990 \
    --content "Replaceable time capsule content" \
    -t x-cap=v1 \
    -t d=my-capsule-id-v1 \
    -t u=threshold,t,2,n,3,T,$FUTURE_TIME \
    -t w="$WITNESS1_PUBKEY","$WITNESS2_PUBKEY","$WITNESS3_PUBKEY" \
    -t w-commit=merkle_root_placeholder2 \
    -t enc=nip44:v1,alg,xchacha20poly1305 \
    -t loc=inline \
    --sec $PRIVKEY \
    $RELAY 2>&1)

REPLACEABLE_ID=$(extract_event_id "$REPLACEABLE_RESPONSE")
if [ ! -z "$REPLACEABLE_ID" ] && check_event_accepted "$REPLACEABLE_RESPONSE"; then
    print_result "Create parameterized replaceable time capsule (kind 31990)" true
    echo -e "   Replaceable Capsule ID: $REPLACEABLE_ID"
else
    print_result "Create parameterized replaceable time capsule (kind 31990)" false "$REPLACEABLE_RESPONSE"
fi

# Test 3: Invalid Time Capsule - Missing Vendor Tag
echo -e "\nTest 3: Invalid Time Capsule - Missing Vendor Tag..."
INVALID_RESPONSE=$(nak event -k 11990 \
    --content "Invalid capsule without vendor tag" \
    -t u=threshold,t,3,n,3,T,$FUTURE_TIME \
    -t w="$WITNESS1_PUBKEY","$WITNESS2_PUBKEY","$WITNESS3_PUBKEY" \
    --sec $PRIVKEY \
    $RELAY 2>&1)

if [[ "$INVALID_RESPONSE" == *"false"* ]] || [[ "$INVALID_RESPONSE" == *"invalid"* ]] || [[ "$INVALID_RESPONSE" == *"vendor"* ]]; then
    print_result "Reject capsule missing vendor tag" true
else
    print_result "Reject capsule missing vendor tag" false "$INVALID_RESPONSE"
fi

# Test 4: Invalid Time Capsule - Past Unlock Time
echo -e "\nTest 4: Invalid Time Capsule - Past Unlock Time..."
PAST_TIME=$(($(date +%s) - 3600))
PAST_RESPONSE=$(nak event -k 11990 \
    --content "Invalid capsule with past unlock time" \
    -t x-cap=v1 \
    -t u=threshold,t,3,n,3,T,$PAST_TIME \
    -t w="$WITNESS1_PUBKEY","$WITNESS2_PUBKEY","$WITNESS3_PUBKEY" \
    --sec $PRIVKEY \
    $RELAY 2>&1)

if [[ "$PAST_RESPONSE" == *"false"* ]] || [[ "$PAST_RESPONSE" == *"future"* ]] || [[ "$PAST_RESPONSE" == *"time"* ]]; then
    print_result "Reject capsule with past unlock time" true
else
    print_result "Reject capsule with past unlock time" false "$PAST_RESPONSE"
fi

# Test 5: Invalid Time Capsule - Threshold Mismatch
echo -e "\nTest 5: Invalid Time Capsule - Threshold Mismatch..."
THRESHOLD_RESPONSE=$(nak event -k 11990 \
    --content "Invalid capsule with threshold > witnesses" \
    -t x-cap=v1 \
    -t u=threshold,t,5,n,3,T,$FUTURE_TIME \
    -t w="$WITNESS1_PUBKEY","$WITNESS2_PUBKEY","$WITNESS3_PUBKEY" \
    --sec $PRIVKEY \
    $RELAY 2>&1)

if [[ "$THRESHOLD_RESPONSE" == *"false"* ]] || [[ "$THRESHOLD_RESPONSE" == *"threshold"* ]] || [[ "$THRESHOLD_RESPONSE" == *"invalid"* ]]; then
    print_result "Reject capsule with invalid threshold" true
else
    print_result "Reject capsule with invalid threshold" false "$THRESHOLD_RESPONSE"
fi

# Test 6: Valid Unlock Share (Kind 11991) - Posted Early (Should Fail)
if [ ! -z "$CAPSULE_ID" ]; then
    echo -e "\nTest 6: Early Unlock Share (Should Fail)..."
    EARLY_SHARE_RESPONSE=$(nak event -k 11991 \
        --content "shamir_share_data_base64" \
        -t x-cap=v1 \
        -t e="$CAPSULE_ID" \
        -t w="$WITNESS1_PUBKEY" \
        -t T="$FUTURE_TIME" \
        --sec $WITNESS1_PRIVKEY \
        $RELAY 2>&1)

    if [[ "$EARLY_SHARE_RESPONSE" == *"false"* ]] || [[ "$EARLY_SHARE_RESPONSE" == *"early"* ]] || [[ "$EARLY_SHARE_RESPONSE" == *"time"* ]]; then
        print_result "Reject early unlock share" true
    else
        print_result "Reject early unlock share" false "$EARLY_SHARE_RESPONSE"
    fi
fi

# Test 7: Invalid Unlock Share - Missing Capsule Reference
echo -e "\nTest 7: Invalid Unlock Share - Missing Capsule Reference..."
MISSING_REF_RESPONSE=$(nak event -k 11991 \
    --content "shamir_share_data_base64" \
    -t x-cap=v1 \
    -t w="$WITNESS1_PUBKEY" \
    -t T="$FUTURE_TIME" \
    --sec $WITNESS1_PRIVKEY \
    $RELAY 2>&1)

if [[ "$MISSING_REF_RESPONSE" == *"false"* ]] || [[ "$MISSING_REF_RESPONSE" == *"reference"* ]] || [[ "$MISSING_REF_RESPONSE" == *"capsule"* ]]; then
    print_result "Reject unlock share missing capsule reference" true
else
    print_result "Reject unlock share missing capsule reference" false "$MISSING_REF_RESPONSE"
fi

# Test 8: Query Time Capsules
echo -e "\nTest 8: Query Time Capsules..."
QUERY_RESPONSE=$(nak req -k 11990,31990 -l 10 $RELAY 2>&1)
if [[ "$QUERY_RESPONSE" == *"\"pubkey\""* ]] && [[ "$QUERY_RESPONSE" == *"\"content\""* ]]; then
    print_result "Query time capsules successfully" true
    # Count how many capsules were found
    CAPSULE_COUNT=$(echo "$QUERY_RESPONSE" | grep -o '"kind":11990\|"kind":31990' | wc -l)
    echo -e "   Found $CAPSULE_COUNT time capsules"
else
    print_result "Query time capsules successfully" false "$QUERY_RESPONSE"
fi

# Test 9: Query Unlock Shares
echo -e "\nTest 9: Query Unlock Shares..."
SHARES_RESPONSE=$(nak req -k 11991 -l 10 $RELAY 2>&1)
if [[ "$SHARES_RESPONSE" == *"\"kind\":11991"* ]] || [[ "$SHARES_RESPONSE" == *"EOSE"* ]]; then
    print_result "Query unlock shares successfully" true
    SHARES_COUNT=$(echo "$SHARES_RESPONSE" | grep -o '"kind":11991' | wc -l)
    echo -e "   Found $SHARES_COUNT unlock shares"
else
    print_result "Query unlock shares successfully" false "$SHARES_RESPONSE"
fi

# Test 10: External Storage Location
echo -e "\nTest 10: External Storage Location..."
EXTERNAL_RESPONSE=$(nak event -k 11990 \
    --content "Short reference content" \
    -t x-cap=v1 \
    -t u=threshold,t,2,n,3,T,$FUTURE_TIME \
    -t w="$WITNESS1_PUBKEY","$WITNESS2_PUBKEY","$WITNESS3_PUBKEY" \
    -t w-commit=merkle_root_external \
    -t enc=nip44:v1,alg,xchacha20poly1305 \
    -t loc=https \
    -t uri=https://example.com/capsule/abc123 \
    --sec $PRIVKEY \
    $RELAY 2>&1)

if check_event_accepted "$EXTERNAL_RESPONSE"; then
    print_result "Create capsule with external storage" true
else
    print_result "Create capsule with external storage" false "$EXTERNAL_RESPONSE"
fi

echo -e "\n${BLUE}Test Summary:${NC}"
echo -e "${GREEN}Passed: $success_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
echo -e "Total: $test_count\n"

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}🎉 All Time Capsules tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Check the relay implementation.${NC}"
    exit 1
fi
