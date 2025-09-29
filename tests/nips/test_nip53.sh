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
HOST_SECRET_KEY="1111111111111111111111111111111111111111111111111111111111111111"
PARTICIPANT_SECRET_KEY="2222222222222222222222222222222222222222222222222222222222222222"

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

echo -e "${BLUE}Starting Shugur Relay NIP-53 Tests${NC}\n"

# Test NIP-53: Live Activities
echo -e "\n${YELLOW}Testing NIP-53: Live Activities${NC}"

# Test 1: Basic live streaming event
LIVE_STREAM=$(nak event -k 30311 -c "" -t d="test-stream-001" -t title="Test Live Stream" -t status="live" --sec $HOST_SECRET_KEY $RELAY)
if [ ! -z "$LIVE_STREAM" ]; then
    print_result "Basic live streaming event" true "53"
else
    print_result "Basic live streaming event" false "53"
fi

# Test 2: Live streaming event with full details
LIVE_STREAM_DETAILED=$(nak event -k 30311 -c "" -t d="detailed-stream" -t title="Advanced Live Stream" -t summary="A comprehensive live stream with all features" -t image="https://example.com/stream.jpg" -t streaming="https://stream.example.com/live.m3u8" -t starts="1735689600" -t status="live" -t current_participants="150" -t total_participants="500" -t p="$(nak key public $HOST_SECRET_KEY),,Host" -t t="livestream" --sec $HOST_SECRET_KEY $RELAY)
if [ ! -z "$LIVE_STREAM_DETAILED" ]; then
    print_result "Live streaming event with full details" true "53"
else
    print_result "Live streaming event with full details" false "53"
fi

# Test 3: Live streaming event with participants and recording
LIVE_STREAM_PARTICIPANTS=$(nak event -k 30311 -c "" -t d="stream-with-participants" -t title="Stream with Participants" -t status="ended" -t recording="https://example.com/recording.mp4" -t p="$(nak key public $HOST_SECRET_KEY),,Host" -t p="$(nak key public $PARTICIPANT_SECRET_KEY),,Speaker" -t p="$(nak key public $TEST_SECRET_KEY),,Participant" -t relays="wss://relay1.com,wss://relay2.com" --sec $HOST_SECRET_KEY $RELAY)
if [ ! -z "$LIVE_STREAM_PARTICIPANTS" ]; then
    print_result "Live streaming event with participants and recording" true "53"
else
    print_result "Live streaming event with participants and recording" false "53"
fi

# Test 4: Live chat message
LIVE_CHAT=$(nak event -k 1311 -c "Hello everyone in the live stream!" -t a="30311:$(nak key public $HOST_SECRET_KEY):test-stream-001" --sec $PARTICIPANT_SECRET_KEY $RELAY)
if [ ! -z "$LIVE_CHAT" ]; then
    print_result "Live chat message" true "53"
else
    print_result "Live chat message" false "53"
fi

# Test 5: Live chat message with reply
LIVE_CHAT_REPLY=$(nak event -k 1311 -c "Thanks for joining!" -t a="30311:$(nak key public $HOST_SECRET_KEY):test-stream-001" -t e="7c9b1fe9a7b2c8e5d3f6a4b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8" --sec $HOST_SECRET_KEY $RELAY)
if [ ! -z "$LIVE_CHAT_REPLY" ]; then
    print_result "Live chat message with reply" true "53"
else
    print_result "Live chat message with reply" false "53"
fi

# Test 6: Live chat message with quote
LIVE_CHAT_QUOTE=$(nak event -k 1311 -c "Great point about the technology!" -t a="30311:$(nak key public $HOST_SECRET_KEY):test-stream-001" -t q="30311:$(nak key public $HOST_SECRET_KEY):detailed-stream,wss://relay.example.com" --sec $PARTICIPANT_SECRET_KEY $RELAY)
if [ ! -z "$LIVE_CHAT_QUOTE" ]; then
    print_result "Live chat message with quote" true "53"
else
    print_result "Live chat message with quote" false "53"
fi

# Test 7: Basic meeting space
MEETING_SPACE=$(nak event -k 30312 -c "" -t d="main-room" -t room="Main Conference Room" -t status="open" -t service="https://meet.example.com/main" -t p="$(nak key public $HOST_SECRET_KEY),,Owner" --sec $HOST_SECRET_KEY $RELAY)
if [ ! -z "$MEETING_SPACE" ]; then
    print_result "Basic meeting space" true "53"
else
    print_result "Basic meeting space" false "53"
fi

