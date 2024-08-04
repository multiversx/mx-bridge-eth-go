package multiversx

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/common"
	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/roleProviders"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKeyGen = signing.NewKeyGenerator(ed25519.NewEd25519())
var pausedBytes = []byte{1}

func createMockClientArgs() ClientArgs {
	privateKey, _ := testKeyGen.PrivateKeyFromByteArray(bytes.Repeat([]byte{1}, 32))
	multisigContractAddress, _ := data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf")
	safeContractAddress, _ := data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqtvnswnzxxz8susupesys0hvg7q2z5nawrcjq06qdus")

	return ClientArgs{
		GasMapConfig: config.MultiversXGasMapConfig{
			Sign:                   10,
			ProposeTransferBase:    20,
			ProposeTransferForEach: 30,
			ProposeStatusBase:      40,
			ProposeStatusForEach:   50,
			PerformActionBase:      60,
			PerformActionForEach:   70,
			ScCallPerByte:          80,
			ScCallPerformForEach:   90,
		},
		Proxy:                        &interactors.ProxyStub{},
		Log:                          logger.GetOrCreate("test"),
		RelayerPrivateKey:            privateKey,
		MultisigContractAddress:      multisigContractAddress,
		SafeContractAddress:          safeContractAddress,
		IntervalToResendTxsInSeconds: 1,
		TokensMapper: &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("converted "), sourceBytes...), nil
			},
		},
		RoleProvider:  &roleproviders.MultiversXRoleProviderStub{},
		StatusHandler: &testsCommon.StatusHandlerStub{},
		AllowDelta:    5,
	}
}

func createMockPendingBatchBytes(numDeposits int) [][]byte {
	pendingBatchBytes := [][]byte{
		big.NewInt(44562).Bytes(),
	}

	generatorByte := byte(0)
	for i := 0; i < numDeposits; i++ {
		pendingBatchBytes = append(pendingBatchBytes, big.NewInt(int64(i)).Bytes())      // block nonce
		pendingBatchBytes = append(pendingBatchBytes, big.NewInt(int64(i+5000)).Bytes()) // deposit nonce

		generatorByte++
		pendingBatchBytes = append(pendingBatchBytes, bytes.Repeat([]byte{generatorByte}, 32)) // from

		generatorByte++
		pendingBatchBytes = append(pendingBatchBytes, bytes.Repeat([]byte{generatorByte}, 20)) // to

		generatorByte++
		pendingBatchBytes = append(pendingBatchBytes, bytes.Repeat([]byte{generatorByte}, 32)) // token

		pendingBatchBytes = append(pendingBatchBytes, big.NewInt(int64((i+1)*10000)).Bytes())
	}

	return pendingBatchBytes
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("nil proxy should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, errNilProxy, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.RelayerPrivateKey = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, clients.ErrNilPrivateKey, err)
	})
	t.Run("nil multisig contract address should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.MultisigContractAddress = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.True(t, errors.Is(err, errNilAddressHandler))
	})
	t.Run("nil safe contract address should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.SafeContractAddress = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.True(t, errors.Is(err, errNilAddressHandler))
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Log = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, clients.ErrNilLogger, err)
	})
	t.Run("nil tokens mapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.TokensMapper = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, clients.ErrNilTokensMapper, err)
	})
	t.Run("gas map invalid value should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.GasMapConfig.PerformActionForEach = 0

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.True(t, errors.Is(err, errInvalidGasValue))
		require.True(t, strings.Contains(err.Error(), "for field PerformActionForEach"))
	})
	t.Run("invalid interval to resend should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.IntervalToResendTxsInSeconds = 0

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "intervalToResend in NewNonceTransactionHandler"))
	})
	t.Run("nil role provider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.RoleProvider = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, errNilRoleProvider, err)
	})
	t.Run("nil status handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.StatusHandler = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, clients.ErrNilStatusHandler, err)
	})
	t.Run("invalid AllowDelta should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.AllowDelta = 0

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.True(t, errors.Is(err, clients.ErrInvalidValue))
		require.True(t, strings.Contains(err.Error(), "for args.AllowedDelta"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		c, err := NewClient(args)

		require.False(t, check.IfNil(c))
		require.Nil(t, err)
	})
}

