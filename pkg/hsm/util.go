package hsm

import (
	"bytes"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/miekg/pkcs11"
)

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