# Test 8: Meeting space with full details
MEETING_SPACE_DETAILED=$(nak event -k 30312 -c "" -t d="advanced-room" -t room="Advanced Meeting Room" -t summary="High-tech meeting space with all features" -t image="https://example.com/room.jpg" -t status="private" -t service="https://meet.example.com/advanced" -t endpoint="https://api.example.com/room/advanced" -t t="meeting" -t t="conference" -t p="$(nak key public $HOST_SECRET_KEY),,Owner" -t p="$(nak key public $PARTICIPANT_SECRET_KEY),,Moderator" -t relays="wss://relay1.com,wss://relay2.com" --sec $HOST_SECRET_KEY $RELAY)
if [ ! -z "$MEETING_SPACE_DETAILED" ]; then
    print_result "Meeting space with full details" true "53"
else
    print_result "Meeting space with full details" false "53"
fi

# Test 9: Meeting room event
MEETING_ROOM_EVENT=$(nak event -k 30313 -c "" -t d="weekly-standup" -t a="30312:$(nak key public $HOST_SECRET_KEY):main-room" -t title="Weekly Team Standup" -t starts="1735689600" -t status="planned" --sec $HOST_SECRET_KEY $RELAY)
if [ ! -z "$MEETING_ROOM_EVENT" ]; then
    print_result "Meeting room event" true "53"
else
    print_result "Meeting room event" false "53"
fi

# Test 10: Meeting room event with full details
MEETING_ROOM_EVENT_DETAILED=$(nak event -k 30313 -c "" -t d="quarterly-review" -t a="30312:$(nak key public $HOST_SECRET_KEY):advanced-room" -t title="Q4 Quarterly Review" -t summary="Quarterly business review meeting" -t image="https://example.com/meeting.jpg" -t starts="1735689600" -t ends="1735693200" -t status="live" -t current_participants="25" -t total_participants="30" -t p="$(nak key public $HOST_SECRET_KEY),,Speaker" -t p="$(nak key public $PARTICIPANT_SECRET_KEY),,Participant" --sec $HOST_SECRET_KEY $RELAY)
if [ ! -z "$MEETING_ROOM_EVENT_DETAILED" ]; then
    print_result "Meeting room event with full details" true "53"
else
    print_result "Meeting room event with full details" false "53"
fi

# Test 11: Room presence
ROOM_PRESENCE=$(nak event -k 10312 -c "" -t a="30312:$(nak key public $HOST_SECRET_KEY):main-room,,root" --sec $PARTICIPANT_SECRET_KEY $RELAY)
if [ ! -z "$ROOM_PRESENCE" ]; then
    print_result "Room presence" true "53"
else
    print_result "Room presence" false "53"
fi

# Test 12: Room presence with hand raised
ROOM_PRESENCE_HAND=$(nak event -k 10312 -c "" -t a="30312:$(nak key public $HOST_SECRET_KEY):advanced-room" -t hand="1" --sec $PARTICIPANT_SECRET_KEY $RELAY)
if [ ! -z "$ROOM_PRESENCE_HAND" ]; then
    print_result "Room presence with hand raised" true "53"
else
    print_result "Room presence with hand raised" false "53"
fi

# Test invalid events (should fail)

