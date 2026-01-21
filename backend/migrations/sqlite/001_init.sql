-- Purpose: Initial schema for SQLite.
-- Notes:
-- - UUID stored as TEXT.
-- - JSON stored as TEXT for portability.

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tracks (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  source_url TEXT NOT NULL,
  duration INTEGER NOT NULL DEFAULT 0,
  added_by_user_id TEXT NULL REFERENCES users(id) ON DELETE SET NULL,
  added_by_nick TEXT NULL,
  metadata_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS queue_entries (
  id TEXT PRIMARY KEY,
  track_id TEXT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
  position INTEGER NOT NULL,
  status TEXT NOT NULL,
  added_at TEXT NOT NULL DEFAULT (datetime('now')),
  is_donation INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_queue_position ON queue_entries(position);
CREATE INDEX IF NOT EXISTS idx_queue_status ON queue_entries(status);

CREATE TABLE IF NOT EXISTS donations (
  id TEXT PRIMARY KEY,
  donor_nick TEXT NOT NULL,
  amount INTEGER NOT NULL DEFAULT 0,
  message TEXT NULL,
  track_url TEXT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS integrations (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  config_json TEXT NOT NULL DEFAULT '{}',
  connected_at TEXT NULL
);

CREATE INDEX IF NOT EXISTS idx_integrations_type ON integrations(type);

CREATE TABLE IF NOT EXISTS playlists (
  id TEXT PRIMARY KEY,
  url TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
