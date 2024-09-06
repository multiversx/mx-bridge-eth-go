package ethereum

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const ethSignatureSize = 64

// ArgsMigrationBatchExecutor is the argument for the NewMigrationBatchExecutor constructor
type ArgsMigrationBatchExecutor struct {
	EthereumChainWrapper    EthereumChainWrapper
	CryptoHandler           CryptoHandler
	Batch                   BatchInfo
	Signatures              []SignatureInfo
	Logger                  logger.Logger
	GasHandler              GasHandler
	TransferGasLimitBase    uint64
	TransferGasLimitForEach uint64
}

type migrationBatchExecutor struct {
	ethereumChainWrapper    EthereumChainWrapper
	cryptoHandler           CryptoHandler
	batch                   BatchInfo
	signatures              []SignatureInfo
	logger                  logger.Logger
	gasHandler              GasHandler
	transferGasLimitBase    uint64
	transferGasLimitForEach uint64
}

// NewMigrationBatchExecutor creates a new instance of type migrationBatchCreator that is able to execute the multisig transfer
func NewMigrationBatchExecutor(args ArgsMigrationBatchExecutor) (*migrationBatchExecutor, error) {
	if check.IfNilReflect(args.EthereumChainWrapper) {
		return nil, errNilEthereumChainWrapper
	}
	if check.IfNil(args.CryptoHandler) {
		return nil, errNilCryptoHandler
	}
	if check.IfNil(args.Logger) {
		return nil, errNilLogger
	}
	if check.IfNil(args.GasHandler) {
		return nil, errNilGasHandler
	}

	return &migrationBatchExecutor{
		ethereumChainWrapper:    args.EthereumChainWrapper,
		cryptoHandler:           args.CryptoHandler,
		batch:                   args.Batch,
		signatures:              args.Signatures,
		logger:                  args.Logger,
		gasHandler:              args.GasHandler,
		transferGasLimitBase:    args.TransferGasLimitBase,
		transferGasLimitForEach: args.TransferGasLimitForEach,
	}, nil
}

// ExecuteTransfer will try to execute the transfer
func (executor *migrationBatchExecutor) ExecuteTransfer(ctx context.Context) error {
	isPaused, err := executor.ethereumChainWrapper.IsPaused(ctx)
	if err != nil {
		return fmt.Errorf("%w in executor.ExecuteTransfer", err)
	}
	if isPaused {
		return fmt.Errorf("%w in executor.ExecuteTransfer", errMultisigContractPaused)
	}

	relayers, err := executor.ethereumChainWrapper.GetRelayers(ctx)
	if err != nil {
		return err
	}

	quorum, err := executor.ethereumChainWrapper.Quorum(ctx)
	if err != nil {
		return err
	}

	signatures, err := executor.checkRelayersSigsAndQuorum(relayers, quorum)
	if err != nil {
		return err
	}

	nonce, err := executor.getNonce(ctx, executor.cryptoHandler.GetAddress())
	if err != nil {
		return err
	}

	chainId, err := executor.ethereumChainWrapper.ChainID(ctx)
	if err != nil {
		return err
	}

	auth, err := executor.cryptoHandler.CreateKeyedTransactor(chainId)
	if err != nil {
		return err
	}

	gasPrice, err := executor.gasHandler.GetCurrentGasPrice()
	if err != nil {
		return err
	}

	tokens, recipients, amounts, depositNonces, batchNonce := executor.extractArgumentsFromBatch()

	auth.Nonce = big.NewInt(nonce)
	auth.Value = big.NewInt(0)
	auth.GasLimit = executor.transferGasLimitBase + uint64(len(tokens))*executor.transferGasLimitForEach
	auth.Context = ctx
	auth.GasPrice = gasPrice

	tx, err := executor.ethereumChainWrapper.ExecuteTransfer(auth, tokens, recipients, amounts, depositNonces, batchNonce, signatures)
	if err != nil {
		return err
	}

	txHash := tx.Hash().String()
	executor.logger.Info("Executed transfer transaction", "batchID", executor.batch.BatchID, "hash", txHash)

	return nil
}

func (executor *migrationBatchExecutor) getNonce(ctx context.Context, fromAddress common.Address) (int64, error) {
	blockNonce, err := executor.ethereumChainWrapper.BlockNumber(ctx)
	if err != nil {
		return 0, fmt.Errorf("%w in getNonce, BlockNumber call", err)
	}

	nonce, err := executor.ethereumChainWrapper.NonceAt(ctx, fromAddress, big.NewInt(int64(blockNonce)))

	return int64(nonce), err
}

func (executor *migrationBatchExecutor) extractArgumentsFromBatch() (
	tokens []common.Address,
	recipients []common.Address,
	amounts []*big.Int,
	nonces []*big.Int,
	batchNonce *big.Int,
) {
	tokens = make([]common.Address, 0, len(executor.batch.DepositsInfo))
	recipients = make([]common.Address, 0, len(executor.batch.DepositsInfo))
	amounts = make([]*big.Int, 0, len(executor.batch.DepositsInfo))
	nonces = make([]*big.Int, 0, len(executor.batch.DepositsInfo))
	batchNonce = big.NewInt(0).SetUint64(executor.batch.BatchID)

	newSafeContractAddress := common.HexToAddress(executor.batch.NewSafeContractAddress)
	for _, deposit := range executor.batch.DepositsInfo {
		tokens = append(tokens, deposit.contractAddress)
		recipients = append(recipients, newSafeContractAddress)
		amounts = append(amounts, deposit.amount)
		nonces = append(nonces, big.NewInt(0).SetUint64(deposit.DepositNonce))
	}

	return
}

