package hsm

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/miekg/pkcs11"
)

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
