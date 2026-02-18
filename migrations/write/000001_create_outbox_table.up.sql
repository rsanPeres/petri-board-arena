CREATE TABLE IF NOT EXISTS outbox_event (
    id                UUID PRIMARY KEY,
    -- Identidade/roteamento do evento
    aggregate_type    TEXT        NOT NULL,
    aggregate_id      TEXT        NOT NULL,
    event_type        TEXT        NOT NULL,
    topic             TEXT        NOT NULL,

    -- Dados do evento
    payload           JSONB       NOT NULL,
    headers           JSONB       NOT NULL DEFAULT '{}'::jsonb,

    -- Ordenação / idempotência / rastreio
    correlation_id    TEXT        NULL,
    causation_id      TEXT        NULL,
    idempotency_key   TEXT        NULL,

    -- Estado de publicação
    status            TEXT        NOT NULL DEFAULT 'PENDING', -- PENDING | PROCESSING | PUBLISHED | FAILED
    attempts          INT         NOT NULL DEFAULT 0,
    max_attempts      INT         NOT NULL DEFAULT 10,

    next_attempt_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at      TIMESTAMPTZ NULL,

    -- Lock para múltiplos workers do outbox
    locked_by         TEXT        NULL,
    locked_at         TIMESTAMPTZ NULL,
    lock_expires_at   TIMESTAMPTZ NULL,

    -- Auditoria
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Garantia de valores válidos no status
ALTER TABLE outbox_event
    ADD CONSTRAINT outbox_event_status_chk
    CHECK (status IN ('PENDING', 'PROCESSING', 'PUBLISHED', 'FAILED'));

-- Índices principais para polling eficiente
CREATE INDEX IF NOT EXISTS idx_outbox_event_poll
    ON outbox_event (status, next_attempt_at, created_at);

-- Índice útil para troubleshooting/replay por aggregate
CREATE INDEX IF NOT EXISTS idx_outbox_event_aggregate
    ON outbox_event (aggregate_type, aggregate_id, created_at);

-- Idempotência: evita duplicar o mesmo comando/evento
CREATE UNIQUE INDEX IF NOT EXISTS ux_outbox_event_idempotency_key
    ON outbox_event (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_event_published_at
    ON outbox_event (published_at)
    WHERE published_at IS NOT NULL;

-- Trigger para manter updated_at automático
CREATE OR REPLACE FUNCTION set_updated_at_outbox_event()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_set_updated_at_outbox_event ON outbox_event;

CREATE TRIGGER trg_set_updated_at_outbox_event
BEFORE UPDATE ON outbox_event
FOR EACH ROW
EXECUTE FUNCTION set_updated_at_outbox_event();


-- dead-letter do outbox (quando excede max_attempts)
CREATE TABLE IF NOT EXISTS outbox_dead_letters (
    id                UUID PRIMARY KEY,
    outbox_event_id    UUID        NULL,
    aggregate_type     TEXT        NOT NULL,
    aggregate_id       TEXT        NOT NULL,
    event_type         TEXT        NOT NULL,
    topic              TEXT        NOT NULL,

    payload            JSONB       NOT NULL,
    headers            JSONB       NOT NULL DEFAULT '{}'::jsonb,

    correlation_id     TEXT        NULL,
    causation_id       TEXT        NULL,
    idempotency_key    TEXT        NULL,

    attempts           INT         NOT NULL,
    last_error         TEXT        NOT NULL,

    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_outbox_dlq_created_at
    ON outbox_dead_letters (created_at);
