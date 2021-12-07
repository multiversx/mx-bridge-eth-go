package elrond

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/interactors"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKeyGen = signing.NewKeyGenerator(ed25519.NewEd25519())

func createMockClientArgs() ClientArgs {
	privateKey, _ := testKeyGen.PrivateKeyFromByteArray(bytes.Repeat([]byte{1}, 32))
	multisigContractAddress, _ := data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqzyuaqg3dl7rqlkudrsnm5ek0j3a97qevd8sszj0glf")

	return ClientArgs{
		GasMapConfig: config.ElrondGasMapConfig{
			Sign:                   10,
			ProposeTransferBase:    20,
			ProposeTransferForEach: 30,
			ProposeStatus:          40,
			PerformActionBase:      50,
			PerformActionForEach:   60,
		},
		Proxy:                        &interactors.ElrondProxyStub{},
		Log:                          logger.GetOrCreate("test"),
		RelayerPrivateKey:            privateKey,
		MultisigContractAddress:      multisigContractAddress,
		IntervalToResendTxsInSeconds: 1,
		TokensMapper: &bridgeV2.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("converted "), sourceBytes...), nil
			},
		},
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
		args := createMockClientArgs()
		args.Proxy = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, errNilProxy, err)
	})
	t.Run("nil private key should error", func(t *testing.T) {
		args := createMockClientArgs()
		args.RelayerPrivateKey = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, errNilPrivateKey, err)
	})
	t.Run("nil multisig contract address should error", func(t *testing.T) {
		args := createMockClientArgs()
		args.MultisigContractAddress = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.True(t, errors.Is(err, errNilAddressHandler))
	})
	t.Run("nil logger should error", func(t *testing.T) {
		args := createMockClientArgs()
		args.Log = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, errNilLogger, err)
	})
	t.Run("nil tokens mapper should error", func(t *testing.T) {
		args := createMockClientArgs()
		args.TokensMapper = nil

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.Equal(t, errNilTokensMapper, err)
	})
	t.Run("gas map invalid value should error", func(t *testing.T) {
		args := createMockClientArgs()
		args.GasMapConfig.PerformActionForEach = 0

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.True(t, errors.Is(err, errInvalidGasValue))
		require.True(t, strings.Contains(err.Error(), "for field PerformActionForEach"))
	})
	t.Run("invalid interval to resend should error", func(t *testing.T) {
		args := createMockClientArgs()
		args.IntervalToResendTxsInSeconds = 0

		c, err := NewClient(args)

		require.True(t, check.IfNil(c))
		require.NotNil(t, err)
		require.True(t, strings.Contains(err.Error(), "intervalToResend in NewNonceTransactionHandler"))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockClientArgs()
		c, err := NewClient(args)

		require.False(t, check.IfNil(c))
		require.Nil(t, err)
	})
}

