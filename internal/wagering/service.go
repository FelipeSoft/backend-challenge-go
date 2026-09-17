package wagering

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetWagerTransaction(wagerTransactionId string) (any, error) {
	return nil, nil
}
