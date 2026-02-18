DROP TRIGGER IF EXISTS trg_set_updated_at_outbox_event ON outbox_event;
DROP FUNCTION IF EXISTS set_updated_at_outbox_event();

DROP TABLE IF EXISTS outbox_dead_letter;
DROP TABLE IF EXISTS outbox_event;
