package eth_hsm

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

func (s *ETHSuite) TestCreateAddress() {
	addr, err := CreateAddress(s.hsm, "test_create_address")
	s.NoError(err)

	found, err := GetAddress(s.hsm, "test_create_address")
	s.NoError(err)

	s.Equal(addr.Hex(), found.Hex())
}

func (s *ETHSuite) TestSignTransaction() {
	_, err := CreateAddress(s.hsm, "test_sign_transaction")
	s.NoError(err)

	tx := types.NewTransaction(1, common.HexToAddress("0xE8B5fBaE723E5A4AAc991dDC54c549c1BaEEAb5e"), big.NewInt(1000), 100, big.NewInt(100), nil)
	_, err = SignTransaction(s.hsm, tx, "test_sign_transaction", big.NewInt(3))
	s.NoError(err)
}
