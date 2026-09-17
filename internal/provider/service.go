package provider

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetExternalTransaction(externalTransactionId string) (any, error) {
	return nil, nil 
}
