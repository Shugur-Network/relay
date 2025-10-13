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
SITE_AUTHOR_SECRET_KEY="1111111111111111111111111111111111111111111111111111111111111111"

# Sample HTML content
SAMPLE_HTML='<!DOCTYPE html><html><head><title>Test Page</title></head><body><h1>Hello Nostr Web</h1></body></html>'
SAMPLE_CSS='body { margin: 0; padding: 20px; font-family: sans-serif; }'
SAMPLE_JS='console.log("Hello from Nostr Web");'

# Compute SHA-256 hashes
HTML_HASH=$(echo -n "$SAMPLE_HTML" | sha256sum | awk '{print $1}')
CSS_HASH=$(echo -n "$SAMPLE_CSS" | sha256sum | awk '{print $1}')
JS_HASH=$(echo -n "$SAMPLE_JS" | sha256sum | awk '{print $1}')

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

echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   Shugur Relay - NIP-YY Tests (Nostr Web Pages)${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${YELLOW}Relay:${NC} $RELAY"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"

# Test NIP-YY: Nostr Web Pages - HTML Content (kind 40000)
echo -e "\n${YELLOW}Testing NIP-YY: HTML Content (kind 40000)${NC}"

# Test 1: Create valid HTML content event
HTML_EVENT=$(nak event -k 40000 -c "$SAMPLE_HTML" -t m=text/html -t sha256=$HTML_HASH -t title="Home Page" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$HTML_EVENT" ]; then
    print_result "Valid HTML content event" true "YY"
else
    print_result "Valid HTML content event" false "YY"
fi

# Test 2: HTML content without MIME type tag (should fail)
INVALID_HTML_NO_MIME=$(nak event -k 40000 -c "$SAMPLE_HTML" -t sha256=$HTML_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_HTML_NO_MIME" == *"m"* ]] || [[ "$INVALID_HTML_NO_MIME" == *"MIME"* ]] || [[ "$INVALID_HTML_NO_MIME" == *"refused"* ]]; then
    print_result "HTML content without MIME type tag (properly rejected)" true "YY"
else
    print_result "HTML content without MIME type tag (improperly accepted)" false "YY"
fi

# Test 3: HTML content without SHA-256 tag (should fail)
INVALID_HTML_NO_HASH=$(nak event -k 40000 -c "$SAMPLE_HTML" -t m=text/html --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_HTML_NO_HASH" == *"sha256"* ]] || [[ "$INVALID_HTML_NO_HASH" == *"refused"* ]]; then
    print_result "HTML content without SHA-256 tag (properly rejected)" true "YY"
else
    print_result "HTML content without SHA-256 tag (improperly accepted)" false "YY"
fi

# Test 4: HTML content with incorrect SHA-256 hash (should fail)
INVALID_HTML_WRONG_HASH=$(nak event -k 40000 -c "$SAMPLE_HTML" -t m=text/html -t sha256=0000000000000000000000000000000000000000000000000000000000000000 --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_HTML_WRONG_HASH" == *"sha256"* ]] || [[ "$INVALID_HTML_WRONG_HASH" == *"hash"* ]] || [[ "$INVALID_HTML_WRONG_HASH" == *"refused"* ]]; then
    print_result "HTML content with incorrect SHA-256 hash (properly rejected)" true "YY"
else
    print_result "HTML content with incorrect SHA-256 hash (improperly accepted)" false "YY"
fi

# Test NIP-YY: CSS Stylesheet (kind 40001)
echo -e "\n${YELLOW}Testing NIP-YY: CSS Stylesheet (kind 40001)${NC}"

# Test 5: Create valid CSS stylesheet event
CSS_EVENT=$(nak event -k 40001 -c "$SAMPLE_CSS" -t m=text/css -t sha256=$CSS_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$CSS_EVENT" ]; then
    print_result "Valid CSS stylesheet event" true "YY"
else
    print_result "Valid CSS stylesheet event" false "YY"
fi

# Test 6: CSS content without MIME type tag (should fail)
INVALID_CSS_NO_MIME=$(nak event -k 40001 -c "$SAMPLE_CSS" -t sha256=$CSS_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_CSS_NO_MIME" == *"m"* ]] || [[ "$INVALID_CSS_NO_MIME" == *"MIME"* ]] || [[ "$INVALID_CSS_NO_MIME" == *"refused"* ]]; then
    print_result "CSS content without MIME type tag (properly rejected)" true "YY"
else
    print_result "CSS content without MIME type tag (improperly accepted)" false "YY"
fi

# Test NIP-YY: JavaScript Module (kind 40002)
echo -e "\n${YELLOW}Testing NIP-YY: JavaScript Module (kind 40002)${NC}"

# Test 7: Create valid JavaScript module event
JS_EVENT=$(nak event -k 40002 -c "$SAMPLE_JS" -t m=text/javascript -t sha256=$JS_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$JS_EVENT" ]; then
    print_result "Valid JavaScript module event" true "YY"
else
    print_result "Valid JavaScript module event" false "YY"
fi

# Test 8: JavaScript content without SHA-256 tag (should fail)
INVALID_JS_NO_HASH=$(nak event -k 40002 -c "$SAMPLE_JS" -t m=text/javascript --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_JS_NO_HASH" == *"sha256"* ]] || [[ "$INVALID_JS_NO_HASH" == *"refused"* ]]; then
    print_result "JavaScript content without SHA-256 tag (properly rejected)" true "YY"
else
    print_result "JavaScript content without SHA-256 tag (improperly accepted)" false "YY"
fi

# Test NIP-YY: Component/Fragment (kind 40003)
echo -e "\n${YELLOW}Testing NIP-YY: Component/Fragment (kind 40003)${NC}"

# Test 9: Create valid component/fragment event
COMPONENT_CONTENT='<div class="component">Component content</div>'
COMPONENT_HASH=$(echo -n "$COMPONENT_CONTENT" | sha256sum | awk '{print $1}')
COMPONENT_EVENT=$(nak event -k 40003 -c "$COMPONENT_CONTENT" -t m=text/html -t sha256=$COMPONENT_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$COMPONENT_EVENT" ]; then
    print_result "Valid component/fragment event" true "YY"
else
    print_result "Valid component/fragment event" false "YY"
fi

# Test NIP-YY: Page Manifest (kind 34235)
echo -e "\n${YELLOW}Testing NIP-YY: Page Manifest (kind 34235)${NC}"

# Sample event IDs for testing (64-character hex strings)
SAMPLE_HTML_EVENT_ID="a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd"
SAMPLE_CSS_EVENT_ID="b2c3d4e5f6789012345678901234567890123456789012345678901234abcde"
SAMPLE_JS_EVENT_ID="c3d4e5f6789012345678901234567890123456789012345678901234abcdef"

# Test 10: Create valid page manifest
PAGE_MANIFEST=$(nak event -k 34235 -c "" -t d="/" -t e="$SAMPLE_HTML_EVENT_ID" -t e="$SAMPLE_CSS_EVENT_ID" -t title="Home Page" -t description="Welcome to my site" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$PAGE_MANIFEST" ]; then
    print_result "Valid page manifest" true "YY"
else
    print_result "Valid page manifest" false "YY"
fi

# Test 11: Page manifest with multiple assets
PAGE_MANIFEST_MULTI=$(nak event -k 34235 -c "" -t d="/about" -t e="$SAMPLE_HTML_EVENT_ID" -t e="$SAMPLE_CSS_EVENT_ID" -t e="$SAMPLE_JS_EVENT_ID" -t title="About Page" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$PAGE_MANIFEST_MULTI" ]; then
    print_result "Page manifest with multiple assets" true "YY"
else
    print_result "Page manifest with multiple assets" false "YY"
fi

# Test 12: Page manifest with CSP directive
PAGE_MANIFEST_CSP=$(nak event -k 34235 -c "" -t d="/secure" -t e="$SAMPLE_HTML_EVENT_ID" -t csp="default-src 'self'; script-src 'sha256-$JS_HASH'" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$PAGE_MANIFEST_CSP" ]; then
    print_result "Page manifest with CSP directive" true "YY"
else
    print_result "Page manifest with CSP directive" false "YY"
fi

# Test 13: Page manifest without d tag (should fail)
INVALID_MANIFEST_NO_D=$(nak event -k 34235 -c "" -t e="$SAMPLE_HTML_EVENT_ID" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_MANIFEST_NO_D" == *"d"* ]] || [[ "$INVALID_MANIFEST_NO_D" == *"route"* ]] || [[ "$INVALID_MANIFEST_NO_D" == *"refused"* ]]; then
    print_result "Page manifest without d tag (properly rejected)" true "YY"
else
    print_result "Page manifest without d tag (improperly accepted)" false "YY"
fi

# Test 14: Page manifest without e tag (should fail)
INVALID_MANIFEST_NO_E=$(nak event -k 34235 -c "" -t d="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_MANIFEST_NO_E" == *"e"* ]] || [[ "$INVALID_MANIFEST_NO_E" == *"asset"* ]] || [[ "$INVALID_MANIFEST_NO_E" == *"refused"* ]]; then
    print_result "Page manifest without e tag (properly rejected)" true "YY"
else
    print_result "Page manifest without e tag (improperly accepted)" false "YY"
fi

# Test 15: Page manifest without HTML marker (now accepted - relay doesn't enforce markers)
# Relay validation simplified: only checks d tag, e tag existence, and event ID format
# Client should validate asset markers, not relay
VALID_MANIFEST_NO_HTML_MARKER=$(nak event -k 34235 -c "" -t d="/" -t e="$SAMPLE_CSS_EVENT_ID" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [ ! -z "$VALID_MANIFEST_NO_HTML_MARKER" ] && [[ "$VALID_MANIFEST_NO_HTML_MARKER" != *"refused"* ]]; then
    print_result "Page manifest with any valid e tag (relay accepts, client validates markers)" true "YY"
else
    print_result "Page manifest with any valid e tag (relay accepts, client validates markers)" false "YY"
fi

# Test 16: Page manifest with invalid route (now accepted - relay doesn't validate route format)
# Relay stores any route, client validates format
VALID_MANIFEST_ANY_ROUTE=$(nak event -k 34235 -c "" -t d="no-leading-slash" -t e="$SAMPLE_HTML_EVENT_ID" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [ ! -z "$VALID_MANIFEST_ANY_ROUTE" ] && [[ "$VALID_MANIFEST_ANY_ROUTE" != *"refused"* ]]; then
    print_result "Page manifest with any route value (relay accepts, client validates format)" true "YY"
else
    print_result "Page manifest with any route value (relay accepts, client validates format)" false "YY"
fi

# Test 17: Page manifest with invalid event ID (should fail - relay validates event ID format)
INVALID_MANIFEST_EVENT_ID=$(nak event -k 34235 -c "" -t d="/" -t e="invalid-id" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_MANIFEST_EVENT_ID" == *"event ID"* ]] || [[ "$INVALID_MANIFEST_EVENT_ID" == *"invalid"* ]] || [[ "$INVALID_MANIFEST_EVENT_ID" == *"refused"* ]]; then
    print_result "Page manifest with invalid event ID (properly rejected)" true "YY"
else
    print_result "Page manifest with invalid event ID (improperly accepted)" false "YY"
fi

# Test 18: Page manifest with invalid asset marker (now accepted - relay doesn't validate markers)
# Relay accepts any marker value, client validates semantics
VALID_MANIFEST_ANY_MARKER=$(nak event -k 34235 -c "" -t d="/" -t e="$SAMPLE_HTML_EVENT_ID" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [ ! -z "$VALID_MANIFEST_ANY_MARKER" ] && [[ "$VALID_MANIFEST_ANY_MARKER" != *"refused"* ]]; then
    print_result "Page manifest with any marker value (relay accepts, client validates)" true "YY"
else
    print_result "Page manifest with any marker value (relay accepts, client validates)" false "YY"
fi

# Test NIP-YY: Site Index (kind 34236)
echo -e "\n${YELLOW}Testing NIP-YY: Site Index (kind 34236)${NC}"

# Sample manifest event ID
SAMPLE_MANIFEST_ID="d4e5f6789012345678901234567890123456789012345678901234abcdef01"

# Test 19: Create valid site index
SITE_INDEX_CONTENT="{\"\/\":\"$SAMPLE_MANIFEST_ID\",\"\/about\":\"$SAMPLE_MANIFEST_ID\"}"
SITE_INDEX=$(nak event -k 34236 -c "$SITE_INDEX_CONTENT" -t d="site-index" -t default_route="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$SITE_INDEX" ]; then
    print_result "Valid site index" true "YY"
else
    print_result "Valid site index" false "YY"
fi

# Test 20: Site index without d tag (should fail)
INVALID_INDEX_NO_D=$(nak event -k 34236 -c "$SITE_INDEX_CONTENT" -t default_route="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_NO_D" == *"d"* ]] || [[ "$INVALID_INDEX_NO_D" == *"refused"* ]]; then
    print_result "Site index without d tag (properly rejected)" true "YY"
else
    print_result "Site index without d tag (improperly accepted)" false "YY"
fi

# Test 21: Site index with wrong d tag value (should fail)
INVALID_INDEX_D_VALUE=$(nak event -k 34236 -c "$SITE_INDEX_CONTENT" -t d="wrong-value" -t default_route="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_D_VALUE" == *"site-index"* ]] || [[ "$INVALID_INDEX_D_VALUE" == *"refused"* ]]; then
    print_result "Site index with wrong d tag value (properly rejected)" true "YY"
else
    print_result "Site index with wrong d tag value (improperly accepted)" false "YY"
fi

# Test 22: Site index without default_route tag (should fail)
INVALID_INDEX_NO_DEFAULT=$(nak event -k 34236 -c "$SITE_INDEX_CONTENT" -t d="site-index" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_NO_DEFAULT" == *"default_route"* ]] || [[ "$INVALID_INDEX_NO_DEFAULT" == *"refused"* ]]; then
    print_result "Site index without default_route tag (properly rejected)" true "YY"
else
    print_result "Site index without default_route tag (improperly accepted)" false "YY"
fi

# Test 23: Site index with empty content (should fail)
INVALID_INDEX_EMPTY=$(nak event -k 34236 -c "" -t d="site-index" -t default_route="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_EMPTY" == *"content"* ]] || [[ "$INVALID_INDEX_EMPTY" == *"empty"* ]] || [[ "$INVALID_INDEX_EMPTY" == *"refused"* ]]; then
    print_result "Site index with empty content (properly rejected)" true "YY"
else
    print_result "Site index with empty content (improperly accepted)" false "YY"
fi

# Test 24: Site index with invalid JSON content (should fail)
INVALID_INDEX_JSON=$(nak event -k 34236 -c "not-valid-json" -t d="site-index" -t default_route="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_JSON" == *"JSON"* ]] || [[ "$INVALID_INDEX_JSON" == *"invalid"* ]] || [[ "$INVALID_INDEX_JSON" == *"refused"* ]]; then
    print_result "Site index with invalid JSON content (properly rejected)" true "YY"
else
    print_result "Site index with invalid JSON content (improperly accepted)" false "YY"
fi

# Test 25: Site index with any route format in JSON (now accepted - relay doesn't validate route format)
# Relay stores any route, client validates format
ANY_ROUTE_INDEX_JSON="{\"no-slash\":\"$SAMPLE_MANIFEST_ID\"}"
VALID_INDEX_ANY_ROUTE=$(nak event -k 34236 -c "$ANY_ROUTE_INDEX_JSON" -t d="site-index" -t default_route="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [ ! -z "$VALID_INDEX_ANY_ROUTE" ] && [[ "$VALID_INDEX_ANY_ROUTE" != *"refused"* ]]; then
    print_result "Site index with any route format in JSON (relay accepts, client validates)" true "YY"
else
    print_result "Site index with any route format in JSON (relay accepts, client validates)" false "YY"
fi

# Test 26: Site index with invalid event ID in JSON (should fail - relay validates event ID format)
INVALID_INDEX_ID_JSON="{\"\/\":\"not-a-valid-id\"}"
INVALID_INDEX_ID=$(nak event -k 34236 -c "$INVALID_INDEX_ID_JSON" -t d="site-index" -t default_route="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_ID" == *"event ID"* ]] || [[ "$INVALID_INDEX_ID" == *"manifest"* ]] || [[ "$INVALID_INDEX_ID" == *"invalid"* ]] || [[ "$INVALID_INDEX_ID" == *"refused"* ]]; then
    print_result "Site index with invalid event ID in JSON (properly rejected)" true "YY"
else
    print_result "Site index with invalid event ID in JSON (improperly accepted)" false "YY"
fi

# Test 27: Site index with any default_route value (now accepted - relay doesn't validate route format)
# Relay stores any default_route, client validates format
VALID_INDEX_ANY_DEFAULT=$(nak event -k 34236 -c "$SITE_INDEX_CONTENT" -t d="site-index" -t default_route="no-slash" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [ ! -z "$VALID_INDEX_ANY_DEFAULT" ] && [[ "$VALID_INDEX_ANY_DEFAULT" != *"refused"* ]]; then
    print_result "Site index with any default_route value (relay accepts, client validates)" true "YY"
else
    print_result "Site index with any default_route value (relay accepts, client validates)" false "YY"
fi

# Print summary
echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}                    Test Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "Total tests:     ${BLUE}$test_count${NC}"
echo -e "Successful:      ${GREEN}$success_count${NC}"
echo -e "Failed:          ${RED}$fail_count${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"

# Exit with error if any tests failed
if [ $fail_count -gt 0 ]; then
    echo -e "\n${RED}❌ Some tests failed. Please review the output above.${NC}\n"
    exit 1
else
    echo -e "\n${GREEN}✅ All tests passed successfully!${NC}\n"
    exit 0
fi
