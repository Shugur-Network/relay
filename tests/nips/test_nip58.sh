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
RELAY=${RELAY:-"ws://localhost:8081"}

# Test secret keys
TEST_SECRET_KEY="26f2ef538bef741566429408b799a7583f6d4a02a2e701fe1b710b3f41055c0c"
RECIPIENT_SECRET_KEY="0000000000000000000000000000000000000000000000000000000000000001"
BADGE_ISSUER_SECRET_KEY="1111111111111111111111111111111111111111111111111111111111111111"

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

echo -e "${BLUE}Starting Shugur Relay NIP-58 Tests${NC}\n"

echo -e "${BLUE}Starting Shugur Relay NIP-58 Tests${NC}\n"

# Test NIP-58: Badges
echo -e "\n${YELLOW}Testing NIP-58: Badges${NC}"

# Test 1: Create a basic badge definition
BADGE_DEF=$(nak event -k 30009 -c "" -t d=bravery -t name="Bravery Badge" -t description="Awarded for acts of courage" --sec $BADGE_ISSUER_SECRET_KEY $RELAY)
if [ ! -z "$BADGE_DEF" ]; then
    print_result "Basic badge definition" true "58"
else
    print_result "Basic badge definition" false "58"
fi

# Test 2: Badge definition with image
BADGE_WITH_IMAGE=$(nak event -k 30009 -c "" -t d=honor -t name="Honor Badge" -t description="Awarded for honorable conduct" -t image="https://example.com/honor.png" --sec $BADGE_ISSUER_SECRET_KEY $RELAY)
if [ ! -z "$BADGE_WITH_IMAGE" ]; then
    print_result "Badge definition with image" true "58"
else
    print_result "Badge definition with image" false "58"
fi

# Test 3: Badge definition with thumbnail and dimensions
BADGE_WITH_THUMB=$(nak event -k 30009 -c "" -t d=excellence -t name="Excellence Badge" -t description="Awarded for exceptional performance" -t image="https://example.com/excellence.png" -t thumb="https://example.com/thumb.png" -t dim="100x100" --sec $BADGE_ISSUER_SECRET_KEY $RELAY)
if [ ! -z "$BADGE_WITH_THUMB" ]; then
    print_result "Badge definition with thumbnail and dimensions" true "58"
else
    print_result "Badge definition with thumbnail and dimensions" false "58"
fi