# Test 13: Live streaming event without d tag (should fail)
INVALID_STREAM_NO_D=$(nak event -k 30311 -c "" -t title="Invalid Stream" -t status="live" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_NO_D" == *"missing required 'd' tag"* ]] || [[ "$INVALID_STREAM_NO_D" == *"refused"* ]]; then
    print_result "Live streaming event without d tag (properly rejected)" true "53"
else
    print_result "Live streaming event without d tag (improperly accepted)" false "53"
fi

# Test 14: Live streaming event with invalid status (should fail)
INVALID_STREAM_STATUS=$(nak event -k 30311 -c "" -t d="invalid-stream" -t title="Invalid Stream" -t status="invalid-status" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_STATUS" == *"status"* ]] || [[ "$INVALID_STREAM_STATUS" == *"refused"* ]]; then
    print_result "Live streaming event with invalid status (properly rejected)" true "53"
else
    print_result "Live streaming event with invalid status (improperly accepted)" false "53"
fi

# Test 15: Live streaming event with invalid participant count (should fail)
INVALID_STREAM_COUNT=$(nak event -k 30311 -c "" -t d="invalid-stream" -t title="Invalid Stream" -t current_participants="invalid-number" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_COUNT" == *"count"* ]] || [[ "$INVALID_STREAM_COUNT" == *"refused"* ]]; then
    print_result "Live streaming event with invalid participant count (properly rejected)" true "53"
else
    print_result "Live streaming event with invalid participant count (improperly accepted)" false "53"
fi

# Test 16: Live streaming event with invalid streaming URL (should fail)
INVALID_STREAM_URL=$(nak event -k 30311 -c "" -t d="invalid-stream" -t title="Invalid Stream" -t streaming="not-a-url" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_URL" == *"streaming"* ]] || [[ "$INVALID_STREAM_URL" == *"refused"* ]]; then
    print_result "Live streaming event with invalid streaming URL (properly rejected)" true "53"
else
    print_result "Live streaming event with invalid streaming URL (improperly accepted)" false "53"
fi

# Test 17: Live streaming event with invalid timestamp (should fail)
INVALID_STREAM_TIME=$(nak event -k 30311 -c "" -t d="invalid-stream" -t title="Invalid Stream" -t starts="invalid-timestamp" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_TIME" == *"timestamp"* ]] || [[ "$INVALID_STREAM_TIME" == *"refused"* ]]; then
    print_result "Live streaming event with invalid timestamp (properly rejected)" true "53"
else
    print_result "Live streaming event with invalid timestamp (improperly accepted)" false "53"
fi

# Test 18: Live streaming event with invalid participant pubkey (should fail)
INVALID_STREAM_PUBKEY=$(nak event -k 30311 -c "" -t d="invalid-stream" -t title="Invalid Stream" -t p="invalid-pubkey" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_PUBKEY" == *"pubkey"* ]] || [[ "$INVALID_STREAM_PUBKEY" == *"refused"* ]]; then
    print_result "Live streaming event with invalid participant pubkey (properly rejected)" true "53"
else
    print_result "Live streaming event with invalid participant pubkey (improperly accepted)" false "53"
fi

# Test 19: Live chat message without a tag (should fail)
INVALID_CHAT_NO_A=$(nak event -k 1311 -c "Invalid chat message" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_CHAT_NO_A" == *"'a' tag"* ]] || [[ "$INVALID_CHAT_NO_A" == *"refused"* ]]; then
    print_result "Live chat message without a tag (properly rejected)" true "53"
else
    print_result "Live chat message without a tag (improperly accepted)" false "53"
fi

# Test 20: Live chat message with invalid activity reference (should fail)
INVALID_CHAT_REF=$(nak event -k 1311 -c "Invalid chat message" -t a="invalid:format" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_CHAT_REF" == *"reference"* ]] || [[ "$INVALID_CHAT_REF" == *"refused"* ]]; then
    print_result "Live chat message with invalid activity reference (properly rejected)" true "53"
else
    print_result "Live chat message with invalid activity reference (improperly accepted)" false "53"
fi

# Test 21: Live chat message with invalid reply event ID (should fail)
INVALID_CHAT_REPLY=$(nak event -k 1311 -c "Invalid chat reply" -t a="30311:$(nak key public $HOST_SECRET_KEY):test-stream-001" -t e="invalid-event-id" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_CHAT_REPLY" == *"event ID"* ]] || [[ "$INVALID_CHAT_REPLY" == *"refused"* ]]; then
    print_result "Live chat message with invalid reply event ID (properly rejected)" true "53"
else
    print_result "Live chat message with invalid reply event ID (improperly accepted)" false "53"
fi

# Test 22: Meeting space without d tag (should fail)
INVALID_SPACE_NO_D=$(nak event -k 30312 -c "" -t room="Invalid Room" -t status="open" -t service="https://example.com" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_SPACE_NO_D" == *"missing required 'd' tag"* ]] || [[ "$INVALID_SPACE_NO_D" == *"refused"* ]]; then
    print_result "Meeting space without d tag (properly rejected)" true "53"
else
    print_result "Meeting space without d tag (improperly accepted)" false "53"
fi

# Test 23: Meeting space without room tag (should fail)
INVALID_SPACE_NO_ROOM=$(nak event -k 30312 -c "" -t d="invalid-room" -t status="open" -t service="https://example.com" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_SPACE_NO_ROOM" == *"room"* ]] || [[ "$INVALID_SPACE_NO_ROOM" == *"refused"* ]]; then
    print_result "Meeting space without room tag (properly rejected)" true "53"
else
    print_result "Meeting space without room tag (improperly accepted)" false "53"
fi

# Test 24: Meeting space without status tag (should fail)
INVALID_SPACE_NO_STATUS=$(nak event -k 30312 -c "" -t d="invalid-room" -t room="Invalid Room" -t service="https://example.com" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_SPACE_NO_STATUS" == *"status"* ]] || [[ "$INVALID_SPACE_NO_STATUS" == *"refused"* ]]; then
    print_result "Meeting space without status tag (properly rejected)" true "53"
else
    print_result "Meeting space without status tag (improperly accepted)" false "53"
fi

# Test 25: Meeting space without service tag (should fail)
INVALID_SPACE_NO_SERVICE=$(nak event -k 30312 -c "" -t d="invalid-room" -t room="Invalid Room" -t status="open" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_SPACE_NO_SERVICE" == *"service"* ]] || [[ "$INVALID_SPACE_NO_SERVICE" == *"refused"* ]]; then
    print_result "Meeting space without service tag (properly rejected)" true "53"
else
    print_result "Meeting space without service tag (improperly accepted)" false "53"
fi

# Test 26: Meeting space with invalid status (should fail)
INVALID_SPACE_STATUS=$(nak event -k 30312 -c "" -t d="invalid-room" -t room="Invalid Room" -t status="invalid-status" -t service="https://example.com" -t p="$(nak key public $HOST_SECRET_KEY),,Host" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_SPACE_STATUS" == *"status"* ]] || [[ "$INVALID_SPACE_STATUS" == *"refused"* ]]; then
    print_result "Meeting space with invalid status (properly rejected)" true "53"
else
    print_result "Meeting space with invalid status (improperly accepted)" false "53"
fi

# Test 27: Meeting space with invalid service URL (should fail)
INVALID_SPACE_SERVICE=$(nak event -k 30312 -c "" -t d="invalid-room" -t room="Invalid Room" -t status="open" -t service="not-a-url" -t p="$(nak key public $HOST_SECRET_KEY),,Host" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_SPACE_SERVICE" == *"service"* ]] || [[ "$INVALID_SPACE_SERVICE" == *"refused"* ]]; then
    print_result "Meeting space with invalid service URL (properly rejected)" true "53"
else
    print_result "Meeting space with invalid service URL (improperly accepted)" false "53"
fi

# Test 28: Meeting room event without d tag (should fail)
INVALID_ROOM_NO_D=$(nak event -k 30313 -c "" -t a="30312:$(nak key public $HOST_SECRET_KEY):main-room" -t title="Invalid Meeting" -t starts="1735689600" -t status="planned" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ROOM_NO_D" == *"missing required 'd' tag"* ]] || [[ "$INVALID_ROOM_NO_D" == *"refused"* ]]; then
    print_result "Meeting room event without d tag (properly rejected)" true "53"
else
    print_result "Meeting room event without d tag (improperly accepted)" false "53"
fi

# Test 29: Meeting room event without a tag (should fail)
INVALID_ROOM_NO_A=$(nak event -k 30313 -c "" -t d="invalid-meeting" -t title="Invalid Meeting" -t starts="1735689600" -t status="planned" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ROOM_NO_A" == *"'a' tag"* ]] || [[ "$INVALID_ROOM_NO_A" == *"refused"* ]]; then
    print_result "Meeting room event without a tag (properly rejected)" true "53"
else
    print_result "Meeting room event without a tag (improperly accepted)" false "53"
fi

# Test 30: Meeting room event without title tag (should fail)
INVALID_ROOM_NO_TITLE=$(nak event -k 30313 -c "" -t d="invalid-meeting" -t a="30312:$(nak key public $HOST_SECRET_KEY):main-room" -t starts="1735689600" -t status="planned" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ROOM_NO_TITLE" == *"title"* ]] || [[ "$INVALID_ROOM_NO_TITLE" == *"refused"* ]]; then
    print_result "Meeting room event without title tag (properly rejected)" true "53"
else
    print_result "Meeting room event without title tag (improperly accepted)" false "53"
fi

# Test 31: Meeting room event without starts tag (should fail)
INVALID_ROOM_NO_STARTS=$(nak event -k 30313 -c "" -t d="invalid-meeting" -t a="30312:$(nak key public $HOST_SECRET_KEY):main-room" -t title="Invalid Meeting" -t status="planned" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ROOM_NO_STARTS" == *"starts"* ]] || [[ "$INVALID_ROOM_NO_STARTS" == *"refused"* ]]; then
    print_result "Meeting room event without starts tag (properly rejected)" true "53"
else
    print_result "Meeting room event without starts tag (improperly accepted)" false "53"
fi

# Test 32: Meeting room event without status tag (should fail)
INVALID_ROOM_NO_STATUS=$(nak event -k 30313 -c "" -t d="invalid-meeting" -t a="30312:$(nak key public $HOST_SECRET_KEY):main-room" -t title="Invalid Meeting" -t starts="1735689600" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ROOM_NO_STATUS" == *"status"* ]] || [[ "$INVALID_ROOM_NO_STATUS" == *"refused"* ]]; then
    print_result "Meeting room event without status tag (properly rejected)" true "53"
else
    print_result "Meeting room event without status tag (improperly accepted)" false "53"
fi

# Test 33: Meeting room event with invalid meeting space reference (should fail)
INVALID_ROOM_REF=$(nak event -k 30313 -c "" -t d="invalid-meeting" -t a="invalid:format" -t title="Invalid Meeting" -t starts="1735689600" -t status="planned" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_ROOM_REF" == *"reference"* ]] || [[ "$INVALID_ROOM_REF" == *"refused"* ]]; then
    print_result "Meeting room event with invalid meeting space reference (properly rejected)" true "53"
else
    print_result "Meeting room event with invalid meeting space reference (improperly accepted)" false "53"
fi

# Test 34: Room presence without a tag (should fail)
INVALID_PRESENCE_NO_A=$(nak event -k 10312 -c "" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_PRESENCE_NO_A" == *"'a' tag"* ]] || [[ "$INVALID_PRESENCE_NO_A" == *"refused"* ]]; then
    print_result "Room presence without a tag (properly rejected)" true "53"
else
    print_result "Room presence without a tag (improperly accepted)" false "53"
fi

# Test 35: Room presence with invalid room reference (should fail)
INVALID_PRESENCE_REF=$(nak event -k 10312 -c "" -t a="invalid:format" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_PRESENCE_REF" == *"reference"* ]] || [[ "$INVALID_PRESENCE_REF" == *"refused"* ]]; then
    print_result "Room presence with invalid room reference (properly rejected)" true "53"
else
    print_result "Room presence with invalid room reference (improperly accepted)" false "53"
fi

# Test 36: Room presence with invalid hand value (should fail)
INVALID_PRESENCE_HAND=$(nak event -k 10312 -c "" -t a="30312:$(nak key public $HOST_SECRET_KEY):main-room" -t hand="invalid" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_PRESENCE_HAND" == *"hand"* ]] || [[ "$INVALID_PRESENCE_HAND" == *"refused"* ]]; then
    print_result "Room presence with invalid hand value (properly rejected)" true "53"
else
    print_result "Room presence with invalid hand value (improperly accepted)" false "53"
fi

# Test 37: Live streaming event with invalid participant proof (should fail)
INVALID_STREAM_PROOF=$(nak event -k 30311 -c "" -t d="invalid-stream" -t title="Invalid Stream" -t p="$(nak key public $HOST_SECRET_KEY),,Host,invalid-proof" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_PROOF" == *"proof"* ]] || [[ "$INVALID_STREAM_PROOF" == *"refused"* ]]; then
    print_result "Live streaming event with invalid participant proof (properly rejected)" true "53"
else
    print_result "Live streaming event with invalid participant proof (improperly accepted)" false "53"
fi

# Test 38: Live streaming event with invalid relay URL in relays tag (should fail)
INVALID_STREAM_RELAYS=$(nak event -k 30311 -c "" -t d="invalid-stream" -t title="Invalid Stream" -t relays="invalid-relay,wss://valid.com" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_RELAYS" == *"relay"* ]] || [[ "$INVALID_STREAM_RELAYS" == *"refused"* ]]; then
    print_result "Live streaming event with invalid relay URL in relays tag (properly rejected)" true "53"
else
    print_result "Live streaming event with invalid relay URL in relays tag (improperly accepted)" false "53"
fi

# Test 39: Live streaming event with invalid pinned message ID (should fail)
INVALID_STREAM_PINNED=$(nak event -k 30311 -c "" -t d="invalid-stream" -t title="Invalid Stream" -t pinned="invalid-event-id" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_STREAM_PINNED" == *"pinned"* ]] || [[ "$INVALID_STREAM_PINNED" == *"refused"* ]]; then
    print_result "Live streaming event with invalid pinned message ID (properly rejected)" true "53"
else
    print_result "Live streaming event with invalid pinned message ID (improperly accepted)" false "53"
fi

# Test 40: Meeting space with empty room name (should fail)
INVALID_SPACE_EMPTY_ROOM=$(nak event -k 30312 -c "" -t d="invalid-room" -t room="" -t status="open" -t service="https://example.com" -t p="$(nak key public $HOST_SECRET_KEY),,Host" --sec $HOST_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_SPACE_EMPTY_ROOM" == *"room"* ]] || [[ "$INVALID_SPACE_EMPTY_ROOM" == *"refused"* ]]; then
    print_result "Meeting space with empty room name (properly rejected)" true "53"
else
    print_result "Meeting space with empty room name (improperly accepted)" false "53"
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