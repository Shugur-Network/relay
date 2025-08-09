-- Shugur Relay Database Schema
-- CockroachDB optimized schema for Nostr relay
-- Database: Defined in constants.DatabaseName

-- Events table - stores all Nostr events with optimized indexes
CREATE TABLE IF NOT EXISTS events (
  id CHAR(64) NOT NULL,
  pubkey CHAR(64) NOT NULL,
  created_at INT8 NOT NULL,
  kind INT8 NOT NULL,
  tags JSONB NULL,
  content STRING NULL,
  sig CHAR(128) NOT NULL,
  CONSTRAINT events_pkey PRIMARY KEY (id ASC),
  INDEX idx_kind (kind ASC),
  INDEX idx_pubkey (pubkey ASC),
  INVERTED INDEX idx_tags (tags),
  INDEX events_pubkey_kind_storing_rec_idx (pubkey ASC, kind ASC) STORING (created_at, tags, content, sig),
  INDEX events_created_at_desc_storing_rec_idx (created_at DESC) STORING (pubkey, kind, tags, content, sig),
  INDEX events_created_at_storing_rec_idx (created_at ASC) STORING (pubkey, kind, tags, content, sig),
  INDEX events_kind_created_at_storing_rec_idx (kind ASC, created_at ASC) STORING (pubkey, tags, content, sig),
  INDEX events_pubkey_created_at_storing_rec_idx (pubkey ASC, created_at ASC) STORING (kind, tags, content, sig),
  INDEX events_pubkey_kind_created_at_storing_rec_idx (pubkey ASC, kind ASC, created_at ASC) STORING (tags, content, sig),
  INDEX idx_events_pubkey_kind (pubkey ASC, kind ASC),
  UNIQUE INDEX uq_replaceable (pubkey ASC, kind ASC) WHERE kind IN (0:::INT8, 3:::INT8, 41:::INT8),
  UNIQUE INDEX uq_addressable (pubkey ASC, kind ASC, (tags->>1:::INT8) ASC) WHERE kind BETWEEN 30000:::INT8 AND 40000:::INT8,
  CONSTRAINT valid_id CHECK (id ~ '^[a-f0-9]{64}$':::STRING),
  CONSTRAINT valid_pubkey CHECK (pubkey ~ '^[a-f0-9]{64}$':::STRING),
  CONSTRAINT valid_sig CHECK (sig ~ '^[a-f0-9]{128}$':::STRING),
  CONSTRAINT kind_range CHECK ((kind >= 0:::INT8) AND (kind <= 65535:::INT8))
);