func TestClient_GetPending(t *testing.T) {
	t.Parallel()

	t.Run("get pending batch failed should error", func(t *testing.T) {
		args := createMockClientArgs()
		expectedErr := errors.New("expected error")
		args.Proxy = &interactors.ElrondProxyStub{
			ExecuteVMQueryCalled: func(ctx context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
				return nil, expectedErr
			},
		}

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())
		assert.Nil(t, batch)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("empty response", func(t *testing.T) {
		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())
		assert.Nil(t, batch)
		assert.Equal(t, ErrNoPendingBatchAvailable, err)
	})
	t.Run("invalid length", func(t *testing.T) {
		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		args.Proxy = createMockProxy(buff[:len(buff)-1])

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errInvalidNumberOfArguments))
		assert.True(t, strings.Contains(err.Error(), "got 12 argument(s)"))

		args.Proxy = createMockProxy([][]byte{{1}})
		c, _ = NewClient(args)

		batch, err = c.GetPending(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errInvalidNumberOfArguments))
		assert.True(t, strings.Contains(err.Error(), "got 1 argument(s)"))
	})
	t.Run("invalid batch ID", func(t *testing.T) {
		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		buff[0] = bytes.Repeat([]byte{1}, 32)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errNotUint64Bytes))
		assert.True(t, strings.Contains(err.Error(), "while parsing batch ID"))
	})
	t.Run("invalid deposit nonce", func(t *testing.T) {
		args := createMockClientArgs()
		buff := createMockPendingBatchBytes(2)
		buff[8] = bytes.Repeat([]byte{1}, 32)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errNotUint64Bytes))
		assert.True(t, strings.Contains(err.Error(), "while parsing the deposit nonce, transfer index 1"))
	})
	t.Run("tokens mapper errors", func(t *testing.T) {
		args := createMockClientArgs()
		expectedErr := errors.New("expected error in convert tokens")
		args.TokensMapper = &bridgeV2.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}
		buff := createMockPendingBatchBytes(2)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, expectedErr))
		assert.True(t, strings.Contains(err.Error(), "while converting token bytes, transfer index 0"))
	})
	t.Run("should create pending batch", func(t *testing.T) {
		args := createMockClientArgs()
		args.TokensMapper = &bridgeV2.TokensMapperStub{
			ConvertTokenCalled: func(ctx context.Context, sourceBytes []byte) ([]byte, error) {
				return append([]byte("converted_"), sourceBytes...), nil
			},
		}
		args.Proxy = createMockProxy(createMockPendingBatchBytes(2))

		tokenBytes1 := bytes.Repeat([]byte{3}, 32)
		tokenBytes2 := bytes.Repeat([]byte{6}, 32)
		expectedBatch := &clients.TransferBatch{
			ID: 44562,
			Deposits: []*clients.DepositTransfer{
				{
					Nonce:               5000,
					ToBytes:             bytes.Repeat([]byte{2}, 20),
					DisplayableTo:       "0x0202020202020202020202020202020202020202",
					FromBytes:           bytes.Repeat([]byte{1}, 32),
					DisplayableFrom:     "erd1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqsl6e0p7",
					TokenBytes:          tokenBytes1,
					ConvertedTokenBytes: append([]byte("converted_"), tokenBytes1...),
					DisplayableToken:    "erd1qvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsh78jz5",
					Amount:              big.NewInt(10000),
				},
				{
					Nonce:               5001,
					ToBytes:             bytes.Repeat([]byte{5}, 20),
					DisplayableTo:       "0x0505050505050505050505050505050505050505",
					FromBytes:           bytes.Repeat([]byte{4}, 32),
					DisplayableFrom:     "erd1qszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqxjfvxn",
					TokenBytes:          tokenBytes2,
					ConvertedTokenBytes: append([]byte("converted_"), tokenBytes2...),
					DisplayableToken:    "erd1qcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqwkh39e",
					Amount:              big.NewInt(20000),
				},
			},
			Statuses: make([]byte, 2),
		}

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())
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
		args := createMockClientArgs()
		c, _ := NewClient(args)

		hash, err := c.ProposeSetStatus(context.Background(), nil)
		assert.Empty(t, hash)
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("should propose set status", func(t *testing.T) {
		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false
		c.txHandler = &bridgeV2.TxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
				sendWasCalled = true

				dataField, err := builder.ToDataString()
				assert.Nil(t, err)
				expectedDataField := proposeSetStatusFuncName + "@" + hex.EncodeToString(big.NewInt(112233).Bytes()) + "@" +
					hex.EncodeToString([]byte{clients.Rejected, clients.Executed})
				assert.Equal(t, expectedDataField, dataField)
				assert.Equal(t, c.gasMapConfig.ProposeStatus, gasLimit)

				return expectedHash, nil
			},
		}

		hash, err := c.ProposeSetStatus(context.Background(), createMockBatch())
		assert.Nil(t, err)
		assert.Equal(t, expectedHash, hash)
		assert.True(t, sendWasCalled)
	})
}

