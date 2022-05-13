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

func (s *service) ValidateSign() error {
	return nil
}

func (s *service) ValidateCreateAddress() error {
	// return errors.New("not allowed to create address")
	return nil
}
