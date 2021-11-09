package interactors

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// Erc20ContractStub -
type Erc20ContractStub struct {
	BalanceOfCalled func(ctx context.Context, account common.Address) (*big.Int, error)
}

// BalanceOf -
func (stub *Erc20ContractStub) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	if stub.BalanceOfCalled != nil {
		return stub.BalanceOfCalled(ctx, account)
	}

	return big.NewInt(0), nil
}

// IsInterfaceNil -
func (stub *Erc20ContractStub) IsInterfaceNil() bool {
	return stub == nil
}
