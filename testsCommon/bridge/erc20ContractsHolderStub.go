package bridge

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ERC20ContractsHolderStub -
type ERC20ContractsHolderStub struct {
	BalanceOfCalled func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error)
}

// BalanceOf -
func (stub *ERC20ContractsHolderStub) BalanceOf(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
	if stub.BalanceOfCalled != nil {
		return stub.BalanceOfCalled(ctx, erc20Address, address)
	}

	return big.NewInt(0), nil
}

// IsInterfaceNil -
func (stub *ERC20ContractsHolderStub) IsInterfaceNil() bool {
	return stub == nil
}
