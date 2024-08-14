package multiversx

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

var testCodec = &testsCommon.TestMultiversXCodec{}

func createMockArgsScCallExecutor() ArgsScCallExecutor {
	return ArgsScCallExecutor{
		ScProxyBech32Address: "erd1qqqqqqqqqqqqqpgqk839entmk46ykukvhpn90g6knskju3dtanaq20f66e",
		Proxy:                &interactors.ProxyStub{},
		Codec:                &testsCommon.MultiversxCodecStub{},
		Filter:               &testsCommon.ScCallsExecuteFilterStub{},
		Log:                  &testsCommon.LoggerStub{},
		ExtraGasToExecute:    100,
		NonceTxHandler:       &testsCommon.TxNonceHandlerV2Stub{},
		PrivateKey:           testCrypto.NewPrivateKeyMock(),
		SingleSigner:         &testCrypto.SingleSignerStub{},
	}
}

func createTestProxySCCompleteCallData(token string) parsers.ProxySCCompleteCallData {
	callData := parsers.ProxySCCompleteCallData{
		RawCallData: testCodec.EncodeCallDataWithLenAndMarker(
			parsers.CallData{
				Type:      1,
				Function:  "callMe",
				GasLimit:  5000000,
				Arguments: []interface{}{"arg1", "arg2"},
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
	t.Run("invalid sc proxy bech32 address should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()
		args.ScProxyBech32Address = "not a valid bech32 address"

		executor, err := NewScCallExecutor(args)
		assert.Nil(t, executor)
		assert.NotNil(t, err)
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
		assert.Equal(t, expectedError, err)
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
		assert.ErrorIs(t, err, errInvalidNumberOfResponseLines)
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
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (parsers.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return parsers.ProxySCCompleteCallData{}, expectedError
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.ErrorIs(t, err, expectedError)
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
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (parsers.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return parsers.ProxySCCompleteCallData{}, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.ErrorIs(t, err, expectedError)
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
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (parsers.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return parsers.ProxySCCompleteCallData{}, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.ErrorIs(t, err, expectedError)
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
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (parsers.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return parsers.ProxySCCompleteCallData{}, nil
			},
		}
		args.SingleSigner = &testCrypto.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return nil, expectedError
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.ErrorIs(t, err, expectedError)
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
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (parsers.ProxySCCompleteCallData, error) {
				assert.Equal(t, []byte{0x03, 0x04}, buff)

				return parsers.ProxySCCompleteCallData{
					RawCallData: []byte("dummy"),
				}, nil
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				assert.Equal(t, "dummy", string(buff))
				return 1000000, nil
			},
		}

		executor, _ := NewScCallExecutor(args)
		err := executor.Execute(context.Background())
		assert.ErrorIs(t, err, expectedError)
		assert.Equal(t, uint32(0), executor.GetNumSentTransaction())
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()

		nonceCounter := uint64(100)
		sendWasCalled := false

		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, args.ScProxyBech32Address, vmRequest.Address)
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
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (parsers.ProxySCCompleteCallData, error) {
				if string(buff) == "ProxySCCompleteCallData 1" {
					return createTestProxySCCompleteCallData("tkn1"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 2" {
					return createTestProxySCCompleteCallData("tkn2"), nil
				}

				return parsers.ProxySCCompleteCallData{}, errors.New("wrong buffer")
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				return 5000000, nil
			},
		}
		args.Filter = &testsCommon.ScCallsExecuteFilterStub{
			ShouldExecuteCalled: func(callData parsers.ProxySCCompleteCallData) bool {
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

				// only the second pending operation gor through the filter
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
	t.Run("should work even if the gas limit decode errors", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsScCallExecutor()

		nonceCounter := uint64(100)
		sendWasCalled := false

		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				assert.Equal(t, args.ScProxyBech32Address, vmRequest.Address)
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
			DecodeProxySCCompleteCallDataCalled: func(buff []byte) (parsers.ProxySCCompleteCallData, error) {
				if string(buff) == "ProxySCCompleteCallData 1" {
					return createTestProxySCCompleteCallData("tkn1"), nil
				}
				if string(buff) == "ProxySCCompleteCallData 2" {
					return createTestProxySCCompleteCallData("tkn2"), nil
				}

				return parsers.ProxySCCompleteCallData{}, errors.New("wrong buffer")
			},
			ExtractGasLimitFromRawCallDataCalled: func(buff []byte) (uint64, error) {
				return 0, expectedError
			},
		}
		args.Filter = &testsCommon.ScCallsExecuteFilterStub{
			ShouldExecuteCalled: func(callData parsers.ProxySCCompleteCallData) bool {
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

				// only the second pending operation gor through the filter
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
}
