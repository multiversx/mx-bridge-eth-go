package bridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// SafeContractStub -
type SafeContractStub struct {
	TokenMintedBalancesCalled       func(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error)
	WhitelistedTokensMintBurnCalled func(opts *bind.CallOpts, arg0 common.Address) (bool, error)
}

// TokenMintedBalances -
func (stub *SafeContractStub) TokenMintedBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	if stub.TokenMintedBalancesCalled != nil {
		return stub.TokenMintedBalancesCalled(opts, arg0)
	}
	return big.NewInt(0), nil // or any other default value
}

// WhitelistedTokensMintBurn -
func (stub *SafeContractStub) WhitelistedTokensMintBurn(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	if stub.WhitelistedTokensMintBurnCalled != nil {
		return stub.WhitelistedTokensMintBurnCalled(opts, arg0)
	}
	return false, nil // or any other default value
}
