DROP TRIGGER IF EXISTS trg_enforce_ledger_immutability ON wallet_ledger_entries;
DROP FUNCTION IF EXISTS prevent_ledger_modification();