func TestClient_GetPendingBatch(t *testing.T) {
	t.Parallel()

	t.Run("get pending batch failed should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		expectedErr := errors.New("expected error")
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		c, _ := NewClient(args)
		batch, err := c.GetPendingBatch(context.Background())
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("empty response", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))

		c, _ := NewClient(args)
		batch, err := c.GetPendingBatch(context.Background())
		assert.Nil(t, batch)
		assert.Equal(t, clients.ErrNoPendingBatchAvailable, err)
	})
	t.Run("invalid length", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		args.Proxy = createMockProxy(buff[:len(buff)-1])

		c, _ := NewClient(args)
		batch, err := c.GetPendingBatch(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errInvalidNumberOfArguments))
		assert.True(t, strings.Contains(err.Error(), "got 12 argument(s)"))

		args.Proxy = createMockProxy([][]byte{{1}})
		c, _ = NewClient(args)

		batch, err = c.GetPendingBatch(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errInvalidNumberOfArguments))
		assert.True(t, strings.Contains(err.Error(), "got 1 argument(s)"))
	})
	t.Run("invalid batch ID", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		buff[0] = bytes.Repeat([]byte{1}, 32)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetPendingBatch(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errNotUint64Bytes))
		assert.True(t, strings.Contains(err.Error(), "while parsing batch ID"))
	})
	t.Run("invalid deposit nonce", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		buff[8] = bytes.Repeat([]byte{1}, 32)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetPendingBatch(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errNotUint64Bytes))
		assert.True(t, strings.Contains(err.Error(), "while parsing the deposit nonce, transfer index 1"))
	})
	t.Run("tokens mapper errors", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		expectedErr := errors.New("expected error in convert tokens")
		args.TokensMapper = &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		buff := createMockPendingBatchBytes(2)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetPendingBatch(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, expectedErr))
		assert.True(t, strings.Contains(err.Error(), "while converting token bytes, transfer index 0"))
	})
	t.Run("should create pending batch", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.TokensMapper = &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("converted_"), sourceBytes...), nil
			},
		}
		args.Proxy = createMockProxy(createMockPendingBatchBytes(2))

		tokenBytes1 := bytes.Repeat([]byte{3}, 32)
		tokenBytes2 := bytes.Repeat([]byte{6}, 32)
		expectedBatch := &common.TransferBatch{
			ID: 44562,
			Deposits: []*common.DepositTransfer{
				{
					Nonce:                 5000,
					ToBytes:               bytes.Repeat([]byte{2}, 20),
					DisplayableTo:         "0x0202020202020202020202020202020202020202",
					FromBytes:             bytes.Repeat([]byte{1}, 32),
					DisplayableFrom:       "erd1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqsl6e0p7",
					SourceTokenBytes:      tokenBytes1,
					DestinationTokenBytes: append([]byte("converted_"), tokenBytes1...),
					DisplayableToken:      string(tokenBytes1),
					Amount:                big.NewInt(10000),
				},
				{
					Nonce:                 5001,
					ToBytes:               bytes.Repeat([]byte{5}, 20),
					DisplayableTo:         "0x0505050505050505050505050505050505050505",
					FromBytes:             bytes.Repeat([]byte{4}, 32),
					DisplayableFrom:       "erd1qszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqxjfvxn",
					SourceTokenBytes:      tokenBytes2,
					DestinationTokenBytes: append([]byte("converted_"), tokenBytes2...),
					DisplayableToken:      string(tokenBytes2),
					Amount:                big.NewInt(20000),
				},
			},
			Statuses: make([]byte, 2),
		}

		c, _ := NewClient(args)
		batch, err := c.GetPendingBatch(context.Background())
		assert.Nil(t, err)

		args.Log.Info("expected batch\n" + expectedBatch.String())
		args.Log.Info("batch\n" + batch.String())

		assert.Equal(t, expectedBatch, batch)
		assert.Nil(t, err)
	})
}

