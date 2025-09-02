#!/usr/bin/env python3
import sys
import hashlib

def compute_merkle_tree_commitment(pubkeys):
    """
    Compute witness commitment per NIP-XX Time Capsules spec
    Domain: "nostr:capsule:witness/v1"
    Leaf: SHA256(domain || LE32(index) || pubkey_bytes)
    Tree: Binary merkle with SHA256(left || right)
    """
    if not pubkeys:
        return ""
    
    domain = b"nostr:capsule:witness/v1"
    leaves = []
    
    for i, pubkey in enumerate(pubkeys):
        # 1-based index as little-endian 4 bytes
        index = i + 1
        index_bytes = index.to_bytes(4, 'little')
        
        # Convert hex pubkey to bytes - must be exactly 64 hex chars (32 bytes)
        try:
            pubkey_clean = pubkey.lower().strip()
            if len(pubkey_clean) != 64:
                print(f"Error: Pubkey must be 64 hex chars, got {len(pubkey_clean)}", file=sys.stderr)
                return ""
            pubkey_bytes = bytes.fromhex(pubkey_clean)
            if len(pubkey_bytes) != 32:
                print(f"Error: Pubkey must decode to 32 bytes, got {len(pubkey_bytes)}", file=sys.stderr)
                return ""
        except ValueError as e:
            print(f"Error: Invalid hex pubkey: {e}", file=sys.stderr)
            return ""
        
        # Leaf = SHA256(domain || LE32(index) || pubkey_bytes)
        leaf_data = domain + index_bytes + pubkey_bytes
        leaf_hash = hashlib.sha256(leaf_data).digest()
        leaves.append(leaf_hash)
    
    # Build merkle tree
    current_level = leaves
    while len(current_level) > 1:
        next_level = []
        for i in range(0, len(current_level), 2):
            left = current_level[i]
            if i + 1 < len(current_level):
                right = current_level[i + 1]
            else:
                # Duplicate last leaf if odd number (per spec)
                right = left
            
            # Combine: SHA256(left || right)
            combined = left + right
            parent_hash = hashlib.sha256(combined).digest()
            next_level.append(parent_hash)
        
        current_level = next_level
    
    # Return root as lowercase hex (per NIP spec)
    return current_level[0].hex().lower() if current_level else ""

if __name__ == "__main__":
    if len(sys.argv) > 1:
        pubkeys = sys.argv[1:]
        result = compute_merkle_tree_commitment(pubkeys)
        if result:
            print(result, end='')
        else:
            sys.exit(1)
    else:
        print("Usage: compute_commitment.py <pubkey1> [pubkey2] ...", file=sys.stderr)
        sys.exit(1)
