package bridge

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
)

// EthereumClientWrapperStub -
type EthereumClientWrapperStub struct {
	core.StatusHandler
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
	TotalBalancesCalled             func(ctx context.Context, account common.Address) (*big.Int, error)
	MintBalancesCalled              func(ctx context.Context, account common.Address) (*big.Int, error)
	BurnBalancesCalled              func(ctx context.Context, account common.Address) (*big.Int, error)
	MintBurnTokensCalled            func(ctx context.Context, account common.Address) (bool, error)
	NativeTokensCalled              func(ctx context.Context, account common.Address) (bool, error)
	WhitelistedTokensCalled         func(ctx context.Context, account common.Address) (bool, error)

	SetIntMetricCalled    func(metric string, value int)
	AddIntMetricCalled    func(metric string, delta int)
	SetStringMetricCalled func(metric string, val string)
	GetAllMetricsCalled   func() core.GeneralMetrics
	NameCalled            func() string
	IsPausedCalled        func(ctx context.Context) (bool, error)
	FilterLogsCalled      func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

// SetIntMetric -
func (stub *EthereumClientWrapperStub) SetIntMetric(metric string, value int) {
	if stub.SetIntMetricCalled != nil {
		stub.SetIntMetricCalled(metric, value)
	}
	if stub.StatusHandler != nil {
		stub.StatusHandler.SetIntMetric(metric, value)
	}
}

// AddIntMetric -
func (stub *EthereumClientWrapperStub) AddIntMetric(metric string, delta int) {
	if stub.AddIntMetricCalled != nil {
		stub.AddIntMetricCalled(metric, delta)
	}
	if stub.StatusHandler != nil {
		stub.StatusHandler.AddIntMetric(metric, delta)
	}
}

// SetStringMetric -
func (stub *EthereumClientWrapperStub) SetStringMetric(metric string, val string) {
	if stub.SetStringMetricCalled != nil {
		stub.SetStringMetricCalled(metric, val)
	}
	if stub.StatusHandler != nil {
		stub.StatusHandler.SetStringMetric(metric, val)
	}
}

// GetAllMetrics -
func (stub *EthereumClientWrapperStub) GetAllMetrics() core.GeneralMetrics {
	if stub.GetAllMetricsCalled != nil {
		return stub.GetAllMetricsCalled()
	}
	if stub.StatusHandler != nil {
		return stub.StatusHandler.GetAllMetrics()
	}

	return make(core.GeneralMetrics)
}

// Name -
func (stub *EthereumClientWrapperStub) Name() string {
	if stub.NameCalled != nil {
		stub.NameCalled()
	}
	if stub.StatusHandler != nil {
		return stub.StatusHandler.Name()
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

// FilterLogs -
func (stub *EthereumClientWrapperStub) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if stub.FilterLogsCalled != nil {
		return stub.FilterLogsCalled(ctx, q)
	}

	return []types.Log{}, nil
}

// IsPaused -
func (stub *EthereumClientWrapperStub) IsPaused(ctx context.Context) (bool, error) {
	if stub.IsPausedCalled != nil {
		return stub.IsPausedCalled(ctx)
	}

	return false, nil
}

// TotalBalances -
func (stub *EthereumClientWrapperStub) TotalBalances(ctx context.Context, account common.Address) (*big.Int, error) {
	if stub.TotalBalancesCalled != nil {
		return stub.TotalBalancesCalled(ctx, account)
	}

	return big.NewInt(0), nil
}

// MintBalances -
func (stub *EthereumClientWrapperStub) MintBalances(ctx context.Context, account common.Address) (*big.Int, error) {
	if stub.MintBalancesCalled != nil {
		return stub.MintBalancesCalled(ctx, account)
	}

	return big.NewInt(0), nil
}

// BurnBalances -
func (stub *EthereumClientWrapperStub) BurnBalances(ctx context.Context, account common.Address) (*big.Int, error) {
	if stub.BurnBalancesCalled != nil {
		return stub.BurnBalancesCalled(ctx, account)
	}

	return big.NewInt(0), nil
}

// MintBurnTokens -
func (stub *EthereumClientWrapperStub) MintBurnTokens(ctx context.Context, account common.Address) (bool, error) {
	if stub.MintBurnTokensCalled != nil {
		return stub.MintBurnTokensCalled(ctx, account)
	}

	return false, nil
}

// NativeTokens -
func (stub *EthereumClientWrapperStub) NativeTokens(ctx context.Context, account common.Address) (bool, error) {
	if stub.NativeTokensCalled != nil {
		return stub.NativeTokensCalled(ctx, account)
	}

	return false, nil
}

// WhitelistedTokens -
func (stub *EthereumClientWrapperStub) WhitelistedTokens(ctx context.Context, account common.Address) (bool, error) {
	if stub.WhitelistedTokensCalled != nil {
		return stub.WhitelistedTokensCalled(ctx, account)
	}

	return false, nil
}

// IsInterfaceNil -
func (stub *EthereumClientWrapperStub) IsInterfaceNil() bool {
	return stub == nil
}