func TestClient_GetBatch(t *testing.T) {
	t.Parallel()

	t.Run("get batch failed should error", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		expectedErr := errors.New("expected error")
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		c, _ := NewClient(args)
		batch, err := c.GetBatch(context.Background(), 37)
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("empty response", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))

		c, _ := NewClient(args)
		batch, err := c.GetBatch(context.Background(), 37)
		assert.Nil(t, batch)
		assert.Equal(t, clients.ErrNoBatchAvailable, err)
	})
	t.Run("invalid length", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		args.Proxy = createMockProxy(buff[:len(buff)-1])

		c, _ := NewClient(args)
		batch, err := c.GetBatch(context.Background(), 37)

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errInvalidNumberOfArguments))
		assert.True(t, strings.Contains(err.Error(), "got 12 argument(s)"))

		args.Proxy = createMockProxy([][]byte{{1}})
		c, _ = NewClient(args)

		batch, err = c.GetBatch(context.Background(), 37)

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errInvalidNumberOfArguments))
		assert.True(t, strings.Contains(err.Error(), "got 1 argument(s)"))
	})
	t.Run("invalid batch ID", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		buff[0] = bytes.Repeat([]byte{1}, 32)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetBatch(context.Background(), 37)

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errNotUint64Bytes))
		assert.True(t, strings.Contains(err.Error(), "while parsing batch ID"))
	})
	t.Run("invalid deposit nonce", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		buff[8] = bytes.Repeat([]byte{1}, 32)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetBatch(context.Background(), 37)

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errNotUint64Bytes))
		assert.True(t, strings.Contains(err.Error(), "while parsing the deposit nonce, transfer index 1"))
	})
	t.Run("tokens mapper errors", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		expectedErr := errors.New("expected error in convert tokens")
		args.TokensMapper = &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		buff := createMockPendingBatchBytes(2)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetBatch(context.Background(), 37)

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, expectedErr))
		assert.True(t, strings.Contains(err.Error(), "while converting token bytes, transfer index 0"))
	})
	t.Run("should create pending batch", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.TokensMapper = &bridgeTests.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("converted_"), sourceBytes...), nil
			},
		}
		args.Proxy = createMockProxy(createMockPendingBatchBytes(2))

		tokenBytes1 := bytes.Repeat([]byte{3}, 32)
		tokenBytes2 := bytes.Repeat([]byte{6}, 32)
		expectedBatch := &common.TransferBatch{
			ID: 44562,
			Deposits: []*common.DepositTransfer{
				{
					Nonce:                 5000,
					ToBytes:               bytes.Repeat([]byte{2}, 20),
					DisplayableTo:         "0x0202020202020202020202020202020202020202",
					FromBytes:             bytes.Repeat([]byte{1}, 32),
					DisplayableFrom:       "erd1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqsl6e0p7",
					SourceTokenBytes:      tokenBytes1,
					DestinationTokenBytes: append([]byte("converted_"), tokenBytes1...),
					DisplayableToken:      string(tokenBytes1),
					Amount:                big.NewInt(10000),
				},
				{
					Nonce:                 5001,
					ToBytes:               bytes.Repeat([]byte{5}, 20),
					DisplayableTo:         "0x0505050505050505050505050505050505050505",
					FromBytes:             bytes.Repeat([]byte{4}, 32),
					DisplayableFrom:       "erd1qszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqxjfvxn",
					SourceTokenBytes:      tokenBytes2,
					DestinationTokenBytes: append([]byte("converted_"), tokenBytes2...),
					DisplayableToken:      string(tokenBytes2),
					Amount:                big.NewInt(20000),
				},
			},
			Statuses: make([]byte, 2),
		}

		c, _ := NewClient(args)
		batch, err := c.GetBatch(context.Background(), 44562)
		assert.Nil(t, err)

		args.Log.Info("expected batch\n" + expectedBatch.String())
		args.Log.Info("batch\n" + batch.String())

		assert.Equal(t, expectedBatch, batch)
		assert.Nil(t, err)
	})
}

