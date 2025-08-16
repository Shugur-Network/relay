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
# RELAY="ws://localhost:8080"
RELAY="wss://shu02.shugur.net"

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

echo -e "${BLUE}Starting Shugur Relay NIP-15 Tests${NC}\n"

echo -e "\n${YELLOW}Testing NIP-15: Stalls${NC}"

# Test 0: Create basic stall
BASIC_STALL=$(nak event -k 30017 -c '{"id":"stall1","name":"Test Stall","description":"A test stall","currency":"USD"}' -t d=stall1 $RELAY)
if [ $? -eq 0 ]; then
    print_result "Create basic stall" true "15"
else
    print_result "Create basic stall" false "15"
fi

# Test 1: Create complete stall with shipping
COMPLETE_STALL=$(nak event -k 30017 -c '{"id":"stall2","name":"Complete Stall","description":"A complete stall listing","currency":"USD","shipping":[{"id":"zone1","name":"US Domestic","cost":500,"regions":["US"]},{"id":"zone2","name":"International","cost":1500,"regions":["CA","EU"]}]}' -t d=stall2 $RELAY)
if [ $? -eq 0 ]; then
    print_result "Create complete stall with shipping" true "15"
else
    print_result "Create complete stall with shipping" false "15"
fi

# Test 2: Create minimal stall
MINIMAL_STALL=$(nak event -k 30017 -c '{"id":"stall3","name":"Minimal Stall","currency":"USD"}' -t d=stall3 $RELAY)
if [ $? -eq 0 ]; then
    print_result "Create minimal stall" true "15"
else
    print_result "Create minimal stall" false "15"
fi

# Test 3: Reject stall with invalid shipping cost
INVALID_SHIPPING=$(nak event -k 30017 -c '{"id":"stall4","name":"Invalid Shipping","currency":"USD","shipping":[{"id":"zone1","name":"Invalid Zone","cost":-500,"regions":["US"]}]}' -t d=stall4 $RELAY 2>&1)
if [[ $INVALID_SHIPPING == *"shipping zone must have a non-negative cost"* ]] || [[ $INVALID_SHIPPING == *"invalid shipping cost"* ]]; then
    print_result "Reject stall with invalid shipping cost" true "15"
else
    print_result "Reject stall with invalid shipping cost" false "15"
fi

# Test 4: Reject stall with missing required fields
MISSING_FIELDS=$(nak event -k 30017 -c '{"description":"Missing required fields"}' -t d=invalid $RELAY 2>&1)
if [[ $MISSING_FIELDS == *"stall must have an id"* ]]; then
    print_result "Reject stall with missing required fields" true "15"
else
    print_result "Reject stall with missing required fields" false "15"
fi

# Test 5: Reject stall with invalid currency
INVALID_CURRENCY=$(nak event -k 30017 -c '{"id":"stall5","name":"Invalid Currency","currency":""}' -t d=stall5 $RELAY 2>&1)
if [[ $INVALID_CURRENCY == *"stall must have a currency"* ]]; then
    print_result "Reject stall with invalid currency" true "15"
else
    print_result "Reject stall with invalid currency" false "15"
fi

# Test 6: Reject stall with invalid shipping regions
INVALID_REGIONS=$(nak event -k 30017 -c '{"id":"stall6","name":"Invalid Regions","currency":"USD","shipping":[{"id":"zone1","name":"Invalid Zone","cost":500,"regions":[]}]}' -t d=stall6 $RELAY 2>&1)
if [[ $INVALID_REGIONS == *"shipping zone must have at least one region"* ]]; then
    print_result "Reject stall with invalid shipping regions" true "15"
else
    print_result "Reject stall with invalid shipping regions" false "15"
fi

# Test 7: Reject stall with mismatched d tag
MISMATCHED_TAG=$(nak event -k 30017 -c '{"id":"stall7","name":"Mismatched Tag","currency":"USD"}' -t d=wrong_id $RELAY 2>&1)
if [[ $MISMATCHED_TAG == *"stall d tag must match stall id"* ]]; then
    print_result "Reject stall with mismatched d tag" true "15"
else
    print_result "Reject stall with mismatched d tag" false "15"
fi

echo -e "\n${YELLOW}Testing NIP-15: Products${NC}"

# Test 8: Create basic product
BASIC_PRODUCT=$(nak event -k 30018 -c '{"id":"prod1","stall_id":"stall1","name":"Test Product","description":"A test product","currency":"USD","price":1000}' -t d=prod1 -t t=test $RELAY)
if [ $? -eq 0 ]; then
    print_result "Create basic product" true "15"
else
    print_result "Create basic product" false "15"
fi