# Test 4: Badge definition without required d tag (should fail)
INVALID_BADGE_NO_D=$(nak event -k 30009 -c "" -t name="Invalid Badge" --sec $BADGE_ISSUER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_BADGE_NO_D" == *"missing required 'd' tag"* ]] || [[ "$INVALID_BADGE_NO_D" == *"refused"* ]]; then
    print_result "Badge definition without d tag (properly rejected)" true "58"
else
    print_result "Badge definition without d tag (improperly accepted)" false "58"
fi

# Test 5: Badge definition without required name tag (should fail)
INVALID_BADGE_NO_NAME=$(nak event -k 30009 -c "" -t d=incomplete --sec $BADGE_ISSUER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_BADGE_NO_NAME" == *"name tag"* ]] || [[ "$INVALID_BADGE_NO_NAME" == *"refused"* ]]; then
    print_result "Badge definition without name tag (properly rejected)" true "58"
else
    print_result "Badge definition without name tag (improperly accepted)" false "58"
fi

# Test 6: Badge definition with invalid image URL (should fail)
INVALID_BADGE_IMAGE=$(nak event -k 30009 -c "" -t d=invalid -t name="Invalid Badge" -t image="not-a-url" --sec $BADGE_ISSUER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_BADGE_IMAGE" == *"invalid"* ]] || [[ "$INVALID_BADGE_IMAGE" == *"refused"* ]]; then
    print_result "Badge definition with invalid image URL (properly rejected)" true "58"
else
    print_result "Badge definition with invalid image URL (improperly accepted)" false "58"
fi

# Test 7: Badge definition with invalid dimensions (should fail)
INVALID_BADGE_DIM=$(nak event -k 30009 -c "" -t d=invalid -t name="Invalid Badge" -t dim="invalid" --sec $BADGE_ISSUER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_BADGE_DIM" == *"dimensions"* ]] || [[ "$INVALID_BADGE_DIM" == *"refused"* ]]; then
    print_result "Badge definition with invalid dimensions (properly rejected)" true "58"
else
    print_result "Badge definition with invalid dimensions (improperly accepted)" false "58"
fi

# Test 8: Basic badge award
BADGE_AWARD=$(nak event -k 8 -c "Badge awarded!" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" -t p="$(nak key public $RECIPIENT_SECRET_KEY)" --sec $BADGE_ISSUER_SECRET_KEY $RELAY)
if [ ! -z "$BADGE_AWARD" ]; then
    print_result "Basic badge award" true "58"
else
    print_result "Basic badge award" false "58"
fi

# Test 9: Badge award with relay hint
BADGE_AWARD_RELAY=$(nak event -k 8 -c "Badge awarded!" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):honor" -t p="$(nak key public $RECIPIENT_SECRET_KEY),wss://relay.example.com" --sec $BADGE_ISSUER_SECRET_KEY $RELAY)
if [ ! -z "$BADGE_AWARD_RELAY" ]; then
    print_result "Badge award with relay hint" true "58"
else
    print_result "Badge award with relay hint" false "58"
fi

# Test 10: Badge award to multiple recipients
BADGE_AWARD_MULTI=$(nak event -k 8 -c "Badge awarded!" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):excellence" -t p="$(nak key public $RECIPIENT_SECRET_KEY)" -t p="$(nak key public $TEST_SECRET_KEY)" --sec $BADGE_ISSUER_SECRET_KEY $RELAY)
if [ ! -z "$BADGE_AWARD_MULTI" ]; then
    print_result "Badge award to multiple recipients" true "58"
else
    print_result "Badge award to multiple recipients" false "58"
fi

# Test 11: Badge award without a tag (should fail)
INVALID_AWARD_NO_A=$(nak event -k 8 -c "Badge awarded!" -t p="$(nak key public $RECIPIENT_SECRET_KEY)" --sec $BADGE_ISSUER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_AWARD_NO_A" == *"'a' tag"* ]] || [[ "$INVALID_AWARD_NO_A" == *"refused"* ]]; then
    print_result "Badge award without a tag (properly rejected)" true "58"
else
    print_result "Badge award without a tag (improperly accepted)" false "58"
fi

# Test 12: Badge award without p tag (should fail)
INVALID_AWARD_NO_P=$(nak event -k 8 -c "Badge awarded!" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" --sec $BADGE_ISSUER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_AWARD_NO_P" == *"'p' tag"* ]] || [[ "$INVALID_AWARD_NO_P" == *"refused"* ]]; then
    print_result "Badge award without p tag (properly rejected)" true "58"
else
    print_result "Badge award without p tag (improperly accepted)" false "58"
fi

# Test 13: Badge award with invalid a tag format (should fail)
INVALID_AWARD_A_FORMAT=$(nak event -k 8 -c "Badge awarded!" -t a="invalid-format" -t p="$(nak key public $RECIPIENT_SECRET_KEY)" --sec $BADGE_ISSUER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_AWARD_A_FORMAT" == *"invalid"* ]] || [[ "$INVALID_AWARD_A_FORMAT" == *"refused"* ]]; then
    print_result "Badge award with invalid a tag format (properly rejected)" true "58"
else
    print_result "Badge award with invalid a tag format (improperly accepted)" false "58"
fi

# Test 14: Badge award with invalid pubkey (should fail)
INVALID_AWARD_PUBKEY=$(nak event -k 8 -c "Badge awarded!" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" -t p="invalid-pubkey" --sec $BADGE_ISSUER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_AWARD_PUBKEY" == *"pubkey"* ]] || [[ "$INVALID_AWARD_PUBKEY" == *"refused"* ]]; then
    print_result "Badge award with invalid pubkey (properly rejected)" true "58"
else
    print_result "Badge award with invalid pubkey (improperly accepted)" false "58"
fi

# Test 15: Basic profile badges
SAMPLE_BADGE_AWARD_ID="7c9b1fe9a7b2c8e5d3f6a4b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8"
PROFILE_BADGES=$(nak event -k 30008 -c "My badges" -t d="badges" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" -t e="$SAMPLE_BADGE_AWARD_ID,wss://relay.example.com" --sec $RECIPIENT_SECRET_KEY $RELAY)
if [ ! -z "$PROFILE_BADGES" ]; then
    print_result "Basic profile badges" true "58"
else
    print_result "Basic profile badges" false "58"
fi

# Test 16: Profile badges with multiple badges
PROFILE_BADGES_MULTI=$(nak event -k 30008 -c "My badges" -t d="badges" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):honor" -t e="$SAMPLE_BADGE_AWARD_ID,wss://relay.example.com" -t e="$SAMPLE_BADGE_AWARD_ID,wss://relay.example.com" --sec $RECIPIENT_SECRET_KEY $RELAY)
if [ ! -z "$PROFILE_BADGES_MULTI" ]; then
    print_result "Profile badges with multiple badges" true "58"
else
    print_result "Profile badges with multiple badges" false "58"
fi

# Test 17: Profile badges without d tag (should fail)
INVALID_PROFILE_NO_D=$(nak event -k 30008 -c "My badges" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" -t e="$SAMPLE_BADGE_AWARD_ID" --sec $RECIPIENT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_PROFILE_NO_D" == *"'d' tag"* ]] || [[ "$INVALID_PROFILE_NO_D" == *"refused"* ]]; then
    print_result "Profile badges without d tag (properly rejected)" true "58"
else
    print_result "Profile badges without d tag (improperly accepted)" false "58"
fi

# Test 18: Profile badges with wrong d tag value (should fail)
INVALID_PROFILE_D_VALUE=$(nak event -k 30008 -c "My badges" -t d="wrong_value" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" -t e="$SAMPLE_BADGE_AWARD_ID" --sec $RECIPIENT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_PROFILE_D_VALUE" == *"profile_badges"* ]] || [[ "$INVALID_PROFILE_D_VALUE" == *"refused"* ]]; then
    print_result "Profile badges with wrong d tag value (properly rejected)" true "58"
else
    print_result "Profile badges with wrong d tag value (improperly accepted)" false "58"
fi

# Test 19: Profile badges with unpaired a/e tags (should fail)
INVALID_PROFILE_UNPAIRED=$(nak event -k 30008 -c "My badges" -t d="profile_badges" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):honor" -t e="$SAMPLE_BADGE_AWARD_ID" --sec $RECIPIENT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_PROFILE_UNPAIRED" == *"paired"* ]] || [[ "$INVALID_PROFILE_UNPAIRED" == *"refused"* ]]; then
    print_result "Profile badges with unpaired a/e tags (properly rejected)" true "58"
else
    print_result "Profile badges with unpaired a/e tags (improperly accepted)" false "58"
fi

# Test 20: Profile badges with invalid event ID (should fail)
INVALID_PROFILE_EVENT_ID=$(nak event -k 30008 -c "My badges" -t d="profile_badges" -t a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" -t e="invalid-event-id" --sec $RECIPIENT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_PROFILE_EVENT_ID" == *"event ID"* ]] || [[ "$INVALID_PROFILE_EVENT_ID" == *"refused"* ]]; then
    print_result "Profile badges with invalid event ID (properly rejected)" true "58"
else
    print_result "Profile badges with invalid event ID (improperly accepted)" false "58"
fi

# Print summary
echo -e "\n${BLUE}Test Summary:${NC}"
echo -e "Total tests: $test_count"
echo -e "${GREEN}Successful: $success_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"

# Exit with error if any tests failed
if [ $fail_count -gt 0 ]; then
    exit 1
else
    exit 0
fi

# Helper functions
test_valid_badge_definition() {
    local description="$1"
    shift
    echo "Testing: $description"
    echo "  Running: nak event --kind 30009 --content '' $@ --sec $BADGE_ISSUER_SECRET_KEY $RELAY_URL"
    if nak event --kind 30009 --content '' "$@" --sec "$BADGE_ISSUER_SECRET_KEY" "$RELAY_URL" >/dev/null 2>&1; then
        echo "  ✓ PASS: Valid badge definition accepted"
    else
        echo "  ✗ FAIL: Valid badge definition rejected"
        return 1
    fi
    echo ""
}

test_valid_badge_award() {
    local description="$1"
    shift
    echo "Testing: $description"
    echo "  Running: nak event --kind 8 --content 'Badge awarded!' $@ --sec $BADGE_ISSUER_SECRET_KEY $RELAY_URL"
    if nak event --kind 8 --content 'Badge awarded!' "$@" --sec "$BADGE_ISSUER_SECRET_KEY" "$RELAY_URL" >/dev/null 2>&1; then
        echo "  ✓ PASS: Valid badge award accepted"
    else
        echo "  ✗ FAIL: Valid badge award rejected"
        return 1
    fi
    echo ""
}

test_valid_profile_badges() {
    local description="$1"
    shift
    echo "Testing: $description"
    echo "  Running: nak event --kind 30008 --content 'My badges' $@ --sec $RECIPIENT_SECRET_KEY $RELAY_URL"
    if nak event --kind 30008 --content 'My badges' "$@" --sec "$RECIPIENT_SECRET_KEY" "$RELAY_URL" >/dev/null 2>&1; then
        echo "  ✓ PASS: Valid profile badges accepted"
    else
        echo "  ✗ FAIL: Valid profile badges rejected"
        return 1
    fi
    echo ""
}

test_invalid_badge_event() {
    local description="$1"
    local kind="$2"
    shift 2
    local content=""
    local secret_key="$TEST_SECRET_KEY"
    
    # Choose appropriate content and key based on kind
    case $kind in
        30009) content="Badge: $description"; secret_key="$BADGE_ISSUER_SECRET_KEY" ;;
        8) content="Badge awarded!"; secret_key="$BADGE_ISSUER_SECRET_KEY" ;;
        30008) content="My badges"; secret_key="$RECIPIENT_SECRET_KEY" ;;
    esac
    
    echo "Testing: $description"
    echo "  Running: nak event --kind $kind --content '$content' $@ --sec $secret_key $RELAY_URL"
    local output=$(nak event --kind "$kind" --content "$content" "$@" --sec "$secret_key" "$RELAY_URL" 2>&1)
    if echo "$output" | grep -q "msg:"; then
        echo "  ✓ PASS: Invalid badge event rejected ($(echo "$output" | grep 'msg:' | sed 's/.*msg: //'))"
    else
        echo "  ✗ FAIL: Invalid badge event accepted"
        echo "  Output: $output"
        return 1
    fi
    echo ""
}

