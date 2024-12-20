package multiversx

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	testCrypto "github.com/multiversx/mx-bridge-eth-go/testsCommon/crypto"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

var testCodec = &parsers.MultiversxCodec{}

func createMockArgsScCallExecutor() ArgsScCallExecutor {
	return ArgsScCallExecutor{
		ScProxyBech32Addresses: []string{
			"erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e",
		},
		Proxy:                           &interactors.ProxyStub{},
		Codec:                           &testsCommon.MultiversxCodecStub{},
		Filter:                          &testsCommon.ScCallsExecuteFilterStub{},
		Log:                             &testsCommon.LoggerStub{},
		ExtraGasToExecute:               100,
		MaxGasLimitToUse:                minGasToExecuteSCCalls,
		GasLimitForOutOfGasTransactions: minGasToExecuteSCCalls,
		NonceTxHandler:                  &testsCommon.TxNonceHandlerV2Stub{},
		PrivateKey:                      testCrypto.NewPrivateKeyMock(),
		SingleSigner:                    &testCrypto.SingleSignerStub{},
		CloseAppChan:                    make(chan struct{}),
	}
}

func createMockCheckConfigs() config.TransactionChecksConfig {
	return config.TransactionChecksConfig{
		CheckTransactionResults:    true,
		TimeInSecondsBetweenChecks: 6,
		ExecutionTimeoutInSeconds:  120,
		CloseAppOnError:            true,
		ExtraDelayInSecondsOnError: 120,
	}
}

func createTestProxySCCompleteCallData(token string) bridgeCore.ProxySCCompleteCallData {
	callData := bridgeCore.ProxySCCompleteCallData{
		RawCallData: testCodec.EncodeCallDataWithLenAndMarker(
			bridgeCore.CallData{
				Type:      1,
				Function:  "callMe",
				GasLimit:  5000000,
				Arguments: []string{"arg1", "arg2"},
			}),
		From:   common.Address{},
		Token:  token,
		Amount: big.NewInt(37),
		Nonce:  1,
	}
	callData.To, _ = data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqnf2w270lhxhlj57jvthxw4tqsunrwnq0anaqm4d4fn")

	return callData
}

func TestNewScCallExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.Proxy = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilProxy, err)
	})
	t.Run("nil codec should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.Codec = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilCodec, err)
	})
	t.Run("nil filter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.Filter = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilFilter, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.Log = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilLogger, err)
	})
	t.Run("nil nonce tx handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.NonceTxHandler = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilNonceTxHandler, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.PrivateKey = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilPrivateKey, err)
	})
	t.Run("nil single signer should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.SingleSigner = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilSingleSigner, err)
	})
	t.Run("empty list of sc proxy bech32 addresses should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ScProxyBech32Addresses = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errEmptyListOfBridgeSCProxy, err)
	})
	t.Run("invalid sc proxy bech32 address should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ScProxyBech32Addresses = append(args.ScProxyBech32Addresses, "not a valid bech32 address")

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.NotNil(t, err)
	})
	t.Run("invalid value for TimeInSecondsBetweenChecks should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 0

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errInvalidValue)
		assert.Contains(t, err.Error(), "for TransactionChecks.TimeInSecondsBetweenChecks, minimum: 1, got: 0")
	})
	t.Run("invalid value for ExecutionTimeoutInSeconds should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.ExecutionTimeoutInSeconds = 0

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errInvalidValue)
		assert.Contains(t, err.Error(), "for TransactionChecks.ExecutionTimeoutInSeconds, minimum: 1, got: 0")
	})
	t.Run("nil close app chan should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.TransactionChecks = createMockCheckConfigs()
		args.CloseAppChan = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errNilCloseAppChannel)
		assert.Contains(t, err.Error(), "while the TransactionChecks.CloseAppOnError is set to true")
	})
	t.Run("invalid MaxGasLimitToUse should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.TransactionChecks = createMockCheckConfigs()
		args.MaxGasLimitToUse = minGasToExecuteSCCalls - 1

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errGasLimitIsLessThanAbsoluteMinimum)
		assert.Contains(t, err.Error(), "provided: 2009999, absolute minimum required: 2010000")
		assert.Contains(t, err.Error(), "MaxGasLimitToUse")
	})
	t.Run("invalid GasLimitForOutOfGasTransactions should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.TransactionChecks = createMockCheckConfigs()
		args.GasLimitForOutOfGasTransactions = minGasToExecuteSCCalls - 1

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errGasLimitIsLessThanAbsoluteMinimum)
		assert.Contains(t, err.Error(), "provided: 2009999, absolute minimum required: 2010000")
		assert.Contains(t, err.Error(), "GasLimitForOutOfGasTransactions")
	})
	t.Run("should work without transaction checks", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()

		executor, err := NewScCallExecutor(args)
		assert.NotNil(t, executor)
		assert.Nil(t, err)
	})
	t.Run("should work with transaction checks", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.TransactionChecks = createMockCheckConfigs()

		executor, err := NewScCallExecutor(args)
		assert.NotNil(t, executor)
		assert.Nil(t, err)
	})
}