func TestClient_ResolveNewDeposits(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		args := createMockClientArgs()
		c, _ := NewClient(args)

		err := c.ResolveNewDeposits(context.Background(), nil)
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("no pending batch should error", func(t *testing.T) {
		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		c, _ := NewClient(args)

		batch := createMockBatch()

		err := c.ResolveNewDeposits(context.Background(), batch)
		assert.True(t, errors.Is(err, ErrNoPendingBatchAvailable))
		assert.True(t, strings.Contains(err.Error(), "while getting new batch in ResolveNewDeposits method"))
		assert.Equal(t, []byte{clients.Rejected, clients.Executed}, batch.Statuses)
	})
	t.Run("should add new statuses to the existing batch", func(t *testing.T) {
		args := createMockClientArgs()
		args.Proxy = createMockProxy(createMockPendingBatchBytes(3))
		c, _ := NewClient(args)

		batch := createMockBatch()

		err := c.ResolveNewDeposits(context.Background(), batch)
		assert.Nil(t, err)
		assert.Equal(t, []byte{clients.Rejected, clients.Executed, clients.Rejected}, batch.Statuses)
	})
}

func TestClient_ProposeTransfer(t *testing.T) {
	t.Parallel()

	t.Run("nil batch", func(t *testing.T) {
		args := createMockClientArgs()
		c, _ := NewClient(args)

		hash, err := c.ProposeTransfer(context.Background(), nil)
		assert.Empty(t, hash)
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("should propose transfer", func(t *testing.T) {
		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false
		batch := createMockBatch()

		c.txHandler = &bridgeV2.TxHandlerStub{
			SendTransactionReturnHashCalled: func(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
				sendWasCalled = true

				dataField, err := builder.ToDataString()
				assert.Nil(t, err)

				dataStrings := []string{
					proposeTransferFuncName,
					hex.EncodeToString(big.NewInt(int64(batch.ID)).Bytes()),
				}
				for _, dt := range batch.Deposits {
					dataStrings = append(dataStrings, depositToStrings(dt)...)
				}

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
}

func depositToStrings(dt *clients.DepositTransfer) []string {
	result := []string{
		hex.EncodeToString(dt.FromBytes),
		hex.EncodeToString(dt.ToBytes),
		hex.EncodeToString(dt.ConvertedTokenBytes),
		hex.EncodeToString(dt.Amount.Bytes()),
		hex.EncodeToString(big.NewInt(int64(dt.Nonce)).Bytes()),
	}

	return result
}

func TestClient_Sign(t *testing.T) {
	t.Parallel()

	args := createMockClientArgs()
	args.Proxy = createMockProxy(make([][]byte, 0))
	expectedHash := "expected hash"
	c, _ := NewClient(args)
	sendWasCalled := false
	actionID := uint64(662528)

	c.txHandler = &bridgeV2.TxHandlerStub{
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
}

func TestClient_PerformAction(t *testing.T) {
	t.Parallel()

	actionID := uint64(662528)
	t.Run("nil batch", func(t *testing.T) {
		args := createMockClientArgs()
		c, _ := NewClient(args)

		hash, err := c.PerformAction(context.Background(), actionID, nil)
		assert.Empty(t, hash)
		assert.Equal(t, errNilBatch, err)
	})
	t.Run("should perform action", func(t *testing.T) {
		args := createMockClientArgs()
		args.Proxy = createMockProxy(make([][]byte, 0))
		expectedHash := "expected hash"
		c, _ := NewClient(args)
		sendWasCalled := false
		batch := createMockBatch()

		c.txHandler = &bridgeV2.TxHandlerStub{
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
				expectedGasdLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Statuses))*c.gasMapConfig.PerformActionForEach
				assert.Equal(t, expectedGasdLimit, gasLimit)

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
	c.txHandler = &bridgeV2.TxHandlerStub{
		CloseCalled: func() error {
			closeCalled = true
			return nil
		},
	}

	err := c.Close()

	assert.Nil(t, err)
	assert.True(t, closeCalled)
}
