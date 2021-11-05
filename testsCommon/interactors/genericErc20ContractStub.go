package interactors

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// GenericErc20ContractStub -
type GenericErc20ContractStub struct {
	BalanceOfCalled func(account common.Address) (*big.Int, error)
}

// BalanceOf -
func (stub *GenericErc20ContractStub) BalanceOf(_ *bind.CallOpts, account common.Address) (*big.Int, error) {
	if stub.BalanceOfCalled != nil {
		return stub.BalanceOfCalled(account)
	}

	return nil, errors.New("GenericErc20ContractStub.BalanceOf not implemented")
}
