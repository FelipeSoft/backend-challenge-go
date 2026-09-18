CREATE TABLE wallet_ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id),
    transaction_id UUID NOT NULL REFERENCES wager_transactions(id),
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('DEBIT', 'CREDIT')),
    amount_value BIGINT NOT NULL CHECK (amount_value > 0),
    amount_currency VARCHAR(3) NOT NULL,
    balance_before_value BIGINT NOT NULL CHECK (balance_before_value >= 0),
    balance_before_currency VARCHAR(3) NOT NULL,
    balance_after_value BIGINT NOT NULL CHECK (balance_after_value >= 0),
    balance_after_currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_wallet_transaction_ledger UNIQUE (wallet_id, transaction_id)
);
CREATE INDEX IF NOT EXISTS idx_wallet_ledger_wallet_id ON wallet_ledger_entries(wallet_id);