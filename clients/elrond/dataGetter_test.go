package elrond

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/interactors"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/stretchr/testify/assert"
)

const (
	returnCode     = "return code"
	returnMessage  = "return message"
	calledFunction = "called function"
)

var calledArgs = []string{"args1", "args2"}

func createMockArgsDataGetter() ArgsDataGetter {
	args := ArgsDataGetter{
		Proxy: &interactors.ElrondProxyStub{},
	}

	args.MultisigContractAddress, _ = data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf")
	args.RelayerAddress, _ = data.NewAddressFromBech32String("erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede")

	return args
}

func createMockProxy(returningBytes [][]byte) *interactors.ElrondProxyStub {
	return &interactors.ElrondProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
			return &data.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: returningBytes,
				},
			}, nil
		},
	}
}

func TestNewDataGetter(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy", func(t *testing.T) {
		args := createMockArgsDataGetter()
		args.Proxy = nil

		dg, err := NewDataGetter(args)
		assert.Equal(t, errNilProxy, err)
		assert.True(t, check.IfNil(dg))
	})
	t.Run("nil multisig contact address", func(t *testing.T) {
		args := createMockArgsDataGetter()
		args.MultisigContractAddress = nil

		dg, err := NewDataGetter(args)
		assert.True(t, errors.Is(err, errNilAddressHandler))
		assert.True(t, strings.Contains(err.Error(), "MultisigContractAddress"))
		assert.True(t, check.IfNil(dg))
	})
	t.Run("nil relayer address", func(t *testing.T) {
		args := createMockArgsDataGetter()
		args.RelayerAddress = nil

		dg, err := NewDataGetter(args)
		assert.True(t, errors.Is(err, errNilAddressHandler))
		assert.True(t, strings.Contains(err.Error(), "RelayerAddress"))
		assert.True(t, check.IfNil(dg))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsDataGetter()

		dg, err := NewDataGetter(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(dg))
	})
}

func TestDataGetter_ExecuteQueryReturningBytes(t *testing.T) {
	t.Parallel()

	args := createMockArgsDataGetter()
	t.Run("nil vm ", func(t *testing.T) {
		dg, _ := NewDataGetter(args)

		result, err := dg.ExecuteQueryReturningBytes(context.Background(), nil)
		assert.Nil(t, result)
		assert.Equal(t, errNilRequest, err)
	})
	t.Run("proxy errors", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		expectedErr := errors.New("expected error")
		dg.proxy = &interactors.ElrondProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		result, err := dg.ExecuteQueryReturningBytes(context.Background(), &data.VmValueRequest{})
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("return code not ok", func(t *testing.T) {
		dg, _ := NewDataGetter(args)

		expectedErr := NewQueryResponseError(returnCode, returnMessage, calledFunction, dg.multisigContractAddress.AddressAsBech32String(), calledArgs...)
		dg.proxy = &interactors.ElrondProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData:      nil,
						ReturnCode:      returnCode,
						ReturnMessage:   returnMessage,
						GasRemaining:    0,
						GasRefund:       nil,
						OutputAccounts:  nil,
						DeletedAccounts: nil,
						TouchedAccounts: nil,
						Logs:            nil,
					},
				}, nil
			},
		}

		request := &data.VmValueRequest{
			Address:    dg.multisigContractAddress.AddressAsBech32String(),
			FuncName:   calledFunction,
			CallerAddr: dg.relayerAddress.AddressAsBech32String(),
			CallValue:  "0",
			Args:       calledArgs,
		}

		result, err := dg.ExecuteQueryReturningBytes(context.Background(), request)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})
	t.Run("should work", func(t *testing.T) {
		dg, _ := NewDataGetter(args)

		retData := [][]byte{[]byte("response 1"), []byte("response 2")}
		dg.proxy = &interactors.ElrondProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnData:      retData,
						ReturnCode:      okCodeAfterExecution,
						ReturnMessage:   returnMessage,
						GasRemaining:    0,
						GasRefund:       nil,
						OutputAccounts:  nil,
						DeletedAccounts: nil,
						TouchedAccounts: nil,
						Logs:            nil,
					},
				}, nil
			},
		}

		request := &data.VmValueRequest{
			Address:    dg.multisigContractAddress.AddressAsBech32String(),
			FuncName:   calledFunction,
			CallerAddr: dg.relayerAddress.AddressAsBech32String(),
			CallValue:  "0",
			Args:       calledArgs,
		}

		result, err := dg.ExecuteQueryReturningBytes(context.Background(), request)
		assert.Nil(t, err)
		assert.Equal(t, retData, result)
	})
}