get_badge_award_event_id() {
    # This would normally query for the actual event ID
    # For testing purposes, return a sample ID
    echo "$SAMPLE_BADGE_AWARD_ID"
}

echo "=== Testing Valid Badge Definitions ==="

# Test basic badge definition
test_valid_badge_definition "Basic badge definition" \
    --tag d="bravery" \
    --tag name="Bravery Badge" \
    --tag description="Awarded for acts of courage"

# Test badge definition with image
test_valid_badge_definition "Badge definition with image" \
    --tag d="honor" \
    --tag name="Honor Badge" \
    --tag description="Awarded for honorable conduct" \
    --tag image="https://example.com/honor.png"

# Test badge definition with thumbnail  
test_valid_badge_definition "Badge definition with thumbnail" \
    --tag d="excellence" \
    --tag name="Excellence Badge" \
    --tag description="Awarded for exceptional performance" \
    --tag image="https://example.com/excellence.png" \
    --tag thumb="https://example.com/thumb.png"

# Test badge definition with dimensions
test_valid_badge_definition "Badge definition with dimensions" \
    --tag d="wisdom" \
    --tag name="Wisdom Badge" \
    --tag description="Awarded for wise counsel" \
    --tag image="https://example.com/wisdom.png" \
    --tag dim="100x100"