# Test 9: Create complete product
COMPLETE_PRODUCT=$(nak event -k 30018 -c '{"id":"prod2","stall_id":"stall1","name":"Complete Product","description":"A complete product","currency":"USD","price":2000,"quantity":10,"images":["https://example.com/img.jpg"],"specs":[["color","red"],["size","large"]],"shipping":[{"id":"zone1","cost":100}]}' -t d=prod2 -t t=test $RELAY)
if [ $? -eq 0 ]; then
    print_result "Create complete product" true "15"
else
    print_result "Create complete product" false "15"
fi

# Test 10: Reject product with invalid price
INVALID_PRICE=$(nak event -k 30018 -c '{"id":"prod3","stall_id":"stall1","name":"Invalid Price","currency":"USD","price":-100}' -t d=prod3 -t t=test $RELAY 2>&1)
if [[ $INVALID_PRICE == *"product must have a positive price"* ]]; then
    print_result "Reject product with invalid price" true "15"
else
    print_result "Reject product with invalid price" false "15"
fi

# Test 11: Reject product without category tag
NO_CATEGORY=$(nak event -k 30018 -c '{"id":"prod4","stall_id":"stall1","name":"No Category","currency":"USD","price":100}' -t d=prod4 $RELAY 2>&1)
if [[ $NO_CATEGORY == *"product must have at least one category tag"* ]]; then
    print_result "Reject product without category tag" true "15"
else
    print_result "Reject product without category tag" false "15"
fi

echo -e "\n${YELLOW}Testing NIP-15: Marketplace${NC}"

# Test 12: Create basic marketplace
BASIC_MARKETPLACE=$(nak event -k 30019 -c '{"name":"Test Market","about":"A test marketplace","ui":{"picture":"https://example.com/logo.jpg","banner":"https://example.com/banner.jpg","theme":"light","darkMode":false}}' -t d=test-market $RELAY)
if [ $? -eq 0 ]; then
    print_result "Create basic marketplace" true "15"
else
    print_result "Create basic marketplace" false "15"
fi

# Test 13: Reject marketplace with invalid URLs
INVALID_URLS=$(nak event -k 30019 -c '{"name":"Invalid URLs","ui":{"picture":"invalid-url","banner":"also-invalid"}}' -t d=invalid-marketplace $RELAY 2>&1)
if [[ $INVALID_URLS == *"marketplace picture must be a valid URL"* ]]; then
    print_result "Reject marketplace with invalid URLs" true "15"
else
    print_result "Reject marketplace with invalid URLs" false "15"
fi

echo -e "\n${YELLOW}Testing NIP-15: Auctions${NC}"

# Test 14: Create basic auction
BASIC_AUCTION=$(nak event -k 30020 -c '{"id":"auction1","stall_id":"stall1","name":"Test Auction","description":"A test auction","starting_bid":1000,"duration":86400}' -t d=auction1 $RELAY)
AUCTION_EVENT_ID=$(echo "$BASIC_AUCTION" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ $? -eq 0 ]; then
    print_result "Create basic auction" true "15"
else
    print_result "Create basic auction" false "15"
fi

# Test 15: Create complete auction
COMPLETE_AUCTION=$(nak event -k 30020 -c '{"id":"auction2","stall_id":"stall1","name":"Complete Auction","description":"A complete auction","images":["https://example.com/img.jpg"],"starting_bid":2000,"start_date":1735689600,"duration":172800,"specs":[["condition","new"],["color","blue"]]}' -t d=auction2 $RELAY)
if [ $? -eq 0 ]; then
    print_result "Create complete auction" true "15"
else
    print_result "Create complete auction" false "15"
fi

# Test 16: Reject auction with invalid duration
INVALID_DURATION=$(nak event -k 30020 -c '{"id":"auction3","stall_id":"stall1","name":"Invalid Duration","starting_bid":1000,"duration":-86400}' -t d=invalid-auction $RELAY 2>&1)
if [[ $INVALID_DURATION == *"auction must have a positive duration"* ]]; then
    print_result "Reject auction with invalid duration" true "15"
else
    print_result "Reject auction with invalid duration" false "15"
fi

echo -e "\n${YELLOW}Testing NIP-15: Bids${NC}"

# Test 17: Create valid bid
VALID_BID=$(nak event -k 1021 -c "1000" -t e=$AUCTION_EVENT_ID $RELAY)
BID_EVENT_ID=$(echo "$VALID_BID" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ $? -eq 0 ]; then
    print_result "Create valid bid" true "15"
else
    print_result "Create valid bid" false "15"
fi

# Test 18: Create bid confirmation
BID_CONFIRMATION=$(nak event -k 1022 -c '{"status":"accepted","message":"Bid accepted"}' -t e=$BID_EVENT_ID -t e=$AUCTION_EVENT_ID $RELAY)
if [ $? -eq 0 ]; then
    print_result "Create bid confirmation" true "15"
else
    print_result "Create bid confirmation" false "15"
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