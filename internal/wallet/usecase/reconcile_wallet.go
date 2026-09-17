package usecase

type ReconcileWallet struct {
}

type ReconcileWalletInput struct {
	WalletID string
}

type ReconcileWalletOutput struct {
	WalletID                  string
	StoredBalanceAmount       string
	StoredBalanceCurrency     string
	CalculatedBalanceAmount   string
	CalculatedBalanceCurrency string
	DifferenceAmount          string
	DifferenceCurrency        string
	Consistent                bool
	CheckedEntries            int
}

func NewReconcileWallet() *ReconcileWallet {
	return &ReconcileWallet{}
}

func (uc *ReconcileWallet) Execute(input ReconcileWalletInput) (ReconcileWalletOutput, error) {
	return ReconcileWalletOutput{}, nil
}
