package elrond

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/interactors"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	logger "github.com/ElrondNetwork/elrond-go-logger"
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
	}
}

func createMockPendingBatchBytes() [][]byte {
	amount1 := big.NewInt(27846)
	amount2 := big.NewInt(28983)

	pendingBatchBytes := [][]byte{
		big.NewInt(44562).Bytes(),

		{0},                         // first transfer: block nonce
		{1},                         // first transfer: deposit nonce
		bytes.Repeat([]byte{1}, 32), // first transfer: from
		bytes.Repeat([]byte{2}, 20), // first transfer: to
		bytes.Repeat([]byte{3}, 32), // first transfer: token
		amount1.Bytes(),

		{2},                         // second transfer: block nonce
		{3},                         // second transfer: deposit nonce
		bytes.Repeat([]byte{4}, 32), // second transfer: from
		bytes.Repeat([]byte{5}, 20), // second transfer: to
		bytes.Repeat([]byte{6}, 32), // second transfer: token
		amount2.Bytes(),
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
		buff := createMockPendingBatchBytes()
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
		buff := createMockPendingBatchBytes()
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
		buff := createMockPendingBatchBytes()
		buff[8] = bytes.Repeat([]byte{1}, 32)
		args.Proxy = createMockProxy(buff)

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())

		assert.Nil(t, batch)
		assert.True(t, errors.Is(err, errNotUint64Bytes))
		assert.True(t, strings.Contains(err.Error(), "while parsing the deposit nonce, transfer index 1"))
	})
	t.Run("should create pending batch", func(t *testing.T) {
		args := createMockClientArgs()
		args.Proxy = createMockProxy(createMockPendingBatchBytes())

		expectedBatch := &clients.TransferBatch{
			ID: 44562,
			Deposits: []*clients.DepositTransfer{
				{
					Nonce:            1,
					ToBytes:          bytes.Repeat([]byte{2}, 20),
					DisplayableTo:    "0x0202020202020202020202020202020202020202",
					FromBytes:        bytes.Repeat([]byte{1}, 32),
					DisplayableFrom:  "erd1qyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqsl6e0p7",
					TokenBytes:       bytes.Repeat([]byte{3}, 32),
					DisplayableToken: "erd1qvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsh78jz5",
					Amount:           big.NewInt(27846),
				},
				{
					Nonce:            3,
					ToBytes:          bytes.Repeat([]byte{5}, 20),
					DisplayableTo:    "0x0505050505050505050505050505050505050505",
					FromBytes:        bytes.Repeat([]byte{4}, 32),
					DisplayableFrom:  "erd1qszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqgpqyqszqxjfvxn",
					TokenBytes:       bytes.Repeat([]byte{6}, 32),
					DisplayableToken: "erd1qcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqvpsxqcrqwkh39e",
					Amount:           big.NewInt(28983),
				},
			},
		}

		c, _ := NewClient(args)
		batch, err := c.GetPending(context.Background())

		args.Log.Info("expected batch\n" + expectedBatch.String())
		args.Log.Info("batch\n" + batch.String())

		assert.Equal(t, expectedBatch, batch)
		assert.Nil(t, err)
	})

}
