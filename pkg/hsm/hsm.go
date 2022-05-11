package hsm

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
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

func (h *hsm) NewSlot(name string) (uint, error) {
	slots, err := h.ctx.GetSlotList(true)
	if err != nil {
		return 0, err
	}

	_, err = findSlotByName(h.ctx, slots, name)
	if err == nil {
		return 0, errors.New("slots already exists")
	}

	slotID := uint(len(slots)) - 1

	if err := h.ctx.InitToken(slotID, h.buildSOPin(), name); err != nil {
		fmt.Println("new_slot: failed to init token ", err)
		return 0, err
	}

	sess, err := newSession(h.ctx, slotID)
	if err != nil {
		fmt.Println("new_slot: failed to open session ", err)
		return 0, err
	}

	defer h.EndSession(&sess)

	if err := h.ctx.Login(sess, pkcs11.CKU_SO, h.buildSOPin()); err != nil {
		fmt.Println("new_slot: failed to login ", err)
		return 0, err
	}

	if err := h.ctx.InitPIN(sess, h.buildPin()); err != nil {
		fmt.Println("new_slot: failed to init pin ", err)
		return 0, err
	}

	return slotID, nil
}

func (h *hsm) SlotExists(name string) (bool, error) {
	slots, err := h.ctx.GetSlotList(true)
	if err != nil {
		return false, err
	}

	_, err = findSlotByName(h.ctx, slots, name)
	return err == nil, err
}

func (h *hsm) NewSlotSession(name string) (*pkcs11.SessionHandle, error) {
	slots, err := h.ctx.GetSlotList(true)
	if err != nil {
		return nil, err
	}

	slot, err := findSlotByName(h.ctx, slots, name)
	if err != nil {
		return nil, err
	}

	// TODO should session params come from function args?
	sess, err := newSession(h.ctx, slot)
	if err != nil {
		return nil, err
	}

	err = h.ctx.Login(sess, pkcs11.CKU_USER, h.buildPin())
	if err != nil {
		return nil, err
	}

	return &sess, nil
}

func (h *hsm) EndSession(sess *pkcs11.SessionHandle) error {
	if err := h.ctx.Logout(*sess); err != nil {
		return err
	}

	if err := h.ctx.CloseSession(*sess); err != nil {
		return err
	}

	if err := h.ctx.FindObjectsFinal(*sess); err != nil {
		return err
	}

	return nil
}

func (h *hsm) PublicKeyHandle(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, func(), error) {

	attr := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_ECDSA),
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PUBLIC_KEY),
	}

	return h.findWithAttributes(session, attr)
}

func (h *hsm) PrivateKeyHandle(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, func(), error) {

	attr := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_ECDSA),
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PRIVATE_KEY),
	}

	return h.findWithAttributes(session, attr)
}

func (h *hsm) GetPublicKey(session pkcs11.SessionHandle, pubKeyHandle pkcs11.ObjectHandle) (ecdsa.PublicKey, error) {
	var pub ecdsa.PublicKey

	// define which attributes we want to get
	// result fills in the nil value
	template := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_ECDSA_PARAMS, nil),
		pkcs11.NewAttribute(pkcs11.CKA_EC_POINT, nil),
	}

	// gets attributes based on template array
	attributes, err := h.ctx.GetAttributeValue(session, pubKeyHandle, template)
	if err != nil {
		return pub, err
	}

	// pub curve could be hard coded, but using this function for extra safety
	// ensures that the attribute adheres to secp256k1
	pub.Curve, err = unmarshalEcParams(attributes[0].Value)
	if err != nil {
		return pub, err
	}

	pub.X, pub.Y, err = unmarshalEcPoint(attributes[1].Value, pub.Curve)
	if err != nil {
		return pub, err
	}

	return pub, nil
}

