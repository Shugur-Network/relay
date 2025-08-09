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
RELAY="ws://localhost:8080"

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

echo -e "${BLUE}Starting Shugur Relay NIP-23 Tests${NC}\n"

# Test NIP-23: Long-form Content
echo -e "\n${YELLOW}Testing NIP-23: Long-form Content${NC}"

# Test 1: Create a basic long-form content event
LONG_FORM_EVENT=$(nak event -k 30023 -c "This is a test long-form content article." -t d=test-article -t title="Test Article" -t summary="A brief summary of the test article" -t published_at=1234567890 -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$LONG_FORM_EVENT" ]; then
    print_result "Create basic long-form content event" true "23"
else
    print_result "Create basic long-form content event" false "23"
fi

# Test 2: Create a long-form content event with all fields
if [ ! -z "$LONG_FORM_EVENT" ]; then
    FULL_EVENT=$(nak event -k 30023 -c "This is a complete long-form content article with all fields." -t d=complete-article -t title="Complete Article" -t summary="A comprehensive summary" -t published_at=1234567890 -t t=test -t t=article -t image=https://example.com/image.jpg -t subject=Technology -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
    if [ ! -z "$FULL_EVENT" ]; then
        print_result "Create long-form content event with all fields" true "23"
    else
        print_result "Create long-form content event with all fields" false "23"
    fi
fi

# Test 3: Create a long-form content event with minimal fields
MINIMAL_EVENT=$(nak event -k 30023 -c "This is a minimal long-form content article." -t d=minimal-article -t title="Minimal Article" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$MINIMAL_EVENT" ]; then
    print_result "Create long-form content event with minimal fields" true "23"
else
    print_result "Create long-form content event with minimal fields" false "23"
fi

# Test 4: Create a long-form content event with long content
LONG_CONTENT=$(printf '%.1000s' "This is a very long article content...")
LONG_CONTENT_EVENT=$(nak event -k 30023 -c "$LONG_CONTENT" -t d=long-content-article -t title="Long Content Article" -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY)
if [ ! -z "$LONG_CONTENT_EVENT" ]; then
    print_result "Create long-form content event with long content" true "23"
else
    print_result "Create long-form content event with long content" false "23"
fi

# Test 5: Attempt to create without title
NO_TITLE_EVENT=$(nak event -k 30023 -c "Attempting to create without title" -t d=no-title -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$NO_TITLE_EVENT" == *"title"* ]] || [[ "$NO_TITLE_EVENT" == *"failed"* ]]; then
    print_result "Reject long-form content event without title" true "23"
else
    print_result "Reject long-form content event without title" false "23"
fi

# Test 6: Attempt to create with invalid published_at
INVALID_DATE_EVENT=$(nak event -k 30023 -c "Attempting to create with invalid date" -t d=invalid-date -t title="Invalid Date Article" -t published_at=invalid -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_DATE_EVENT" == *"invalid"* ]] || [[ "$INVALID_DATE_EVENT" == *"❌"* ]]; then
    print_result "Reject long-form content event with invalid published_at" true "23"
else
    print_result "Reject long-form content event with invalid published_at" false "23"
fi

# Test 7: Attempt to create with invalid recipient
INVALID_RECIPIENT_EVENT=$(nak event -k 30023 -c "Attempting to create with invalid recipient" -t d=invalid-recipient -t title="Invalid Recipient Article" -t p=invalid_pubkey $RELAY 2>&1)
if [[ "$INVALID_RECIPIENT_EVENT" == *"invalid"* ]] || [[ "$INVALID_RECIPIENT_EVENT" == *"❌"* ]]; then
    print_result "Reject long-form content event with invalid recipient" true "23"
else
    print_result "Reject long-form content event with invalid recipient" false "23"
fi

# Test 8: Attempt to create with invalid image URL
INVALID_IMAGE_EVENT=$(nak event -k 30023 -c "Attempting to create with invalid image URL" -t d=invalid-image -t title="Invalid Image Article" -t image=invalid_url -t p=79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798 $RELAY 2>&1)
if [[ "$INVALID_IMAGE_EVENT" == *"invalid"* ]] || [[ "$INVALID_IMAGE_EVENT" == *"❌"* ]]; then
    print_result "Reject long-form content event with invalid image URL" true "23"
else
    print_result "Reject long-form content event with invalid image URL" false "23"
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