package multiversx

import (
	"bytes"
	"context"
	"errors"
	"testing"

	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	cryptoMock "github.com/multiversx/mx-bridge-eth-go/testsCommon/crypto"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/interactors"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/roleProviders"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

var (
	testSigner          = &singlesig.Ed25519Signer{}
	skBytes             = bytes.Repeat([]byte{1}, 32)
	testMultisigAddress = "erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede"
	relayerAddress      = "erd132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqvjzv73"
)

func createTransactionHandlerWithMockComponents() *transactionHandler {
	sk, _ := testKeyGen.PrivateKeyFromByteArray(skBytes)
	pk := sk.GeneratePublic()
	pkBytes, _ := pk.ToByteArray()

	return &transactionHandler{
		proxy:                   &interactors.ProxyStub{},
		relayerAddress:          data.NewAddressFromBytes(pkBytes),
		multisigAddressAsBech32: testMultisigAddress,
		nonceTxHandler:          &bridgeTests.NonceTransactionsHandlerStub{},
		relayerPrivateKey:       sk,
		singleSigner:            testSigner,
		roleProvider:            &roleproviders.MultiversXRoleProviderStub{},
	}
}

func TestTransactionHandler_SendTransactionReturnHash(t *testing.T) {
	t.Parallel()

	builder := builders.NewTxDataBuilder().Function("function").ArgBytes([]byte("buff")).ArgInt64(22)
	gasLimit := uint64(2000000)

	t.Run("get network configs errors", func(t *testing.T) {
		expectedErr := errors.New("expected error in get network configs")
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHandlerInstance.proxy = &interactors.ProxyStub{
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return nil, expectedErr
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)
		assert.Empty(t, hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("get nonce errors", func(t *testing.T) {
		expectedErr := errors.New("expected error in get nonce")
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHandlerInstance.nonceTxHandler = &bridgeTests.NonceTransactionsHandlerStub{
			GetNonceCalled: func(ctx context.Context, address core.AddressHandler) (uint64, error) {
				return 0, expectedErr
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)
		assert.Empty(t, hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("builder errors", func(t *testing.T) {
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		erroredBuilder := builders.NewTxDataBuilder().ArgAddress(nil)

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), erroredBuilder, gasLimit)
		assert.Empty(t, hash)
		assert.NotNil(t, err)
		assert.Equal(t, "nil address handler in builder.checkAddress", err.Error())
	})
	t.Run("signer errors", func(t *testing.T) {
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		expectedErr := errors.New("expected error in single signer")
		txHandlerInstance.singleSigner = &cryptoMock.SingleSignerStub{
			SignCalled: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
				return nil, expectedErr
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)
		assert.Empty(t, hash)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("relayer not whitelisted", func(t *testing.T) {
		wasWhiteListedCalled := false
		wasSendTransactionCalled := false
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHandlerInstance.roleProvider = &roleproviders.MultiversXRoleProviderStub{
			IsWhitelistedCalled: func(address core.AddressHandler) bool {
				wasWhiteListedCalled = true
				return false
			},
		}
		txHandlerInstance.nonceTxHandler = &bridgeTests.NonceTransactionsHandlerStub{
			SendTransactionCalled: func(ctx context.Context, tx *data.Transaction) (string, error) {
				wasSendTransactionCalled = true
				return "", nil
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)
		assert.Empty(t, hash)
		assert.Equal(t, errRelayerNotWhitelisted, err)
		assert.True(t, wasWhiteListedCalled)
		assert.False(t, wasSendTransactionCalled)
	})
	t.Run("should work", func(t *testing.T) {
		nonce := uint64(55273)
		txHandlerInstance := createTransactionHandlerWithMockComponents()
		txHash := "tx hash"
		sendWasCalled := false

		chainID := "chain ID"
		minGasPrice := uint64(12234)
		minTxVersion := uint32(122)

		txHandlerInstance.proxy = &interactors.ProxyStub{
			GetNetworkConfigCalled: func(ctx context.Context) (*data.NetworkConfig, error) {
				return &data.NetworkConfig{
					ChainID:               chainID,
					MinGasPrice:           minGasPrice,
					MinTransactionVersion: minTxVersion,
				}, nil
			},
		}

		txHandlerInstance.nonceTxHandler = &bridgeTests.NonceTransactionsHandlerStub{
			GetNonceCalled: func(ctx context.Context, address core.AddressHandler) (uint64, error) {
				if address.AddressAsBech32String() == relayerAddress {
					return nonce, nil
				}

				return 0, errors.New("unexpected address to fetch the nonce")
			},
			SendTransactionCalled: func(ctx context.Context, tx *data.Transaction) (string, error) {
				sendWasCalled = true
				assert.Equal(t, relayerAddress, tx.SndAddr)
				assert.Equal(t, testMultisigAddress, tx.RcvAddr)
				assert.Equal(t, nonce, tx.Nonce)
				assert.Equal(t, "0", tx.Value)
				assert.Equal(t, "function@62756666@16", string(tx.Data))
				assert.Equal(t, "fdbd51691e8179da15b22b133ab7e2d9f67faef585f6f4d9859ae176e7b6c2d7bb7f930de753fb7f8a377cd460ff41b54f8cfb0c720f586fbbfbee680edb310b", tx.Signature)
				assert.Equal(t, chainID, tx.ChainID)
				assert.Equal(t, gasLimit, tx.GasLimit)
				assert.Equal(t, minGasPrice, tx.GasPrice)
				assert.Equal(t, minTxVersion, tx.Version)

				return txHash, nil
			},
		}

		hash, err := txHandlerInstance.SendTransactionReturnHash(context.Background(), builder, gasLimit)

		assert.Nil(t, err)
		assert.Equal(t, txHash, hash)
		assert.True(t, sendWasCalled)
	})
}
