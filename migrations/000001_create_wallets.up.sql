CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL,
    currency VARCHAR(3) NOT NULL,
    balance_amount BIGINT NOT NULL CHECK (balance_amount >= 0),
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_wallet_player_currency UNIQUE (player_id, currency)
);
CREATE INDEX idx_wallets_player_id ON wallets(player_id);