#!/usr/bin/env python3
"""
NIP-XX Time Capsules Validation Script
======================================

Strict implementation following NIP-XX specification with real drand and real encryption.

Requirements:
1. Creates 2 capsules: public and private
2. Uses real drand networks and real age/tlock encryption 
3. Waits for timelock expiration and decrypts to show results
4. Posts events to relay for validation

Dependencies:
- tle (tlock encryption): go install github.com/drand/tlock/cmd/tle@latest  
- nak (nostr events): go install github.com/fiatjaf/nak@latest
- python packages: websocket-client requests
"""

import json
import time
import secrets
import subprocess
import websocket
import requests
import hashlib
import sys
from datetime import datetime


class NIPXXValidator:
    """NIP-XX Time Capsules Validator - Strict specification compliance"""
    
    def __init__(self, relay_url="wss://shu01.shugur.net"):
        self.relay_url = relay_url
        
        # Real drand networks (unchained production networks only)
        self.drand_networks = {
            "quicknet": {
                "name": "Quicknet", 
                "hash": "52db9ba70e0cc0f6eaf7803dd07447a1f5477735fd3f661792ba94600c84e971",
                "api": "https://api.drand.sh",
                "period": 3,  # 3 second rounds
                "description": "Primary unchained drand network"
            }
        }
    
    def check_dependencies(self):
        """Verify required tools are available"""
        missing = []
        
        for tool in ["tle", "nak"]:
            try:
                subprocess.run([tool, "--help"], capture_output=True, timeout=5)
            except (subprocess.TimeoutExpired, FileNotFoundError):
                missing.append(tool)
        
        try:
            import websocket
            import requests
        except ImportError as e:
            missing.append(f"python-{e.name}")
        
        if missing:
            raise RuntimeError(f"Missing dependencies: {', '.join(missing)}")
        
        print("✅ All dependencies available")

    def get_current_round(self, chain_hash, api_url):
        """Get current drand round from network"""
        try:
            response = requests.get(f"{api_url}/{chain_hash}/public/latest", timeout=10)
            response.raise_for_status()
            data = response.json()
            return int(data["round"])
        except Exception as e:
            raise RuntimeError(f"Failed to get current round: {e}")

    def calculate_target_round(self, unlock_time, chain_hash, api_url):
        """Calculate target round for given unlock time"""
        try:
            # Get current round and info
            current_round = self.get_current_round(chain_hash, api_url)
            
            # Get network info for period  
            info_response = requests.get(f"{api_url}/{chain_hash}/info", timeout=10)
            info_response.raise_for_status()
            info = info_response.json()
            period = info.get("period", 3)
            
            # Calculate rounds needed
            current_time = int(time.time())
            seconds_needed = max(0, unlock_time - current_time)
            rounds_needed = (seconds_needed + period - 1) // period  # Round up
            
            return current_round + rounds_needed
        except Exception as e:
            raise RuntimeError(f"Failed to calculate target round: {e}")

    def tlock_encrypt(self, plaintext, target_round, chain_hash):
        """Encrypt plaintext using tlock with real drand network"""
        try:
            # Use tle for real tlock encryption - binary input/output
            result = subprocess.run([
                "tle", "--encrypt",
                "--chain", chain_hash,
                "--round", str(target_round)
            ], input=plaintext.encode('utf-8'), capture_output=True, timeout=30)
            
            if result.returncode != 0:
                raise RuntimeError(f"tlock encryption failed: {result.stderr.decode()}")
            
            return result.stdout  # Return binary data
        except Exception as e:
            raise RuntimeError(f"tlock encryption error: {e}")

    def tlock_decrypt(self, ciphertext, chain_hash):
        """Decrypt ciphertext using tlock with real drand network"""
        try:
            # Use tle for real tlock decryption
            result = subprocess.run([
                "tle", "--decrypt",
                "--chain", chain_hash
            ], input=ciphertext, capture_output=True, timeout=30)
            
            if result.returncode != 0:
                raise RuntimeError(f"tlock decryption failed: {result.stderr.decode()}")
            
            return result.stdout.decode('utf-8')
        except Exception as e:
            raise RuntimeError(f"tlock decryption error: {e}")

    def nip44_encrypt(self, plaintext, sender_privkey, recipient_pubkey):
        """Encrypt using NIP-44 via nak"""
        try:
            result = subprocess.run([
                "nak", "encrypt",
                "--sec", sender_privkey,
                "--recipient-pubkey", recipient_pubkey,
                plaintext
            ], capture_output=True, text=True, timeout=10)
            
            if result.returncode != 0:
                raise RuntimeError(f"NIP-44 encryption failed: {result.stderr}")
            
            return result.stdout.strip()
        except Exception as e:
            raise RuntimeError(f"NIP-44 encryption error: {e}")

    def nip44_decrypt(self, ciphertext_b64, receiver_privkey, sender_pubkey):
        """Decrypt using NIP-44 via nak"""
        try:
            result = subprocess.run([
                "nak", "decrypt", 
                "--sec", receiver_privkey,
                "--sender-pubkey", sender_pubkey,
                ciphertext_b64
            ], capture_output=True, text=True, timeout=10)
            
            if result.returncode != 0:
                raise RuntimeError(f"NIP-44 decryption failed: {result.stderr}")
            
            return result.stdout.strip()
        except Exception as e:
            raise RuntimeError(f"NIP-44 decryption error: {e}")

    def privkey_to_pubkey(self, privkey_hex):
        """Convert private key to public key using nak"""
        try:
            result = subprocess.run([
                "nak", "key", "public", privkey_hex
            ], capture_output=True, text=True, timeout=10)
            
            if result.returncode != 0:
                raise RuntimeError(f"Key conversion failed: {result.stderr}")
            
            return result.stdout.strip()
        except Exception as e:
            raise RuntimeError(f"Key conversion error: {e}")

    def calculate_event_id(self, event):
        """Calculate event ID according to NIP-01"""
        serialized = json.dumps([
            0,  # Reserved
            event["pubkey"], 
            event["created_at"],
            event["kind"],
            event["tags"],
            event["content"]
        ], separators=(',', ':'), ensure_ascii=False)
        
        return hashlib.sha256(serialized.encode('utf-8')).hexdigest()

    def sign_event(self, event, privkey_hex):
        """Sign event using nak"""
        try:
            event_json = json.dumps(event, separators=(',', ':'), ensure_ascii=False)
            result = subprocess.run([
                "nak", "event", "--sec", privkey_hex
            ], input=event_json, text=True, capture_output=True, timeout=10)
            
            if result.returncode != 0:
                raise RuntimeError(f"Event signing failed: {result.stderr}")
            
            return json.loads(result.stdout.strip())
        except Exception as e:
            raise RuntimeError(f"Event signing error: {e}")

    def create_public_capsule(self, message, target_round, author_privkey, chain_hash):
        """Create public time capsule (kind 1041) per NIP-XX specification"""
        import base64
        
        print(f"📝 Creating public capsule...")
        print(f"   Message: {message}")
        print(f"   Target round: {target_round}")
        print(f"   Chain: {chain_hash[:8]}...")
        
        # Step 1: Encrypt with tlock (real drand network)
        tlock_blob = self.tlock_encrypt(message, target_round, chain_hash)
        print(f"   Encrypted blob size: {len(tlock_blob)} bytes")
        
        # Step 2: Create event per NIP-XX specification
        event = {
            "kind": 1041,
            "content": base64.b64encode(tlock_blob).decode('utf-8'),  # Direct age v1 binary
            "tags": [
                ["tlock", chain_hash, str(target_round)],  # 3-element format
                ["alt", "NIP-XX public time capsule"] 
            ],
            "created_at": int(time.time()),
            "pubkey": self.privkey_to_pubkey(author_privkey)
        }
        
        # Step 3: Sign event
        signed_event = self.sign_event(event, author_privkey)
        print(f"   Event ID: {signed_event['id']}")
        
        return signed_event

    def create_private_capsule(self, message, target_round, author_privkey, recipient_pubkey, chain_hash):
        """Create private time capsule using NIP-59 gift wrapping per NIP-XX specification"""
        import base64
        
        print(f"🔒 Creating private capsule...")
        print(f"   Message: {message}")
        print(f"   Target round: {target_round}")
        print(f"   Recipient: {recipient_pubkey[:16]}...")
        
        # Step 1: Create rumor (unsigned kind 1041)
        tlock_blob = self.tlock_encrypt(message, target_round, chain_hash)
        rumor = {
            "kind": 1041,
            "content": base64.b64encode(tlock_blob).decode('utf-8'),
            "tags": [
                ["tlock", chain_hash, str(target_round)],
                ["alt", "NIP-XX private time capsule rumor"]
            ],
            "created_at": int(time.time()),
            "pubkey": self.privkey_to_pubkey(author_privkey)
        }
        rumor["id"] = self.calculate_event_id(rumor)
        
        # Step 2: Create seal (kind 13) - encrypt rumor with NIP-44
        rumor_json = json.dumps(rumor, separators=(',', ':'), ensure_ascii=False)
        seal_content = self.nip44_encrypt(rumor_json, author_privkey, recipient_pubkey)
        
        seal = {
            "kind": 13,
            "content": seal_content,
            "tags": [],  # Must be empty per NIP-59
            "created_at": int(time.time()),
            "pubkey": self.privkey_to_pubkey(author_privkey)
        }
        signed_seal = self.sign_event(seal, author_privkey)
        
        # Step 3: Create gift wrap (kind 1059) - encrypt seal with ephemeral key
        ephemeral_privkey = secrets.token_hex(32)
        ephemeral_pubkey = self.privkey_to_pubkey(ephemeral_privkey)
        
        seal_json = json.dumps(signed_seal, separators=(',', ':'), ensure_ascii=False)
        gift_wrap_content = self.nip44_encrypt(seal_json, ephemeral_privkey, recipient_pubkey)
        
        gift_wrap = {
            "kind": 1059,
            "content": gift_wrap_content,
            "tags": [["p", recipient_pubkey]],  # Routing tag
            "created_at": int(time.time()),
            "pubkey": ephemeral_pubkey
        }
        
        signed_gift_wrap = self.sign_event(gift_wrap, ephemeral_privkey)
        print(f"   Gift wrap ID: {signed_gift_wrap['id']}")
        
        return signed_gift_wrap, author_privkey  # Return author key for later decryption

    def post_to_relay(self, event):
        """Post event to relay and return success status"""
        try:
            ws = websocket.create_connection(self.relay_url, timeout=10)
            req = json.dumps(["EVENT", event])
            ws.send(req)
            
            response = ws.recv()
            result = json.loads(response)
            ws.close()
            
            if result[0] == "OK" and result[2]:
                print(f"   ✅ Posted to relay: {event['id'][:16]}...")
                return True
            else:
                error_msg = result[3] if len(result) > 3 else "Unknown error"
                print(f"   ❌ Rejected by relay: {error_msg}")
                return False
                
        except Exception as e:
            print(f"   ❌ Relay error: {e}")
            return False

    def wait_for_unlock(self, target_round, chain_hash, api_url, network_name):
        """Wait for drand round to be reached"""
        print(f"⏳ Waiting for {network_name} round {target_round}...")
        
        start_time = time.time()
        while True:
            try:
                current_round = self.get_current_round(chain_hash, api_url)
                
                if current_round >= target_round:
                    elapsed = time.time() - start_time
                    print(f"   ✅ Round {target_round} reached! (waited {elapsed:.1f}s)")
                    return True
                
                remaining = target_round - current_round
                print(f"   ⏰ Round {current_round}/{target_round} ({remaining} rounds remaining)")
                
                # Wait for next round (with some buffer)
                time.sleep(5)
                
            except Exception as e:
                print(f"   ⚠️ Error checking round: {e}")
                time.sleep(10)

    def decrypt_public_capsule(self, event, chain_hash, api_url):
        """Decrypt public time capsule"""
        import base64
        
        print(f"🔓 Decrypting public capsule {event['id'][:16]}...")
        
        # Extract and decode tlock blob
        tlock_blob = base64.b64decode(event["content"])
        
        # Decrypt using real drand network
        decrypted = self.tlock_decrypt(tlock_blob, chain_hash)
        
        print(f"   ✅ Decrypted: {decrypted}")
        return decrypted

    def decrypt_private_capsule(self, gift_wrap_event, author_privkey, recipient_privkey):
        """Decrypt private time capsule from NIP-59 gift wrap"""
        print(f"🔓 Decrypting private capsule {gift_wrap_event['id'][:16]}...")
        
        # Step 1: Get ephemeral pubkey and decrypt gift wrap
        ephemeral_pubkey = gift_wrap_event["pubkey"]
        seal_json = self.nip44_decrypt(gift_wrap_event["content"], recipient_privkey, ephemeral_pubkey)
        seal_event = json.loads(seal_json)
        
        # Step 2: Decrypt seal to get rumor  
        author_pubkey = self.privkey_to_pubkey(author_privkey)
        rumor_json = self.nip44_decrypt(seal_event["content"], recipient_privkey, author_pubkey)
        rumor_event = json.loads(rumor_json)
        
        # Step 3: Extract tlock parameters and decrypt
        tlock_tag = None
        for tag in rumor_event["tags"]:
            if tag[0] == "tlock":
                tlock_tag = tag
                break
        
        if not tlock_tag or len(tlock_tag) != 3:
            raise ValueError("Invalid tlock tag in rumor")
        
        chain_hash = tlock_tag[1]
        
        # Find API for this chain
        api_url = None
        for network in self.drand_networks.values():
            if network["hash"] == chain_hash:
                api_url = network["api"] 
                break
        
        if not api_url:
            raise ValueError(f"Unknown drand chain: {chain_hash}")
        
        # Decrypt the time capsule content
        import base64
        tlock_blob = base64.b64decode(rumor_event["content"])
        decrypted = self.tlock_decrypt(tlock_blob, chain_hash)
        
        print(f"   ✅ Decrypted: {decrypted}")
        return decrypted

    def run_validation(self):
        """Run complete NIP-XX validation with real drand and encryption"""
        print("🕐 NIP-XX Time Capsules Validation")
        print("=" * 50)
        print("Strict specification compliance with real drand networks")
        print()
        
        # Check dependencies
        try:
            self.check_dependencies()
        except RuntimeError as e:
            print(f"❌ {e}")
            return False
        
        # Generate keys
        print("🔑 Generating keys...")
        author_privkey = secrets.token_hex(32)
        recipient_privkey = secrets.token_hex(32)
        author_pubkey = self.privkey_to_pubkey(author_privkey)
        recipient_pubkey = self.privkey_to_pubkey(recipient_privkey)
        
        print(f"   Author: {author_pubkey}")
        print(f"   Recipient: {recipient_pubkey}")
        print()
        
        # Use quicknet for testing
        network = self.drand_networks["quicknet"]
        chain_hash = network["hash"]
        api_url = network["api"]
        network_name = network["name"]
        
        print(f"🌐 Using {network_name} drand network")
        print(f"   Chain: {chain_hash}")
        print(f"   API: {api_url}")
        print()
        
        # Calculate target rounds (short delay for demo)
        current_time = int(time.time())
        unlock_time = current_time + 30  # 30 second delay
        
        try:
            target_round = self.calculate_target_round(unlock_time, chain_hash, api_url)
            current_round = self.get_current_round(chain_hash, api_url)
            
            print(f"⏰ Timing information:")
            print(f"   Current round: {current_round}")
            print(f"   Target round: {target_round}")
            print(f"   Unlock time: {datetime.fromtimestamp(unlock_time)}")
            print()
            
        except Exception as e:
            print(f"❌ Failed to setup timing: {e}")
            return False
        
        # Create capsules
        created_events = []
        
        # 1. Create public capsule
        try:
            public_event = self.create_public_capsule(
                "Hello from public time capsule! 🕐",
                target_round, 
                author_privkey,
                chain_hash
            )
            
            if self.post_to_relay(public_event):
                created_events.append({
                    "type": "public", 
                    "event": public_event,
                    "chain_hash": chain_hash,
                    "api_url": api_url
                })
            print()
            
        except Exception as e:
            print(f"❌ Public capsule creation failed: {e}")
            return False
        
        # 2. Create private capsule  
        try:
            private_event, author_key = self.create_private_capsule(
                "Secret message in private capsule! 🔒",
                target_round,
                author_privkey,
                recipient_pubkey, 
                chain_hash
            )
            
            if self.post_to_relay(private_event):
                created_events.append({
                    "type": "private",
                    "event": private_event, 
                    "author_privkey": author_key,
                    "recipient_privkey": recipient_privkey
                })
            print()
            
        except Exception as e:
            print(f"❌ Private capsule creation failed: {e}")
            return False
        
        if not created_events:
            print("❌ No events were successfully created")
            return False
        
        print(f"📊 Created {len(created_events)} time capsules")
        print()
        
        # Wait for timelock expiration
        self.wait_for_unlock(target_round, chain_hash, api_url, network_name)
        print()
        
        # Decrypt capsules
        print("🔓 Decrypting time capsules...")
        success_count = 0
        
        for item in created_events:
            try:
                if item["type"] == "public":
                    decrypted = self.decrypt_public_capsule(
                        item["event"],
                        item["chain_hash"], 
                        item["api_url"]
                    )
                else:  # private
                    decrypted = self.decrypt_private_capsule(
                        item["event"],
                        item["author_privkey"],
                        item["recipient_privkey"]
                    )
                
                success_count += 1
                
            except Exception as e:
                print(f"   ❌ Decryption failed: {e}")
        
        print()
        
        # Final results
        print("=" * 50)
        print("🎉 NIP-XX Validation Results:")
        print(f"   ✅ Created: {len(created_events)}/2 capsules")
        print(f"   ✅ Decrypted: {success_count}/{len(created_events)} capsules")
        print(f"   ✅ Real drand network: {network_name}")
        print(f"   ✅ Real tlock encryption: age v1 + drand")
        print(f"   ✅ Real NIP-44 encryption: for private capsules")
        
        success = success_count == len(created_events) == 2
        
        if success:
            print("🏆 ALL VALIDATIONS PASSED!")
            print("   NIP-XX specification fully validated with real encryption")
        else:
            print("❌ Some validations failed")
        
        return success


def main():
    """Main entry point"""
    validator = NIPXXValidator()
    
    try:
        success = validator.run_validation()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        print("\n⚠️ Validation interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n💥 Unexpected error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
