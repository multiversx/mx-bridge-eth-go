package ethereum

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("executors/ethereum_test")

func createMockArgsMigrationBatchExecutor() ArgsMigrationBatchExecutor {
	return ArgsMigrationBatchExecutor{
		EthereumChainWrapper:    &bridge.EthereumClientWrapperStub{},
		CryptoHandler:           &bridge.CryptoHandlerStub{},
		Batch:                   BatchInfo{},
		Signatures:              make([]SignatureInfo, 0),
		Logger:                  log,
		GasHandler:              &testsCommon.GasHandlerStub{},
		TransferGasLimitBase:    100,
		TransferGasLimitForEach: 10,
	}
}

func createPrivateKeys(tb testing.TB, num int) []*ecdsa.PrivateKey {
	keys := make([]*ecdsa.PrivateKey, 0, num)

	for i := 0; i < num; i++ {
		skBytes := make([]byte, 32)
		_, _ = rand.Read(skBytes)

		privateKey, err := ethCrypto.HexToECDSA(hex.EncodeToString(skBytes))
		require.Nil(tb, err)

		keys = append(keys, privateKey)
	}

	return keys
}

func sign(tb testing.TB, sk *ecdsa.PrivateKey, msgHash common.Hash) []byte {
	sig, err := ethCrypto.Sign(msgHash.Bytes(), sk)
	require.Nil(tb, err)

	return sig
}

func TestNewMigrationBatchExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil Ethereum chain wrapper should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.EthereumChainWrapper = nil

		executor, err := NewMigrationBatchExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilEthereumChainWrapper, err)
	})
	t.Run("nil crypto handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.CryptoHandler = nil

		executor, err := NewMigrationBatchExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilCryptoHandler, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Logger = nil

		executor, err := NewMigrationBatchExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilLogger, err)
	})
	t.Run("nil gas handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.GasHandler = nil

		executor, err := NewMigrationBatchExecutor(args)
		assert.Nil(t, executor)
		assert.Equal(t, errNilGasHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		executor, err := NewMigrationBatchExecutor(args)
		assert.NotNil(t, executor)
		assert.Nil(t, err)
	})
}

func TestMigrationBatchExecutor_checkRelayersSigsAndQuorum(t *testing.T) {
	t.Parallel()

	t.Run("quorum not satisfied should error", func(t *testing.T) {
		t.Parallel()

		privateKeys := createPrivateKeys(t, 3)

		testMsgHash := common.HexToHash(strings.Repeat("1", 64))

		signatures := []SignatureInfo{
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[0].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[0], testMsgHash)),
			},
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[1], testMsgHash)),
			},
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[2].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[2], testMsgHash)),
			},
		}

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = BatchInfo{
			MessageHash: testMsgHash,
		}
		args.Signatures = signatures

		executor, _ := NewMigrationBatchExecutor(args)
		whitelistedRelayers := []common.Address{
			ethCrypto.PubkeyToAddress(privateKeys[0].PublicKey),
			ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey),
			ethCrypto.PubkeyToAddress(privateKeys[2].PublicKey),
		}

		verifiedSigs, err := executor.checkRelayersSigsAndQuorum(whitelistedRelayers, big.NewInt(4))
		assert.ErrorIs(t, err, errQuorumNotReached)
		assert.Contains(t, err.Error(), "minimum 4, got 3")
		assert.Empty(t, verifiedSigs)
	})
	t.Run("should work with wrong sig info elements", func(t *testing.T) {
		t.Parallel()

		privateKeys := createPrivateKeys(t, 6)

		testMsgHash := common.HexToHash(strings.Repeat("1", 64))
		wrongMsgHash := common.HexToHash(strings.Repeat("2", 64))

		correctSigForFifthElement := hex.EncodeToString(sign(t, privateKeys[5], testMsgHash))
		signatures := []SignatureInfo{
			// wrong message hash
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[4].PublicKey).String(),
				MessageHash: wrongMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[4], wrongMsgHash)),
			},
			// wrong signature: another message hash
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[5].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[5], wrongMsgHash)),
			},
			// wrong signature: not a hex string
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[5].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   "not a hex string",
			},
			// wrong signature: malformed signature
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[5].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   strings.Replace(correctSigForFifthElement, "1", "2", -1),
			},
			// repeated good sig[1]
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[1], testMsgHash)),
			},
			// good sigs
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[0].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[0], testMsgHash)),
			},
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[1], testMsgHash)),
			},
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[2].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[2], testMsgHash)),
			},
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[3].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[3], testMsgHash)),
			},
		}

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = BatchInfo{
			MessageHash: testMsgHash,
		}
		args.Signatures = signatures

		executor, _ := NewMigrationBatchExecutor(args)
		whitelistedRelayers := []common.Address{
			// all but private key[3] are whitelisted
			ethCrypto.PubkeyToAddress(privateKeys[0].PublicKey),
			ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey),
			ethCrypto.PubkeyToAddress(privateKeys[2].PublicKey),
			ethCrypto.PubkeyToAddress(privateKeys[4].PublicKey),
			ethCrypto.PubkeyToAddress(privateKeys[5].PublicKey),
		}

		verifiedSigs, err := executor.checkRelayersSigsAndQuorum(whitelistedRelayers, big.NewInt(3))
		assert.Nil(t, err)
		assert.Equal(t, 3, len(verifiedSigs))
	})
	t.Run("should work with correct sig elements", func(t *testing.T) {
		t.Parallel()

		privateKeys := createPrivateKeys(t, 3)

		testMsgHash := common.HexToHash(strings.Repeat("1", 64))

		signatures := []SignatureInfo{
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[0].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[0], testMsgHash)),
			},
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[1], testMsgHash)),
			},
			{
				Address:     ethCrypto.PubkeyToAddress(privateKeys[2].PublicKey).String(),
				MessageHash: testMsgHash.String(),
				Signature:   hex.EncodeToString(sign(t, privateKeys[2], testMsgHash)),
			},
		}

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = BatchInfo{
			MessageHash: testMsgHash,
		}
		args.Signatures = signatures

		executor, _ := NewMigrationBatchExecutor(args)
		whitelistedRelayers := []common.Address{
			ethCrypto.PubkeyToAddress(privateKeys[0].PublicKey),
			ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey),
			ethCrypto.PubkeyToAddress(privateKeys[2].PublicKey),
		}

		verifiedSigs, err := executor.checkRelayersSigsAndQuorum(whitelistedRelayers, big.NewInt(3))
		assert.Nil(t, err)
		assert.Equal(t, 3, len(verifiedSigs))
	})
}

