-- +goose Up
CREATE TABLE webhook_events (
    id BIGSERIAL PRIMARY KEY,
    provider VARCHAR(30) NOT NULL DEFAULT 'razorpay',
    event_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK ( status IN (
            'pending',
            'processing',
            'processed',
            'failed'
        )),
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,

    UNIQUE (provider, event_id)
);

-- +goose Down
DROP TABLE webhook_events;