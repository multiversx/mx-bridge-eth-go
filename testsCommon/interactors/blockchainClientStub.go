package interactors

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BlockchainClientStub -
type BlockchainClientStub struct {
	BlockNumberCalled func(ctx context.Context) (uint64, error)
	NonceAtCalled     func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainIDCalled     func(ctx context.Context) (*big.Int, error)
}

// BlockNumber -
func (bcs *BlockchainClientStub) BlockNumber(ctx context.Context) (uint64, error) {
	if bcs.BlockNumberCalled != nil {
		return bcs.BlockNumberCalled(ctx)
	}

	return 0, nil
}

// NonceAt -
func (bcs *BlockchainClientStub) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	if bcs.NonceAtCalled != nil {
		return bcs.NonceAtCalled(ctx, account, blockNumber)
	}

	return 0, nil
}

// ChainID -
func (bcs *BlockchainClientStub) ChainID(ctx context.Context) (*big.Int, error) {
	if bcs.ChainIDCalled != nil {
		return bcs.ChainIDCalled(ctx)
	}

	return big.NewInt(0), nil
}
