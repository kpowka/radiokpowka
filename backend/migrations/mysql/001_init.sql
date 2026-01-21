-- Purpose: Initial schema for MySQL 8+.
-- Notes:
-- - UUID stored as CHAR(36) for portability.
-- - JSON supported in MySQL 5.7+/8+.

CREATE TABLE IF NOT EXISTS users (
  id CHAR(36) PRIMARY KEY,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(32) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tracks (
  id CHAR(36) PRIMARY KEY,
  title VARCHAR(512) NOT NULL,
  source_url TEXT NOT NULL,
  duration INT NOT NULL DEFAULT 0,
  added_by_user_id CHAR(36) NULL,
  added_by_nick VARCHAR(128) NULL,
  metadata_json JSON NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_tracks_user FOREIGN KEY (added_by_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS queue_entries (
  id CHAR(36) PRIMARY KEY,
  track_id CHAR(36) NOT NULL,
  position INT NOT NULL,
  status VARCHAR(16) NOT NULL,
  added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  is_donation BOOLEAN NOT NULL DEFAULT FALSE,
  CONSTRAINT fk_queue_track FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
  INDEX idx_queue_position (position),
  INDEX idx_queue_status (status)
);

CREATE TABLE IF NOT EXISTS donations (
  id CHAR(36) PRIMARY KEY,
  donor_nick VARCHAR(128) NOT NULL,
  amount BIGINT NOT NULL DEFAULT 0,
  message TEXT NULL,
  track_url TEXT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS integrations (
  id CHAR(36) PRIMARY KEY,
  type VARCHAR(64) NOT NULL,
  config_json JSON NOT NULL,
  connected_at TIMESTAMP NULL,
  INDEX idx_integrations_type (type)
);

CREATE TABLE IF NOT EXISTS playlists (
  id CHAR(36) PRIMARY KEY,
  url TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
