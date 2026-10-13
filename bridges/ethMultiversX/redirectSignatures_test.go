package ethmultiversx

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-bridge-eth-go/core/batchRedirect"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// digests computed independently with viem, the same way recovery-5956/recover.mjs builds the message
const (
	viemRedirectedDigest = "0x2ed413d0cd051d34e31cc307f38e9d1070a1b7a13260f9cc72d6161a61425999"
	viemOriginalDigest   = "0x157d31c60c7b7cf8fe681882011e1c16f1c8fe459ae27bd9d02fdb039a5b4a9a"
	testRecoveryAddress  = "0x1234567890abcdef1234567890abcdef12345678"
)

func batch5956FromMultiversX() *bridgeCore.TransferBatch {
	from, _ := hex.DecodeString("30adcf1257a0c227befca62f5d073b90a6c9875ee7e80e48ec0c3bbd5f622d13")
	to := common.HexToAddress("0x717fb1435f5fa0cc1c7e541e5d92ef71342fcd41").Bytes()
	usdc := common.HexToAddress("0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d").Bytes()
	unit := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)
	deposit := func(nonce uint64, units int64) *bridgeCore.DepositTransfer {
		return &bridgeCore.DepositTransfer{
			Nonce:                 nonce,
			FromBytes:             from,
			ToBytes:               to,
			SourceTokenBytes:      []byte("BSCUSDC-887875"),
			DestinationTokenBytes: usdc,
			DisplayableToken:      "BSCUSDC-887875",
			Amount:                big.NewInt(0).Mul(big.NewInt(units), unit),
		}
	}

	return &bridgeCore.TransferBatch{
		ID:       5956,
		Deposits: []*bridgeCore.DepositTransfer{deposit(7717, 49_999), deposit(7718, 30_440)},
		Statuses: make([]byte, 2),
	}
}

func createRedirectExecutor(t *testing.T, recoveryAddress string, signatures map[string][][]byte) *bridgeExecutor {
	redirector, err := batchRedirect.NewRedirector(recoveryAddress)
	require.Nil(t, err)

	args := createMockExecutorArgs()
	args.BatchRedirector = redirector
	args.MultiversXClient = &bridgeTests.MultiversXClientStub{
		GetPendingBatchCalled: func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
			return batch5956FromMultiversX(), nil
		},
	}
	args.EthereumClient = &bridgeTests.EthereumClientStub{
		GenerateMessageHashCalled: func(batch *batchProcessor.ArgListsBatch, batchID uint64) (common.Hash, error) {
			return ethereum.GenerateMessageHash(batch, batchID)
		},
	}
	args.SignaturesHolder = &testsCommon.SignaturesHolderStub{
		SignaturesCalled: func(messageHash []byte) [][]byte {
			return signatures[common.BytesToHash(messageHash).Hex()]
		},
	}
	executor, err := NewBridgeExecutor(args)
	require.Nil(t, err)

	return executor
}

func TestBridgeExecutor_RedirectedBatchMatchesManualRecoveryMessage(t *testing.T) {
	t.Parallel()

	executor := createRedirectExecutor(t, testRecoveryAddress, nil)
	batch, err := executor.GetBatchFromMultiversX(context.Background())
	require.Nil(t, err)
	for _, dt := range batch.Deposits {
		assert.Equal(t, common.HexToAddress(testRecoveryAddress).Bytes(), dt.ToBytes)
	}

	require.Nil(t, executor.StoreBatchFromMultiversX(batch))
	require.Nil(t, executor.SignTransferOnEthereum())
	assert.Equal(t, viemRedirectedDigest, executor.msgHash.Hex())
	require.NotNil(t, executor.redirectSignatures)
	assert.Equal(t, viemOriginalDigest, executor.redirectSignatures.originalHash.Hex())

	argLists := batchProcessor.ExtractListMvxToEth(executor.batch)
	for _, recipient := range argLists.Recipients {
		assert.Equal(t, common.HexToAddress(testRecoveryAddress), recipient)
	}
}

func TestBridgeExecutor_RefusesRedirectedBatchWithoutRecoveryAddress(t *testing.T) {
	t.Parallel()

	executor := createRedirectExecutor(t, "", nil)
	batch, err := executor.GetBatchFromMultiversX(context.Background())
	assert.ErrorIs(t, err, batchRedirect.ErrRecoveryAddressNotSet)
	assert.Nil(t, batch)
}

func TestBridgeExecutor_RedirectSignatureStatsIdentifySigners(t *testing.T) {
	t.Parallel()

	patchedA, _ := crypto.GenerateKey()
	patchedB, _ := crypto.GenerateKey()
	unpatched, _ := crypto.GenerateKey()
	sign := func(digest string, key *ecdsa.PrivateKey) []byte {
		signature, err := crypto.Sign(common.HexToHash(digest).Bytes(), key)
		require.Nil(t, err)
		return signature
	}
	signatures := map[string][][]byte{
		viemRedirectedDigest: {sign(viemRedirectedDigest, patchedA), sign(viemRedirectedDigest, patchedB)},
		viemOriginalDigest:   {sign(viemOriginalDigest, unpatched)},
	}

	executor := createRedirectExecutor(t, testRecoveryAddress, signatures)
	batch, err := executor.GetBatchFromMultiversX(context.Background())
	require.Nil(t, err)
	require.Nil(t, executor.StoreBatchFromMultiversX(batch))
	require.Nil(t, executor.SignTransferOnEthereum())

	tracked := executor.redirectSignatures
	redirected := executor.logSignaturesFor(tracked, tracked.redirectedHash, false)
	original := executor.logSignaturesFor(tracked, tracked.originalHash, true)

	expectedRedirected := []string{crypto.PubkeyToAddress(patchedA.PublicKey).Hex(), crypto.PubkeyToAddress(patchedB.PublicKey).Hex()}
	assert.ElementsMatch(t, expectedRedirected, redirected)
	assert.Equal(t, []string{crypto.PubkeyToAddress(unpatched.PublicKey).Hex()}, original)

	executor.logRedirectSignatures()
	assert.NotEmpty(t, tracked.lastSummary)
	assert.Len(t, tracked.logged, 3)
}

func TestBridgeExecutor_UnaffectedBatchKeepsOriginalBehaviour(t *testing.T) {
	t.Parallel()

	executor := createRedirectExecutor(t, "", nil)
	other := batch5956FromMultiversX()
	other.ID = 5957
	other.Deposits[0].ToBytes = bytes.Repeat([]byte{0xaa}, 20)
	other.Deposits = other.Deposits[:1]
	executor.multiversXClient = &bridgeTests.MultiversXClientStub{
		GetPendingBatchCalled: func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
			return other, nil
		},
	}

	batch, err := executor.GetBatchFromMultiversX(context.Background())
	require.Nil(t, err)
	assert.Same(t, other, batch)
	require.Nil(t, executor.StoreBatchFromMultiversX(batch))
	require.Nil(t, executor.SignTransferOnEthereum())
	assert.Nil(t, executor.redirectSignatures)
}
