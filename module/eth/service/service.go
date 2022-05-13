package eth_svc

import (
	"open_custodial/pkg/hsm"

	"open_custodial/pkg/eth"

	"github.com/ethereum/go-ethereum/common"
)

type ETHService interface {
	CreateAddress(label string) (a Address, err error)
	GetAddressByLabel(label string) (a Address, err error)
	GetSlotAddress(slotID uint) (a Address, err error)
}

type service struct {
	hsm hsm.HSM
}

func NewETHService(h hsm.HSM) ETHService {
	return &service{hsm: h}
}

type Address struct {
	Addr  common.Address `json:"address"`
	Label string         `json:"label"`
}

func (s *service) CreateAddress(label string) (a Address, err error) {
	a.Label = label
	a.Addr, err = eth.CreateAddress(s.hsm, label)
	if err != nil {
		return a, err
	}

	return a, nil
}

func (s *service) GetAddressByLabel(label string) (a Address, err error) {
	a.Label = label
	a.Addr, err = eth.GetAddress(s.hsm, label)
	if err != nil {
		return a, err
	}

	return a, nil
}

func (s *service) GetSlotAddress(slotID uint) (a Address, err error) {
	a.Addr, err = eth.GetSlotAddress(s.hsm, slotID)
	if err != nil {
		return a, err
	}

	return a, nil
}
