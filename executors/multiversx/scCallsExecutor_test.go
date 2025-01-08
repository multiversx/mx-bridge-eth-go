package multiversx

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

var testCodec = &parsers.MultiversxCodec{}

func createMockArgsScCallExecutor() ArgsScCallExecutor {
	return ArgsScCallExecutor{
		ScProxyBech32Addresses: []string{
			"erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e",
		},
		Proxy:               &interactors.ProxyStub{},
		TransactionExecutor: &testsCommon.TransactionExecutorStub{},
		Codec:               &testsCommon.MultiversxCodecStub{},
		Filter:              &testsCommon.ScCallsExecuteFilterStub{},
		Log:                 &testsCommon.LoggerStub{},
		ExecutorConfig: config.ScCallsExecutorConfig{
			ExtraGasToExecute:               100,
			MaxGasLimitToUse:                minGasToExecuteSCCalls,
			GasLimitForOutOfGasTransactions: minGasToExecuteSCCalls,
			TTLForFailedRefundIdInSeconds:   1,
		},
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
	t.Run("nil transaction executor should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.TransactionExecutor = nil

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilTransactionExecutor, err)
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
	t.Run("invalid MaxGasLimitToUse should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ExecutorConfig.MaxGasLimitToUse = minGasToExecuteSCCalls - 1

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errGasLimitIsLessThanAbsoluteMinimum)
		assert.Contains(t, err.Error(), "provided: 2009999, absolute minimum required: 2010000")
		assert.Contains(t, err.Error(), "MaxGasLimitToUse")
	})
	t.Run("invalid GasLimitForOutOfGasTransactions should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ExecutorConfig.GasLimitForOutOfGasTransactions = minGasToExecuteSCCalls - 1

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errGasLimitIsLessThanAbsoluteMinimum)
		assert.Contains(t, err.Error(), "provided: 2009999, absolute minimum required: 2010000")
		assert.Contains(t, err.Error(), "GasLimitForOutOfGasTransactions")
	})
	t.Run("invalid TTLForFailedRefundID should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ExecutorConfig.TTLForFailedRefundIdInSeconds = 0

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.ErrorIs(t, err, errInvalidValue)
		assert.Contains(t, err.Error(), "provided: 0s, absolute minimum required: 1s")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()

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
	argsForErrors.TransactionExecutor = &testsCommon.TransactionExecutorStub{
		ExecuteTransactionCalled: func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error {
			assert.Fail(t, "should have not called ExecuteTransactionCalled")
			return runError
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
	})
	t.Run("get pending returns an invalid vm values response (nil and nil), should error", func(t *testing.T) {
		t.Parallel()

		args := argsForErrors // value copy
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "nil response data")
	})
	t.Run("get pending returns an invalid vm values response (nil data), should error", func(t *testing.T) {
		t.Parallel()

		args := argsForErrors // value copy
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{}, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "nil response data")
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
	})
	t.Run("SendTransaction errors, should error and register the failed id", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ExecutorConfig.TTLForFailedRefundIdInSeconds = 60
		numExecuted := 0
		args.TransactionExecutor = &testsCommon.TransactionExecutorStub{
			ExecuteTransactionCalled: func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error {
				numExecuted++
				return expectedError
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

		//re-run the same call, no errors should be found
		err = executor.Execute(context.Background())
		assert.Nil(t, err)

		//only one sent transaction should be
		assert.Equal(t, 1, numExecuted)
		assert.True(t, executor.isFailed(1))
	})
	t.Run("SendTransaction errors, should error and register the failed id, after TTL expires should call again", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ExecutorConfig.TTLForFailedRefundIdInSeconds = 2
		numExecuted := 0
		args.TransactionExecutor = &testsCommon.TransactionExecutorStub{
			ExecuteTransactionCalled: func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error {
				numExecuted++
				return expectedError
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

		//wait for TTL to expire
		time.Sleep(time.Second * 3)

		//re-run the same call, same error should be found
		err = executor.Execute(context.Background())
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
		assert.Contains(t, err.Error(), "errors found during execution")

		//only one sent transaction should be
		assert.Equal(t, 2, numExecuted)
		assert.True(t, executor.isFailed(1))
	})
	t.Run("should not execute transactions with high gas limit usage", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ExecutorConfig.MaxGasLimitToUse = 5000000

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
				return 5000000, nil
			},
		}
		args.TransactionExecutor = &testsCommon.TransactionExecutorStub{
			ExecuteTransactionCalled: func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error {
				assert.Fail(t, "should have not called execute transaction")

				return nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.Nil(t, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ExecutorConfig.MaxGasLimitToUse = 250000000
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
		args.TransactionExecutor = &testsCommon.TransactionExecutorStub{
			ExecuteTransactionCalled: func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error {
				assert.Equal(t, "TEST", networkConfig.ChainID)
				assert.Equal(t, uint32(111), networkConfig.MinTransactionVersion)
				assert.Equal(t, args.ExecutorConfig.ExtraGasToExecute+5000000, gasLimit)
				assert.Equal(t, "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e", receiver)
				assert.Equal(t, scCallTxType, transactionType)

				expectedData := scProxyCallFunction + "@02"
				assert.Equal(t, expectedData, string(dataBytes))

				sendWasCalled = true

				return nil
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)
	})
	t.Run("should work with one two proxy address", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ScProxyBech32Addresses = append(args.ScProxyBech32Addresses, "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf")
		args.ExecutorConfig.MaxGasLimitToUse = 250000000

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
		}
		args.Codec = &testsCommon.MultiversxCodecStub{
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (bridgeCore.ProxySCCompleteCallData, error) {
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
				return callData.Token == "tkn2" || callData.Token == "tkn4"
			},
		}

		type sentTxValues struct {
			receiver        string
			transactionType string
			gasLimit        uint64
			dataBytes       []byte
		}
		sentTransactions := make([]*sentTxValues, 0)
		args.TransactionExecutor = &testsCommon.TransactionExecutorStub{
			ExecuteTransactionCalled: func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error {
				tx := &sentTxValues{
					receiver:        receiver,
					transactionType: transactionType,
					gasLimit:        gasLimit,
					dataBytes:       dataBytes,
				}
				sentTransactions = append(sentTransactions, tx)

				return nil
			},
		}

		expectedSentTransactions := []*sentTxValues{
			{
				receiver:        "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e",
				transactionType: scCallTxType,
				gasLimit:        args.ExecutorConfig.ExtraGasToExecute + 5000000,
				dataBytes:       []byte(scProxyCallFunction + "@02"),
			},
			{
				receiver:        "erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf",
				transactionType: scCallTxType,
				gasLimit:        args.ExecutorConfig.ExtraGasToExecute + 5000000,
				dataBytes:       []byte(scProxyCallFunction + "@04"),
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedSentTransactions, sentTransactions)
	})
	t.Run("should work even if the gas limit decode errors", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
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
		args.TransactionExecutor = &testsCommon.TransactionExecutorStub{
			ExecuteTransactionCalled: func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error {
				assert.Equal(t, "TEST", networkConfig.ChainID)
				assert.Equal(t, uint32(111), networkConfig.MinTransactionVersion)
				assert.Equal(t, args.ExecutorConfig.ExtraGasToExecute, gasLimit)
				assert.Equal(t, "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e", receiver)
				assert.Equal(t, scCallTxType, transactionType)

				expectedData := scProxyCallFunction + "@02"
				assert.Equal(t, expectedData, string(dataBytes))

				sendWasCalled = true

				return nil
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)
	})
	t.Run("should work if the gas limit is above the contract threshold", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
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
		args.TransactionExecutor = &testsCommon.TransactionExecutorStub{
			ExecuteTransactionCalled: func(ctx context.Context, networkConfig *data.NetworkConfig, receiver string, transactionType string, gasLimit uint64, dataBytes []byte) error {
				assert.Equal(t, "TEST", networkConfig.ChainID)
				assert.Equal(t, uint32(111), networkConfig.MinTransactionVersion)
				assert.Equal(t, args.ExecutorConfig.GasLimitForOutOfGasTransactions, gasLimit)
				assert.Equal(t, "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e", receiver)
				assert.Equal(t, scCallTxType, transactionType)

				expectedData := scProxyCallFunction + "@02"
				assert.Equal(t, expectedData, string(dataBytes))

				sendWasCalled = true

				return nil
			},
		}

		executor, _ := NewScCallExecutor(args)

		err := executor.Execute(context.Background())
		assert.Nil(t, err)
		assert.True(t, sendWasCalled)
	})
}
