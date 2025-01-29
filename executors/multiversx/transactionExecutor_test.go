package multiversx

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	testCrypto "github.com/multiversx/mx-bridge-eth-go/testsCommon/crypto"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

const testAddress = "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e"

func createMockArgsTransactionExecutor() ArgsTransactionExecutor {
	return ArgsTransactionExecutor{
		Proxy:          &interactors.ProxyStub{},
		Log:            &testsCommon.LoggerStub{},
		NonceTxHandler: &testsCommon.TxNonceHandlerV2Stub{},
		PrivateKey:     testCrypto.NewPrivateKeyMock(),
		SingleSigner:   &testCrypto.SingleSignerStub{},
		TransactionChecks: config.TransactionChecksConfig{
			TimeInSecondsBetweenChecks: 6,
			ExecutionTimeoutInSeconds:  120,
		},
	}
}

func TestNewExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.Proxy = nil

		executor, err := NewTransactionExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilProxy, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.Log = nil

		executor, err := NewTransactionExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilLogger, err)
	})
	t.Run("nil nonce tx handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.NonceTxHandler = nil

		executor, err := NewTransactionExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilNonceTxHandler, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.PrivateKey = nil

		executor, err := NewTransactionExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilPrivateKey, err)
	})
	t.Run("nil single signer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.SingleSigner = nil

		executor, err := NewTransactionExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilSingleSigner, err)
	})
	t.Run("invalid value for TimeInSecondsBetweenChecks should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 0

		executor, err := NewTransactionExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errInvalidValue)
		assert.Contains(t, err.Error(), "for TransactionChecks.TimeInSecondsBetweenChecks, minimum: 1, got: 0")
	})
	t.Run("invalid value for ExecutionTimeoutInSeconds should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.TransactionChecks.ExecutionTimeoutInSeconds = 0

		executor, err := NewTransactionExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errInvalidValue)
		assert.Contains(t, err.Error(), "for TransactionChecks.ExecutionTimeoutInSeconds, minimum: 1, got: 0")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()

		executor, err := NewTransactionExecutor(args)
		assert.NotNil(t, executor)
		assert.Nil(t, err)
	})
}

func TestExecutor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *transactionExecutor
	assert.True(t, instance.IsInterfaceNil())

	instance = &transactionExecutor{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestTransactionExecutor_ExecuteTransaction(t *testing.T) {
	t.Parallel()

	runError := errors.New("run error")
	expectedError := errors.New("expected error")

	argsForErrors := createMockArgsTransactionExecutor()
	argsForErrors.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
		ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
			assert.Fail(t, "should have not called ApplyNonceAndGasPriceCalled")
			return runError
		},
		SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
			assert.Fail(t, "should have not called SendTransactionCalled")
			return "", runError
		},
	}

	t.Run("nil network configs, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()

		executor, _ := NewTransactionExecutor(args)
		err := executor.ExecuteTransaction(context.Background(), nil, testAddress, "test", 0, nil)
		assert.ErrorIs(t, err, builders.ErrNilNetworkConfig)
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("not a valid address, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()

		executor, _ := NewTransactionExecutor(args)
		err := executor.ExecuteTransaction(context.Background(), nil, "invalid address", "test", 0, nil)
		assert.ErrorIs(t, err, builders.ErrNilNetworkConfig)
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("ApplyNonceAndGasPrice errors, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
				return expectedError
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should have not called SendTransactionCalled")
				return "", runError
			},
		}

		executor, _ := NewTransactionExecutor(args)
		err := executor.ExecuteTransaction(context.Background(), &data.NetworkConfig{}, testAddress, "test", 0, nil)
		assert.ErrorIs(t, err, expectedError)
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("Sign errors, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
				return nil
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should have not called SendTransactionCalled")
				return "", runError
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return nil, expectedError
			},
		}

		executor, _ := NewTransactionExecutor(args)
		err := executor.ExecuteTransaction(context.Background(), &data.NetworkConfig{}, testAddress, "test", 0, nil)
		assert.ErrorIs(t, err, expectedError)
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("SendTransaction errors, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
				return nil
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				return "", expectedError
			},
		}

		executor, _ := NewTransactionExecutor(args)
		err := executor.ExecuteTransaction(context.Background(), &data.NetworkConfig{}, testAddress, "test", 0, nil)
		assert.ErrorIs(t, err, expectedError)
		assert.Equal(t, uint32(0), executor.GetNumSentTransaction())
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1
		txHash := "tx hash"
		processTransactionStatusCalled := false

		nonceCounter := uint64(100)
		sendWasCalled := false
		gasLimit := uint64(20000000)
		txData := "testTxData"
		configs := &data.NetworkConfig{
			ChainID:               "TEST",
			MinTransactionVersion: 111,
		}

		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				assert.Equal(t, txHash, hexTxHash)
				processTransactionStatusCalled = true

				return transaction.TxStatusSuccess, nil
			},
		}
		args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
				tx.Nonce = nonceCounter
				tx.GasPrice = 101010
				nonceCounter++
				return nil
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Equal(t, "TEST", tx.ChainID)
				assert.Equal(t, uint32(111), tx.Version)
				assert.Equal(t, gasLimit, tx.GasLimit)
				assert.Equal(t, nonceCounter-1, tx.Nonce)
				assert.Equal(t, uint64(101010), tx.GasPrice)
				assert.Equal(t, hex.EncodeToString([]byte("sig")), tx.Signature)
				_, err := data.NewAddressFromBech32String(tx.Sender)
				assert.Nil(t, err)
				assert.Equal(t, "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e", tx.Receiver)
				assert.Equal(t, "0", tx.Value)
				assert.Equal(t, txData, string(tx.Data))

				sendWasCalled = true

				return txHash, nil
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return []byte("sig"), nil
			},
		}

		executor, _ := NewTransactionExecutor(args)

		err := executor.ExecuteTransaction(context.Background(), configs, testAddress, "test", gasLimit, []byte(txData))
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)
		assert.Equal(t, uint32(1), executor.GetNumSentTransaction())
		assert.True(t, processTransactionStatusCalled)
	})
}

