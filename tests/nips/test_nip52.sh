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
ORGANIZER_SECRET_KEY="1111111111111111111111111111111111111111111111111111111111111111"
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

echo -e "${BLUE}Starting Shugur Relay NIP-52 Tests${NC}\n"

# Test NIP-52: Calendar Events
echo -e "\n${YELLOW}Testing NIP-52: Calendar Events${NC}"

# Test 1: Create a basic date-based calendar event
DATE_BASED_EVENT=$(nak event -k 31922 -c "Annual team retreat" -t d="retreat-2025" -t title="Team Retreat 2025" -t start="2025-10-15" -t end="2025-10-17" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$DATE_BASED_EVENT" ]; then
    print_result "Basic date-based calendar event" true "52"
else
    print_result "Basic date-based calendar event" false "52"
fi

# Test 2: Date-based event with summary and location
DATE_EVENT_DETAILED=$(nak event -k 31922 -c "Join us for our annual company retreat in the mountains" -t d="retreat-detailed" -t title="Company Mountain Retreat" -t summary="Three days of team building and planning" -t start="2025-11-01" -t end="2025-11-03" -t location="Mountain Resort, Colorado" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$DATE_EVENT_DETAILED" ]; then
    print_result "Date-based event with details" true "52"
else
    print_result "Date-based event with details" false "52"
fi

# Test 3: Date-based event with geohash and participants
DATE_EVENT_GEO=$(nak event -k 31922 -c "Local conference" -t d="conf-2025" -t title="Tech Conference 2025" -t start="2025-12-05" -t g="9q9hvu" -t p="$(nak key public $PARTICIPANT_SECRET_KEY),wss://relay.example.com,speaker" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$DATE_EVENT_GEO" ]; then
    print_result "Date-based event with geohash and participants" true "52"
else
    print_result "Date-based event with geohash and participants" false "52"
fi

# Test 4: Date-based event with image and references
DATE_EVENT_MEDIA=$(nak event -k 31922 -c "Annual hackathon" -t d="hackathon-2025" -t title="Code for Good Hackathon" -t start="2025-09-20" -t end="2025-09-22" -t image="https://example.com/hackathon.jpg" -t r="https://example.com/hackathon-info" -t t="hackathon" -t t="coding" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$DATE_EVENT_MEDIA" ]; then
    print_result "Date-based event with media and references" true "52"
else
    print_result "Date-based event with media and references" false "52"
fi

# Test 5: Date-based event with calendar reference
DATE_EVENT_CALENDAR=$(nak event -k 31922 -c "Quarterly team meeting" -t d="q1-meeting" -t title="Q1 Team Meeting" -t start="2025-03-15" -t a="31924:$(nak key public $ORGANIZER_SECRET_KEY):work-calendar" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$DATE_EVENT_CALENDAR" ]; then
    print_result "Date-based event with calendar reference" true "52"
else
    print_result "Date-based event with calendar reference" false "52"
fi

# Test 6: Time-based calendar event
TIME_BASED_EVENT=$(nak event -k 31923 -c "Weekly team standup" -t d="standup-001" -t title="Team Standup" -t start="1735689600" -t end="1735691400" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$TIME_BASED_EVENT" ]; then
    print_result "Basic time-based calendar event" true "52"
else
    print_result "Basic time-based calendar event" false "52"
fi

# Test 7: Time-based event with timezone
TIME_EVENT_TZ=$(nak event -k 31923 -c "Important client meeting" -t d="client-meeting-001" -t title="Client Presentation" -t summary="Q4 results presentation" -t start="1735689600" -t end="1735693200" -t start_tzid="America/New_York" -t end_tzid="America/New_York" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$TIME_EVENT_TZ" ]; then
    print_result "Time-based event with timezone" true "52"
else
    print_result "Time-based event with timezone" false "52"
fi

# Test 8: Time-based event with participants and location
TIME_EVENT_PARTICIPANTS=$(nak event -k 31923 -c "Project kickoff meeting" -t d="kickoff-001" -t title="Project Alpha Kickoff" -t start="1735689600" -t end="1735693200" -t location="Conference Room A" -t p="$(nak key public $PARTICIPANT_SECRET_KEY),,project-manager" -t p="$(nak key public $TEST_SECRET_KEY),,developer" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$TIME_EVENT_PARTICIPANTS" ]; then
    print_result "Time-based event with participants" true "52"
else
    print_result "Time-based event with participants" false "52"
fi

# Test 9: Calendar (kind 31924)
CALENDAR=$(nak event -k 31924 -c "Work calendar for team events and meetings" -t d="work-calendar" -t title="Work Calendar" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$CALENDAR" ]; then
    print_result "Basic calendar" true "52"
else
    print_result "Basic calendar" false "52"
fi

# Test 10: Calendar with event references
CALENDAR_WITH_EVENTS=$(nak event -k 31924 -c "Personal events calendar" -t d="personal-calendar" -t title="Personal Events" -t a="31922:$(nak key public $ORGANIZER_SECRET_KEY):retreat-2025,wss://relay.example.com" -t a="31923:$(nak key public $ORGANIZER_SECRET_KEY):standup-001" --sec $ORGANIZER_SECRET_KEY $RELAY)
if [ ! -z "$CALENDAR_WITH_EVENTS" ]; then
    print_result "Calendar with event references" true "52"
else
    print_result "Calendar with event references" false "52"
fi

# Test 11: Calendar Event RSVP - Accepted
RSVP_ACCEPTED=$(nak event -k 31925 -c "Looking forward to it!" -t d="rsvp-retreat-001" -t a="31922:$(nak key public $ORGANIZER_SECRET_KEY):retreat-2025" -t status="accepted" -t fb="busy" -t p="$(nak key public $ORGANIZER_SECRET_KEY)" --sec $PARTICIPANT_SECRET_KEY $RELAY)
if [ ! -z "$RSVP_ACCEPTED" ]; then
    print_result "RSVP - accepted" true "52"
else
    print_result "RSVP - accepted" false "52"
fi

# Test 12: Calendar Event RSVP - Declined
RSVP_DECLINED=$(nak event -k 31925 -c "Unfortunately I cannot attend" -t d="rsvp-conf-001" -t a="31922:$(nak key public $ORGANIZER_SECRET_KEY):conf-2025" -t status="declined" -t p="$(nak key public $ORGANIZER_SECRET_KEY)" --sec $PARTICIPANT_SECRET_KEY $RELAY)
if [ ! -z "$RSVP_DECLINED" ]; then
    print_result "RSVP - declined" true "52"
else
    print_result "RSVP - declined" false "52"
fi

# Test 13: Calendar Event RSVP - Tentative
RSVP_TENTATIVE=$(nak event -k 31925 -c "I'll try to make it" -t d="rsvp-hackathon-001" -t a="31922:$(nak key public $ORGANIZER_SECRET_KEY):hackathon-2025" -t status="tentative" -t fb="free" -t p="$(nak key public $ORGANIZER_SECRET_KEY)" --sec $PARTICIPANT_SECRET_KEY $RELAY)
if [ ! -z "$RSVP_TENTATIVE" ]; then
    print_result "RSVP - tentative" true "52"
else
    print_result "RSVP - tentative" false "52"
fi

# Test 14: RSVP with event ID reference
SAMPLE_EVENT_ID="7c9b1fe9a7b2c8e5d3f6a4b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8"
RSVP_WITH_EVENT_ID=$(nak event -k 31925 -c "RSVP with event reference" -t d="rsvp-with-e-tag" -t a="31923:$(nak key public $ORGANIZER_SECRET_KEY):standup-001" -t e="$SAMPLE_EVENT_ID,wss://relay.example.com" -t status="accepted" -t p="$(nak key public $ORGANIZER_SECRET_KEY)" --sec $PARTICIPANT_SECRET_KEY $RELAY)
if [ ! -z "$RSVP_WITH_EVENT_ID" ]; then
    print_result "RSVP with event ID reference" true "52"
else
    print_result "RSVP with event ID reference" false "52"
fi

# Test invalid events (should fail)

# Test 15: Date-based event without d tag (should fail)
INVALID_DATE_NO_D=$(nak event -k 31922 -c "Invalid event" -t title="Invalid Event" -t start="2025-10-15" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_DATE_NO_D" == *"missing required 'd' tag"* ]] || [[ "$INVALID_DATE_NO_D" == *"refused"* ]]; then
    print_result "Date-based event without d tag (properly rejected)" true "52"
else
    print_result "Date-based event without d tag (improperly accepted)" false "52"
fi

# Test 16: Date-based event without title tag (should fail)
INVALID_DATE_NO_TITLE=$(nak event -k 31922 -c "Invalid event" -t d="invalid-event" -t start="2025-10-15" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_DATE_NO_TITLE" == *"title"* ]] || [[ "$INVALID_DATE_NO_TITLE" == *"refused"* ]]; then
    print_result "Date-based event without title tag (properly rejected)" true "52"
else
    print_result "Date-based event without title tag (improperly accepted)" false "52"
fi

# Test 17: Date-based event without start tag (should fail)
INVALID_DATE_NO_START=$(nak event -k 31922 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_DATE_NO_START" == *"start"* ]] || [[ "$INVALID_DATE_NO_START" == *"refused"* ]]; then
    print_result "Date-based event without start tag (properly rejected)" true "52"
else
    print_result "Date-based event without start tag (improperly accepted)" false "52"
fi

# Test 18: Date-based event with invalid date format (should fail)
INVALID_DATE_FORMAT=$(nak event -k 31922 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="invalid-date" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_DATE_FORMAT" == *"date"* ]] || [[ "$INVALID_DATE_FORMAT" == *"refused"* ]]; then
    print_result "Date-based event with invalid date format (properly rejected)" true "52"
else
    print_result "Date-based event with invalid date format (improperly accepted)" false "52"
fi

# Test 19: Date-based event with end before start (should fail)
INVALID_DATE_ORDER=$(nak event -k 31922 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="2025-10-17" -t end="2025-10-15" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_DATE_ORDER" == *"before"* ]] || [[ "$INVALID_DATE_ORDER" == *"refused"* ]]; then
    print_result "Date-based event with invalid date order (properly rejected)" true "52"
else
    print_result "Date-based event with invalid date order (improperly accepted)" false "52"
fi

# Test 20: Time-based event without d tag (should fail)
INVALID_TIME_NO_D=$(nak event -k 31923 -c "Invalid event" -t title="Invalid Event" -t start="1735689600" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_TIME_NO_D" == *"missing required 'd' tag"* ]] || [[ "$INVALID_TIME_NO_D" == *"refused"* ]]; then
    print_result "Time-based event without d tag (properly rejected)" true "52"
else
    print_result "Time-based event without d tag (improperly accepted)" false "52"
fi

# Test 21: Time-based event with invalid timestamp (should fail)
INVALID_TIME_TIMESTAMP=$(nak event -k 31923 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="invalid-timestamp" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_TIME_TIMESTAMP" == *"timestamp"* ]] || [[ "$INVALID_TIME_TIMESTAMP" == *"refused"* ]]; then
    print_result "Time-based event with invalid timestamp (properly rejected)" true "52"
else
    print_result "Time-based event with invalid timestamp (improperly accepted)" false "52"
fi

# Test 22: Time-based event with end before start (should fail)
INVALID_TIME_ORDER=$(nak event -k 31923 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="1735693200" -t end="1735689600" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_TIME_ORDER" == *"less than"* ]] || [[ "$INVALID_TIME_ORDER" == *"refused"* ]]; then
    print_result "Time-based event with invalid time order (properly rejected)" true "52"
else
    print_result "Time-based event with invalid time order (improperly accepted)" false "52"
fi

# Test 23: Time-based event with invalid timezone (should fail)
INVALID_TIMEZONE=$(nak event -k 31923 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="1735689600" -t start_tzid="Invalid/Timezone" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_TIMEZONE" == *"timezone"* ]] || [[ "$INVALID_TIMEZONE" == *"refused"* ]]; then
    print_result "Time-based event with invalid timezone (properly rejected)" true "52"
else
    print_result "Time-based event with invalid timezone (improperly accepted)" false "52"
fi

# Test 24: Calendar without d tag (should fail)
INVALID_CALENDAR_NO_D=$(nak event -k 31924 -c "Invalid calendar" -t title="Invalid Calendar" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_CALENDAR_NO_D" == *"missing required 'd' tag"* ]] || [[ "$INVALID_CALENDAR_NO_D" == *"refused"* ]]; then
    print_result "Calendar without d tag (properly rejected)" true "52"
else
    print_result "Calendar without d tag (improperly accepted)" false "52"
fi

# Test 25: Calendar without title tag (should fail)
INVALID_CALENDAR_NO_TITLE=$(nak event -k 31924 -c "Invalid calendar" -t d="invalid-calendar" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_CALENDAR_NO_TITLE" == *"title"* ]] || [[ "$INVALID_CALENDAR_NO_TITLE" == *"refused"* ]]; then
    print_result "Calendar without title tag (properly rejected)" true "52"
else
    print_result "Calendar without title tag (improperly accepted)" false "52"
fi

# Test 26: RSVP without d tag (should fail)
INVALID_RSVP_NO_D=$(nak event -k 31925 -c "Invalid RSVP" -t a="31922:$(nak key public $ORGANIZER_SECRET_KEY):retreat-2025" -t status="accepted" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_RSVP_NO_D" == *"missing required 'd' tag"* ]] || [[ "$INVALID_RSVP_NO_D" == *"refused"* ]]; then
    print_result "RSVP without d tag (properly rejected)" true "52"
else
    print_result "RSVP without d tag (improperly accepted)" false "52"
fi

# Test 27: RSVP without a tag (should fail)
INVALID_RSVP_NO_A=$(nak event -k 31925 -c "Invalid RSVP" -t d="invalid-rsvp" -t status="accepted" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_RSVP_NO_A" == *"'a' tag"* ]] || [[ "$INVALID_RSVP_NO_A" == *"refused"* ]]; then
    print_result "RSVP without a tag (properly rejected)" true "52"
else
    print_result "RSVP without a tag (improperly accepted)" false "52"
fi

# Test 28: RSVP without status tag (should fail)
INVALID_RSVP_NO_STATUS=$(nak event -k 31925 -c "Invalid RSVP" -t d="invalid-rsvp" -t a="31922:$(nak key public $ORGANIZER_SECRET_KEY):retreat-2025" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_RSVP_NO_STATUS" == *"status"* ]] || [[ "$INVALID_RSVP_NO_STATUS" == *"refused"* ]]; then
    print_result "RSVP without status tag (properly rejected)" true "52"
else
    print_result "RSVP without status tag (improperly accepted)" false "52"
fi

# Test 29: RSVP with invalid status (should fail)
INVALID_RSVP_STATUS=$(nak event -k 31925 -c "Invalid RSVP" -t d="invalid-rsvp" -t a="31922:$(nak key public $ORGANIZER_SECRET_KEY):retreat-2025" -t status="invalid-status" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_RSVP_STATUS" == *"status"* ]] || [[ "$INVALID_RSVP_STATUS" == *"refused"* ]]; then
    print_result "RSVP with invalid status (properly rejected)" true "52"
else
    print_result "RSVP with invalid status (improperly accepted)" false "52"
fi

# Test 30: RSVP with invalid fb value (should fail)
INVALID_RSVP_FB=$(nak event -k 31925 -c "Invalid RSVP" -t d="invalid-rsvp" -t a="31922:$(nak key public $ORGANIZER_SECRET_KEY):retreat-2025" -t status="accepted" -t fb="invalid-fb" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_RSVP_FB" == *"fb"* ]] || [[ "$INVALID_RSVP_FB" == *"refused"* ]]; then
    print_result "RSVP with invalid fb value (properly rejected)" true "52"
else
    print_result "RSVP with invalid fb value (improperly accepted)" false "52"
fi

# Test 31: Event with invalid geohash (should fail)
INVALID_GEOHASH=$(nak event -k 31922 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="2025-10-15" -t g="invalid!hash" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_GEOHASH" == *"geohash"* ]] || [[ "$INVALID_GEOHASH" == *"refused"* ]]; then
    print_result "Event with invalid geohash (properly rejected)" true "52"
else
    print_result "Event with invalid geohash (improperly accepted)" false "52"
fi

# Test 32: Event with invalid participant pubkey (should fail)
INVALID_PARTICIPANT=$(nak event -k 31922 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="2025-10-15" -t p="invalid-pubkey" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_PARTICIPANT" == *"pubkey"* ]] || [[ "$INVALID_PARTICIPANT" == *"refused"* ]]; then
    print_result "Event with invalid participant pubkey (properly rejected)" true "52"
else
    print_result "Event with invalid participant pubkey (improperly accepted)" false "52"
fi

# Test 33: Event with invalid image URL (should fail)
INVALID_IMAGE=$(nak event -k 31922 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="2025-10-15" -t image="not-a-url" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_IMAGE" == *"image"* ]] || [[ "$INVALID_IMAGE" == *"refused"* ]]; then
    print_result "Event with invalid image URL (properly rejected)" true "52"
else
    print_result "Event with invalid image URL (improperly accepted)" false "52"
fi

# Test 34: Event with invalid reference URL (should fail)
INVALID_REFERENCE=$(nak event -k 31922 -c "Invalid event" -t d="invalid-event" -t title="Invalid Event" -t start="2025-10-15" -t r="ftp://invalid-scheme.com" --sec $ORGANIZER_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_REFERENCE" == *"reference"* ]] || [[ "$INVALID_REFERENCE" == *"refused"* ]]; then
    print_result "Event with invalid reference URL (properly rejected)" true "52"
else
    print_result "Event with invalid reference URL (improperly accepted)" false "52"
fi

# Test 35: RSVP with invalid calendar event reference (should fail)
INVALID_RSVP_REFERENCE=$(nak event -k 31925 -c "Invalid RSVP" -t d="invalid-rsvp" -t a="invalid:format" -t status="accepted" --sec $PARTICIPANT_SECRET_KEY $RELAY 2>&1)
if [[ "$INVALID_RSVP_REFERENCE" == *"reference"* ]] || [[ "$INVALID_RSVP_REFERENCE" == *"refused"* ]]; then
    print_result "RSVP with invalid calendar event reference (properly rejected)" true "52"
else
    print_result "RSVP with invalid calendar event reference (improperly accepted)" false "52"
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