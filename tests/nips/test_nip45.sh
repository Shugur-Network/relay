#!/bin/bash

# NIP-45 COUNT Command Test Script
# Tests the COUNT command implementation

echo "=== NIP-45 COUNT Command Tests ==="
echo

# Test 1: Basic COUNT request
echo "1. Testing basic COUNT request for kind 1 events..."
nak count wss://shu01.shugur.net -k 1
echo

# Test 2: COUNT with author filter
echo "2. Testing COUNT with author filter..."
nak count wss://shu01.shugur.net -k 1 -a 79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798
echo

# Test 3: COUNT for kind 10002 (relay lists)
echo "3. Testing COUNT for kind 10002 (relay lists)..."
nak count wss://shu01.shugur.net -k 10002
echo

# Test 4: COUNT with multiple kinds
echo "4. Testing COUNT with multiple kinds..."
nak count wss://shu01.shugur.net -k 1 -k 10002
echo

# Test 5: COUNT with time range
echo "5. Testing COUNT with time range (last 24 hours)..."
SINCE=$(date -d "24 hours ago" +%s)
nak count wss://shu01.shugur.net -k 1 --since $SINCE
echo

echo "=== NIP-45 Tests Complete ==="
