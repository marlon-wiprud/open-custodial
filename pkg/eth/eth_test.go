package eth

import (
	"fmt"
	"open_custodial/pkg/config"
	"open_custodial/pkg/hsm"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ETHSuite struct {
	suite.Suite
	hsm hsm.HSM
}

func TestETHSuite(t *testing.T) {
	suite.Run(t, new(ETHSuite))
}

func (s *ETHSuite) SetupSuite() {
	c := config.NewConfig()
	h, err := hsm.NewHSM(c.HSMLibPath, c.CU_USERNAME, c.CU_PASSWORD, c.SO_PASSWORD)
	s.NoError(err)
	s.hsm = h
}

func (s *ETHSuite) TestGetAddress() {
	addr, err := GetAddress(s.hsm, "tester")
	s.NoError(err)
	fmt.Println("got address:", addr)
}

func (s *ETHSuite) TestCreateAddress() {
	addr, err := CreateAddress(s.hsm, "tester")
	s.NoError(err)
	fmt.Println("got address:", addr)
}