// TODO - the key generated here will not be verified correctly. This makes testing pretty tricky.
func (h *hsm) GenerateKeyECDSA_secp256k1(session pkcs11.SessionHandle) (pkcs11.ObjectHandle, pkcs11.ObjectHandle, error) {

	oid, err := secp256k1OID()
	if err != nil {
		return 0, 0, errors.New("unable to get secp256 OID")
	}

	pubTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, pkcs11.CKO_PUBLIC_KEY),
		pkcs11.NewAttribute(pkcs11.CKA_KEY_TYPE, pkcs11.CKK_ECDSA),
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_VERIFY, true),
		pkcs11.NewAttribute(pkcs11.CKA_ECDSA_PARAMS, oid),
	}

	privTemplate := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_TOKEN, true),
		pkcs11.NewAttribute(pkcs11.CKA_SIGN, true),
		pkcs11.NewAttribute(pkcs11.CKA_SENSITIVE, true),
		pkcs11.NewAttribute(pkcs11.CKA_EXTRACTABLE, false),
		pkcs11.NewAttribute(pkcs11.CKA_SIGN_RECOVER, true),
	}

	// important to have the right mechanism
	mech := []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_ECDSA_KEY_PAIR_GEN, nil)}

	return h.ctx.GenerateKeyPair(session, mech, pubTemplate, privTemplate)
}

// asn1 object identiefier based on
// http://oid-info.com/get/1.3.132.0.10
func secp256k1OID() ([]byte, error) {
	return asn1.Marshal(asn1.ObjectIdentifier{1, 3, 132, 0, 10})
}

func unmarshalEcPoint(b []byte, c elliptic.Curve) (*big.Int, *big.Int, error) {
	var pointBytes []byte
	extra, err := asn1.Unmarshal(b, &pointBytes)
	if err != nil {
		return nil, nil, err
	}

	if len(extra) > 0 {
		return nil, nil, errors.New("unexpected data found when parsing elliptic curve point")
	}

	x, y := elliptic.Unmarshal(c, pointBytes)
	if x == nil || y == nil {
		return nil, nil, errors.New("failed to parse elliptic curve point")
	}
	return x, y, nil
}

func (h *hsm) buildSOPin() string {
	return fmt.Sprintf("%s:%s", h.cuUsername, h.cuPassword)
}

func (h *hsm) buildPin() string {
	return fmt.Sprintf("%s:%s", h.cuUsername, h.cuPassword)
}

func unmarshalEcParams(b []byte) (elliptic.Curve, error) {
	oid, err := secp256k1OID()
	if err != nil {
		return nil, err
	}

	if bytes.Equal(b, oid) {
		return secp256k1.S256(), nil
	}

	return nil, errors.New("ec params do not match secp256k1")
}

func (h *hsm) findWithAttributes(session pkcs11.SessionHandle, attr []*pkcs11.Attribute) (pkcs11.ObjectHandle, func(), error) {
	finisher := func() {
		h.ctx.FindObjectsFinal(session)
	}

	err := h.ctx.FindObjectsInit(session, attr)
	if err != nil {
		return pkcs11.ObjectHandle(1), finisher, err
	}

	obj, _, err := h.ctx.FindObjects(session, 1)
	if err != nil {
		return pkcs11.ObjectHandle(1), finisher, err
	}

	if len(obj) == 0 {
		return pkcs11.ObjectHandle(1), finisher, errors.New("no objects found")
	}

	return obj[0], finisher, nil
}

func findSlotByName(ctx *pkcs11.Ctx, slots []uint, name string) (uint, error) {
	for _, s := range slots {
		info, err := ctx.GetTokenInfo(s)
		if err != nil {
			return 0, err
		}

		if info.Label == name {
			return s, nil
		}
	}

	return 0, fmt.Errorf("unable to find slots with name %s", name)
}

func newSession(ctx *pkcs11.Ctx, slotID uint) (pkcs11.SessionHandle, error) {
	// TODO should session params come from function args?
	return ctx.OpenSession(slotID, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
}
