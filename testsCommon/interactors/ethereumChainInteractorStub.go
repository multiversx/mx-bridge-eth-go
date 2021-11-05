package interactors

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthereumChainInteractorStub -
type EthereumChainInteractorStub struct {
	core.StatusHandler

	NameCalled                      func() string
	GetNextPendingBatchCalled       func(ctx context.Context) (contract.Batch, error)
	WasBatchExecutedCalled          func(ctx context.Context, batchNonce *big.Int) (bool, error)
	WasBatchFinishedCalled          func(ctx context.Context, batchNonce *big.Int) (bool, error)
	GetStatusesAfterExecutionCalled func(ctx context.Context, batchNonceElrondETH *big.Int) ([]uint8, error)
	ChainIDCalled                   func(ctx context.Context) (*big.Int, error)
	BlockNumberCalled               func(ctx context.Context) (uint64, error)
	NonceAtCalled                   func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	FinishCurrentPendingBatchCalled func(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error)
	ExecuteTransferCalled           func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	QuorumCalled                    func(ctx context.Context) (*big.Int, error)
	GetRelayersCalled               func(ctx context.Context) ([]common.Address, error)
}

// GetNextPendingBatch -
func (stub *EthereumChainInteractorStub) GetNextPendingBatch(ctx context.Context) (contract.Batch, error) {
	if stub.GetNextPendingBatchCalled != nil {
		return stub.GetNextPendingBatchCalled(ctx)
	}

	return contract.Batch{}, nil
}

// WasBatchExecuted -
func (stub *EthereumChainInteractorStub) WasBatchExecuted(ctx context.Context, batchNonce *big.Int) (bool, error) {
	if stub.WasBatchExecutedCalled != nil {
		return stub.WasBatchExecutedCalled(ctx, batchNonce)
	}

	return false, nil
}

// WasBatchFinished -
func (stub *EthereumChainInteractorStub) WasBatchFinished(ctx context.Context, batchNonce *big.Int) (bool, error) {
	if stub.WasBatchFinishedCalled != nil {
		return stub.WasBatchFinishedCalled(ctx, batchNonce)
	}

	return false, nil
}

// GetStatusesAfterExecution -
func (stub *EthereumChainInteractorStub) GetStatusesAfterExecution(ctx context.Context, batchNonceElrondETH *big.Int) ([]uint8, error) {
	if stub.GetStatusesAfterExecutionCalled != nil {
		return stub.GetStatusesAfterExecutionCalled(ctx, batchNonceElrondETH)
	}

	return make([]byte, 0), nil
}

// ChainID -
func (stub *EthereumChainInteractorStub) ChainID(ctx context.Context) (*big.Int, error) {
	if stub.ChainIDCalled != nil {
		return stub.ChainIDCalled(ctx)
	}

	return big.NewInt(0), nil
}

// BlockNumber -
func (stub *EthereumChainInteractorStub) BlockNumber(ctx context.Context) (uint64, error) {
	if stub.BlockNumberCalled != nil {
		return stub.BlockNumberCalled(ctx)
	}

	return 0, nil
}

// NonceAt -
func (stub *EthereumChainInteractorStub) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	if stub.NonceAtCalled != nil {
		return stub.NonceAtCalled(ctx, account, blockNumber)
	}

	return 0, nil
}

// FinishCurrentPendingBatch -
func (stub *EthereumChainInteractorStub) FinishCurrentPendingBatch(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error) {
	if stub.FinishCurrentPendingBatchCalled != nil {
		return stub.FinishCurrentPendingBatchCalled(opts, batchNonce, newDepositStatuses, signatures)
	}

	return nil, errNotImplemented
}

// ExecuteTransfer -
func (stub *EthereumChainInteractorStub) ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	if stub.ExecuteTransferCalled != nil {
		return stub.ExecuteTransferCalled(opts, tokens, recipients, amounts, batchNonce, signatures)
	}

	return nil, errNotImplemented
}

// Quorum -
func (stub *EthereumChainInteractorStub) Quorum(ctx context.Context) (*big.Int, error) {
	if stub.QuorumCalled != nil {
		return stub.QuorumCalled(ctx)
	}

	return big.NewInt(0), nil
}

// GetRelayers -
func (stub *EthereumChainInteractorStub) GetRelayers(ctx context.Context) ([]common.Address, error) {
	if stub.GetRelayersCalled != nil {
		return stub.GetRelayersCalled(ctx)
	}

	return nil, nil
}

// IsInterfaceNil -
func (stub *EthereumChainInteractorStub) IsInterfaceNil() bool {
	return stub == nil
}
