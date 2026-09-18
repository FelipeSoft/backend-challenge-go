CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    correlation_id UUID NOT NULL,
    causation_id UUID NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    aggregate_type VARCHAR(100) NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    next_delivery_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE NULL,
    payload JSONB NOT NULL,
    locked_until TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_outbox_pending ON outbox(next_delivery_at) WHERE published_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed ON outbox(created_at) WHERE published_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_outbox_correlation ON outbox(correlation_id);