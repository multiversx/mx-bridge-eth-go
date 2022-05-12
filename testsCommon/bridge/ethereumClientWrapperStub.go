package bridge

import (
	"context"
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthereumClientWrapperStub -
type EthereumClientWrapperStub struct {
	GetBatchCalled         func(ctx context.Context, batchNonce *big.Int) (contract.Batch, error)
	GetBatchDepositsCalled func(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, error)
	GetRelayersCalled      func(ctx context.Context) ([]common.Address, error)
	WasBatchExecutedCalled func(ctx context.Context, batchNonce *big.Int) (bool, error)
	ChainIDCalled          func(ctx context.Context) (*big.Int, error)
	BlockNumberCalled      func(ctx context.Context) (uint64, error)
	NonceAtCalled          func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	ExecuteTransferCalled  func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address,
		amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	QuorumCalled                    func(ctx context.Context) (*big.Int, error)
	GetStatusesAfterExecutionCalled func(ctx context.Context, batchID *big.Int) ([]byte, error)
	BalanceAtCalled                 func(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)

	SetIntMetricCalled    func(metric string, value int)
	AddIntMetricCalled    func(metric string, delta int)
	SetStringMetricCalled func(metric string, val string)
	GetAllMetricsCalled   func() core.GeneralMetrics
	NameCalled            func() string
}

// SetIntMetric -
func (stub *EthereumClientWrapperStub) SetIntMetric(metric string, value int) {
	if stub.SetIntMetricCalled != nil {
		stub.SetIntMetricCalled(metric, value)
	}
}

// AddIntMetric -
func (stub *EthereumClientWrapperStub) AddIntMetric(metric string, delta int) {
	if stub.AddIntMetricCalled != nil {
		stub.AddIntMetricCalled(metric, delta)
	}
}

// SetStringMetric -
func (stub *EthereumClientWrapperStub) SetStringMetric(metric string, val string) {
	if stub.SetStringMetricCalled != nil {
		stub.SetStringMetricCalled(metric, val)
	}
}

// GetAllMetrics -
func (stub *EthereumClientWrapperStub) GetAllMetrics() core.GeneralMetrics {
	if stub.GetAllMetricsCalled != nil {
		stub.GetAllMetricsCalled()
	}
	return make(core.GeneralMetrics)
}

// Name -
func (stub *EthereumClientWrapperStub) Name() string {
	if stub.NameCalled != nil {
		stub.NameCalled()
	}
	return ""
}

// GetBatch -
func (stub *EthereumClientWrapperStub) GetBatch(ctx context.Context, batchNonce *big.Int) (contract.Batch, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchCalled(ctx, batchNonce)
	}

	return contract.Batch{}, nil
}

// GetBatchDeposits -
func (stub *EthereumClientWrapperStub) GetBatchDeposits(ctx context.Context, batchNonce *big.Int) ([]contract.Deposit, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchDepositsCalled(ctx, batchNonce)
	}

	return make([]contract.Deposit, 0), nil
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

// GetStatusesAfterExecution -
func (stub *EthereumClientWrapperStub) GetStatusesAfterExecution(ctx context.Context, batchID *big.Int) ([]byte, error) {
	if stub.GetStatusesAfterExecutionCalled != nil {
		return stub.GetStatusesAfterExecutionCalled(ctx, batchID)
	}

	return make([]byte, 0), nil
}

// BalanceAt -
func (stub *EthereumClientWrapperStub) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	if stub.BalanceAtCalled != nil {
		return stub.BalanceAtCalled(ctx, account, blockNumber)
	}

	return big.NewInt(0), nil
}

// IsInterfaceNil -
func (stub *EthereumClientWrapperStub) IsInterfaceNil() bool {
	return stub == nil
}
