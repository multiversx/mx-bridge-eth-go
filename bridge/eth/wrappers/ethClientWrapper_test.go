package wrappers

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/interactors"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func createMockArgsEthClientWrapper() (ArgsEthClientWrapper, *testsCommon.StatusHandlerMock) {
	statusHandler := testsCommon.NewStatusHandlerMock()

	return ArgsEthClientWrapper{
		BridgeContract:   &interactors.BridgeContractStub{},
		BlockchainClient: &interactors.BlockchainClientStub{},
		StatusHandler:    statusHandler,
	}, statusHandler
}

func TestNewEthClientWrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil bridge contract", func(t *testing.T) {
		args, _ := createMockArgsEthClientWrapper()
		args.BridgeContract = nil

		wrapper, err := NewEthClientWrapper(args)
		assert.True(t, check.IfNil(wrapper))
		assert.Equal(t, ErrNilBrdgeContract, err)
	})
	t.Run("nil blockchain client", func(t *testing.T) {
		args, _ := createMockArgsEthClientWrapper()
		args.BlockchainClient = nil

		wrapper, err := NewEthClientWrapper(args)
		assert.True(t, check.IfNil(wrapper))
		assert.Equal(t, ErrNilBlockchainClient, err)
	})
	t.Run("nil status handler", func(t *testing.T) {
		args, _ := createMockArgsEthClientWrapper()
		args.StatusHandler = nil

		wrapper, err := NewEthClientWrapper(args)
		assert.True(t, check.IfNil(wrapper))
		assert.Equal(t, ErrNilStatusHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		args, _ := createMockArgsEthClientWrapper()

		wrapper, err := NewEthClientWrapper(args)
		assert.False(t, check.IfNil(wrapper))
		assert.Nil(t, err)
	})
}

func TestEthClientWrapper_GetNextPendingBatch(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BridgeContract = &interactors.BridgeContractStub{
		GetNextPendingBatchCalled: func(opts *bind.CallOpts) (contract.Batch, error) {
			handlerCalled = true
			return contract.Batch{}, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	batch, err := wrapper.GetNextPendingBatch(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, contract.Batch{}, batch)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_GetRelayers(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BridgeContract = &interactors.BridgeContractStub{
		GetRelayersCalled: func(opts *bind.CallOpts) ([]common.Address, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	relayers, err := wrapper.GetRelayers(context.TODO())
	assert.Nil(t, err)
	assert.Nil(t, relayers)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_WasBatchExecuted(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BridgeContract = &interactors.BridgeContractStub{
		WasBatchExecutedCalled: func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error) {
			handlerCalled = true
			return false, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	executed, err := wrapper.WasBatchExecuted(context.TODO(), nil)
	assert.Nil(t, err)
	assert.False(t, executed)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_WasBatchFinished(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BridgeContract = &interactors.BridgeContractStub{
		WasBatchFinishedCalled: func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error) {
			handlerCalled = true
			return false, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	executed, err := wrapper.WasBatchFinished(context.TODO(), nil)
	assert.Nil(t, err)
	assert.False(t, executed)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_GetStatusesAfterExecution(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BridgeContract = &interactors.BridgeContractStub{
		GetStatusesAfterExecutionCalled: func(opts *bind.CallOpts, batchNonceElrondETH *big.Int) ([]uint8, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	statuses, err := wrapper.GetStatusesAfterExecution(context.TODO(), nil)
	assert.Nil(t, err)
	assert.Nil(t, statuses)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_ChainID(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BlockchainClient = &interactors.BlockchainClientStub{
		ChainIDCalled: func(ctx context.Context) (*big.Int, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	chainID, err := wrapper.ChainID(context.TODO())
	assert.Nil(t, err)
	assert.Nil(t, chainID)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_BlockNumber(t *testing.T) {
	t.Parallel()

	t.Run("block number call errors", func(t *testing.T) {
		args, statusHandler := createMockArgsEthClientWrapper()
		handlerCalled := false
		expectedError := errors.New("expected error")
		args.BlockchainClient = &interactors.BlockchainClientStub{
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				handlerCalled = true
				return 0, expectedError
			},
		}
		lastBlockNum := 3343
		statusHandler.SetIntMetric(core.MetricLastQueriedEthereumBlockNumber, lastBlockNum)

		wrapper, _ := NewEthClientWrapper(args)
		blockNum, err := wrapper.BlockNumber(context.TODO())
		assert.Equal(t, expectedError, err)
		assert.Equal(t, uint64(0), blockNum)
		assert.True(t, handlerCalled)
		assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
		assert.Equal(t, lastBlockNum, statusHandler.GetIntMetric(core.MetricLastQueriedEthereumBlockNumber))
	})
	t.Run("block number call returns a value", func(t *testing.T) {
		args, statusHandler := createMockArgsEthClientWrapper()
		handlerCalled := false
		newBlockNum := 772537
		args.BlockchainClient = &interactors.BlockchainClientStub{
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				handlerCalled = true
				return uint64(newBlockNum), nil
			},
		}
		lastBlockNum := 3343
		statusHandler.SetIntMetric(core.MetricLastQueriedEthereumBlockNumber, lastBlockNum)

		wrapper, _ := NewEthClientWrapper(args)
		blockNum, err := wrapper.BlockNumber(context.TODO())
		assert.Nil(t, err)
		assert.Equal(t, uint64(newBlockNum), blockNum)
		assert.True(t, handlerCalled)
		assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
		assert.Equal(t, newBlockNum, statusHandler.GetIntMetric(core.MetricLastQueriedEthereumBlockNumber))
	})
}

func TestEthClientWrapper_NonceAt(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BlockchainClient = &interactors.BlockchainClientStub{
		NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
			handlerCalled = true
			return 0, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	nonce, err := wrapper.NonceAt(context.TODO(), common.Address{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), nonce)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_FinishCurrentPendingBatch(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BridgeContract = &interactors.BridgeContractStub{
		FinishCurrentPendingBatchCalled: func(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	tx, err := wrapper.FinishCurrentPendingBatch(nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Nil(t, tx)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientTransactions))
}

func TestEthClientWrapper_ExecuteTransfer(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BridgeContract = &interactors.BridgeContractStub{
		ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	tx, err := wrapper.ExecuteTransfer(nil, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Nil(t, tx)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientTransactions))
}

func TestEthClientWrapper_Quorum(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthClientWrapper()
	handlerCalled := false
	args.BridgeContract = &interactors.BridgeContractStub{
		QuorumCalled: func(opts *bind.CallOpts) (*big.Int, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthClientWrapper(args)
	tx, err := wrapper.Quorum(context.TODO())
	assert.Nil(t, err)
	assert.Nil(t, tx)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}
