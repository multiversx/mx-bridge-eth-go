package wrappers

import (
	"context"
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func createMockArgsEthereumChainWrapper() (ArgsEthereumChainWrapper, *testsCommon.StatusHandlerMock) {
	statusHandler := testsCommon.NewStatusHandlerMock("mock")

	return ArgsEthereumChainWrapper{
		MultiSigContract:    &bridgeTests.MultiSigContractStub{},
		BlockchainClient:    &interactors.BlockchainClientStub{},
		SCExecProxyContract: &bridgeTests.SCExecProxyContractStub{},
		StatusHandler:       statusHandler,
	}, statusHandler
}

func TestNewMultiSigContractWrapper(t *testing.T) {
	t.Parallel()

	t.Run("nil status handler", func(t *testing.T) {
		t.Parallel()

		args, _ := createMockArgsEthereumChainWrapper()
		args.StatusHandler = nil

		wrapper, err := NewEthereumChainWrapper(args)
		assert.True(t, check.IfNil(wrapper))
		assert.Equal(t, clients.ErrNilStatusHandler, err)
	})
	t.Run("nil blockchain client", func(t *testing.T) {
		t.Parallel()

		args, _ := createMockArgsEthereumChainWrapper()
		args.BlockchainClient = nil

		wrapper, err := NewEthereumChainWrapper(args)
		assert.True(t, check.IfNil(wrapper))
		assert.Equal(t, errNilBlockchainClient, err)
	})
	t.Run("nil multisig contract", func(t *testing.T) {
		t.Parallel()

		args, _ := createMockArgsEthereumChainWrapper()
		args.MultiSigContract = nil

		wrapper, err := NewEthereumChainWrapper(args)
		assert.True(t, check.IfNil(wrapper))
		assert.Equal(t, errNilMultiSigContract, err)
	})
	t.Run("nil sc exec contract", func(t *testing.T) {
		t.Parallel()

		args, _ := createMockArgsEthereumChainWrapper()
		args.MultiSigContract = nil

		wrapper, err := NewEthereumChainWrapper(args)
		assert.True(t, check.IfNil(wrapper))
		assert.Equal(t, errNilMultiSigContract, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args, _ := createMockArgsEthereumChainWrapper()

		wrapper, err := NewEthereumChainWrapper(args)
		assert.False(t, check.IfNil(wrapper))
		assert.Nil(t, err)
	})
}

func TestEthClientWrapper_GetBatch(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	providedBatchID := big.NewInt(223)
	args.MultiSigContract = &bridgeTests.MultiSigContractStub{
		GetBatchCalled: func(opts *bind.CallOpts, batchNonce *big.Int) (contract.Batch, error) {
			handlerCalled = true
			assert.Equal(t, providedBatchID, batchNonce)
			return contract.Batch{}, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	batch, err := wrapper.GetBatch(context.Background(), providedBatchID)
	assert.Nil(t, err)
	assert.Equal(t, contract.Batch{}, batch)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_GetRelayers(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	args.MultiSigContract = &bridgeTests.MultiSigContractStub{
		GetRelayersCalled: func(opts *bind.CallOpts) ([]common.Address, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	relayers, err := wrapper.GetRelayers(context.Background())
	assert.Nil(t, err)
	assert.Nil(t, relayers)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_WasBatchExecuted(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	args.MultiSigContract = &bridgeTests.MultiSigContractStub{
		WasBatchExecutedCalled: func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error) {
			handlerCalled = true
			return false, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	executed, err := wrapper.WasBatchExecuted(context.Background(), nil)
	assert.Nil(t, err)
	assert.False(t, executed)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_ChainID(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	args.BlockchainClient = &interactors.BlockchainClientStub{
		ChainIDCalled: func(ctx context.Context) (*big.Int, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	chainID, err := wrapper.ChainID(context.Background())
	assert.Nil(t, err)
	assert.Nil(t, chainID)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_BlockNumber(t *testing.T) {
	t.Parallel()

	t.Run("block number call returns error", func(t *testing.T) {
		args, statusHandler := createMockArgsEthereumChainWrapper()
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

		wrapper, _ := NewEthereumChainWrapper(args)
		blockNum, err := wrapper.BlockNumber(context.Background())
		assert.Equal(t, expectedError, err)
		assert.Equal(t, uint64(0), blockNum)
		assert.True(t, handlerCalled)
		assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
		assert.Equal(t, lastBlockNum, statusHandler.GetIntMetric(core.MetricLastQueriedEthereumBlockNumber))
	})
	t.Run("block number call returns a value", func(t *testing.T) {
		args, statusHandler := createMockArgsEthereumChainWrapper()
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

		wrapper, _ := NewEthereumChainWrapper(args)
		blockNum, err := wrapper.BlockNumber(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, uint64(newBlockNum), blockNum)
		assert.True(t, handlerCalled)
		assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
		assert.Equal(t, newBlockNum, statusHandler.GetIntMetric(core.MetricLastQueriedEthereumBlockNumber))
	})
}

func TestEthClientWrapper_NonceAt(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	args.BlockchainClient = &interactors.BlockchainClientStub{
		NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
			handlerCalled = true
			return 0, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	nonce, err := wrapper.NonceAt(context.Background(), common.Address{}, nil)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), nonce)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_ExecuteTransfer(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	args.MultiSigContract = &bridgeTests.MultiSigContractStub{
		ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address,
			amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {

			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	tx, err := wrapper.ExecuteTransfer(nil, nil, nil, nil, nil, nil, nil)
	assert.Nil(t, err)
	assert.Nil(t, tx)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientTransactions))
}

func TestEthClientWrapper_Quorum(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	args.MultiSigContract = &bridgeTests.MultiSigContractStub{
		QuorumCalled: func(opts *bind.CallOpts) (*big.Int, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	tx, err := wrapper.Quorum(context.Background())
	assert.Nil(t, err)
	assert.Nil(t, tx)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthClientWrapper_GetStatusesAfterExecution(t *testing.T) {
	t.Parallel()

	args, statusHandler := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	args.MultiSigContract = &bridgeTests.MultiSigContractStub{
		GetStatusesAfterExecutionCalled: func(opts *bind.CallOpts, batchNonceMultiversXETH *big.Int) ([]uint8, error) {
			handlerCalled = true
			return nil, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	statuses, err := wrapper.GetStatusesAfterExecution(context.Background(), nil)
	assert.Nil(t, err)
	assert.Nil(t, statuses)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, statusHandler.GetIntMetric(core.MetricNumEthClientRequests))
}

func TestEthereumChainWrapper_IsPaused(t *testing.T) {
	t.Parallel()

	args, _ := createMockArgsEthereumChainWrapper()
	handlerCalled := false
	args.MultiSigContract = &bridgeTests.MultiSigContractStub{
		PausedCalled: func(opts *bind.CallOpts) (bool, error) {
			handlerCalled = true
			return true, nil
		},
	}
	wrapper, _ := NewEthereumChainWrapper(args)
	result, err := wrapper.IsPaused(context.Background())

	assert.Nil(t, err)
	assert.True(t, result)
	assert.True(t, handlerCalled)
}

func TestEthereumChainWrapper_FilterLogs(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("expected error")
	args, _ := createMockArgsEthereumChainWrapper()

	t.Run("returns error", func(t *testing.T) {
		args.BlockchainClient = &interactors.BlockchainClientStub{
			FilterLogsCalled: func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
				return nil, expectedError
			},
		}

		wrapper, _ := NewEthereumChainWrapper(args)

		logs, err := wrapper.FilterLogs(context.Background(), ethereum.FilterQuery{})

		assert.Nil(t, logs)
		assert.Equal(t, expectedError, err)
	})

	t.Run("returns expected logs", func(t *testing.T) {
		expectedLogs := []types.Log{
			{
				Index: 1,
			},
		}
		args.BlockchainClient = &interactors.BlockchainClientStub{
			FilterLogsCalled: func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
				return expectedLogs, nil
			},
		}

		wrapper, _ := NewEthereumChainWrapper(args)
		logs, err := wrapper.FilterLogs(context.Background(), ethereum.FilterQuery{})

		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(expectedLogs, logs))
	})

}