func TestMigrationBatchExecutor_ExecuteTransfer(t *testing.T) {
	t.Parallel()

	testMsgHash := common.HexToHash(strings.Repeat("1", 64))
	newSafeContractAddress := common.HexToAddress("A6504Cc508889bbDBd4B748aFf6EA6b5D0d2684c")
	batchInfo := BatchInfo{
		OldSafeContractAddress: "3009d97FfeD62E57d444e552A9eDF9Ee6Bc8644c",
		NewSafeContractAddress: newSafeContractAddress.String(),
		BatchID:                4432,
		MessageHash:            testMsgHash,
		DepositsInfo: []*DepositInfo{
			{
				DepositNonce:          37,
				Token:                 "tkn1",
				ContractAddressString: common.BytesToAddress(tkn1Erc20Address).String(),
				ContractAddress:       common.BytesToAddress(tkn1Erc20Address),
				Amount:                big.NewInt(112),
				AmountString:          "112",
			},
			{
				DepositNonce:          38,
				Token:                 "tkn2",
				ContractAddressString: common.BytesToAddress(tkn2Erc20Address).String(),
				ContractAddress:       common.BytesToAddress(tkn2Erc20Address),
				Amount:                big.NewInt(113),
				AmountString:          "113",
			},
		},
	}

	privateKeys := createPrivateKeys(t, 3)
	signatures := []SignatureInfo{
		{
			Address:     ethCrypto.PubkeyToAddress(privateKeys[0].PublicKey).String(),
			MessageHash: testMsgHash.String(),
			Signature:   hex.EncodeToString(sign(t, privateKeys[0], testMsgHash)),
		},
		{
			Address:     ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey).String(),
			MessageHash: testMsgHash.String(),
			Signature:   hex.EncodeToString(sign(t, privateKeys[1], testMsgHash)),
		},
		{
			Address:     ethCrypto.PubkeyToAddress(privateKeys[2].PublicKey).String(),
			MessageHash: testMsgHash.String(),
			Signature:   hex.EncodeToString(sign(t, privateKeys[2], testMsgHash)),
		},
	}

	whitelistedRelayers := []common.Address{
		ethCrypto.PubkeyToAddress(privateKeys[0].PublicKey),
		ethCrypto.PubkeyToAddress(privateKeys[1].PublicKey),
		ethCrypto.PubkeyToAddress(privateKeys[2].PublicKey),
	}
	testBlockNumber := uint64(1000000)
	senderNonce := uint64(3377)
	testChainId := big.NewInt(2222)
	testGasPrice := big.NewInt(112233)

	expectedTokens := []common.Address{
		common.BytesToAddress(tkn1Erc20Address),
		common.BytesToAddress(tkn2Erc20Address),
	}
	expectedRecipients := []common.Address{
		newSafeContractAddress,
		newSafeContractAddress,
	}
	expectedAmounts := []*big.Int{
		big.NewInt(112),
		big.NewInt(113),
	}
	expectedNonces := []*big.Int{
		big.NewInt(37),
		big.NewInt(38),
	}
	expectedSignatures := make([][]byte, 0, len(signatures))
	for _, sigInfo := range signatures {
		sig, err := hex.DecodeString(sigInfo.Signature)
		require.Nil(t, err)
		expectedSignatures = append(expectedSignatures, sig)
	}
	expectedErr := errors.New("expected error")

	t.Run("is paused query errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			IsPausedCalled: func(ctx context.Context) (bool, error) {
				return false, expectedErr
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("is paused should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			IsPausedCalled: func(ctx context.Context) (bool, error) {
				return true, nil
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, errMultisigContractPaused)
	})
	t.Run("get relayers errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return nil, expectedErr
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("get quorum errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return nil, expectedErr
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.Equal(t, expectedErr, err)
	})
	t.Run("checking the signatures and relayers errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures // no whitelisted relayers
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, errQuorumNotReached)
	})
	t.Run("get block errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedRelayers, nil
			},
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return 0, expectedErr
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("get nonce errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedRelayers, nil
			},
			NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
				return 0, expectedErr
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("chain ID errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedRelayers, nil
			},
			ChainIDCalled: func(ctx context.Context) (*big.Int, error) {
				return nil, expectedErr
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("create keyed transactor errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedRelayers, nil
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}
		args.CryptoHandler = &bridge.CryptoHandlerStub{
			CreateKeyedTransactorCalled: func(chainId *big.Int) (*bind.TransactOpts, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("get gas price errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedRelayers, nil
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Fail(t, "should have not called execute transfer")

				return nil, nil
			},
		}
		args.CryptoHandler = &bridge.CryptoHandlerStub{
			CreateKeyedTransactorCalled: func(chainId *big.Int) (*bind.TransactOpts, error) {
				return &bind.TransactOpts{}, nil
			},
		}
		args.GasHandler = &testsCommon.GasHandlerStub{
			GetCurrentGasPriceCalled: func() (*big.Int, error) {
				return nil, expectedErr
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("execute transfer errors should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedRelayers, nil
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				return nil, expectedErr
			},
		}
		args.CryptoHandler = &bridge.CryptoHandlerStub{
			CreateKeyedTransactorCalled: func(chainId *big.Int) (*bind.TransactOpts, error) {
				return &bind.TransactOpts{}, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)
		err := executor.ExecuteTransfer(context.Background())
		assert.ErrorIs(t, err, expectedErr)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsMigrationBatchExecutor()
		args.Batch = batchInfo
		args.Signatures = signatures
		executeWasCalled := false
		args.EthereumChainWrapper = &bridge.EthereumClientWrapperStub{
			GetRelayersCalled: func(ctx context.Context) ([]common.Address, error) {
				return whitelistedRelayers, nil
			},
			BlockNumberCalled: func(ctx context.Context) (uint64, error) {
				return testBlockNumber, nil
			},
			NonceAtCalled: func(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
				assert.Equal(t, big.NewInt(0).SetUint64(testBlockNumber), blockNumber)
				return senderNonce, nil
			},
			ChainIDCalled: func(ctx context.Context) (*big.Int, error) {
				return testChainId, nil
			},
			QuorumCalled: func(ctx context.Context) (*big.Int, error) {
				return big.NewInt(3), nil
			},
			ExecuteTransferCalled: func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
				assert.Equal(t, big.NewInt(0).SetUint64(senderNonce), opts.Nonce)
				assert.Equal(t, big.NewInt(0), opts.Value)
				assert.Equal(t, uint64(100+10+10), opts.GasLimit) // base + 2 deposits
				assert.Equal(t, testGasPrice, opts.GasPrice)
				assert.Equal(t, expectedTokens, tokens)
				assert.Equal(t, expectedRecipients, recipients)
				assert.Equal(t, expectedAmounts, amounts)
				assert.Equal(t, expectedNonces, nonces)
				assert.Equal(t, big.NewInt(4432), batchNonce)
				assert.ElementsMatch(t, expectedSignatures, signatures)
				executeWasCalled = true

				txData := &types.LegacyTx{
					Nonce: 0,
					Data:  []byte("mocked data"),
				}
				tx := types.NewTx(txData)

				return tx, nil
			},
		}
		args.GasHandler = &testsCommon.GasHandlerStub{
			GetCurrentGasPriceCalled: func() (*big.Int, error) {
				return testGasPrice, nil
			},
		}
		args.CryptoHandler = &bridge.CryptoHandlerStub{
			CreateKeyedTransactorCalled: func(chainId *big.Int) (*bind.TransactOpts, error) {
				assert.Equal(t, testChainId, chainId)
				return &bind.TransactOpts{}, nil
			},
		}

		executor, _ := NewMigrationBatchExecutor(args)

		err := executor.ExecuteTransfer(context.Background())
		assert.Nil(t, err)
		assert.True(t, executeWasCalled)
	})
}
