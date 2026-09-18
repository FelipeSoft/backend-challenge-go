package usecase

import (
	"context"
	"fmt"

	"github.com/FelipeSoft/backend-challenge-go/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReconcileWallet struct {
	db *pgxpool.Pool
}

type ReconcileWalletInput struct {
	WalletID string
}

type ReconcileWalletOutput struct {
	WalletID                  string       `json:"walletId"`
	StoredBalance             domain.Money `json:"storedBalance"`
	CalculatedBalance         domain.Money `json:"calculatedBalance"`
	Difference                domain.Money `json:"difference"`
	Consistent                bool         `json:"consistent"`
	CheckedEntries            int          `json:"checkedEntries"`
}

func NewReconcileWallet(db *pgxpool.Pool) *ReconcileWallet {
	return &ReconcileWallet{
		db: db,
	}
}

func (uc *ReconcileWallet) Execute(ctx context.Context, input ReconcileWalletInput) (ReconcileWalletOutput, error) {
	if input.WalletID == "" {
		return ReconcileWalletOutput{}, fmt.Errorf("walletId cannot be empty")
	}
	var storedBalanceAmount int64
	var currency string
	walletQuery := `SELECT balance_amount, currency FROM wallets WHERE id = $1`
	err := uc.db.QueryRow(ctx, walletQuery, input.WalletID).Scan(&storedBalanceAmount, &currency)
	if err != nil {
		return ReconcileWalletOutput{}, fmt.Errorf("wallet not found: %w", err)
	}
	storedBalance, err := domain.NewMoneyFromInt(storedBalanceAmount, currency)
	if err != nil {
		return ReconcileWalletOutput{}, fmt.Errorf("error creating stored Money: %w", err)
	}
	ledgerQuery := `
		SELECT 
			COALESCE(
				SUM(
					CASE 
						WHEN direction = 'CREDIT' THEN amount_value 
						WHEN direction = 'DEBIT' THEN -amount_value 
						ELSE 0 
					END
				), 0
			) AS calculated_balance,
			COUNT(*) AS checked_entries
		FROM wallet_ledger_entries 
		WHERE wallet_id = $1
	`
	var calculatedBalanceAmount int64
	var checkedEntries int
	err = uc.db.QueryRow(ctx, ledgerQuery, input.WalletID).Scan(&calculatedBalanceAmount, &checkedEntries)
	if err != nil {
		return ReconcileWalletOutput{}, fmt.Errorf("error calculating ledger balance: %w", err)
	}
	calculatedBalance, err := domain.NewMoneyFromInt(calculatedBalanceAmount, currency)
	if err != nil {
		return ReconcileWalletOutput{}, fmt.Errorf("error creating calculated Money: %w", err)
	}
	diffAmount := storedBalanceAmount - calculatedBalanceAmount
	difference, err := domain.NewMoneyFromInt(diffAmount, currency)
	if err != nil {
		return ReconcileWalletOutput{}, fmt.Errorf("error creating difference Money: %w", err)
	}
	consistent := storedBalanceAmount == calculatedBalanceAmount
	if !consistent {
		fmt.Printf("[RECONCILIATION_MISMATCH] WalletID: %s | Stored: %d | Calculated: %d | Diff: %d | CheckedEntries: %d\n",
			input.WalletID, storedBalanceAmount, calculatedBalanceAmount, diffAmount, checkedEntries)
	}
	return ReconcileWalletOutput{
		WalletID:          input.WalletID,
		StoredBalance:     storedBalance,
		CalculatedBalance: calculatedBalance,
		Difference:        difference,
		Consistent:        consistent,
		CheckedEntries:    checkedEntries,
	}, nil
}