#!/usr/bin/env python3
import sys
import hashlib

def compute_merkle_tree_commitment(pubkeys):
    if not pubkeys:
        return ""
    
    domain = b"nostr:capsule:witness/v1"
    leaves = []
    
    for i, pubkey in enumerate(pubkeys):
        # 1-based index as little-endian 4 bytes
        index = i + 1
        index_bytes = index.to_bytes(4, 'little')
        
        # Convert hex pubkey to bytes
        try:
            pubkey_bytes = bytes.fromhex(pubkey)
        except ValueError:
            return ""  # Invalid hex
        
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
                right = left  # Duplicate if odd number
            
            # Combine: SHA256(left || right)
            combined = left + right
            parent_hash = hashlib.sha256(combined).digest()
            next_level.append(parent_hash)
        
        current_level = next_level
    
    return current_level[0].hex() if current_level else ""

if len(sys.argv) > 1:
    pubkeys = sys.argv[1:]
    result = compute_merkle_tree_commitment(pubkeys)
    print(result, end='')
