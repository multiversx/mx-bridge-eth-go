package bridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// SafeContractStub -
type SafeContractStub struct {
	TotalBalancesCalled     func(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error)
	MintBalancesCalled      func(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error)
	BurnBalancesCalled      func(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error)
	MintBurnTokensCalled    func(opts *bind.CallOpts, arg0 common.Address) (bool, error)
	NativeTokensCalled      func(opts *bind.CallOpts, arg0 common.Address) (bool, error)
	WhiteListedTokensCalled func(opts *bind.CallOpts, arg0 common.Address) (bool, error)
}

// TotalBalances -
func (stub *SafeContractStub) TotalBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	if stub.TotalBalancesCalled != nil {
		return stub.TotalBalancesCalled(opts, arg0)
	}

	return big.NewInt(0), nil
}

// MintBalances -
func (stub *SafeContractStub) MintBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	if stub.MintBalancesCalled != nil {
		return stub.MintBalancesCalled(opts, arg0)
	}

	return big.NewInt(0), nil
}

// BurnBalances -
func (stub *SafeContractStub) BurnBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	if stub.BurnBalancesCalled != nil {
		return stub.BurnBalancesCalled(opts, arg0)
	}

	return big.NewInt(0), nil
}

// MintBurnTokens -
func (stub *SafeContractStub) MintBurnTokens(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	if stub.MintBurnTokensCalled != nil {
		return stub.MintBurnTokensCalled(opts, arg0)
	}

	return false, nil
}

// NativeTokens -
func (stub *SafeContractStub) NativeTokens(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	if stub.NativeTokensCalled != nil {
		return stub.NativeTokensCalled(opts, arg0)
	}

	return false, nil
}

// WhiteListedTokens -
func (stub *SafeContractStub) WhiteListedTokens(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	if stub.WhiteListedTokensCalled != nil {
		return stub.WhiteListedTokensCalled(opts, arg0)
	}

	return false, nil
}
