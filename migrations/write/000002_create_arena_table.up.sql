-- 000001_create_arena_table.up.sql

CREATE TABLE IF NOT EXISTS arena (
  id           UUID PRIMARY KEY,
  name         TEXT NOT NULL,
  status       TEXT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL,
  started_at   TIMESTAMPTZ NULL,
  finished_at  TIMESTAMPTZ NULL,

  tick         BIGINT NOT NULL DEFAULT 0,

  config       JSONB NOT NULL,

  -- housekeeping
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE arenas
  ADD CONSTRAINT arenas_status_chk
  CHECK (status IN ('PENDING', 'RUNNING', 'PAUSED', 'FINISHED'));

CREATE INDEX IF NOT EXISTS arenas_status_idx ON arenas (status);
CREATE INDEX IF NOT EXISTS arenas_created_at_idx ON arenas (created_at DESC);
