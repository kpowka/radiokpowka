-- Purpose: Initial schema for Postgres.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  username VARCHAR(64) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS tracks (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  title VARCHAR(512) NOT NULL,
  source_url TEXT NOT NULL,
  duration INTEGER NOT NULL DEFAULT 0,
  added_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
  added_by_nick VARCHAR(128) NULL,
  metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS queue_entries (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  track_id UUID NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
  position INTEGER NOT NULL,
  status VARCHAR(16) NOT NULL, -- prev|current|next
  added_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  is_donation BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_queue_position ON queue_entries(position);
CREATE INDEX IF NOT EXISTS idx_queue_status ON queue_entries(status);

CREATE TABLE IF NOT EXISTS donations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  donor_nick VARCHAR(128) NOT NULL,
  amount BIGINT NOT NULL DEFAULT 0,
  message TEXT NULL,
  track_url TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS integrations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  type VARCHAR(64) NOT NULL,
  config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  connected_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_integrations_type ON integrations(type);

CREATE TABLE IF NOT EXISTS playlists (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  url TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