func TestDataGetter_ExecuteQueryReturningBool(t *testing.T) {
	t.Parallel()

	args := createMockArgsDataGetter()
	t.Run("nil request", func(t *testing.T) {
		dg, _ := NewDataGetter(args)

		result, err := dg.ExecuteQueryReturningBool(context.Background(), nil)
		assert.False(t, result)
		assert.Equal(t, errNilRequest, err)
	})
	t.Run("empty response", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		dg.proxy = createMockProxy(make([][]byte, 0))

		result, err := dg.ExecuteQueryReturningBool(context.Background(), &data.VmValueRequest{})
		assert.False(t, result)
		assert.Nil(t, err)
	})
	t.Run("empty byte slice on first element", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		dg.proxy = createMockProxy([][]byte{make([]byte, 0)})

		result, err := dg.ExecuteQueryReturningBool(context.Background(), &data.VmValueRequest{})
		assert.False(t, result)
		assert.Nil(t, err)
	})
	t.Run("not a bool result", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		dg.proxy = createMockProxy([][]byte{[]byte("random bytes")})

		expectedError := NewQueryResponseError(
			internalError,
			`error converting the received bytes to bool, strconv.ParseBool: parsing "114": invalid syntax`,
			"",
			"",
		)

		result, err := dg.ExecuteQueryReturningBool(context.Background(), &data.VmValueRequest{})
		assert.False(t, result)
		assert.Equal(t, expectedError, err)
	})
	t.Run("should work", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		dg.proxy = createMockProxy([][]byte{{1}})

		result, err := dg.ExecuteQueryReturningBool(context.Background(), &data.VmValueRequest{})
		assert.True(t, result)
		assert.Nil(t, err)

		dg.proxy = createMockProxy([][]byte{{0}})

		result, err = dg.ExecuteQueryReturningBool(context.Background(), &data.VmValueRequest{})
		assert.False(t, result)
		assert.Nil(t, err)
	})
}

func TestDataGetter_ExecuteQueryReturningUint64(t *testing.T) {
	t.Parallel()

	args := createMockArgsDataGetter()
	t.Run("nil request", func(t *testing.T) {
		dg, _ := NewDataGetter(args)

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), nil)
		assert.Zero(t, result)
		assert.Equal(t, errNilRequest, err)
	})
	t.Run("empty response", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		dg.proxy = createMockProxy(make([][]byte, 0))

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), &data.VmValueRequest{})
		assert.Zero(t, result)
		assert.Nil(t, err)
	})
	t.Run("empty byte slice on first element", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		dg.proxy = createMockProxy([][]byte{make([]byte, 0)})

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), &data.VmValueRequest{})
		assert.Zero(t, result)
		assert.Nil(t, err)
	})
	t.Run("large buffer", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		dg.proxy = createMockProxy([][]byte{[]byte("random bytes")})

		expectedError := NewQueryResponseError(
			internalError,
			errNotUint64Bytes.Error(),
			"",
			"",
		)

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), &data.VmValueRequest{})
		assert.Zero(t, result)
		assert.Equal(t, expectedError, err)
	})
	t.Run("should work", func(t *testing.T) {
		dg, _ := NewDataGetter(args)
		dg.proxy = createMockProxy([][]byte{{1}})

		result, err := dg.ExecuteQueryReturningUint64(context.Background(), &data.VmValueRequest{})
		assert.Equal(t, uint64(1), result)
		assert.Nil(t, err)

		dg.proxy = createMockProxy([][]byte{{0xFF, 0xFF}})

		result, err = dg.ExecuteQueryReturningUint64(context.Background(), &data.VmValueRequest{})
		assert.Equal(t, uint64(65535), result)
		assert.Nil(t, err)
	})
}

func TestDataGetter_GetCurrentBatchAsDataBytes(t *testing.T) {
	t.Parallel()

	args := createMockArgsDataGetter()
	returningBytes := [][]byte{[]byte("buff0"), []byte("buff1"), []byte("buff2")}
	args.Proxy = &interactors.ElrondProxyStub{
		ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
			assert.Equal(t, args.MultisigContractAddress.AddressAsBech32String(), vmRequest.Address)
			assert.Equal(t, args.RelayerAddress.AddressAsBech32String(), vmRequest.CallerAddr)
			assert.Equal(t, 0, len(vmRequest.CallValue))
			assert.Equal(t, getCurrentTxBatchFuncName, vmRequest.FuncName)

			return &data.VmValuesResponseData{
				Data: &vm.VMOutputApi{
					ReturnCode: okCodeAfterExecution,
					ReturnData: returningBytes,
				},
			}, nil
		},
	}
	dg, _ := NewDataGetter(args)

	result, err := dg.GetCurrentBatchAsDataBytes(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, returningBytes, result)
}
