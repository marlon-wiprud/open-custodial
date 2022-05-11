package hsm

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/miekg/pkcs11"
)

type HSM interface {
	NewSlotSession(name string) (*pkcs11.SessionHandle, error)
	EndSession(sess *pkcs11.SessionHandle) error
	PublicKeyHandle(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, func(), error)
	GetPublicKey(session pkcs11.SessionHandle, pubKeyHandle pkcs11.ObjectHandle) (ecdsa.PublicKey, error)
	GenerateKeyECDSA_secp256k1(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, pkcs11.ObjectHandle, error)
	SignECDSA_secp256k1(msg []byte, sess pkcs11.SessionHandle, privKey pkcs11.ObjectHandle) ([]byte, error)
	NewSlot(name string) (uint, error)
	PrivateKeyHandle(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, func(), error)
}

type hsm struct {
	ctx        *pkcs11.Ctx
	cuUsername string
	cuPassword string
	soPassword string
}

func NewHSM(libPath, cuUsername, cuPassword, soPassword string) (HSM, error) {
	p := pkcs11.New(libPath)
	if p == nil {
		return nil, fmt.Errorf("unable to initialize pkcs11 module with path %s", libPath)
	}

	if err := p.Initialize(); err != nil {
		return nil, err
	}

	return &hsm{p, cuUsername, cuPassword, soPassword}, nil
}