func TestClient_ProposeSetStatus(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		c, _ := NewClient(args)

		hash, err := c.ProposeSetStatus(context.Background(), nil)
		assert.Empty(t, hash)
		assert.Equal(t, clients.ErrNilBatch, err)
	})
	t.Run("check is paused failed", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		expectedErr := errors.New("expected error")
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewClient(args)

		hash, err := c.ProposeSetStatus(context.Background(), &common.TransferBatch{})
		assert.Empty(t, hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("contract is paused", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{pausedBytes},
					},
				}, nil
			},
		}
		c, _ := NewClient(args)

		hash, err := c.ProposeSetStatus(context.Background(), &common.TransferBatch{})
		assert.Empty(t, hash)
		assert.True(t, errors.Is(err, clients.ErrMultisigContractPaused))
	})
	t.Run("should propose set status", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false
		c.txHandler = &bridgeTests.TxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
				sendWasCalled = true

				dataField, err := builder.ToDataString()
				assert.Nil(t, err)

				expectedArgs := []string{
					proposeSetStatusFuncName,
					hex.EncodeToString(big.NewInt(112233).Bytes()),
				}
				expectedStatus := []byte{clients.Rejected, clients.Executed}
				for _, stat := range expectedStatus {
					expectedArgs = append(expectedArgs, hex.EncodeToString([]byte{stat}))
				}

				expectedDataField := strings.Join(expectedArgs, "@")
				assert.Equal(t, expectedDataField, dataField)
				expectedGasLimit := c.gasMapConfig.ProposeStatusBase + uint64(len(expectedStatus))*c.gasMapConfig.ProposeStatusForEach
				assert.Equal(t, gasLimit, expectedGasLimit)

				return expectedHash, nil
			},
		}

		hash, err := c.ProposeSetStatus(context.Background(), createMockBatch())
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
		assert.True(t, sendWasCalled)
	})
}

func TestClient_ProposeTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		c, _ := NewClient(args)

		hash, err := c.ProposeTransfer(context.Background(), nil)
		assert.Empty(t, hash)
		assert.Equal(t, clients.ErrNilBatch, err)
	})
	t.Run("check is paused failed", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		expectedErr := errors.New("expected error")
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewClient(args)

		hash, err := c.ProposeTransfer(context.Background(), &common.TransferBatch{})
		assert.Empty(t, hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("contract is paused", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{pausedBytes},
					},
				}, nil
			},
		}
		c, _ := NewClient(args)

		hash, err := c.ProposeTransfer(context.Background(), &common.TransferBatch{})
		assert.Empty(t, hash)
		assert.True(t, errors.Is(err, clients.ErrMultisigContractPaused))
	})
	t.Run("should propose transfer", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false
		batch := createMockBatch()

		c.txHandler = &bridgeTests.TxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
				sendWasCalled = true

				dataField, err := builder.ToDataString()
				assert.Nil(t, err)

				dataStrings := []string{
					proposeTransferFuncName,
					hex.EncodeToString(big.NewInt(int64(batch.ID)).Bytes()),
				}
				depositsString := ""
				for _, dt := range batch.Deposits {
					depositsString = depositsString + depositToString(dt)
				}
				dataStrings = append(dataStrings, depositsString)

				expectedDataField := strings.Join(dataStrings, "@")
				assert.Equal(t, expectedDataField, dataField)

				expectedGasLimit := c.gasMapConfig.ProposeTransferBase + uint64(len(batch.Deposits))*c.gasMapConfig.ProposeTransferForEach
				assert.Equal(t, expectedGasLimit, gasLimit)

				return expectedHash, nil
			},
		}

		hash, err := c.ProposeTransfer(context.Background(), batch)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
		assert.True(t, sendWasCalled)
	})
	t.Run("should propose transfer with SC call", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false
		batch := createMockBatch()
		batch.Deposits[0].Data = bridgeTests.CallDataMock
		var err error
		batch.Deposits[0].DisplayableData, err = ethmultiversx.ConvertToDisplayableData(batch.Deposits[0].Data)
		require.Nil(t, err)

		c.txHandler = &bridgeTests.TxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
				sendWasCalled = true

				dataField, errConvert := builder.ToDataString()
				assert.Nil(t, errConvert)

				dataStrings := []string{
					proposeTransferFuncName,
					hex.EncodeToString(big.NewInt(int64(batch.ID)).Bytes()),
				}
				extraGas := uint64(0)
				depositsString := ""
				for _, dt := range batch.Deposits {
					depositsString = depositsString + depositToString(dt)
					if bytes.Equal(dt.Data, []byte{parsers.MissingDataProtocolMarker}) {
						continue
					}
					extraGas += (uint64(len(dt.Data))*2 + 1) * args.GasMapConfig.ScCallPerByte
				}
				dataStrings = append(dataStrings, depositsString)

				expectedDataField := strings.Join(dataStrings, "@")
				assert.Equal(t, expectedDataField, dataField)

				expectedGasLimit := c.gasMapConfig.ProposeTransferBase + uint64(len(batch.Deposits))*c.gasMapConfig.ProposeTransferForEach + extraGas
				assert.Equal(t, expectedGasLimit, gasLimit)

				return expectedHash, nil
			},
		}

		hash, err := c.ProposeTransfer(context.Background(), batch)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
		assert.True(t, sendWasCalled)
	})
}