func TestScCallExecutor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *scCallExecutor
	assert.True(t, instance.IsInterfaceNil())

	instance = &scCallExecutor{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestScCallExecutor_Execute(t *testing.T) {
	t.Parallel()

	runError := errors.New("run error")
	expectedError := errors.New("expected error")

	argsForErrors := createMockArgsScCallExecutor()
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

	t.Run("get pending errors, should error", func(t *testing.T) {
		t.Parallel()

		args := argsForErrors // value copy
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedError
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
		assert.Contains(t, err.Error(), "errors found during execution")
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("get pending returns a not ok status, should error", func(t *testing.T) {
		t.Parallel()

		args := argsForErrors // value copy
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: "NOT OK",
					},
				}, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "got response code 'NOT OK'")
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("get pending returns an odd number of lines, should error", func(t *testing.T) {
		t.Parallel()

		args := argsForErrors // value copy
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
						},
					},
				}, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), errInvalidNumberOfResponseLines.Error())
		assert.Contains(t, err.Error(), "errors found during execution")
		assert.Contains(t, err.Error(), "expected an even number, got 1")
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("decoder errors, should error", func(t *testing.T) {
		t.Parallel()

		args := argsForErrors // value copy
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							{0x03, 0x04},
						},
					},
				}, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return bridgeCore.ProxySCCompleteCallData{
					To: data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32)),
				}, expectedError
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
		assert.Contains(t, err.Error(), "errors found during execution")
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("get network configs errors, should error", func(t *testing.T) {
		t.Parallel()

		args := argsForErrors // value copy
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							{0x03, 0x04},
						},
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return nil, expectedError
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return bridgeCore.ProxySCCompleteCallData{
					To: data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32)),
				}, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
		assert.Contains(t, err.Error(), "errors found during execution")
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("ApplyNonceAndGasPrice errors, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
				return expectedError
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should have not called SendTransactionCalled")
				return "", runError
			},
		}
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							{0x03, 0x04},
						},
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{}, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return bridgeCore.ProxySCCompleteCallData{
					To: data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32)),
				}, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
		assert.Contains(t, err.Error(), "errors found during execution")
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("Sign errors, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
				return nil
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should have not called SendTransactionCalled")
				return "", runError
			},
		}
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							{0x03, 0x04},
						},
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{}, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return bridgeCore.ProxySCCompleteCallData{
					To: data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32)),
				}, nil
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return nil, expectedError
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
		assert.Contains(t, err.Error(), "errors found during execution")
		assert.Zero(t, executor.GetNumSentTransaction())
	})
	t.Run("SendTransaction errors, should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
				return nil
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				return "", expectedError
			},
		}
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							{0x03, 0x04},
						},
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{}, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return bridgeCore.ProxySCCompleteCallData{
					RawCallData: []byte("dummy"),
					To:          data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32)),
				}, nil
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				assert.Equal(t, "dummy", string(buff))
				return 1000000, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
		assert.Contains(t, err.Error(), "errors found during execution")
		assert.Equal(t, uint32(0), executor.GetNumSentTransaction())
	})
	t.Run("should work with one SC proxy address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.MaxGasLimitToUse = 250000000
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1
		txHash := "tx hash"
		processTransactionStatusCalled := false

		nonceCounter := uint64(100)
		sendWasCalled := false

		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, args.ScProxyBech32Addresses[0], vmRequest.Address)
				assert.Equal(t, getPendingTransactionsFunction, vmRequest.FuncName)

				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							[]byte("ProxySCCompleteCallData 1"),
							{0x02},
							[]byte("ProxySCCompleteCallData 2"),
						},
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					ChainID:               "TEST",
					MinTransactionVersion: 111,
				}, nil
			},
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				assert.Equal(t, txHash, hexTxHash)
				processTransactionStatusCalled = true

				return transaction.TxStatusSuccess, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				if string(buff) == "ProxySCCompleteCallData 1" {
					return createTestProxySCCompleteCallData("tkn1"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 2" {
					return createTestProxySCCompleteCallData("tkn2"), nil
				}

				return bridgeCore.ProxySCCompleteCallData{
					To: data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32)),
				}, errors.New("wrong buffer")
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				return 5000000, nil
			},
		}
		args.Filter = &testsCommon.ScCallsExecuteFilterStub{
			ShouldExecuteCalled: func(callData bridgeCore.ProxySCCompleteCallData) bool {
				return callData.Token == "tkn2"
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
				assert.Equal(t, args.ExtraGasToExecute+5000000, tx.GasLimit)
				assert.Equal(t, nonceCounter-1, tx.Nonce)
				assert.Equal(t, uint64(101010), tx.GasPrice)
				assert.Equal(t, hex.EncodeToString([]byte("sig")), tx.Signature)
				_, err := data.NewAddressFromBech32String(tx.Sender)
				assert.Nil(t, err)
				assert.Equal(t, "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e", tx.Receiver)
				assert.Equal(t, "0", tx.Value)

				// only the second pending operation got through the filter
				expectedData := scProxyCallFunction + "@02"
				assert.Equal(t, expectedData, string(tx.Data))

				sendWasCalled = true

				return txHash, nil
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return []byte("sig"), nil
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)
		assert.Equal(t, uint32(1), executor.GetNumSentTransaction())
		assert.True(t, processTransactionStatusCalled)
	})
	t.Run("should work with one two proxy address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		pk := args.PrivateKey.GeneratePublic()
		pkBuff, _ := pk.ToByteArray()
		sender := data.NewAddressFromBytes(pkBuff)
		senderAddress, _ := sender.AddressAsBech32String()

		args.ScProxyBech32Addresses = append(args.ScProxyBech32Addresses, "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf")
		args.MaxGasLimitToUse = 250000000
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1
		txHash := "tx hash"
		numProcessTransactionStatusCalled := 0

		nonceCounter := uint64(100)

		sentTransactions := make([]*transaction.FrontendTransaction, 0)
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, getPendingTransactionsFunction, vmRequest.FuncName)

				returnData := make([][]byte, 4)
				switch vmRequest.Address {
				case args.ScProxyBech32Addresses[0]:
					returnData[0] = []byte{0x01}
					returnData[1] = []byte("ProxySCCompleteCallData 1")
					returnData[2] = []byte{0x02}
					returnData[3] = []byte("ProxySCCompleteCallData 2")
				case args.ScProxyBech32Addresses[1]:
					returnData[0] = []byte{0x03}
					returnData[1] = []byte("ProxySCCompleteCallData 3")
					returnData[2] = []byte{0x04}
					returnData[3] = []byte("ProxySCCompleteCallData 4")
				}
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: returnData,
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					ChainID:               "TEST",
					MinTransactionVersion: 111,
				}, nil
			},
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				assert.Contains(t, hexTxHash, txHash)
				numProcessTransactionStatusCalled++

				return transaction.TxStatusSuccess, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (parsers.ProxySCCompleteCallData, error) {
				if string(buff) == "ProxySCCompleteCallData 1" {
					return createTestProxySCCompleteCallData("tkn1"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 2" {
					return createTestProxySCCompleteCallData("tkn2"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 3" {
					return createTestProxySCCompleteCallData("tkn3"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 4" {
					return createTestProxySCCompleteCallData("tkn4"), nil
				}

				return parsers.ProxySCCompleteCallData{
					To: data.NewAddressFromBytes(bytes.Repeat([]byte{1}, 32)),
				}, errors.New("wrong buffer")
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				return 5000000, nil
			},
		}
		args.Filter = &testsCommon.ScCallsExecuteFilterStub{
			ShouldExecuteCalled: func(callData parsers.ProxySCCompleteCallData) bool {
				return callData.Token == "tkn2" || callData.Token == "tkn4"
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
				sentTransactions = append(sentTransactions, tx)

				return fmt.Sprintf("%s - %d", txHash, tx.Nonce), nil
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return []byte("sig"), nil
			},
		}

		expectedSentTransactions := []*transaction.FrontendTransaction{
			{
				Nonce:     100,
				Value:     "0",
				Receiver:  "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e",
				Sender:    senderAddress,
				GasPrice:  101010,
				GasLimit:  args.ExtraGasToExecute + 5000000,
				Data:      []byte(scProxyCallFunction + "@02"),
				Signature: hex.EncodeToString([]byte("sig")),
				ChainID:   "TEST",
				Version:   111,
			},
			{
				Nonce:     101,
				Value:     "0",
				Receiver:  "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
				Sender:    senderAddress,
				GasPrice:  101010,
				GasLimit:  args.ExtraGasToExecute + 5000000,
				Data:      []byte(scProxyCallFunction + "@04"),
				Signature: hex.EncodeToString([]byte("sig")),
				ChainID:   "TEST",
				Version:   111,
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, uint32(2), executor.GetNumSentTransaction())
		assert.Equal(t, expectedSentTransactions, sentTransactions)
		assert.Equal(t, 2, numProcessTransactionStatusCalled)
	})
	t.Run("should work even if the gas limit decode errors", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()

		nonceCounter := uint64(100)
		sendWasCalled := false

		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, args.ScProxyBech32Addresses[0], vmRequest.Address)
				assert.Equal(t, getPendingTransactionsFunction, vmRequest.FuncName)

				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							[]byte("ProxySCCompleteCallData 1"),
							{0x02},
							[]byte("ProxySCCompleteCallData 2"),
						},
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					ChainID:               "TEST",
					MinTransactionVersion: 111,
				}, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				if string(buff) == "ProxySCCompleteCallData 1" {
					return createTestProxySCCompleteCallData("tkn1"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 2" {
					return createTestProxySCCompleteCallData("tkn2"), nil
				}

				return bridgeCore.ProxySCCompleteCallData{}, errors.New("wrong buffer")
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				return 0, expectedError
			},
		}
		args.Filter = &testsCommon.ScCallsExecuteFilterStub{
			ShouldExecuteCalled: func(callData bridgeCore.ProxySCCompleteCallData) bool {
				return callData.Token == "tkn2"
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
				assert.Equal(t, args.ExtraGasToExecute, tx.GasLimit) // no 5000000 added gas limit because it wasn't extracted
				assert.Equal(t, nonceCounter-1, tx.Nonce)
				assert.Equal(t, uint64(101010), tx.GasPrice)
				assert.Equal(t, hex.EncodeToString([]byte("sig")), tx.Signature)
				_, err := data.NewAddressFromBech32String(tx.Sender)
				assert.Nil(t, err)
				assert.Equal(t, "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e", tx.Receiver)
				assert.Equal(t, "0", tx.Value)

				// only the second pending operation got through the filter
				expectedData := scProxyCallFunction + "@02"
				assert.Equal(t, expectedData, string(tx.Data))

				sendWasCalled = true

				return "", nil
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return []byte("sig"), nil
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)
		assert.Equal(t, uint32(1), executor.GetNumSentTransaction())
	})
	t.Run("should work if the gas limit is above the contract threshold", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()

		nonceCounter := uint64(100)
		sendWasCalled := false

		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, args.ScProxyBech32Addresses[0], vmRequest.Address)
				assert.Equal(t, getPendingTransactionsFunction, vmRequest.FuncName)

				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							[]byte("ProxySCCompleteCallData 1"),
							{0x02},
							[]byte("ProxySCCompleteCallData 2"),
						},
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					ChainID:               "TEST",
					MinTransactionVersion: 111,
				}, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				if string(buff) == "ProxySCCompleteCallData 1" {
					return createTestProxySCCompleteCallData("tkn1"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 2" {
					return createTestProxySCCompleteCallData("tkn2"), nil
				}

				return bridgeCore.ProxySCCompleteCallData{}, errors.New("wrong buffer")
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				return contractMaxGasLimit + 1, nil
			},
		}
		args.Filter = &testsCommon.ScCallsExecuteFilterStub{
			ShouldExecuteCalled: func(callData bridgeCore.ProxySCCompleteCallData) bool {
				return callData.Token == "tkn2"
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
				assert.Equal(t, args.GasLimitForOutOfGasTransactions, tx.GasLimit) // the gas limit was replaced
				assert.Equal(t, nonceCounter-1, tx.Nonce)
				assert.Equal(t, uint64(101010), tx.GasPrice)
				assert.Equal(t, hex.EncodeToString([]byte("sig")), tx.Signature)
				_, err := data.NewAddressFromBech32String(tx.Sender)
				assert.Nil(t, err)
				assert.Equal(t, "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e", tx.Receiver)
				assert.Equal(t, "0", tx.Value)

				// only the second pending operation got through the filter
				expectedData := scProxyCallFunction + "@02"
				assert.Equal(t, expectedData, string(tx.Data))

				sendWasCalled = true

				return "", nil
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return []byte("sig"), nil
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)
		assert.Equal(t, uint32(1), executor.GetNumSentTransaction())
	})
	t.Run("should skip execution if the gas limit exceeds the maximum allowed", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()

		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, args.ScProxyBech32Addresses[0], vmRequest.Address)
				assert.Equal(t, getPendingTransactionsFunction, vmRequest.FuncName)

				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{
							{0x01},
							[]byte("ProxySCCompleteCallData 1"),
							{0x02},
							[]byte("ProxySCCompleteCallData 2"),
						},
					},
				}, nil
			},
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					ChainID:               "TEST",
					MinTransactionVersion: 111,
				}, nil
			},
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
				if string(buff) == "ProxySCCompleteCallData 1" {
					return createTestProxySCCompleteCallData("tkn1"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 2" {
					return createTestProxySCCompleteCallData("tkn2"), nil
				}

				return bridgeCore.ProxySCCompleteCallData{}, errors.New("wrong buffer")
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				return args.MaxGasLimitToUse - args.ExtraGasToExecute + 1, nil
			},
		}
		args.Filter = &testsCommon.ScCallsExecuteFilterStub{
			ShouldExecuteCalled: func(callData bridgeCore.ProxySCCompleteCallData) bool {
				return callData.Token == "tkn2"
			},
		}
		args.NonceTxHandler = &testsCommon.TxNonceHandlerV2Stub{
			ApplyNonceAndGasPriceCalled: func(ctx context.Context, address core.AddressHandler, tx *transaction.FrontendTransaction) error {
				assert.Fail(t, "should have not apply nonce")
				return nil
			},
			SendTransactionCalled: func(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
				assert.Fail(t, "should have not called send")

				return "", nil
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return []byte("sig"), nil
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, uint32(0), executor.GetNumSentTransaction())
	})
}

