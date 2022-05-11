package hsm

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/miekg/pkcs11"
)

func (h *hsm) SignECDSA_secp256k1(msg []byte, sess pkcs11.SessionHandle, privKey pkcs11.ObjectHandle) ([]byte, error) {

	curve := secp256k1.S256()
	halfN := new(big.Int).Div(curve.N, big.NewInt(2))

	// arbitrary limit
	// TODO - flip S value instead of retrying till it works
	for i := 0; i < 20; i++ {

		signature, err := h.Sign(sess, privKey, msg)
		if err != nil {
			return nil, err
		}

		// s value is the last 32 bytes of an ECDSA signature
		// the first 32 bytes represent r value
		s := new(big.Int).SetBytes(signature[32:64])

		// check if s value is less than half of N
		if s.Cmp(halfN) == -1 {
			return signature, nil
		}
	}

	return nil, errors.New("unable to calculate signature within 20 attempts")
}

func (h *hsm) Sign(sess pkcs11.SessionHandle, privKey pkcs11.ObjectHandle, message []byte) ([]byte, error) {

	mech := []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_ECDSA, nil)}

	if err := h.ctx.SignInit(sess, mech, privKey); err != nil {
		return nil, err
	}

	defer h.ctx.SignFinal(sess)

	b, err := h.ctx.Sign(sess, message)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func VerifySignature(message, signature, expectedPublicKey []byte) ([]byte, error) {
	// attempt recovery ID of 0
	sig := append(signature, 0)

	err := recoverPublicKey(expectedPublicKey, message, sig)
	// if there is no error, recovery was successful.
	if err == nil {
		return sig, nil
	}

	// if recovery ID 0 did not work, try 1
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
