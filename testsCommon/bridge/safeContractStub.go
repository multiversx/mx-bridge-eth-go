package bridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// SafeContractStub -
type SafeContractStub struct {
	MintBalancesCalled      func(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error)
	WhitelistedTokensCalled func(opts *bind.CallOpts, arg0 common.Address) (bool, error)
}

// MintBalances -
func (stub *SafeContractStub) MintBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	if stub.MintBalancesCalled != nil {
		return stub.MintBalancesCalled(opts, arg0)
	}
	return big.NewInt(0), nil // or any other default value
}

// WhitelistedTokens -
func (stub *SafeContractStub) WhitelistedTokens(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	if stub.WhitelistedTokensCalled != nil {
		return stub.WhitelistedTokensCalled(opts, arg0)
	}
	return false, nil // or any other default value
}