func TestTransactionExecutor_checkResultsUntilDone(t *testing.T) {
	t.Parallel()

	testHash := "test hash"
	t.Run("timeout before process transaction called", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				assert.Fail(t, "should have not called ProcessTransactionStatusCalled")

				return transaction.TxStatusFail, nil
			},
		}

		executor, _ := NewTransactionExecutor(args)

		workingCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := executor.checkResultsUntilDone(workingCtx, testHash)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("transaction not found should continuously request the status", func(t *testing.T) {
		t.Parallel()

		numRequests := uint64(0)
		args := createMockArgsTransactionExecutor()
		chDone := make(chan struct{}, 1)
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				atomic.AddUint64(&numRequests, 1)
				if atomic.LoadUint64(&numRequests) > 3 {
					chDone <- struct{}{}
				}

				return transaction.TxStatusInvalid, errors.New("transaction not found")
			},
		}
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1

		executor, _ := NewTransactionExecutor(args)

		go func() {
			err := executor.checkResultsUntilDone(context.Background(), testHash)
			assert.ErrorIs(t, err, context.DeadlineExceeded) // this will be the actual error when the function finishes
		}()

		select {
		case <-chDone:
			return
		case <-time.After(time.Second * 30):
			assert.Fail(t, "timeout")
		}
	})
	t.Run("transaction is still pending should continuously request the status", func(t *testing.T) {
		t.Parallel()

		numRequests := uint64(0)
		args := createMockArgsTransactionExecutor()
		chDone := make(chan struct{}, 1)
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				atomic.AddUint64(&numRequests, 1)
				if atomic.LoadUint64(&numRequests) > 3 {
					chDone <- struct{}{}
				}

				return transaction.TxStatusPending, nil
			},
		}
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1

		executor, _ := NewTransactionExecutor(args)

		go func() {
			err := executor.checkResultsUntilDone(context.Background(), testHash)
			assert.ErrorIs(t, err, context.DeadlineExceeded) // this will be the actual error when the function finishes
		}()

		select {
		case <-chDone:
			return
		case <-time.After(time.Second * 30):
			assert.Fail(t, "timeout")
		}
	})
	t.Run("error while requesting the status should return the error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsTransactionExecutor()
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				return transaction.TxStatusInvalid, expectedErr
			},
		}
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1

		executor, _ := NewTransactionExecutor(args)

		start := time.Now()
		err := executor.checkResultsUntilDone(context.Background(), testHash)
		assert.Equal(t, expectedErr, err)
		end := time.Now()

		assert.GreaterOrEqual(t, end.Sub(start), time.Second)
	})
	t.Run("transaction failed, should get more info and signal error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsTransactionExecutor()
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				return transaction.TxStatusFail, nil
			},
			GetTransactionInfoWithResultsCalled: func(ctx context.Context, txHash string) (*data.TransactionInfo, error) {
				return &data.TransactionInfo{}, nil
			},
		}
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1

		executor, _ := NewTransactionExecutor(args)

		err := executor.checkResultsUntilDone(context.Background(), testHash)
		assert.ErrorIs(t, err, errTransactionFailed)
	})
	t.Run("transaction failed, get more info fails, should signal error and not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
			}
		}()

		args := createMockArgsTransactionExecutor()
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				return transaction.TxStatusFail, nil
			},
			GetTransactionInfoWithResultsCalled: func(ctx context.Context, txHash string) (*data.TransactionInfo, error) {
				return nil, fmt.Errorf("random error")
			},
		}
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1

		executor, _ := NewTransactionExecutor(args)

		err := executor.checkResultsUntilDone(context.Background(), testHash)
		assert.ErrorIs(t, err, errTransactionFailed)
	})
}

func TestTransactionExecutor_ExecuteTransactionInParallelShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockArgsTransactionExecutor()
	args.TransactionChecks.TimeInSecondsBetweenChecks = 1

	nonceCounter := uint64(100)
	gasLimit := uint64(20000000)
	txData := "testTxData"
	configs := &data.NetworkConfig{
		ChainID:               "TEST",
		MinTransactionVersion: 111,
	}

	sentTransaction := make(map[uint64]struct{})

	args.Proxy = &interactors.ProxyStub{
		ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
			assert.Contains(t, hexTxHash, "tx hash ")

			time.Sleep(time.Second) // simulate a delay in processing

			return transaction.TxStatusSuccess, nil
		},
	}
	args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
		ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
			// do not use here a mutex. The executor should automatically protect the calls
			tx.Nonce = nonceCounter
			tx.GasPrice = 101010
			nonceCounter++

			return nil
		},
		SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
			// do not use here a mutex. The executor should automatically protect the calls
			sentTransaction[tx.Nonce] = struct{}{}

			return fmt.Sprintf("tx hash %d", tx.Nonce), nil
		},
	}
	args.SingleSigner = &testCrypto.SingleSignerStub{
		SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte("sig"), nil
		},
	}

	executor, _ := NewTransactionExecutor(args)

	numExecuteCalls := 100
	wg := sync.WaitGroup{}
	wg.Add(numExecuteCalls)
	for i := 0; i < numExecuteCalls; i++ {
		go func() {
			err := executor.ExecuteTransaction(context.Background(), configs, testAddress, "test", gasLimit, []byte(txData))
			assert.Nil(t, err)

			wg.Done()
		}()
	}

	wg.Wait()
}
