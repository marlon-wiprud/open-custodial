package eth_svc

import (
	"math/big"
	"open_custodial/pkg/hsm"

	validator_svc "open_custodial/module/validator/service"
	eth "open_custodial/pkg/eth_hsm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ETHService interface {
	CreateAddress(label string) (a Address, err error)
	GetAddressByLabel(label string) (a Address, err error)
	GetSlotAddress(slotID uint) (a Address, err error)
	SignTransaction(tx *types.Transaction, chainID *big.Int, label string) (*types.Transaction, error)
}

type service struct {
	hsm       hsm.HSM
	validator validator_svc.ValidatorService
}

func NewETHService(h hsm.HSM, v validator_svc.ValidatorService) ETHService {
	return &service{hsm: h, validator: v}
}

type Address struct {
	Addr  common.Address `json:"address"`
	Label string         `json:"label"`
}

func (s *service) CreateAddress(label string) (a Address, err error) {
	if err := s.validator.ValidateCreateAddress(); err != nil {
		return a, err
	}

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

func (s *service) SignTransaction(tx *types.Transaction, chainID *big.Int, label string) (*types.Transaction, error) {
	if err := s.validator.ValidateSign(); err != nil {
		return nil, err
	}
	return eth.SignTransaction(s.hsm, tx, label, chainID)
}
