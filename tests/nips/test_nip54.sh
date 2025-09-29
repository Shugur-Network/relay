#!/bin/bash

# Test script for NIP-54 Wiki 
# Tests all event types: Wiki Articles (30818), Merge Requests (818), and Redirects (30819)

# Color definitions for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

RELAY_URL="ws://localhost:8081"
NIP="NIP-54"

# Test counter
test_count=0
failed_count=0
passed_count=0

# Generate a test pri# Test 18: Multiple redirects chaining
redirect_chain='{
    "kind": 30819,
    "content": "First redirect in chain",
    "tags": [
        ["d", "redirect-chain-first"],
        ["redirect", "30819:'$AUTHOR_PUBKEY':redirect-chain-second"]
    ]
}'consistently
TEST_PRIVATE_KEY="0000000000000000000000000000000000000000000000000000000000000001"
TEST_AUTHOR_PUBKEY="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
TARGET_AUTHOR_PUBKEY="79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
FORK_AUTHOR_PUBKEY="c6047f9441ed7d6d3045406e95c07cd85c778e4b8cef3ca7abac09b95c709ee5"
REVIEWER_PUBKEY="f9308a019258c31049344f85f89d5229b531c845836f99b08601f113bce036f9"

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
    output=$(echo "$event_data" | timeout 10s nak event "$RELAY_URL" --sec "$TEST_PRIVATE_KEY" 2>&1)
    exit_code=$?
    
    if [ "$should_succeed" = true ]; then
        # Event should succeed - check if it was published
        if [ $exit_code -eq 0 ] && [[ ! "$output" =~ "failed to publish" ]] && [[ ! "$output" =~ "validation failed" ]] && [[ ! "$output" =~ "error" ]] && [[ ! "$output" =~ "failed:" ]]; then
            return 0
        else
            echo "    Expected success but got failure: $output"
            return 1
        fi
    else
        # Event should fail - check if it was rejected
        if [ $exit_code -ne 0 ] || [[ "$output" =~ "failed to publish" ]] || [[ "$output" =~ "validation failed" ]] || [[ "$output" =~ "error" ]] || [[ "$output" =~ "failed:" ]]; then
            return 0
        else
            echo "    Expected failure but got success: $output"
            return 1
        fi
    fi
}

echo -e "${BLUE}Testing $NIP: Wiki${NC}"
echo

# ========================================
# Valid Wiki Article Tests (Kind 30818)
# ========================================

echo -e "${YELLOW}Valid Wiki Article Tests${NC}"

# Test 1: Basic valid wiki article
wiki_article='{
    "kind": 30818,
    "content": "This is a test wiki article.\n\nIt has multiple paragraphs and [[wikilinks]].",
    "tags": [
        ["d", "test-article"],
        ["title", "Test Article"],
        ["summary", "A simple test article"],
        ["a", "30818:'$TEST_AUTHOR_PUBKEY':test-article", ""]
    ]
}'

if run_nak_test "$wiki_article" true; then
    print_result "Basic valid wiki article" true "$NIP"
else
    print_result "Basic valid wiki article" false "$NIP"
fi

# Test 2: Wiki article with normalized d-tag
wiki_normalized='{
    "kind": 30818,
    "content": "Article with normalized title",
    "tags": [
        ["d", "article-with-spaces"],
        ["title", "Article With Spaces"], 
        ["a", "30818:'$TEST_AUTHOR_PUBKEY':article-with-spaces", ""]
    ]
}'

if run_nak_test "$wiki_normalized" true; then
    print_result "Wiki article with d-tag normalization" true "$NIP"
else
    print_result "Wiki article with d-tag normalization" false "$NIP"
fi

# Test 3: Wiki article with wikilinks and references  
wiki_with_links='{
    "kind": 30818,
    "content": "This article references [[Other Article]] and has a link to [[external:example.com]].",
    "tags": [
        ["d", "article-with-links"],
        ["title", "Article With Links"],
        ["a", "30818:'$TEST_AUTHOR_PUBKEY':other-article", ""], 
        ["r", "https://example.com", "External Link"]
    ]
}'

if run_nak_test "$wiki_with_links" true; then
    print_result "Wiki article with wikilinks and references" true "$NIP"
else
    print_result "Wiki article with wikilinks and references" false "$NIP"
fi

# ========================================
# Valid Merge Request Tests (Kind 818)  
# ========================================

echo
echo -e "${YELLOW}Valid Merge Request Tests${NC}"

# Test 4: Basic merge request
merge_request='{
    "kind": 818,
    "content": "Please merge this change to fix the typo in line 5.",
    "tags": [
        ["a", "30818:'$TARGET_AUTHOR_PUBKEY':target-article", ""],
        ["e", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "", "source"],
        ["p", "'$TARGET_AUTHOR_PUBKEY'", ""],
        ["fork", "30818:'$FORK_AUTHOR_PUBKEY':forked-article"],
        ["merge", "30818:'$TARGET_AUTHOR_PUBKEY':target-article"]
    ]
}'

