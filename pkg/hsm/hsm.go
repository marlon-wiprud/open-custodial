package hsm

import (
	"crypto/ecdsa"
	"fmt"
	"sync"

	"github.com/miekg/pkcs11"
)

type HSM interface {
	NewSlotSession(name string) (*pkcs11.SessionHandle, error)
	EndSession(sess *pkcs11.SessionHandle) error
	GetPublicKey(session pkcs11.SessionHandle, pubKeyHandle pkcs11.ObjectHandle) (ecdsa.PublicKey, error)
	PublicKeyHandle(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, error)
	PrivateKeyHandle(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, error)
	GenerateKeyECDSA_secp256k1(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, pkcs11.ObjectHandle, error)
	SignECDSA_secp256k1(msg []byte, sess pkcs11.SessionHandle, privKey pkcs11.ObjectHandle) ([]byte, error)
	NewSlot(name string) (uint, error)
	NewSession(slotID uint) (pkcs11.SessionHandle, error)
	GetSlotID(label string) (uint, error)
}

type hsm struct {
	ctx        *pkcs11.Ctx
	cuUsername string
	cuPassword string
	soPassword string
	slotIndex  *slotIndex
}

func NewHSM(libPath, cuUsername, cuPassword, soPassword string) (HSM, error) {
	p := pkcs11.New(libPath)
	if p == nil {
		return nil, fmt.Errorf("unable to initialize pkcs11 module with path %s", libPath)
	}

	if err := p.Initialize(); err != nil {
		return nil, err
	}

	slotIdx := &slotIndex{labelSlotIdx: make(map[string]uint)}
	h := &hsm{p, cuUsername, cuPassword, soPassword, slotIdx}

	if err := h.buildSlotIndex(); err != nil {
		return nil, err
	}

	return h, nil
}

type slotIndex struct {
	mu           sync.Mutex
	labelSlotIdx map[string]uint
}

func (s *slotIndex) Set(label string, slotID uint) {
	s.mu.Lock()
	s.labelSlotIdx[label] = slotID
	s.mu.Unlock()
}

func (s *slotIndex) Get(label string) (uint, bool) {
	slotID, ok := s.labelSlotIdx[label]
	return slotID, ok
}
