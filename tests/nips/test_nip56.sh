#!/bin/bash
# NIP-56 Reporting Tests
# Tests for kind 1984 report events

set -e

RELAY_URL="ws://localhost:8081"
TEST_PRIVATE_KEY="0000000000000000000000000000000000000000000000000000000000000001"
TEST_PUBKEY="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"

echo "🧪 Testing NIP-56 Reporting Implementation"
echo "Relay: $RELAY_URL"
echo "Test pubkey: $TEST_PUBKEY"
echo ""

# Function to test valid report event
test_valid_report() {
    local report_type="$1"
    local description="$2"
    local tags="$3"
    
    echo "Testing VALID: $description"
    
    # Create and publish report event
    local result=$(nak event -k 1984 -c "$description" $tags --sec $TEST_PRIVATE_KEY $RELAY_URL 2>&1)
    
    if echo "$result" | grep -q "success"; then
        echo "✅ $report_type report accepted"
    else
        echo "❌ $report_type report rejected: $result"
    fi
    echo ""
}

# Function to test invalid report (should be rejected)
test_invalid_report() {
    local description="$1"
    local tags="$2"
    
    echo "Testing INVALID: $description"
    
    # Create and publish invalid report event
    local result=$(nak event -k 1984 -c "$description" $tags --sec $TEST_PRIVATE_KEY $RELAY_URL 2>&1)
    
    if echo "$result" | grep -q "msg:" && ! echo "$result" | grep -q "success"; then
        echo "✅ Invalid report correctly rejected"
    else
        echo "❌ Invalid report incorrectly accepted: $result"
    fi
    echo ""
}

echo "=== Testing Valid Report Types ==="

# Test nudity report
test_valid_report "nudity" "Nudity content report" "-t p=$TEST_PUBKEY;nudity"

# Test malware report
test_valid_report "malware" "Malware content report" "-t p=$TEST_PUBKEY;malware"

# Test profanity report
test_valid_report "profanity" "Profanity content report" "-t p=$TEST_PUBKEY;profanity"

# Test illegal report
test_valid_report "illegal" "Illegal content report" "-t p=$TEST_PUBKEY;illegal"

# Test spam report
test_valid_report "spam" "Spam content report" "-t p=$TEST_PUBKEY;spam"

# Test impersonation report
test_valid_report "impersonation" "Impersonation report" "-t p=$TEST_PUBKEY;impersonation"

# Test other report
test_valid_report "other" "Other violation report" "-t p=$TEST_PUBKEY;other"

echo "=== Testing Event Reports ==="

# Test event report with e tag (and required p tag)
test_valid_report "event-spam" "Event spam report" "-t p=$TEST_PUBKEY;spam -t e=84af8cfb06bb1da8c438d74c59d52ffbdda993f1df6eca37abeb928136d64216"

# Test event report with both p and e tags
test_valid_report "event-nudity" "Nudity in event by user" "-t p=$TEST_PUBKEY;nudity -t e=84af8cfb06bb1da8c438d74c59d52ffbdda993f1df6eca37abeb928136d64216"

echo "=== Testing NIP-32 Label Reports ==="

# Test with L and l tags
test_valid_report "nip32-labels" "Custom label report" "-t p=$TEST_PUBKEY;other -t L=com.example.content-warning -t l=graphic-violence;com.example.content-warning"

echo "=== Testing Multiple Tag Combinations ==="

# Test multiple p tags
test_valid_report "multi-user" "Multiple user spam report" "-t p=$TEST_PUBKEY;spam -t p=0000000000000000000000000000000000000000000000000000000000000002"

# Test comprehensive report with multiple tags
test_valid_report "comprehensive" "Comprehensive illegal content report" "-t p=$TEST_PUBKEY;illegal -t e=84af8cfb06bb1da8c438d74c59d52ffbdda993f1df6eca37abeb928136d64216 -t L=com.example.severity -t l=high;com.example.severity"

echo "=== Testing Invalid Reports ==="

# Test missing required tags
test_invalid_report "No tags report" ""

# Test invalid report type
test_invalid_report "Invalid report type" "-t p=$TEST_PUBKEY;invalid_type"

# Test empty report type
test_invalid_report "Empty report type" "-t p=$TEST_PUBKEY;"

# Test only e tag (missing required p tag)
test_invalid_report "Only e tag (missing p tag)" "-t e=84af8cfb06bb1da8c438d74c59d52ffbdda993f1df6eca37abeb928136d64216;malware"

# Test invalid pubkey format
test_invalid_report "Invalid pubkey format" "-t p=invalid_pubkey;spam"

# Test invalid event ID format
test_invalid_report "Invalid event ID" "-t e=invalid_event_id;malware -t p=$TEST_PUBKEY"

# Test NIP-32 l tag without L tag
test_invalid_report "NIP-32 l tag without L tag" "-t p=$TEST_PUBKEY;other -t l=violence"

echo "=== Testing Edge Cases ==="

# Test empty content
test_valid_report "empty-content" "" "-t p=$TEST_PUBKEY;spam"

# Test special characters in content
test_valid_report "special-chars" "Report with special chars: àáâãäåæçèéêë" "-t p=$TEST_PUBKEY;profanity"

# Test multiple report types in same event
test_valid_report "multi-concerns" "Multiple report concerns" "-t p=$TEST_PUBKEY;nudity -t p=$TEST_PUBKEY;spam"

echo "=== Testing Report Queries ==="

# Test querying reports
echo "Querying recent reports..."
nak req -k 1984 --limit 5 $RELAY_URL 2>/dev/null | head -10

echo ""
echo "🏁 NIP-56 Reporting tests completed!"
echo "Review the results above to verify all valid reports were accepted"
echo "and all invalid reports were properly rejected by the relay."