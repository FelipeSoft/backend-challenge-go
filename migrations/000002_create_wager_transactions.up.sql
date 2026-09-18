CREATE TABLE wager_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id VARCHAR(100) NULL,
    external_transaction_id VARCHAR(255) NULL,
    idempotency_key VARCHAR(255) NULL,
    payload_hash VARCHAR(64) NOT NULL,
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    player_id UUID NOT NULL,
    round_id VARCHAR(255) NULL,
    game_id VARCHAR(255) NULL,
    kind VARCHAR(50) NOT NULL,
    amount_value BIGINT NOT NULL CHECK (amount_value >= 0),
    amount_currency VARCHAR(3) NOT NULL,
    reference_external_transaction_id VARCHAR(255) NULL,
    status VARCHAR(50) NOT NULL,
    failure_code VARCHAR(100) NULL,
    result_amount_value BIGINT NULL,
    result_amount_currency VARCHAR(3) NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_provider_external_tx UNIQUE (provider_id, external_transaction_id),
    CONSTRAINT uk_idempotency_key UNIQUE (idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_wager_transactions_wallet_id ON wager_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wager_transactions_status ON wager_transactions(status);
CREATE INDEX IF NOT EXISTS idx_wager_transactions_ref ON wager_transactions(provider_id, reference_external_transaction_id);
CREATE INDEX IF NOT EXISTS idx_wager_transactions_pending_ref ON wager_transactions(status, created_at) WHERE status = 'PENDING_REFERENCE';