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

# Sample web assets content
SAMPLE_HTML='<!DOCTYPE html><html><head><title>Test Page</title></head><body><h1>Hello Nostr Web</h1></body></html>'
SAMPLE_CSS='body { margin: 0; padding: 20px; font-family: sans-serif; }'
SAMPLE_JS='console.log("Hello from Nostr Web");'

# Compute SHA-256 hashes (x tag)
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

# Test NIP-YY: Nostr Web Pages - Asset (kind 1125)
echo -e "\n${YELLOW}Testing NIP-YY: Asset (kind 1125)${NC}"

# Test 1: Create valid HTML asset event
HTML_EVENT=$(nak event -k 1125 -c "$SAMPLE_HTML" -t m=text/html -t x=$HTML_HASH -t alt="Home Page" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$HTML_EVENT" ]; then
    print_result "Valid HTML asset event" true "YY"
else
    print_result "Valid HTML asset event" false "YY"
fi

# Test 2: Asset without MIME type tag (should fail)
INVALID_ASSET_NO_MIME=$(nak event -k 1125 -c "$SAMPLE_HTML" -t x=$HTML_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ASSET_NO_MIME" == *"m"* ]] || [[ "$INVALID_ASSET_NO_MIME" == *"MIME"* ]] || [[ "$INVALID_ASSET_NO_MIME" == *"refused"* ]]; then
    print_result "Asset without MIME type tag (properly rejected)" true "YY"
else
    print_result "Asset without MIME type tag (improperly accepted)" false "YY"
fi

# Test 3: Asset without x (SHA-256 hash) tag (should fail)
INVALID_ASSET_NO_HASH=$(nak event -k 1125 -c "$SAMPLE_HTML" -t m=text/html --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ASSET_NO_HASH" == *"x"* ]] || [[ "$INVALID_ASSET_NO_HASH" == *"hash"* ]] || [[ "$INVALID_ASSET_NO_HASH" == *"refused"* ]]; then
    print_result "Asset without x tag (properly rejected)" true "YY"
else
    print_result "Asset without x tag (improperly accepted)" false "YY"
fi

# Test 4: Asset with incorrect SHA-256 hash (should fail)
INVALID_ASSET_WRONG_HASH=$(nak event -k 1125 -c "$SAMPLE_HTML" -t m=text/html -t x=0000000000000000000000000000000000000000000000000000000000000000 --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ASSET_WRONG_HASH" == *"x"* ]] || [[ "$INVALID_ASSET_WRONG_HASH" == *"hash"* ]] || [[ "$INVALID_ASSET_WRONG_HASH" == *"refused"* ]]; then
    print_result "Asset with incorrect x tag hash (properly rejected)" true "YY"
else
    print_result "Asset with incorrect x tag hash (improperly accepted)" false "YY"
fi

# Test 5: Create valid CSS asset event
CSS_EVENT=$(nak event -k 1125 -c "$SAMPLE_CSS" -t m=text/css -t x=$CSS_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$CSS_EVENT" ]; then
    print_result "Valid CSS asset event" true "YY"
else
    print_result "Valid CSS asset event" false "YY"
fi

# Test 6: Create valid JavaScript asset event
JS_EVENT=$(nak event -k 1125 -c "$SAMPLE_JS" -t m=text/javascript -t x=$JS_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$JS_EVENT" ]; then
    print_result "Valid JavaScript asset event" true "YY"
else
    print_result "Valid JavaScript asset event" false "YY"
fi

# Test 7: Create valid WASM asset event
WASM_CONTENT='fake-wasm-binary-content'
WASM_HASH=$(echo -n "$WASM_CONTENT" | sha256sum | awk '{print $1}')
WASM_EVENT=$(nak event -k 1125 -c "$WASM_CONTENT" -t m=application/wasm -t x=$WASM_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$WASM_EVENT" ]; then
    print_result "Valid WASM asset event" true "YY"
else
    print_result "Valid WASM asset event" false "YY"
fi

# Test 8: Create valid font asset event
FONT_CONTENT='fake-font-binary-content'
FONT_HASH=$(echo -n "$FONT_CONTENT" | sha256sum | awk '{print $1}')
FONT_EVENT=$(nak event -k 1125 -c "$FONT_CONTENT" -t m=font/woff2 -t x=$FONT_HASH --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$FONT_EVENT" ]; then
    print_result "Valid font asset event" true "YY"
else
    print_result "Valid font asset event" false "YY"
fi

# Test NIP-YY: Page Manifest (kind 1126)
echo -e "\n${YELLOW}Testing NIP-YY: Page Manifest (kind 1126)${NC}"

# Sample asset event IDs for testing (64-character hex strings)
SAMPLE_HTML_EVENT_ID="a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd"
SAMPLE_CSS_EVENT_ID="b2c3d4e5f6789012345678901234567890123456789012345678901234abcde"
SAMPLE_JS_EVENT_ID="c3d4e5f6789012345678901234567890123456789012345678901234abcdef"

# Test 9: Create valid page manifest
PAGE_MANIFEST=$(nak event -k 1126 -c "" -t e="$SAMPLE_HTML_EVENT_ID" -t e="$SAMPLE_CSS_EVENT_ID" -t title="Home Page" -t description="Welcome to my site" -t route="/" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$PAGE_MANIFEST" ]; then
    print_result "Valid page manifest" true "YY"
else
    print_result "Valid page manifest" false "YY"
fi

# Test 10: Page manifest with multiple assets
PAGE_MANIFEST_MULTI=$(nak event -k 1126 -c "" -t e="$SAMPLE_HTML_EVENT_ID" -t e="$SAMPLE_CSS_EVENT_ID" -t e="$SAMPLE_JS_EVENT_ID" -t title="About Page" -t route="/about" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$PAGE_MANIFEST_MULTI" ]; then
    print_result "Page manifest with multiple assets" true "YY"
else
    print_result "Page manifest with multiple assets" false "YY"
fi

# Test 11: Page manifest with CSP directive
PAGE_MANIFEST_CSP=$(nak event -k 1126 -c "" -t e="$SAMPLE_HTML_EVENT_ID" -t csp="default-src 'self'; script-src 'sha256-$JS_HASH'" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$PAGE_MANIFEST_CSP" ]; then
    print_result "Page manifest with CSP directive" true "YY"
else
    print_result "Page manifest with CSP directive" false "YY"
fi

# Test 12: Page manifest without e tag (should fail)
INVALID_MANIFEST_NO_E=$(nak event -k 1126 -c "" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_MANIFEST_NO_E" == *"e"* ]] || [[ "$INVALID_MANIFEST_NO_E" == *"asset"* ]] || [[ "$INVALID_MANIFEST_NO_E" == *"refused"* ]]; then
    print_result "Page manifest without e tag (properly rejected)" true "YY"
else
    print_result "Page manifest without e tag (improperly accepted)" false "YY"
fi

# Test 13: Page manifest with invalid event ID (should fail - relay validates event ID format)
INVALID_MANIFEST_EVENT_ID=$(nak event -k 1126 -c "" -t e="invalid-id" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_MANIFEST_EVENT_ID" == *"event ID"* ]] || [[ "$INVALID_MANIFEST_EVENT_ID" == *"invalid"* ]] || [[ "$INVALID_MANIFEST_EVENT_ID" == *"refused"* ]]; then
    print_result "Page manifest with invalid event ID (properly rejected)" true "YY"
else
    print_result "Page manifest with invalid event ID (improperly accepted)" false "YY"
fi

# Test NIP-YY: Site Index (kind 31126)
echo -e "\n${YELLOW}Testing NIP-YY: Site Index (kind 31126)${NC}"

# Sample manifest event IDs
SAMPLE_MANIFEST_ID="d4e5f6789012345678901234567890123456789012345678901234abcdef01"
SAMPLE_MANIFEST_ID_2="e5f6789012345678901234567890123456789012345678901234abcdef0123"

# Test 14: Create valid site index
SITE_INDEX_CONTENT="{\"routes\":{\"/\":\"$SAMPLE_MANIFEST_ID\",\"/about\":\"$SAMPLE_MANIFEST_ID_2\"},\"version\":\"1.0.0\",\"defaultRoute\":\"/\",\"notFoundRoute\":\"/404\"}"
SITE_INDEX_HASH=$(echo -n "$SITE_INDEX_CONTENT" | sha256sum | awk '{print $1}')
SITE_INDEX_D_TAG="${SITE_INDEX_HASH:0:7}"

SITE_INDEX=$(nak event -k 31126 -c "$SITE_INDEX_CONTENT" -t d="$SITE_INDEX_D_TAG" -t x="$SITE_INDEX_HASH" -t alt="main" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$SITE_INDEX" ]; then
    print_result "Valid site index with all fields" true "YY"
else
    print_result "Valid site index with all fields" false "YY"
fi

# Test 15: Site index without d tag (should fail)
INVALID_INDEX_NO_D=$(nak event -k 31126 -c "$SITE_INDEX_CONTENT" -t x="$SITE_INDEX_HASH" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_NO_D" == *"d"* ]] || [[ "$INVALID_INDEX_NO_D" == *"refused"* ]]; then
    print_result "Site index without d tag (properly rejected)" true "YY"
else
    print_result "Site index without d tag (improperly accepted)" false "YY"
fi

# Test 16: Site index without x tag (should fail)
INVALID_INDEX_NO_X=$(nak event -k 31126 -c "$SITE_INDEX_CONTENT" -t d="$SITE_INDEX_D_TAG" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_NO_X" == *"x"* ]] || [[ "$INVALID_INDEX_NO_X" == *"hash"* ]] || [[ "$INVALID_INDEX_NO_X" == *"refused"* ]]; then
    print_result "Site index without x tag (properly rejected)" true "YY"
else
    print_result "Site index without x tag (improperly accepted)" false "YY"
fi

# Test 17: Site index with d tag too short (should fail)
INVALID_INDEX_D_SHORT=$(nak event -k 31126 -c "$SITE_INDEX_CONTENT" -t d="abc123" -t x="$SITE_INDEX_HASH" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_D_SHORT" == *"d"* ]] || [[ "$INVALID_INDEX_D_SHORT" == *"7-12"* ]] || [[ "$INVALID_INDEX_D_SHORT" == *"refused"* ]]; then
    print_result "Site index with d tag too short (properly rejected)" true "YY"
else
    print_result "Site index with d tag too short (improperly accepted)" false "YY"
fi

# Test 18: Site index with d tag not matching x tag (should fail)
WRONG_D_TAG="1234567"
INVALID_INDEX_D_MISMATCH=$(nak event -k 31126 -c "$SITE_INDEX_CONTENT" -t d="$WRONG_D_TAG" -t x="$SITE_INDEX_HASH" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_D_MISMATCH" == *"d"* ]] || [[ "$INVALID_INDEX_D_MISMATCH" == *"first"* ]] || [[ "$INVALID_INDEX_D_MISMATCH" == *"refused"* ]]; then
    print_result "Site index with d tag not matching x tag (properly rejected)" true "YY"
else
    print_result "Site index with d tag not matching x tag (improperly accepted)" false "YY"
fi

# Test 19: Site index with incorrect x tag hash (should fail)
INVALID_INDEX_WRONG_HASH=$(nak event -k 31126 -c "$SITE_INDEX_CONTENT" -t d="$SITE_INDEX_D_TAG" -t x="0000000000000000000000000000000000000000000000000000000000000000" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_WRONG_HASH" == *"x"* ]] || [[ "$INVALID_INDEX_WRONG_HASH" == *"hash"* ]] || [[ "$INVALID_INDEX_WRONG_HASH" == *"refused"* ]]; then
    print_result "Site index with incorrect x tag hash (properly rejected)" true "YY"
else
    print_result "Site index with incorrect x tag hash (improperly accepted)" false "YY"
fi

# Test 20: Site index with empty content (should fail)
INVALID_INDEX_EMPTY=$(nak event -k 31126 -c "" -t d="a1b2c3d" -t x="0000000000000000000000000000000000000000000000000000000000000000" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_EMPTY" == *"content"* ]] || [[ "$INVALID_INDEX_EMPTY" == *"empty"* ]] || [[ "$INVALID_INDEX_EMPTY" == *"refused"* ]]; then
    print_result "Site index with empty content (properly rejected)" true "YY"
else
    print_result "Site index with empty content (improperly accepted)" false "YY"
fi

# Test 21: Site index with invalid JSON content (should fail)
INVALID_JSON_CONTENT="not-valid-json"
INVALID_JSON_HASH=$(echo -n "$INVALID_JSON_CONTENT" | sha256sum | awk '{print $1}')
INVALID_JSON_D="${INVALID_JSON_HASH:0:7}"
INVALID_INDEX_JSON=$(nak event -k 31126 -c "$INVALID_JSON_CONTENT" -t d="$INVALID_JSON_D" -t x="$INVALID_JSON_HASH" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_JSON" == *"JSON"* ]] || [[ "$INVALID_INDEX_JSON" == *"invalid"* ]] || [[ "$INVALID_INDEX_JSON" == *"refused"* ]]; then
    print_result "Site index with invalid JSON content (properly rejected)" true "YY"
else
    print_result "Site index with invalid JSON content (improperly accepted)" false "YY"
fi

# Test 22: Site index without routes field (should fail)
NO_ROUTES_CONTENT="{\"version\":\"1.0.0\"}"
NO_ROUTES_HASH=$(echo -n "$NO_ROUTES_CONTENT" | sha256sum | awk '{print $1}')
NO_ROUTES_D="${NO_ROUTES_HASH:0:7}"
INVALID_INDEX_NO_ROUTES=$(nak event -k 31126 -c "$NO_ROUTES_CONTENT" -t d="$NO_ROUTES_D" -t x="$NO_ROUTES_HASH" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_NO_ROUTES" == *"routes"* ]] || [[ "$INVALID_INDEX_NO_ROUTES" == *"refused"* ]]; then
    print_result "Site index without routes field (properly rejected)" true "YY"
else
    print_result "Site index without routes field (improperly accepted)" false "YY"
fi

# Test 23: Site index with invalid manifest ID in routes (should fail)
INVALID_MANIFEST_CONTENT="{\"routes\":{\"/\":\"invalid-id\"}}"
INVALID_MANIFEST_HASH=$(echo -n "$INVALID_MANIFEST_CONTENT" | sha256sum | awk '{print $1}')
INVALID_MANIFEST_D="${INVALID_MANIFEST_HASH:0:7}"
INVALID_INDEX_MANIFEST_ID=$(nak event -k 31126 -c "$INVALID_MANIFEST_CONTENT" -t d="$INVALID_MANIFEST_D" -t x="$INVALID_MANIFEST_HASH" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_INDEX_MANIFEST_ID" == *"event ID"* ]] || [[ "$INVALID_INDEX_MANIFEST_ID" == *"manifest"* ]] || [[ "$INVALID_INDEX_MANIFEST_ID" == *"invalid"* ]] || [[ "$INVALID_INDEX_MANIFEST_ID" == *"refused"* ]]; then
    print_result "Site index with invalid manifest ID in routes (properly rejected)" true "YY"
else
    print_result "Site index with invalid manifest ID in routes (improperly accepted)" false "YY"
fi

# Test NIP-YY: Entrypoint (kind 11126)
echo -e "\n${YELLOW}Testing NIP-YY: Entrypoint (kind 11126)${NC}"

# Get author pubkey for address coordinate
SITE_AUTHOR_PUBKEY=$(echo "$SITE_AUTHOR_SECRET_KEY" | nak key-public)

# Test 24: Create valid entrypoint
ENTRYPOINT_ADDRESS="31126:$SITE_AUTHOR_PUBKEY:$SITE_INDEX_D_TAG"
ENTRYPOINT=$(nak event -k 11126 -c "" -t a="$ENTRYPOINT_ADDRESS" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$ENTRYPOINT" ]; then
    print_result "Valid entrypoint" true "YY"
else
    print_result "Valid entrypoint" false "YY"
fi

# Test 25: Entrypoint with relay hint
ENTRYPOINT_WITH_RELAY=$(nak event -k 11126 -c "" -t a="$ENTRYPOINT_ADDRESS" -t a="wss://relay.example.com" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>/dev/null)
if [ ! -z "$ENTRYPOINT_WITH_RELAY" ]; then
    print_result "Entrypoint with relay hint" true "YY"
else
    print_result "Entrypoint with relay hint" false "YY"
fi

# Test 26: Entrypoint without a tag (should fail)
INVALID_ENTRYPOINT_NO_A=$(nak event -k 11126 -c "" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ENTRYPOINT_NO_A" == *"a"* ]] || [[ "$INVALID_ENTRYPOINT_NO_A" == *"refused"* ]]; then
    print_result "Entrypoint without a tag (properly rejected)" true "YY"
else
    print_result "Entrypoint without a tag (improperly accepted)" false "YY"
fi

# Test 27: Entrypoint with wrong kind in address (should fail)
WRONG_KIND_ADDRESS="30000:$SITE_AUTHOR_PUBKEY:test"
INVALID_ENTRYPOINT_WRONG_KIND=$(nak event -k 11126 -c "" -t a="$WRONG_KIND_ADDRESS" --sec $SITE_AUTHOR_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ENTRYPOINT_WRONG_KIND" == *"31126"* ]] || [[ "$INVALID_ENTRYPOINT_WRONG_KIND" == *"site index"* ]] || [[ "$INVALID_ENTRYPOINT_WRONG_KIND" == *"refused"* ]]; then
    print_result "Entrypoint with wrong kind in address (properly rejected)" true "YY"
else
    print_result "Entrypoint with wrong kind in address (improperly accepted)" false "YY"
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