if run_nak_test "$merge_request" true; then
    print_result "Basic merge request" true "$NIP"
else
    print_result "Basic merge request" false "$NIP"
fi

# Test 5: Merge request with defer marker
merge_with_defer='{
    "kind": 818, 
    "content": "Proposed changes with defer marker for review.",
    "tags": [
        ["a", "30818:'$TARGET_AUTHOR_PUBKEY':target-article", ""],
        ["e", "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", "", "source"],
        ["p", "'$TARGET_AUTHOR_PUBKEY'", ""],
        ["fork", "30818:'$FORK_AUTHOR_PUBKEY':forked-article"],
        ["merge", "30818:'$TARGET_AUTHOR_PUBKEY':target-article"],
        ["defer", "30818:'$REVIEWER_PUBKEY':review-article", "Please review before merging"]
    ]
}'

if run_nak_test "$merge_with_defer" true; then
    print_result "Merge request with defer marker" true "$NIP"
else
    print_result "Merge request with defer marker" false "$NIP"
fi

# ========================================
# Valid Redirect Tests (Kind 30819)
# ========================================

echo
echo -e "${YELLOW}Valid Redirect Tests${NC}"

# Test 6: Basic redirect
redirect_basic='{
    "kind": 30819,
    "content": "",
    "tags": [
        ["d", "old-article-name"],
        ["redirect", "30818:'$TEST_AUTHOR_PUBKEY':new-article-name"]
    ]
}'

if run_nak_test "$redirect_basic" true; then
    print_result "Basic wiki redirect" true "$NIP"
else
    print_result "Basic wiki redirect" false "$NIP"
fi

# Test 7: Redirect with reason
redirect_with_reason='{
    "kind": 30819,
    "content": "This article has been moved.",
    "tags": [
        ["d", "moved-article"],
        ["redirect", "30818:'$TEST_AUTHOR_PUBKEY':new-location"]
    ]
}'

if run_nak_test "$redirect_with_reason" true; then
    print_result "Wiki redirect with reason" true "$NIP"
else
    print_result "Wiki redirect with reason" false "$NIP"
fi

# ========================================
# Invalid Event Tests
# ========================================

echo
echo -e "${YELLOW}Invalid Event Tests${NC}"

# Test 8: Wiki article missing required d-tag
wiki_no_d_tag='{
    "kind": 30818,
    "content": "Article without d-tag",
    "tags": [
        ["title", "Missing D-Tag Article"]
    ]
}'

if run_nak_test "$wiki_no_d_tag" false; then
    print_result "Wiki article missing d-tag (should fail)" true "$NIP"
else
    print_result "Wiki article missing d-tag (should fail)" false "$NIP"
fi

# Test 9: Wiki article with empty d-tag
wiki_empty_d_tag='{
    "kind": 30818,
    "content": "Article with empty d-tag",
    "tags": [
        ["d", ""],
        ["title", "Empty D-Tag Article"]
    ]
}'

if run_nak_test "$wiki_empty_d_tag" false; then
    print_result "Wiki article with empty d-tag (should fail)" true "$NIP"
else
    print_result "Wiki article with empty d-tag (should fail)" false "$NIP"
fi

# Test 10: Merge request missing required a-tag
merge_no_a_tag='{
    "kind": 818,
    "content": "Merge request without a-tag",
    "tags": [
        ["p", "'$TARGET_AUTHOR_PUBKEY'", ""],
        ["fork", "30818:'$FORK_AUTHOR_PUBKEY':forked-article"]
    ]
}'

if run_nak_test "$merge_no_a_tag" false; then
    print_result "Merge request missing a-tag (should fail)" true "$NIP"
else
    print_result "Merge request missing a-tag (should fail)" false "$NIP"
fi

# Test 11: Merge request missing required p-tag  
merge_no_p_tag='{
    "kind": 818,
    "content": "Merge request without p-tag",
    "tags": [
        ["a", "30818:'$TARGET_AUTHOR_PUBKEY':target-article", ""],
        ["fork", "30818:'$FORK_AUTHOR_PUBKEY':forked-article"]
    ]
}'

if run_nak_test "$merge_no_p_tag" false; then
    print_result "Merge request missing p-tag (should fail)" true "$NIP"
else
    print_result "Merge request missing p-tag (should fail)" false "$NIP"
fi

# Test 12: Redirect missing required redirect tag
redirect_no_redirect_tag='{
    "kind": 30819,
    "content": "",
    "tags": [
        ["d", "broken-redirect"]
    ]
}'

if run_nak_test "$redirect_no_redirect_tag" false; then
    print_result "Redirect missing redirect tag (should fail)" true "$NIP"
else
    print_result "Redirect missing redirect tag (should fail)" false "$NIP"
fi

