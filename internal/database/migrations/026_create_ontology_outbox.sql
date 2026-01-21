-- +goose Up
CREATE TABLE IF NOT EXISTS ontology_outbox (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ontology_outbox_created_at ON ontology_outbox (created_at);

-- +goose Down
DROP TABLE IF EXISTS ontology_outbox;
