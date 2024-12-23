package multiversx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	minCheckValues               = 1
	transactionNotFoundErrString = "transaction not found"
	minGasToExecuteSCCalls       = 2010000 // the absolut minimum gas limit to do a SC call
)

// ArgsTransactionExecutor represents the DTO struct for creating a new instance of transaction executor
type ArgsTransactionExecutor struct {
	Proxy             Proxy
	Log               logger.Logger
	NonceTxHandler    NonceTransactionsHandler
	PrivateKey        crypto.PrivateKey
	SingleSigner      crypto.SingleSigner
	TransactionChecks config.TransactionChecksConfig
	CloseAppChan      chan struct{}
}

type transactionExecutor struct {
	proxy                   Proxy
	nonceTxHandler          NonceTransactionsHandler
	numSentTransactions     uint32
	privateKey              crypto.PrivateKey
	singleSigner            crypto.SingleSigner
	senderAddress           core.AddressHandler
	log                     logger.Logger
	timeBetweenChecks       time.Duration
	closeAppOnError         bool
	extraDelayOnError       time.Duration
	closeAppChan            chan struct{}
	checkTransactionResults bool
	mutCriticalSection      sync.Mutex
}

// NewTransactionExecutor creates a new executor instance that is able to send transactions & handle results
func NewTransactionExecutor(args ArgsTransactionExecutor) (*transactionExecutor, error) {
	err := checkTransactionExecutorArgs(args)
	if err != nil {
		return nil, err
	}

	publicKey := args.PrivateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}
	senderAddress := data.NewAddressFromBytes(publicKeyBytes)

	return &transactionExecutor{
		proxy:                   args.Proxy,
		log:                     args.Log,
		nonceTxHandler:          args.NonceTxHandler,
		privateKey:              args.PrivateKey,
		singleSigner:            args.SingleSigner,
		senderAddress:           senderAddress,
		checkTransactionResults: args.TransactionChecks.CheckTransactionResults,
		timeBetweenChecks:       time.Second * time.Duration(args.TransactionChecks.TimeInSecondsBetweenChecks),
		closeAppOnError:         args.TransactionChecks.CloseAppOnError,
		extraDelayOnError:       time.Second * time.Duration(args.TransactionChecks.ExtraDelayInSecondsOnError),
		closeAppChan:            args.CloseAppChan,
	}, nil
}

func checkTransactionExecutorArgs(args ArgsTransactionExecutor) error {
	if check.IfNil(args.Proxy) {
		return errNilProxy
	}

	if check.IfNil(args.Log) {
		return errNilLogger
	}
	if check.IfNil(args.NonceTxHandler) {
		return errNilNonceTxHandler
	}
	if check.IfNil(args.PrivateKey) {
		return errNilPrivateKey
	}
	if check.IfNil(args.SingleSigner) {
		return errNilSingleSigner
	}
	err := checkTransactionChecksConfig(args.TransactionChecks, args.Log)
	if err != nil {
		return err
	}

	if args.CloseAppChan == nil && args.TransactionChecks.CloseAppOnError {
		return fmt.Errorf("%w while the TransactionChecks.CloseAppOnError is set to true", errNilCloseAppChannel)
	}

	return nil
}

func checkTransactionChecksConfig(args config.TransactionChecksConfig, log logger.Logger) error {
	if !args.CheckTransactionResults {
		log.Warn("transaction checks are disabled! This can lead to funds being drained in case of a repetitive error")
		return nil
	}

	if args.TimeInSecondsBetweenChecks < minCheckValues {
		return fmt.Errorf("%w for TransactionChecks.TimeInSecondsBetweenChecks, minimum: %d, got: %d",
			errInvalidValue, minCheckValues, args.TimeInSecondsBetweenChecks)
	}
	if args.ExecutionTimeoutInSeconds < minCheckValues {
		return fmt.Errorf("%w for TransactionChecks.ExecutionTimeoutInSeconds, minimum: %d, got: %d",
			errInvalidValue, minCheckValues, args.ExecutionTimeoutInSeconds)
	}

	return nil
}

