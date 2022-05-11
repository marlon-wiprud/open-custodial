package eth

import (
	"fmt"
	"math/big"
	"open_custodial/pkg/hsm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func GetAddress(h hsm.HSM, name string) (addr common.Address, err error) {

	sess, err := h.NewSlotSession(name)
	if err != nil {
		fmt.Println("get_addr: failed to open slot session", err)
		return addr, err
	}

	defer h.EndSession(sess)

	pubHandle, err := h.PublicKeyHandle(*sess)
	if err != nil {
		fmt.Println("get_addr: failed to get public key handle", err)
		return addr, err
	}

	pub, err := h.GetPublicKey(*sess, pubHandle)
	if err != nil {
		return addr, err
	}

	return crypto.PubkeyToAddress(pub), nil
}

func CreateAddress(h hsm.HSM, name string) (addr common.Address, err error) {
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

func SignTransaction(h hsm.HSM, tx *types.Transaction, label string, chainID *big.Int) (*types.Transaction, error) {
	sess, err := h.NewSlotSession(label)
	if err != nil {
		return nil, err
	}

	defer h.EndSession(sess)

	pubHandle, err := h.PublicKeyHandle(*sess)
	if err != nil {
		return nil, err
	}

	pubKey, err := h.GetPublicKey(*sess, pubHandle)
	if err != nil {
		return nil, err
	}

	privHandle, err := h.PrivateKeyHandle(*sess)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := crypto.FromECDSAPub(&pubKey)
	signer := types.NewLondonSigner(chainID)
	message := signer.Hash(tx).Bytes()

	signature, err := h.SignECDSA_secp256k1(message, *sess, privHandle)
	if err != nil {
		return nil, err
	}

	verifiedSig, err := hsm.VerifySignature(message, signature, pubKeyBytes)
	if err != nil {
		return nil, err
	}

	return tx.WithSignature(signer, verifiedSig)
}
