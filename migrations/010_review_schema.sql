CREATE TABLE IF NOT EXISTS templates (
    payload jsonb NOT NULL,
    idempotency_key text NOT NULL,
    revision bigint NOT NULL CHECK (revision > 0),
    status text NOT NULL CHECK (status IN ('draft','reviewing','published','deprecated')),
    version integer NOT NULL CHECK (version > 0),
    id text NOT NULL,
    CONSTRAINT templates_identity PRIMARY KEY (id, version),
    CONSTRAINT templates_request_once UNIQUE (idempotency_key)
);

CREATE TABLE IF NOT EXISTS batches (
    payload jsonb NOT NULL,
    template_version integer,
    template_id text,
    idempotency_key text NOT NULL,
    revision bigint NOT NULL CHECK (revision > 0),
    status text NOT NULL CHECK (status IN ('draft','started','submitted','returned','re_review','completed','voided')),
    id text NOT NULL PRIMARY KEY,
    CONSTRAINT batches_request_once UNIQUE (idempotency_key),
    CONSTRAINT batches_template_snapshot FOREIGN KEY (template_id, template_version)
        REFERENCES templates (id, version) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS reviews (
    payload jsonb NOT NULL,
    reviewer_id text NOT NULL,
    material_id text NOT NULL,
    batch_id text NOT NULL,
    id text NOT NULL PRIMARY KEY,
    CONSTRAINT reviews_one_seat UNIQUE (batch_id, material_id, reviewer_id),
    CONSTRAINT reviews_batch FOREIGN KEY (batch_id) REFERENCES batches (id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS audit_events (
    created_at timestamptz NOT NULL,
    payload jsonb NOT NULL,
    hash text NOT NULL,
    previous_hash text NOT NULL,
    aggregate_id text NOT NULL,
    id text NOT NULL PRIMARY KEY,
    CONSTRAINT audit_hash_once UNIQUE (hash)
);

CREATE INDEX IF NOT EXISTS batch_status_idx ON batches (status, template_id, template_version);
CREATE INDEX IF NOT EXISTS review_panel_idx ON reviews (batch_id, material_id, reviewer_id);
CREATE INDEX IF NOT EXISTS audit_chain_idx ON audit_events (aggregate_id, created_at DESC);