# Test 13: Invalid a-tag format in wiki article
wiki_invalid_a_tag='{
    "kind": 30818,
    "content": "Article with invalid a-tag",
    "tags": [
        ["d", "invalid-a-tag-article"],
        ["title", "Invalid A-Tag"],
        ["a", "invalid-format", ""]
    ]
}'

if run_nak_test "$wiki_invalid_a_tag" false; then
    print_result "Wiki article with invalid a-tag format (should fail)" true "$NIP"
else
    print_result "Wiki article with invalid a-tag format (should fail)" false "$NIP"
fi

# Test 14: Invalid redirect tag format
redirect_invalid_format='{
    "kind": 30819,
    "content": "",
    "tags": [
        ["d", "invalid-redirect"],
        ["redirect", "invalid-format", ""]
    ]
}'

if run_nak_test "$redirect_invalid_format" false; then
    print_result "Redirect with invalid redirect tag format (should fail)" true "$NIP"
else
    print_result "Redirect with invalid redirect tag format (should fail)" false "$NIP"
fi

# Test 15: Wiki article with malformed wikilink syntax
wiki_malformed_wikilink='{
    "kind": 30818,
    "content": "This has malformed [[wikilink syntax and [[another broken link",
    "tags": [
        ["d", "malformed-wikilinks"],
        ["title", "Malformed Wikilinks"]
    ]
}'

if run_nak_test "$wiki_malformed_wikilink" false; then
    print_result "Wiki article with malformed wikilinks (should fail)" true "$NIP"
else
    print_result "Wiki article with malformed wikilinks (should fail)" false "$NIP"
fi

# ========================================
# Edge Case Tests
# ========================================

echo
echo -e "${YELLOW}Edge Case Tests${NC}"

# Test 16: Wiki article with maximum allowed content length
max_content=$(printf 'a%.0s' {1..2000})  # 2000 character content
wiki_max_content="{
    \"kind\": 30818,
    \"content\": \"$max_content\",
    \"tags\": [
        [\"d\", \"max-content-article\"],
        [\"title\", \"Maximum Content Article\"]
    ]
}"

if run_nak_test "$wiki_max_content" true; then
    print_result "Wiki article with maximum content length" true "$NIP"
else
    print_result "Wiki article with maximum content length" false "$NIP"
fi

# Test 17: Wiki article with special characters in d-tag
wiki_special_chars='{
    "kind": 30818,
    "content": "Article with special characters in identifier",
    "tags": [
        ["d", "special-chars"],
        ["title", "Special Characters Test"]
    ]
}'

if run_nak_test "$wiki_special_chars" true; then
    print_result "Wiki article with special characters in d-tag" true "$NIP"
else
    print_result "Wiki article with special characters in d-tag" false "$NIP"
fi

# Test 18: Multiple redirects chaining
redirect_chain="{
    \"kind\": 30819,
    \"content\": \"First redirect in chain\",
    \"tags\": [
        [\"d\", \"redirect-chain-first\"],
        [\"redirect\", \"30819:$TEST_AUTHOR_PUBKEY:redirect-chain-second\"]
    ]
}"

if run_nak_test "$redirect_chain" true; then
    print_result "Redirect chain handling" true "$NIP"
else
    print_result "Redirect chain handling" false "$NIP"
fi

# Test 19: Merge request with complex fork lineage
complex_merge='{
    "kind": 818,
    "content": "Complex merge with multiple references",
    "tags": [
        ["a", "30818:'$TARGET_AUTHOR_PUBKEY':target-article", ""],
        ["e", "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321", "", "source"],
        ["p", "'$TARGET_AUTHOR_PUBKEY'", ""],
        ["p", "'$REVIEWER_PUBKEY'", ""],
        ["fork", "30818:'$FORK_AUTHOR_PUBKEY':forked-article"],
        ["fork", "30818:'$TEST_AUTHOR_PUBKEY':original-article"],
        ["merge", "30818:'$TARGET_AUTHOR_PUBKEY':target-article"],
        ["r", "https://github.com/example/diff", "View Diff"]
    ]
}'

if run_nak_test "$complex_merge" true; then
    print_result "Complex merge request with multiple references" true "$NIP"
else
    print_result "Complex merge request with multiple references" false "$NIP"
fi

# Test 20: Wiki article with Asciidoc-style content
wiki_asciidoc='{
    "kind": 30818,
    "content": "= Article Title\n\n== Section 1\n\nThis is content with *bold* and _italic_ text.\n\n[source,code]\n----\ncode block\n----\n\n== Section 2\n\nMore content with [[wikilinks]].",
    "tags": [
        ["d", "asciidoc-article"],
        ["title", "Asciidoc Article"],
        ["format", "asciidoc"]
    ]
}'

if run_nak_test "$wiki_asciidoc" true; then
    print_result "Wiki article with Asciidoc formatting" true "$NIP"
else
    print_result "Wiki article with Asciidoc formatting" false "$NIP"
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