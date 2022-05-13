package eth_http

import (
	"math/big"
	"open_custodial/pkg/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
)

type SignTxForm struct {
	Nonce    uint64         `json:"nonce"`
	To       common.Address `json:"to"`
	Amount   *big.Int       `json:"amount"`
	GasLimit uint64         `json:"gasLimit"`
	GasPrice *big.Int       `json:"gasPrice"`
	Data     []byte         `json:"data"`
	ChainID  *big.Int       `json:"chaindID"`
	Label    string         `json:"label"`
}

func newSignTxForm(c *gin.Context) (f SignTxForm, err error) {
	err = c.BindJSON(&f)
	return f, err
}

type SignTxResp struct {
	SerializedTransaction []byte `json:"serializedTransaction"`
}

func NewSignTxResp(tx *types.Transaction) (f SignTxResp, err error) {
	b, err := eth.RawTransaction(tx)
	if err != nil {
		return f, err
	}
	f.SerializedTransaction = b
	return f, nil
}
