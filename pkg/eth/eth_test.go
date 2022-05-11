package eth

import (
	"math/big"
	"open_custodial/pkg/config"
	"open_custodial/pkg/hsm"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

// func (s *ETHSuite) TestGetAddress() {
// 	addr, err := GetAddress(s.hsm, "tester")
// 	s.NoError(err)
// 	fmt.Println("got address:", addr)
// }

// func (s *ETHSuite) TestCreateAddress() {
// 	addr, err := CreateAddress(s.hsm, "tester")
// 	s.NoError(err)
// 	fmt.Println("got address:", addr)
// }

func (s *ETHSuite) TestSignTransaction() {
	tx := types.NewTransaction(1, common.HexToAddress("0xE8B5fBaE723E5A4AAc991dDC54c549c1BaEEAb5e"), big.NewInt(1000), 100, big.NewInt(100), nil)
	_, err := SignTransaction(s.hsm, tx, "tester", big.NewInt(3))
	s.NoError(err)
}