# Test badge definition with all optional fields
test_valid_badge_definition "Badge definition with all fields" \
    --tag d="leadership" \
    --tag name="Leadership Badge" \
    --tag description="Awarded for exceptional leadership" \
    --tag image="https://example.com/leadership.png" \
    --tag thumb="https://example.com/leadership_thumb.png" \
    --tag dim="150x150"

echo ""
echo "=== Testing Invalid Badge Definitions ==="

# Test badge definition without d tag
test_invalid_badge_event "Badge definition without d tag" 30009 \
    --tag name="Invalid Badge"

# Test badge definition without name
test_invalid_badge_event "Badge definition without name" 30009 \
    --tag d="incomplete"

# Test badge definition with invalid image URL
test_invalid_badge_event "Badge definition with invalid image" 30009 \
    --tag d="invalid" \
    --tag name="Invalid Badge" \
    --tag image="not-a-url"

# Test badge definition with invalid dimensions
test_invalid_badge_event "Badge definition with invalid dimensions" 30009 \
    --tag d="invalid" \
    --tag name="Invalid Badge" \
    --tag dim="invalid"

echo ""
echo "=== Testing Valid Badge Awards ==="

# Test basic badge award
test_valid_badge_award "Basic badge award" \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag p="$(nak key public $RECIPIENT_SECRET_KEY)"

