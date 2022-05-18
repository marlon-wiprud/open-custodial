package hsm

import (
	"errors"
	"open_custodial/pkg/config"
	"testing"

	"github.com/miekg/pkcs11"
	"github.com/stretchr/testify/suite"
)

type HSMSuite struct {
	suite.Suite
	hsm *hsm
}

func TestHSMSuite(t *testing.T) {
	suite.Run(t, new(HSMSuite))
}

func (s *HSMSuite) SetupSuite() {
	c := config.NewConfig()
	p := pkcs11.New(c.HSMLibPath)
	if p == nil {
		s.NoError(errors.New("unable to init pkcs11"))
	}

	err := p.Initialize()
	s.NoError(err)

	s.hsm = &hsm{p, c.CU_USERNAME, c.CU_PASSWORD, c.SO_PASSWORD, nil}
}

func (s *HSMSuite) TestNewSlot() {
	_, err := s.hsm.NewSlot("test_new_slot")
	s.NoError(err)
}
