package eth_hsm

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"open_custodial/pkg/hsm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/miekg/pkcs11"
)

func GetSlotAddress(h hsm.HSM, slotID uint) (addr common.Address, err error) {
	sess, err := h.NewSession(slotID)
	if err != nil {
		return addr, err
	}

	defer h.EndSession(&sess)

	pubHandle, err := h.PublicKeyHandle(sess)
	if err != nil {
		return addr, err
	}

	pub, err := h.GetPublicKey(sess, pubHandle)
	if err != nil {
		return addr, err
	}

	return crypto.PubkeyToAddress(pub), nil
}

func GetAddress(h hsm.HSM, name string) (addr common.Address, err error) {

	sess, err := h.NewSlotSession(name)
	if err != nil {
		return addr, err
	}

	defer h.EndSession(sess)

	pubHandle, err := h.PublicKeyHandle(*sess)
	if err != nil {
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
	signer := types.NewLondonSigner(chainID)
	message := signer.Hash(tx).Bytes()

	sess, err := h.NewSlotSession(label)
	if err != nil {
		return nil, err
	}

	defer h.EndSession(sess)

	pubKeyBytes, privHandle, err := getSigningKeys(h, sess, label)
	if err != nil {
		return nil, err
	}

	signature, err := h.SignECDSA_secp256k1(message, *sess, privHandle)
	if err != nil {
		return nil, err
	}

	verifiedSig, err := VerifySignature(message, signature, pubKeyBytes)
	if err != nil {
		return nil, err
	}

	return tx.WithSignature(signer, verifiedSig)
}

func getSigningKeys(h hsm.HSM, sess *pkcs11.SessionHandle, label string) (b []byte, privKey pkcs11.ObjectHandle, err error) {
	pubHandle, err := h.PublicKeyHandle(*sess)
	if err != nil {
		return b, privKey, err
	}

	if err := h.ReleaseHandle(*sess); err != nil {
		return b, privKey, err
	}

	pubKey, err := h.GetPublicKey(*sess, pubHandle)
	if err != nil {
		return b, privKey, err
	}

	privHandle, err := h.PrivateKeyHandle(*sess)
	if err != nil {
		return b, privKey, err
	}

	if err := h.ReleaseHandle(*sess); err != nil {
		return b, privKey, err
	}

	return crypto.FromECDSAPub(&pubKey), privHandle, nil
}

func RawTransaction(tx *types.Transaction) ([]byte, error) {
	var b bytes.Buffer

	if err := tx.EncodeRLP(&b); err != nil {
		return nil, fmt.Errorf("unable to encode signed transaction %v", err)
	}

	return b.Bytes(), nil
}

func VerifySignature(message, signature, expectedPublicKey []byte) ([]byte, error) {
	sig := append(signature, 0)

	err := recoverPublicKey(expectedPublicKey, message, sig)
	if err == nil {
		return sig, nil
	}

	sig[64] = 1
	return sig, recoverPublicKey(expectedPublicKey, message, sig)
}

func recoverPublicKey(expectedPubKey, msg, sig []byte) error {

	recoveredPubKey, err := crypto.Ecrecover(msg, sig)
	if err != nil {
		return err
	}

	if !bytes.Equal(recoveredPubKey, expectedPubKey) {
		return errors.New("expected public key and recovered public key do not match")
	}

	return nil
}