# Test badge award with relay hint
test_valid_badge_award "Badge award with relay hint" \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):honor" \
    --tag p="$(nak key public $RECIPIENT_SECRET_KEY),wss://relay.example.com"

# Test badge award to multiple recipients
test_valid_badge_award "Badge award to multiple recipients" \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):excellence" \
    --tag p="$(nak key public $RECIPIENT_SECRET_KEY)" \
    --tag p="$(nak key public $TEST_SECRET_KEY)"

echo ""
echo "=== Testing Valid Profile Badges ==="

# Test basic profile badges
test_valid_profile_badges "Basic profile badges" \
    --tag d="badges" \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag e="$(get_badge_award_event_id $(nak key public $BADGE_ISSUER_SECRET_KEY) $(nak key public $RECIPIENT_SECRET_KEY) bravery),wss://relay.example.com"

# Test profile badges with multiple badges
test_valid_profile_badges "Profile badges with multiple badges" \
    --tag d="badges" \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):honor" \
    --tag e="$(get_badge_award_event_id $(nak key public $BADGE_ISSUER_SECRET_KEY) $(nak key public $RECIPIENT_SECRET_KEY) bravery),wss://relay.example.com" \
    --tag e="$(get_badge_award_event_id $(nak key public $BADGE_ISSUER_SECRET_KEY) $(nak key public $RECIPIENT_SECRET_KEY) honor),wss://relay.example.com"

echo ""
echo "=== Testing Invalid Badge Awards ==="

# Test badge award without a tag
test_invalid_badge_event "Badge award without a tag" 8 \
    --tag p="$(nak key public $RECIPIENT_SECRET_KEY)"

# Test badge award without p tag
test_invalid_badge_event "Badge award without p tag" 8 \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery"

# Test badge award with invalid a tag format
test_invalid_badge_event "Badge award with invalid a tag" 8 \
    --tag a="invalid-format" \
    --tag p="$(nak key public $RECIPIENT_SECRET_KEY)"

# Test badge award with wrong kind in a tag
test_invalid_badge_event "Badge award with wrong kind in a tag" 8 \
    --tag a="30008:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag p="$(nak key public $RECIPIENT_SECRET_KEY)"

# Test badge award with invalid pubkey
test_invalid_badge_event "Badge award with invalid pubkey" 8 \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag p="invalid-pubkey"

# Test badge award with invalid relay URL
test_invalid_badge_event "Badge award with invalid relay URL" 8 \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag p="$(nak key public $RECIPIENT_SECRET_KEY),invalid-relay"

echo ""
echo "=== Testing Invalid Profile Badges ==="

# Test profile badges without d tag
test_invalid_badge_event "Profile badges without d tag" 30008 \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag e="$SAMPLE_BADGE_AWARD_ID"

# Test profile badges with wrong d tag value
test_invalid_badge_event "Profile badges with wrong d tag" 30008 \
    --tag d=wrong_value \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag e="$SAMPLE_BADGE_AWARD_ID"

# Test profile badges with unpaired a/e tags (more a tags)
test_invalid_badge_event "Profile badges with unpaired tags (extra a)" 30008 \
    --tag d=profile_badges \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):honor" \
    --tag e="$SAMPLE_BADGE_AWARD_ID"

# Test profile badges with unpaired a/e tags (more e tags)
test_invalid_badge_event "Profile badges with unpaired tags (extra e)" 30008 \
    --tag d=profile_badges \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag e="$SAMPLE_BADGE_AWARD_ID" \
    --tag e="$SAMPLE_BADGE_AWARD_ID"

# Test profile badges with invalid a tag format
test_invalid_badge_event "Profile badges with invalid a tag" 30008 \
    --tag d=profile_badges \
    --tag a="invalid-format" \
    --tag e="$SAMPLE_BADGE_AWARD_ID"

# Test profile badges with invalid event ID
test_invalid_badge_event "Profile badges with invalid event ID" 30008 \
    --tag d=profile_badges \
    --tag a="30009:$(nak key public $BADGE_ISSUER_SECRET_KEY):bravery" \
    --tag e="invalid-event-id"

echo ""
echo "=== NIP-58 Badges Test Suite Complete ==="