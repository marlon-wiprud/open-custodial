package validator_svc

type ValidatorService interface {
	ValidateSign() error
	ValidateCreateAddress() error
}

type service struct {
}

func NewValidatorService() ValidatorService {
	return &service{}
}

// TODO - find and invoke validator webhook
func (s *service) ValidateSign() error {
	return nil
}

// TODO - find and invoke validator webhook
func (s *service) ValidateCreateAddress() error {
	return nil
}