func depositToString(dt *common.DepositTransfer) string {
	result := hex.EncodeToString(dt.FromBytes)
	result = result + hex.EncodeToString(dt.ToBytes)

	tokenLength := len(dt.DestinationTokenBytes)
	result = result + encodeLenAsHex(tokenLength) + hex.EncodeToString(dt.DestinationTokenBytes)

	amountLength := len(dt.Amount.Bytes())
	result = result + encodeLenAsHex(amountLength) + hex.EncodeToString(dt.Amount.Bytes())

	result = result + encodeUint64AsHex(dt.Nonce)
	result = result + hex.EncodeToString(dt.Data)

	return result
}

func encodeLenAsHex(length int) string {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, uint32(length))

	return hex.EncodeToString(buff)
}

func encodeUint64AsHex(value uint64) string {
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, value)

	return hex.EncodeToString(buff)
}

func TestClient_Sign(t *testing.T) {
	t.Parallel()

	actionID := uint64(662528)
	t.Run("check is paused failed", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		expectedErr := errors.New("expected error")
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewClient(args)

		hash, err := c.Sign(context.Background(), actionID)
		assert.Empty(t, hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("contract is paused", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{pausedBytes},
					},
				}, nil
			},
		}
		c, _ := NewClient(args)

		hash, err := c.Sign(context.Background(), actionID)
		assert.Empty(t, hash)
		assert.True(t, errors.Is(err, clients.ErrMultisigContractPaused))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false

		c.txHandler = &bridgeTests.TxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
				sendWasCalled = true

				dataField, err := builder.ToDataString()
				assert.Nil(t, err)

				expectedDataField := signFuncName + "@" + hex.EncodeToString(big.NewInt(int64(actionID)).Bytes())
				assert.Equal(t, expectedDataField, dataField)
				assert.Equal(t, c.gasMapConfig.Sign, gasLimit)

				return expectedHash, nil
			},
		}

		hash, err := c.Sign(context.Background(), actionID)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
		assert.True(t, sendWasCalled)
	})
}

