package eth

import (
	"open_custodial/pkg/hsm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetAddress(h hsm.HSM, name string) (addr common.Address, err error) {

	sess, err := h.NewSlotSession(name)
	if err != nil {
		return addr, err
	}

	defer h.EndSession(sess)

	pubHandle, done, err := h.PublicKeyHandle(*sess)
	if err != nil {
		return addr, err
	}

	done()

	pub, err := h.GetPublicKey(*sess, pubHandle)
	if err != nil {
		return addr, err
	}

	return crypto.PubkeyToAddress(pub), nil
}

func NewAddress(h hsm.HSM, name string) (addr common.Address, err error) {
	_, err = h.NewSlot(name)
	if err != nil {
		return addr, err
	}

	sess, err := h.NewSlotSession(name)
	if err != nil {
		return addr, err
	}

	defer h.EndSession(sess)

	pubHandle, _, err := h.GenerateKeyECDSA_secp256k1(*sess)
	if err != nil {
		return addr, err
	}

	pub, err := h.GetPublicKey(*sess, pubHandle)
	if err != nil {
		return addr, err
	}

	return crypto.PubkeyToAddress(pub), nil
}
