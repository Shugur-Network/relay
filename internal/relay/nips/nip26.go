package nips

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	nostr "github.com/nbd-wtf/go-nostr"
)

// DelegationInfo holds the parsed delegation data
type DelegationInfo struct {
	MasterPubkey string
	Conditions   string
	Sig          string
}

// ExtractDelegationTag finds and returns the `delegation` tag if present
func ExtractDelegationTag(evt nostr.Event) *DelegationInfo {
	for _, tag := range evt.Tags {
		if len(tag) >= 4 && tag[0] == "delegation" {
			return &DelegationInfo{
				MasterPubkey: tag[1],
				Conditions:   tag[2],
				Sig:          tag[3],
			}
		}
	}
	return nil
}

// ValidateDelegation verifies that the delegation signature is valid and conditions match
func ValidateDelegation(evt *nostr.Event, del *DelegationInfo) error {
	// 1) Check the Schnorr signature
	if !checkSig(del.MasterPubkey, del.Sig, del.Conditions, evt.PubKey) {
		return errors.New("invalid delegation signature")
	}

	// 2) Check conditions (e.g. "kind=1&created_at>1670000000")
	if err := checkConditions(del.Conditions, evt); err != nil {
		return fmt.Errorf("delegation conditions not met: %w", err)
	}
	return nil
}

// checkSig does a Schnorr verification using the method on the *Signature object
// (signature.Verify(hash, pubKey)).
func checkSig(masterPub, sig, conditions, delegatePub string) bool {
	// Build the message = (conditions + ":" + delegatePub)
	msg := []byte(conditions + ":" + delegatePub)
	h := sha256.Sum256(msg)

	// Decode masterPub from hex to bytes
	pubKeyBytes, err := hex.DecodeString(masterPub)
	if err != nil {
		return false
	}
	// Parse the master pubkey (BIP-340 x-only)
	pubKey, err := schnorr.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false
	}

	// Decode the signature from hex
	sigBytes, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	parsedSig, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return false
	}

	// Use the *Signature's Verify method
	// Verify(hash []byte, pubKey *btcec.PublicKey) bool
	ok := parsedSig.Verify(h[:], pubKey)
	return ok
}

// checkConditions parses something like "kind=1&created_at>1670000000"
// and compares with evt.Kind, evt.CreatedAt, etc.
func checkConditions(conds string, evt *nostr.Event) error {
	if conds == "" {
		// no conditions => no restriction
		return nil
	}
	parts := strings.Split(conds, "&")
	for _, p := range parts {
		if err := checkSingleCondition(p, evt); err != nil {
			return err
		}
	}
	return nil
}

// checkSingleCondition handles e.g. "kind=1" or "created_at>1670000000"
func checkSingleCondition(cond string, evt *nostr.Event) error {
	switch {
	case strings.HasPrefix(cond, "kind="):
		val := strings.TrimPrefix(cond, "kind=")
		wantKind, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid kind in delegation conditions: %s", val)
		}
		if evt.Kind != wantKind {
			return fmt.Errorf("event kind %d != required %d", evt.Kind, wantKind)
		}

	case strings.HasPrefix(cond, "created_at>"):
		val := strings.TrimPrefix(cond, "created_at>")
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid created_at> in delegation: %s", val)
		}
		if evt.CreatedAt.Time().Unix() <= num {
			return fmt.Errorf("event created_at %d is not > %d",
				evt.CreatedAt.Time().Unix(), num)
		}

	case strings.HasPrefix(cond, "created_at<"):
		val := strings.TrimPrefix(cond, "created_at<")
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid created_at< in delegation: %s", val)
		}
		if evt.CreatedAt.Time().Unix() >= num {
			return fmt.Errorf("event created_at %d is not < %d",
				evt.CreatedAt.Time().Unix(), num)
		}

	default:
		return fmt.Errorf("unsupported delegation condition: %s", cond)
	}
	return nil
}
