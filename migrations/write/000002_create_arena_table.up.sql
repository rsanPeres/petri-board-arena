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

ALTER TABLE arena
  ADD CONSTRAINT arena_status_chk
  CHECK (status IN ('PENDING', 'RUNNING', 'PAUSED', 'FINISHED'));

CREATE INDEX IF NOT EXISTS arena_status_idx ON arena (status);
CREATE INDEX IF NOT EXISTS arena_created_at_idx ON arena (created_at DESC);