func (executor *migrationBatchExecutor) checkRelayersSigsAndQuorum(relayers []common.Address, quorum *big.Int) ([][]byte, error) {
	sameMessageHashSignatures := executor.getSameMessageHashSignatures()
	validSignatures := executor.getValidSignatures(sameMessageHashSignatures)
	return executor.checkQuorum(relayers, quorum, validSignatures)
}

func (executor *migrationBatchExecutor) getSameMessageHashSignatures() []SignatureInfo {
	filtered := make([]SignatureInfo, 0, len(executor.signatures))
	for _, sigInfo := range executor.signatures {
		if sigInfo.MessageHash != executor.batch.MessageHash.String() {
			executor.logger.Warn("found a signature info that was not carried on the same message hash",
				"local message hash", executor.batch.MessageHash.String(),
				"address", sigInfo.Address, "message hash", sigInfo.MessageHash)

			continue
		}

		filtered = append(filtered, sigInfo)
	}

	return filtered
}

func (executor *migrationBatchExecutor) getValidSignatures(provided []SignatureInfo) []SignatureInfo {
	filtered := make([]SignatureInfo, 0, len(provided))
	for _, sigInfo := range provided {
		hash := common.HexToHash(sigInfo.MessageHash)
		sig, err := hex.DecodeString(sigInfo.Signature)
		if err != nil {
			executor.logger.Warn("found a non valid signature info (can not unhex the signature)",
				"address", sigInfo.Address, "message hash", sigInfo.MessageHash, "signature", sigInfo.Signature, "error", err)
			continue
		}

		err = verifySignature(hash, sig, common.HexToAddress(sigInfo.Address))
		if err != nil {
			executor.logger.Warn("found a non valid signature info",
				"address", sigInfo.Address, "message hash", sigInfo.MessageHash, "signature", sigInfo.Signature, "error", err)
			continue
		}

		filtered = append(filtered, sigInfo)
	}

	return filtered
}

func verifySignature(messageHash common.Hash, signature []byte, address common.Address) error {
	pkBytes, err := crypto.Ecrecover(messageHash.Bytes(), signature)
	if err != nil {
		return err
	}

	pk, err := crypto.UnmarshalPubkey(pkBytes)
	if err != nil {
		return err
	}

	addressFromPk := crypto.PubkeyToAddress(*pk)
	if addressFromPk.String() != address.String() {
		// we need to check that the recovered public key matched the one provided in order to make sure
		// that the signature, hash and public key match
		return errInvalidSignature
	}

	if len(signature) > ethSignatureSize {
		// signatures might contain the recovery byte
		signature = signature[:ethSignatureSize]
	}

	sigOk := crypto.VerifySignature(pkBytes, messageHash.Bytes(), signature)
	if !sigOk {
		return errInvalidSignature
	}

	return nil
}

func (executor *migrationBatchExecutor) checkQuorum(relayers []common.Address, quorum *big.Int, signatures []SignatureInfo) ([][]byte, error) {
	whitelistedRelayers := make(map[common.Address]SignatureInfo)

	for _, sigInfo := range signatures {
		if !isWhitelistedRelayer(sigInfo, relayers) {
			executor.logger.Warn("found a non whitelisted relayer",
				"address", sigInfo.Address)
			continue
		}

		relayerAddress := common.HexToAddress(sigInfo.Address)
		_, found := whitelistedRelayers[relayerAddress]
		if found {
			executor.logger.Warn("found a multiple relayer sig info, ignoring",
				"address", sigInfo.Address)
			continue
		}

		whitelistedRelayers[relayerAddress] = sigInfo
	}

	result := make([][]byte, 0, len(whitelistedRelayers))
	for _, sigInfo := range whitelistedRelayers {
		sig, err := hex.DecodeString(sigInfo.Signature)
		if err != nil {
			return nil, fmt.Errorf("internal error: %w while decoding this string %s that should have been hexed encoded", err, sigInfo.Signature)
		}

		result = append(result, sig)
		executor.logger.Info("valid signature recorded for whitelisted relayer", "relayer", sigInfo.Address)
	}

	if uint64(len(result)) < quorum.Uint64() {
		return nil, fmt.Errorf("%w: minimum %d, got %d", errQuorumNotReached, quorum.Uint64(), len(result))
	}

	return result, nil
}

func isWhitelistedRelayer(sigInfo SignatureInfo, relayers []common.Address) bool {
	relayerAddress := common.HexToAddress(sigInfo.Address)
	for _, relayer := range relayers {
		if bytes.Equal(relayer.Bytes(), relayerAddress.Bytes()) {
			return true
		}
	}

	return false
}
