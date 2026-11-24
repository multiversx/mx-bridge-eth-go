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

// first four deposits of mainnet batch 6185; digests computed independently with viem, as above
const (
	viemBatch6185RedirectedDigest = "0x8fcae496f41f14e9a10c150dc52ae129f3c0a763c60e5106f623bbc6181f777e"
	viemBatch6185OriginalDigest   = "0x2a9121a7d7205802f650c5fc69c1202b5fbd8eebd7410430f418a8a255aa7079"
)

var batch6185OtherRecipients = []common.Address{
	common.HexToAddress("0x7f341040a9708bdbbecffa1ac997653447bdbe1b"),
	common.HexToAddress("0x86a3dc5b726b513af1e1e630e39987c19b68379a"),
	common.HexToAddress("0x97ce2c128ce3cb2705ead8d1df95528a3cf1476e"),
}

func batch6185FromMultiversX() *bridgeCore.TransferBatch {
	redirectedSender, _ := hex.DecodeString("d1e0e0a269ae700450e8dda25a4e31da6154f3a1f5e3adb4ffeb856eaf47b210")
	redirectedRecipient := common.HexToAddress("0x3099ba48d4daf3fad0b874b40f1bb55f8a6b4107")
	otherSender := bytes.Repeat([]byte{0x11}, 32)
	usdc := common.HexToAddress("0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d").Bytes()
	tada := common.HexToAddress("0x9b26e318bc6a2c8b45f5daea2cc14697e0e0f8b5").Bytes()
	deposit := func(nonce uint64, from []byte, to common.Address, ticker string, token []byte, amount string) *bridgeCore.DepositTransfer {
		value, _ := big.NewInt(0).SetString(amount, 10)
		return &bridgeCore.DepositTransfer{
			Nonce:                 nonce,
			FromBytes:             from,
			ToBytes:               to.Bytes(),
			SourceTokenBytes:      []byte(ticker),
			DestinationTokenBytes: token,
			DisplayableToken:      ticker,
			Amount:                value,
		}
	}

	return &bridgeCore.TransferBatch{
		ID: 6185,
		Deposits: []*bridgeCore.DepositTransfer{
			deposit(7961, redirectedSender, redirectedRecipient, "BSCUSDC-887875", usdc, "8853225198000000000000"),
			deposit(7962, otherSender, batch6185OtherRecipients[0], "BSCUSDC-887875", usdc, "120692637000000000000"),
			deposit(7963, otherSender, batch6185OtherRecipients[1], "BSCUSDC-887875", usdc, "39000000000000000000"),
			deposit(7964, otherSender, batch6185OtherRecipients[2], "BSCTADA-f64757", tada, "964222143204175498875647"),
		},
		Statuses: make([]byte, 4),
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

func TestBridgeExecutor_RedirectsOnlyTheListedDepositOfBatch6185(t *testing.T) {
	t.Parallel()

	executor := createRedirectExecutor(t, testRecoveryAddress, nil)
	executor.multiversXClient = &bridgeTests.MultiversXClientStub{
		GetPendingBatchCalled: func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
			return batch6185FromMultiversX(), nil
		},
	}

	batch, err := executor.GetBatchFromMultiversX(context.Background())
	require.Nil(t, err)
	require.Nil(t, executor.StoreBatchFromMultiversX(batch))
	require.Nil(t, executor.SignTransferOnEthereum())
	assert.Equal(t, viemBatch6185RedirectedDigest, executor.msgHash.Hex())
	require.NotNil(t, executor.redirectSignatures)
	assert.Equal(t, viemBatch6185OriginalDigest, executor.redirectSignatures.originalHash.Hex())

	expectedRecipients := append([]common.Address{common.HexToAddress(testRecoveryAddress)}, batch6185OtherRecipients...)
	assert.Equal(t, expectedRecipients, batchProcessor.ExtractListMvxToEth(executor.batch).Recipients)
}

func TestBridgeExecutor_RefusesBatch6185WithoutRecoveryAddress(t *testing.T) {
	t.Parallel()

	executor := createRedirectExecutor(t, "", nil)
	executor.multiversXClient = &bridgeTests.MultiversXClientStub{
		GetPendingBatchCalled: func(ctx context.Context) (*bridgeCore.TransferBatch, error) {
			return batch6185FromMultiversX(), nil
		},
	}

	batch, err := executor.GetBatchFromMultiversX(context.Background())
	assert.ErrorIs(t, err, batchRedirect.ErrRecoveryAddressNotSet)
	assert.Nil(t, batch)
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