func TestClient_PerformAction(t *testing.T) {
	t.Parallel()

	actionID := uint64(662528)
	t.Run("nil batch", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		c, _ := NewClient(args)

		hash, err := c.PerformAction(context.Background(), actionID, nil)
		assert.Empty(t, hash)
		assert.Equal(t, clients.ErrNilBatch, err)
	})
	t.Run("check is paused failed", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		expectedErr := errors.New("expected error")
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}
		c, _ := NewClient(args)

		hash, err := c.PerformAction(context.Background(), actionID, &common.TransferBatch{})
		assert.Empty(t, hash)
		assert.True(t, errors.Is(err, expectedErr))
	})
	t.Run("contract is paused", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = &interactors.ProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return &data.VmValuesResponseData{
					Data: &vm.VMOutputApi{
						ReturnCode: okCodeAfterExecution,
						ReturnData: [][]byte{pausedBytes},
					},
				}, nil
			},
		}
		c, _ := NewClient(args)

		hash, err := c.PerformAction(context.Background(), actionID, &common.TransferBatch{})
		assert.Empty(t, hash)
		assert.True(t, errors.Is(err, clients.ErrMultisigContractPaused))
	})
	t.Run("should perform action", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false
		batch := createMockBatch()

		c.txHandler = &bridgeTests.TxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
				sendWasCalled = true

				dataField, err := builder.ToDataString()
				assert.Nil(t, err)

				dataStrings := []string{
					performActionFuncName,
					hex.EncodeToString(big.NewInt(int64(actionID)).Bytes()),
				}
				expectedDataField := strings.Join(dataStrings, "@")
				assert.Equal(t, expectedDataField, dataField)
				expectedGasLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Statuses))*c.gasMapConfig.PerformActionForEach
				assert.Equal(t, expectedGasLimit, gasLimit)

				return expectedHash, nil
			},
		}

		hash, err := c.PerformAction(context.Background(), actionID, batch)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
		assert.True(t, sendWasCalled)
	})
	t.Run("should perform action with SC call", func(t *testing.T) {
		t.Parallel()

		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false
		batch := createMockBatch()
		batch.Deposits[0].Data = bridgeTests.CallDataMock
		var err error
		batch.Deposits[0].DisplayableData, err = ethmultiversx.ConvertToDisplayableData(batch.Deposits[0].Data)
		require.Nil(t, err)

		c.txHandler = &bridgeTests.TxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
				sendWasCalled = true

				dataField, errConvert := builder.ToDataString()
				assert.Nil(t, errConvert)

				dataStrings := []string{
					performActionFuncName,
					hex.EncodeToString(big.NewInt(int64(actionID)).Bytes()),
				}
				expectedDataField := strings.Join(dataStrings, "@")
				assert.Equal(t, expectedDataField, dataField)
				depositsString := ""

				extraGas := uint64(0)
				for _, dt := range batch.Deposits {
					depositsString = depositsString + depositToString(dt)
					if bytes.Equal(dt.Data, []byte{parsers.MissingDataProtocolMarker}) {
						continue
					}
					extraGas += (uint64(len(dt.Data))*2 + 1) * args.GasMapConfig.ScCallPerByte
					extraGas += args.GasMapConfig.ScCallPerformForEach
				}
				dataStrings = append(dataStrings, depositsString)

				expectedGasLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Statuses))*c.gasMapConfig.PerformActionForEach
				expectedGasLimit += extraGas
				assert.Equal(t, expectedGasLimit, gasLimit)

				return expectedHash, nil
			},
		}

		hash, err := c.PerformAction(context.Background(), actionID, batch)
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
		assert.True(t, sendWasCalled)
	})
}

func TestClient_Close(t *testing.T) {
	t.Parallel()

	args := createMockClientArgs()
	c, _ := NewClient(args)

	closeCalled := false
	c.txHandler = &bridgeTests.TxHandlerStub{
		CloseCalled: func() error {
			closeCalled = true
			return nil
		},
	}

	err := c.Close()

	assert.Nil(t, err)
	assert.True(t, closeCalled)
}