func TestScCallExecutor_handleResults(t *testing.T) {
	t.Parallel()

	testHash := "test hash"
	t.Run("checkTransactionResults false should not check and return nil", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				assert.Fail(t, "should have not called ProcessTransactionStatusCalled")

				return transaction.TxStatusFail, nil
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.handleResults(context.Background(), testHash)
		assert.Nil(t, err)
	})
	t.Run("timeout before process transaction called", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				assert.Fail(t, "should have not called ProcessTransactionStatusCalled")

				return transaction.TxStatusFail, nil
			},
		}
		args.TransactionChecks = createMockCheckConfigs()

		executor, _ := NewScCallExecutor(args)

		workingCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := executor.handleResults(workingCtx, testHash)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("transaction not found should continuously request the status", func(t *testing.T) {
		t.Parallel()

		numRequests := uint64(0)
		args := createMockArgsScCallExecutor()
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
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1

		executor, _ := NewScCallExecutor(args)

		go func() {
			err := executor.handleResults(context.Background(), testHash)
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
		args := createMockArgsScCallExecutor()
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
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1

		executor, _ := NewScCallExecutor(args)

		go func() {
			err := executor.handleResults(context.Background(), testHash)
			assert.ErrorIs(t, err, context.DeadlineExceeded) // this will be the actual error when the function finishes
		}()

		select {
		case <-chDone:
			return
		case <-time.After(time.Second * 30):
			assert.Fail(t, "timeout")
		}
	})
	t.Run("error while requesting the status should return the error and wait", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsScCallExecutor()
		args.CloseAppChan = make(chan struct{}, 1)
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				return transaction.TxStatusInvalid, expectedErr
			},
		}
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1
		args.TransactionChecks.ExtraDelayInSecondsOnError = 6

		executor, _ := NewScCallExecutor(args)

		start := time.Now()
		err := executor.handleResults(context.Background(), testHash)
		assert.Equal(t, expectedErr, err)
		end := time.Now()

		assert.GreaterOrEqual(t, end.Sub(start), time.Second*6)
		select {
		case <-args.CloseAppChan:
		default:
			assert.Fail(t, "failed to write on the close app chan")
		}
	})
	t.Run("error while requesting the status should not write on the close app chan, if not enabled", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		args := createMockArgsScCallExecutor()
		args.CloseAppChan = make(chan struct{}, 1)
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				return transaction.TxStatusInvalid, expectedErr
			},
		}
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1
		args.TransactionChecks.ExtraDelayInSecondsOnError = 1
		args.TransactionChecks.CloseAppOnError = false

		executor, _ := NewScCallExecutor(args)

		err := executor.handleResults(context.Background(), testHash)
		assert.Equal(t, expectedErr, err)

		select {
		case <-args.CloseAppChan:
			assert.Fail(t, "should have not written on the close chan")
		default:
		}
	})
	t.Run("transaction failed, should get more info and signal error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.CloseAppChan = make(chan struct{}, 1)
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				return transaction.TxStatusFail, nil
			},
			GetTransactionInfoWithResultsCalled: func(ctx context.Context, txHash string) (*data.TransactionInfo, error) {
				return &data.TransactionInfo{}, nil
			},
		}
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1
		args.TransactionChecks.ExtraDelayInSecondsOnError = 1

		executor, _ := NewScCallExecutor(args)

		err := executor.handleResults(context.Background(), testHash)
		assert.ErrorIs(t, err, errTransactionFailed)

		select {
		case <-args.CloseAppChan:
		default:
			assert.Fail(t, "failed to write on the close app chan")
		}
	})
	t.Run("transaction failed, get more info fails, should signal error and not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, fmt.Sprintf("should have not panicked %v", r))
			}
		}()

		args := createMockArgsScCallExecutor()
		args.CloseAppChan = make(chan struct{}, 1)
		args.Proxy = &interactors.ProxyStub{
			ProcessTransactionStatusCalled: func(ctx context.Context, hexTxHash string) (transaction.TxStatus, error) {
				return transaction.TxStatusFail, nil
			},
			GetTransactionInfoWithResultsCalled: func(ctx context.Context, txHash string) (*data.TransactionInfo, error) {
				return nil, fmt.Errorf("random error")
			},
		}
		args.TransactionChecks = createMockCheckConfigs()
		args.TransactionChecks.TimeInSecondsBetweenChecks = 1
		args.TransactionChecks.ExtraDelayInSecondsOnError = 1

		executor, _ := NewScCallExecutor(args)

		err := executor.handleResults(context.Background(), testHash)
		assert.ErrorIs(t, err, errTransactionFailed)

		select {
		case <-args.CloseAppChan:
		default:
			assert.Fail(t, "failed to write on the close app chan")
		}
	})
}
