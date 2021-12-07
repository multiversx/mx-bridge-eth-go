package bridgeV2

import (
	"context"
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthereumClientWrapperStub -
type EthereumClientWrapperStub struct {
	core.StatusHandler

	GetBatchCalled         func(ctx context.Context, batchNonce *big.Int) (contract.Batch, error)
	GetRelayersCalled      func(ctx context.Context) ([]common.Address, error)
	WasBatchExecutedCalled func(ctx context.Context, batchNonce *big.Int) (bool, error)
	ChainIDCalled          func(ctx context.Context) (*big.Int, error)
	BlockNumberCalled      func(ctx context.Context) (uint64, error)
	NonceAtCalled          func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ExecuteTransferCalled  func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address,
		amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	QuorumCalled func(ctx context.Context) (*big.Int, error)
}

// GetBatch -
func (stub *EthereumClientWrapperStub) GetBatch(ctx context.Context, batchNonce *big.Int) (contract.Batch, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchCalled(ctx, batchNonce)
	}

	return contract.Batch{}, nil
}

// GetRelayers -
func (stub *EthereumClientWrapperStub) GetRelayers(ctx context.Context) ([]common.Address, error) {
	if stub.GetRelayersCalled != nil {
		return stub.GetRelayersCalled(ctx)
	}

	return make([]common.Address, 0), nil
}

// WasBatchExecuted -
func (stub *EthereumClientWrapperStub) WasBatchExecuted(ctx context.Context, batchNonce *big.Int) (bool, error) {
	if stub.WasBatchExecutedCalled != nil {
		return stub.WasBatchExecutedCalled(ctx, batchNonce)
	}

	return true, nil
}

// ChainID -
func (stub *EthereumClientWrapperStub) ChainID(ctx context.Context) (*big.Int, error) {
	if stub.ChainIDCalled != nil {
		return stub.ChainIDCalled(ctx)
	}

	return big.NewInt(0), nil
}

// BlockNumber -
func (stub *EthereumClientWrapperStub) BlockNumber(ctx context.Context) (uint64, error) {
	if stub.BlockNumberCalled != nil {
		return stub.BlockNumberCalled(ctx)
	}

	return 0, nil
}

// NonceAt -
func (stub *EthereumClientWrapperStub) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	if stub.NonceAtCalled != nil {
		return stub.NonceAtCalled(ctx, account, blockNumber)
	}

	return 0, nil
}

// ExecuteTransfer -
func (stub *EthereumClientWrapperStub) ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	if stub.ExecuteTransferCalled != nil {
		return stub.ExecuteTransferCalled(opts, tokens, recipients, amounts, nonces, batchNonce, signatures)
	}

	return nil, errors.New("not implemented")
}

// Quorum -
func (stub *EthereumClientWrapperStub) Quorum(ctx context.Context) (*big.Int, error) {
	if stub.QuorumCalled != nil {
		return stub.QuorumCalled(ctx)
	}

	return big.NewInt(0), nil
}

// IsInterfaceNil -
func (stub *EthereumClientWrapperStub) IsInterfaceNil() bool {
	return stub == nil
}