func TestClient_CheckClientAvailability(t *testing.T) {
	t.Parallel()

	currentNonce := uint64(0)
	incrementor := uint64(1)
	args := createMockClientArgs()
	statusHandler := testsCommon.NewStatusHandlerMock("test")
	args.StatusHandler = statusHandler
	expectedErr := errors.New("expected error")
	args.Proxy = &interactors.ProxyStub{
		GetShardOfAddressCalled: func(ctx context.Context, bech32Address string) (uint32, error) {
			return 0, nil
		},
		GetNetworkStatusCalled: func(ctx context.Context, shardID uint32) (*data.NetworkStatus, error) {
			currentNonce += incrementor
			return &data.NetworkStatus{
				Nonce: currentNonce,
			}, nil
		},
	}
	c, _ := NewClient(args)
	t.Run("different current nonce should update - 10 times", func(t *testing.T) {
		resetClient(c)
		for i := 0; i < 10; i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethmultiversx.Available, "")
		}
		assert.True(t, statusHandler.GetIntMetric(bridgeCore.MetricLastBlockNonce) > 0)
	})
	t.Run("same current nonce should error after a while", func(t *testing.T) {
		resetClient(c)
		_ = c.CheckClientAvailability(context.Background())

		incrementor = 0

		// place a random message as to test it is reset
		statusHandler.SetStringMetric(bridgeCore.MetricMultiversXClientStatus, ethmultiversx.ClientStatus(3).String())
		statusHandler.SetStringMetric(bridgeCore.MetricLastMultiversXClientError, "random")

		// this will just increment the retry counter
		for i := 0; i < int(args.AllowDelta); i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethmultiversx.Available, "")
		}

		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("nonce %d fetched for %d times in a row", currentNonce, args.AllowDelta+uint64(i+1))
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethmultiversx.Unavailable, message)
		}
	})
	t.Run("same current nonce should error after a while and then recovers", func(t *testing.T) {
		resetClient(c)
		_ = c.CheckClientAvailability(context.Background())

		incrementor = 0

		// this will just increment the retry counter
		for i := 0; i < int(args.AllowDelta); i++ {
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethmultiversx.Available, "")
		}

		for i := 0; i < 10; i++ {
			message := fmt.Sprintf("nonce %d fetched for %d times in a row", currentNonce, args.AllowDelta+uint64(i+1))
			err := c.CheckClientAvailability(context.Background())
			assert.Nil(t, err)
			checkStatusHandler(t, statusHandler, ethmultiversx.Unavailable, message)
		}

		incrementor = 1

		err := c.CheckClientAvailability(context.Background())
		assert.Nil(t, err)
		checkStatusHandler(t, statusHandler, ethmultiversx.Available, "")
	})
	t.Run("get current nonce errors", func(t *testing.T) {
		resetClient(c)
		c.proxy = &interactors.ProxyStub{
			GetShardOfAddressCalled: func(ctx context.Context, bech32Address string) (uint32, error) {
				return 0, nil
			},
			GetNetworkStatusCalled: func(ctx context.Context, shardID uint32) (*data.NetworkStatus, error) {
				return nil, expectedErr
			},
		}

		err := c.CheckClientAvailability(context.Background())
		checkStatusHandler(t, statusHandler, ethmultiversx.Unavailable, expectedErr.Error())
		assert.Equal(t, expectedErr, err)
	})
}

func resetClient(c *client) {
	c.mut.Lock()
	c.retriesAvailabilityCheck = 0
	c.mut.Unlock()
	c.statusHandler.SetStringMetric(bridgeCore.MetricMultiversXClientStatus, "")
	c.statusHandler.SetStringMetric(bridgeCore.MetricLastMultiversXClientError, "")
	c.statusHandler.SetIntMetric(bridgeCore.MetricLastBlockNonce, 0)
}

func checkStatusHandler(t *testing.T, statusHandler *testsCommon.StatusHandlerMock, status ethmultiversx.ClientStatus, message string) {
	assert.Equal(t, status.String(), statusHandler.GetStringMetric(bridgeCore.MetricMultiversXClientStatus))
	assert.Equal(t, message, statusHandler.GetStringMetric(bridgeCore.MetricLastMultiversXClientError))
}
