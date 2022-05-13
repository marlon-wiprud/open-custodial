package hsm

import (
	"fmt"

	"github.com/miekg/pkcs11"
)

func (h *hsm) NewSession(slotID uint) (pkcs11.SessionHandle, error) {
	return newSession(h.ctx, slotID)
}

func newSession(ctx *pkcs11.Ctx, slotID uint) (pkcs11.SessionHandle, error) {
	// TODO should session params come from function args?
	return ctx.OpenSession(slotID, pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
}

func (h *hsm) NewSlotSession(name string) (*pkcs11.SessionHandle, error) {
	slot, err := h.GetSlotID(name)
	if err != nil {
		return nil, err
	}

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

func (h *hsm) buildSOPin() string {
	return fmt.Sprintf("%s:%s", h.cuUsername, h.cuPassword)
}

func (h *hsm) buildPin() string {
	return fmt.Sprintf("%s:%s", h.cuUsername, h.cuPassword)
}
