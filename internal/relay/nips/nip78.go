package nips

import (
	"encoding/json"
	"fmt"

	nostr "github.com/nbd-wtf/go-nostr"
)

// NIP-78: Application-specific data
// https://github.com/nostr-protocol/nips/blob/master/78.md

// ValidateApplicationSpecificData validates NIP-78 application-specific data events (kind 30078)
func ValidateApplicationSpecificData(evt *nostr.Event) error {
	if evt.Kind != 30078 {
		return fmt.Errorf("invalid event kind for application-specific data: %d", evt.Kind)
	}

	// Must have "d" tag for parameterized replaceable events
	hasDTag := false
	hasPTag := false

	for _, tag := range evt.Tags {
		if len(tag) >= 2 {
			switch tag[0] {
			case "d":
				hasDTag = true
			case "p":
				hasPTag = true
				// Validate pubkey format
				if len(tag[1]) != 64 {
					return fmt.Errorf("invalid pubkey in 'p' tag: %s", tag[1])
				}
			}
		}
	}

	if !hasDTag {
		return fmt.Errorf("application-specific data must have 'd' tag")
	}

	if !hasPTag {
		return fmt.Errorf("application-specific data must have 'p' tag")
	}

	// Parse and validate JSON content
	var appData map[string]interface{}
	if err := json.Unmarshal([]byte(evt.Content), &appData); err != nil {
		return fmt.Errorf("invalid JSON content: %v", err)
	}

	// Must have "name" field
	name, hasName := appData["name"]
	if !hasName {
		return fmt.Errorf("application-specific data must have 'name' field")
	}
	nameStr, ok := name.(string)
	if !ok {
		return fmt.Errorf("'name' field must be a string")
	}
	
	// Validate name field constraints
	if len(nameStr) > 100 {
		return fmt.Errorf("application-specific data 'name' field too long (max 100 characters)")
	}

	// Must have "data" field
	data, hasData := appData["data"]
	if !hasData {
		return fmt.Errorf("application-specific data must have 'data' field")
	}

	// "data" must be an object/map, not a string or other primitive
	if _, ok := data.(map[string]interface{}); !ok {
		return fmt.Errorf("'data' field must be an object")
	}

	// Content can be any application-specific data
	return nil
}

// IsApplicationSpecificData checks if an event is application-specific data
func IsApplicationSpecificData(evt *nostr.Event) bool {
	return evt.Kind == 30078
}

// GetApplicationDataIdentifier returns the "d" tag value (application identifier)
func GetApplicationDataIdentifier(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "d" {
			return tag[1]
		}
	}
	return ""
}

// GetApplicationDataTarget returns the "p" tag value (target pubkey)
func GetApplicationDataTarget(evt *nostr.Event) string {
	for _, tag := range evt.Tags {
		if len(tag) >= 2 && tag[0] == "p" {
			return tag[1]
		}
	}
	return ""
}
