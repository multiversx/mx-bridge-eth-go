package bridge

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// CryptoHandlerStub -
type CryptoHandlerStub struct {
	SignCalled                  func(msgHash common.Hash) ([]byte, error)
	GetAddressCalled            func() common.Address
	CreateKeyedTransactorCalled func(chainId *big.Int) (*bind.TransactOpts, error)
}

// Sign -
func (stub *CryptoHandlerStub) Sign(msgHash common.Hash) ([]byte, error) {
	if stub.SignCalled != nil {
		return stub.SignCalled(msgHash)
	}

	return make([]byte, 0), nil
}

// GetAddress -
func (stub *CryptoHandlerStub) GetAddress() common.Address {
	if stub.GetAddressCalled != nil {
		return stub.GetAddressCalled()
	}

	return common.BytesToAddress(make([]byte, 0))
}

// CreateKeyedTransactor -
func (stub *CryptoHandlerStub) CreateKeyedTransactor(chainId *big.Int) (*bind.TransactOpts, error) {
	if stub.CreateKeyedTransactorCalled != nil {
		return stub.CreateKeyedTransactorCalled(chainId)
	}

	return nil, fmt.Errorf("not implemented")
}

// IsInterfaceNil -
func (stub *CryptoHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