// ExecuteTransaction will try to execute a transaction. It also can handle the results.
// Concurrent safe function.
func (executor *transactionExecutor) ExecuteTransaction(
	ctx context.Context,
	networkConfig *data.NetworkConfig,
	receiver string,
	transactionType string,
	gasLimit uint64,
	dataBytes []byte,
) error {
	if networkConfig == nil {
		return builders.ErrNilNetworkConfig
	}
	_, err := data.NewAddressFromBech32String(receiver)
	if err != nil {
		return err
	}

	bech32Address, err := executor.senderAddress.AddressAsBech32String()
	if err != nil {
		return err
	}

	tx := &transaction.FrontendTransaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: gasLimit,
		Data:     dataBytes,
		Sender:   bech32Address,
		Receiver: receiver,
		Value:    "0",
	}

	hash, err := executor.executeAsCriticalSection(ctx, tx)
	if err != nil {
		return err
	}

	executor.log.Info("executeOperation: sent transaction from executor",
		"type", transactionType,
		"hash", hash,
		"nonce", tx.Nonce,
		"data", dataBytes,
		"gas provided", gasLimit,
		"sender", bech32Address)

	atomic.AddUint32(&executor.numSentTransactions, 1)

	return executor.handleResults(ctx, hash)
}

func (executor *transactionExecutor) executeAsCriticalSection(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
	executor.mutCriticalSection.Lock()
	defer executor.mutCriticalSection.Unlock()

	err := executor.nonceTxHandler.ApplyNonceAndGasPrice(ctx, executor.senderAddress, tx)
	if err != nil {
		return "", err
	}

	err = executor.signTransactionWithPrivateKey(tx)
	if err != nil {
		return "", err
	}

	return executor.nonceTxHandler.SendTransaction(ctx, tx)
}

// signTransactionWithPrivateKey signs a transaction with the client's private key
func (executor *transactionExecutor) signTransactionWithPrivateKey(tx *transaction.FrontendTransaction) error {
	tx.Signature = ""
	bytes, err := json.Marshal(&tx)
	if err != nil {
		return err
	}

	signature, err := executor.singleSigner.Sign(executor.privateKey, bytes)
	if err != nil {
		return err
	}

	tx.Signature = hex.EncodeToString(signature)

	return nil
}

func (executor *transactionExecutor) handleResults(ctx context.Context, hash string) error {
	if !executor.checkTransactionResults {
		return nil
	}

	err := executor.checkResultsUntilDone(ctx, hash)
	executor.waitForExtraDelay(ctx, err)
	return err
}

func (executor *transactionExecutor) checkResultsUntilDone(ctx context.Context, hash string) error {
	timer := time.NewTimer(executor.timeBetweenChecks)
	defer timer.Stop()

	for {
		timer.Reset(executor.timeBetweenChecks)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			err, shouldStop := executor.checkResults(ctx, hash)
			if shouldStop {
				executor.handleError(ctx, err)
				return err
			}
		}
	}
}

func (executor *transactionExecutor) checkResults(ctx context.Context, hash string) (error, bool) {
	txStatus, err := executor.proxy.ProcessTransactionStatus(ctx, hash)
	if err != nil {
		if err.Error() == transactionNotFoundErrString {
			return nil, false
		}

		return err, true
	}

	if txStatus == transaction.TxStatusSuccess {
		return nil, true
	}
	if txStatus == transaction.TxStatusPending {
		return nil, false
	}

	executor.logFullTransaction(ctx, hash)
	return fmt.Errorf("%w for tx hash %s", errTransactionFailed, hash), true
}

func (executor *transactionExecutor) logFullTransaction(ctx context.Context, hash string) {
	txData, err := executor.proxy.GetTransactionInfoWithResults(ctx, hash)
	if err != nil {
		executor.log.Error("error getting the transaction for display", "error", err)
		return
	}

	txDataString, err := json.MarshalIndent(txData.Data.Transaction, "", "  ")
	if err != nil {
		executor.log.Error("error preparing transaction for display", "error", err)
		return
	}

	executor.log.Error("transaction failed", "hash", hash, "full transaction details", string(txDataString))
}

func (executor *transactionExecutor) handleError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	if !executor.closeAppOnError {
		return
	}

	go func() {
		// wait here until we could write in the close app chan
		// ... or the context expired (application might close)
		select {
		case <-ctx.Done():
		case executor.closeAppChan <- struct{}{}:
		}
	}()
}

func (executor *transactionExecutor) waitForExtraDelay(ctx context.Context, err error) {
	if err == nil {
		return
	}

	timer := time.NewTimer(executor.extraDelayOnError)
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// GetNumSentTransaction returns the total sent transactions
func (executor *transactionExecutor) GetNumSentTransaction() uint32 {
	return atomic.LoadUint32(&executor.numSentTransactions)
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *transactionExecutor) IsInterfaceNil() bool {
	return executor == nil
}
