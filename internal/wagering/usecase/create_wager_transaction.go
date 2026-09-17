package usecase

type CreateWagerTransaction struct {
}

type CreateWagerTransactionInput struct {
	ProviderID                     string
	ExternalTransactionID          string
	PlayerID                       string
	WalletID                       string
	RoundID                        string
	GameID                         string
	Kind                           string
	MoneyAmount                    string
	MoneyCurrency                  string
	ReferenceExternalTransactionId *string
}

type CreateWagerTransactionOutput struct {
	TransactionID    string
	Status           string
	BalanceAmount    string
	BalanceCurrency  string
	IdempotentReplay bool
}

func NewCreateWagerTransaction() *CreateWagerTransaction {
	return &CreateWagerTransaction{}
}

func (uc *CreateWagerTransaction) Execute(input CreateWagerTransactionInput) (CreateWagerTransactionOutput, error) {
	return CreateWagerTransactionOutput{}, nil
}
