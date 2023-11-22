package interactors

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BlockchainClientStub -
type BlockchainClientStub struct {
	BlockNumberCalled func(ctx context.Context) (uint64, error)
	NonceAtCalled     func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ChainIDCalled     func(ctx context.Context) (*big.Int, error)
	BalanceAtCalled   func(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	FilterLogsCalled  func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
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

// BalanceAt -
func (bcs *BlockchainClientStub) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	if bcs.BalanceAtCalled != nil {
		return bcs.BalanceAtCalled(ctx, account, blockNumber)
	}

	return big.NewInt(0), nil
}

// FilterLogs -
func (bcs *BlockchainClientStub) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if bcs.FilterLogsCalled != nil {
		return bcs.FilterLogsCalled(ctx, q)
	}

	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (bcs *BlockchainClientStub) IsInterfaceNil() bool {
	return bcs == nil
}
