package multiversx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/errors"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	getPendingTransactionsFunction = "getPendingTransactions"
	okCodeAfterExecution           = "ok"
	scProxyCallFunction            = "execute"
	minCheckValues                 = 1
	transactionNotFoundErrString   = "transaction not found"
	minGasToExecuteSCCalls         = 2010000 // the absolut minimum gas limit to do a SC call
)

// ArgsScCallExecutor represents the DTO struct for creating a new instance of type scCallExecutor
type ArgsScCallExecutor struct {
	ScProxyBech32Address string
	Proxy                Proxy
	Codec                Codec
	Filter               ScCallsExecuteFilter
	Log                  logger.Logger
	ExtraGasToExecute    uint64
	MaxGasLimitToUse     uint64
	NonceTxHandler       NonceTransactionsHandler
	PrivateKey           crypto.PrivateKey
	SingleSigner         crypto.SingleSigner
	TransactionChecks    config.TransactionChecksConfig
	CloseAppChan         chan struct{}
}

type scCallExecutor struct {
	scProxyBech32Address    string
	proxy                   Proxy
	codec                   Codec
	filter                  ScCallsExecuteFilter
	log                     logger.Logger
	extraGasToExecute       uint64
	maxGasLimitToUse        uint64
	nonceTxHandler          NonceTransactionsHandler
	privateKey              crypto.PrivateKey
	singleSigner            crypto.SingleSigner
	senderAddress           core.AddressHandler
	numSentTransactions     uint32
	checkTransactionResults bool
	timeBetweenChecks       time.Duration
	executionTimeout        time.Duration
	closeAppOnError         bool
	extraDelayOnError       time.Duration
	closeAppChan            chan struct{}
}

// NewScCallExecutor creates a new instance of type scCallExecutor
func NewScCallExecutor(args ArgsScCallExecutor) (*scCallExecutor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	publicKey := args.PrivateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}
	senderAddress := data.NewAddressFromBytes(publicKeyBytes)

	return &scCallExecutor{
		scProxyBech32Address:    args.ScProxyBech32Address,
		proxy:                   args.Proxy,
		codec:                   args.Codec,
		filter:                  args.Filter,
		log:                     args.Log,
		extraGasToExecute:       args.ExtraGasToExecute,
		maxGasLimitToUse:        args.MaxGasLimitToUse,
		nonceTxHandler:          args.NonceTxHandler,
		privateKey:              args.PrivateKey,
		singleSigner:            args.SingleSigner,
		senderAddress:           senderAddress,
		checkTransactionResults: args.TransactionChecks.CheckTransactionResults,
		timeBetweenChecks:       time.Second * time.Duration(args.TransactionChecks.TimeInSecondsBetweenChecks),
		executionTimeout:        time.Second * time.Duration(args.TransactionChecks.ExecutionTimeoutInSeconds),
		closeAppOnError:         args.TransactionChecks.CloseAppOnError,
		extraDelayOnError:       time.Second * time.Duration(args.TransactionChecks.ExtraDelayInSecondsOnError),
		closeAppChan:            args.CloseAppChan,
	}, nil
}

func checkArgs(args ArgsScCallExecutor) error {
	if check.IfNil(args.Proxy) {
		return errNilProxy
	}
	if check.IfNil(args.Codec) {
		return errNilCodec
	}
	if check.IfNil(args.Filter) {
		return errNilFilter
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
	if args.MaxGasLimitToUse < minGasToExecuteSCCalls {
		return fmt.Errorf("%w: provided: %d, absolute minimum required: %d", errMaxGasLimitIsLessThanRequired, args.MaxGasLimitToUse, minGasToExecuteSCCalls)
	}
	err := checkTransactionChecksConfig(args)
	if err != nil {
		return err
	}

	_, err = data.NewAddressFromBech32String(args.ScProxyBech32Address)

	return err
}

func checkTransactionChecksConfig(args ArgsScCallExecutor) error {
	if !args.TransactionChecks.CheckTransactionResults {
		args.Log.Warn("transaction checks are disabled! This can lead to funds being drained in case of a repetitive error")
		return nil
	}

	if args.TransactionChecks.TimeInSecondsBetweenChecks < minCheckValues {
		return fmt.Errorf("%w for TransactionChecks.TimeInSecondsBetweenChecks, minimum: %d, got: %d",
			errInvalidValue, minCheckValues, args.TransactionChecks.TimeInSecondsBetweenChecks)
	}
	if args.TransactionChecks.ExecutionTimeoutInSeconds < minCheckValues {
		return fmt.Errorf("%w for TransactionChecks.ExecutionTimeoutInSeconds, minimum: %d, got: %d",
			errInvalidValue, minCheckValues, args.TransactionChecks.ExecutionTimeoutInSeconds)
	}
	if args.CloseAppChan == nil && args.TransactionChecks.CloseAppOnError {
		return fmt.Errorf("%w while the TransactionChecks.CloseAppOnError is set to true", errNilCloseAppChannel)
	}

	return nil
}

// Execute will execute one step: get all pending operations, call the filter and send execution transactions
func (executor *scCallExecutor) Execute(ctx context.Context) error {
	pendingOperations, err := executor.getPendingOperations(ctx)
	if err != nil {
		return err
	}

	filteredPendingOperations := executor.filterOperations(pendingOperations)

	return executor.executeOperations(ctx, filteredPendingOperations)
}

func (executor *scCallExecutor) getPendingOperations(ctx context.Context) (map[uint64]parsers.ProxySCCompleteCallData, error) {
	request := &data.VmValueRequest{
		Address:  executor.scProxyBech32Address,
		FuncName: getPendingTransactionsFunction,
	}

	response, err := executor.proxy.ExecuteVMQuery(ctx, request)
	if err != nil {
		executor.log.Error("got error on VMQuery", "FuncName", request.FuncName,
			"Args", request.Args, "SC address", request.Address, "Caller", request.CallerAddr, "error", err)
		return nil, err
	}
	if response.Data.ReturnCode != okCodeAfterExecution {
		return nil, errors.NewQueryResponseError(
			response.Data.ReturnCode,
			response.Data.ReturnMessage,
			request.FuncName,
			request.Address,
			request.Args...,
		)
	}

	return executor.parseResponse(response)
}

func (executor *scCallExecutor) parseResponse(response *data.VmValuesResponseData) (map[uint64]parsers.ProxySCCompleteCallData, error) {
	numResponseLines := len(response.Data.ReturnData)
	if numResponseLines%2 != 0 {
		return nil, fmt.Errorf("%w: expected an even number, got %d", errInvalidNumberOfResponseLines, numResponseLines)
	}

	result := make(map[uint64]parsers.ProxySCCompleteCallData, numResponseLines/2)

	for i := 0; i < numResponseLines; i += 2 {
		pendingOperationID := big.NewInt(0).SetBytes(response.Data.ReturnData[i])
		callData, err := executor.codec.DecodeProxySCCompleteCallData(response.Data.ReturnData[i+1])
		if err != nil {
			return nil, fmt.Errorf("%w for ReturnData at index %d", err, i+1)
		}

		result[pendingOperationID.Uint64()] = callData
	}

	return result, nil
}

func (executor *scCallExecutor) filterOperations(pendingOperations map[uint64]parsers.ProxySCCompleteCallData) map[uint64]parsers.ProxySCCompleteCallData {
	result := make(map[uint64]parsers.ProxySCCompleteCallData)
	for id, callData := range pendingOperations {
		if executor.filter.ShouldExecute(callData) {
			result[id] = callData
		}
	}

	executor.log.Debug("scCallExecutor.filterOperations", "input pending ops", len(pendingOperations), "result pending ops", len(result))

	return result
}

func (executor *scCallExecutor) executeOperations(ctx context.Context, pendingOperations map[uint64]parsers.ProxySCCompleteCallData) error {
	networkConfig, err := executor.proxy.GetNetworkConfig(ctx)
	if err != nil {
		return fmt.Errorf("%w while fetching network configs", err)
	}

	for id, callData := range pendingOperations {
		workingCtx, cancel := context.WithTimeout(ctx, executor.executionTimeout)

		executor.log.Debug("scCallExecutor.executeOperations", "executing ID", id, "call data", callData,
			"maximum timeout", executor.executionTimeout)
		err = executor.executeOperation(workingCtx, id, callData, networkConfig)
		cancel()

		if err != nil {
			return fmt.Errorf("%w for call data: %s", err, callData)
		}
	}

	return nil
}

func (executor *scCallExecutor) executeOperation(
	ctx context.Context,
	id uint64,
	callData parsers.ProxySCCompleteCallData,
	networkConfig *data.NetworkConfig,
) error {
	txBuilder := builders.NewTxDataBuilder()
	txBuilder.Function(scProxyCallFunction).ArgInt64(int64(id))

	dataBytes, err := txBuilder.ToDataBytes()
	if err != nil {
		return err
	}

	bech32Address, err := executor.senderAddress.AddressAsBech32String()
	if err != nil {
		return err
	}

	gasLimit, err := executor.codec.ExtractGasLimitFromRawCallData(callData.RawCallData)
	if err != nil {
		executor.log.Warn("scCallExecutor.executeOperation found a non-parsable raw call data",
			"raw call data", callData.RawCallData, "error", err)
		gasLimit = 0
	}

	tx := &transaction.FrontendTransaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: gasLimit + executor.extraGasToExecute,
		Data:     dataBytes,
		Sender:   bech32Address,
		Receiver: executor.scProxyBech32Address,
		Value:    "0",
	}

	if tx.GasLimit > executor.maxGasLimitToUse {
		to, _ := callData.To.AddressAsBech32String()
		executor.log.Warn("can not execute transaction because the provided gas limit on the SC call exceeds "+
			"the maximum gas limit allowance for this executor, WILL SKIP the execution",
			"computed gas limit", tx.GasLimit,
			"max allowed", executor.maxGasLimitToUse,
			"data", dataBytes,
			"from", callData.From.Hex(),
			"to", to,
			"token", callData.Token,
			"amount", callData.Amount,
			"nonce", callData.Nonce,
		)

		return nil
	}

	err = executor.nonceTxHandler.ApplyNonceAndGasPrice(ctx, executor.senderAddress, tx)
	if err != nil {
		return err
	}

	err = executor.signTransactionWithPrivateKey(tx)
	if err != nil {
		return err
	}

	hash, err := executor.nonceTxHandler.SendTransaction(ctx, tx)
	if err != nil {
		return err
	}

	executor.log.Info("scCallExecutor.executeOperation: sent transaction from executor",
		"hash", hash,
		"tx ID", id,
		"call data", callData.String(),
		"extra gas", executor.extraGasToExecute,
		"sender", bech32Address)

	atomic.AddUint32(&executor.numSentTransactions, 1)

	return executor.handleResults(ctx, hash)
}

func (executor *scCallExecutor) handleResults(ctx context.Context, hash string) error {
	if !executor.checkTransactionResults {
		return nil
	}

	err := executor.checkResultsUntilDone(ctx, hash)
	executor.waitForExtraDelay(ctx, err)
	return err
}

// signTransactionWithPrivateKey signs a transaction with the client's private key
func (executor *scCallExecutor) signTransactionWithPrivateKey(tx *transaction.FrontendTransaction) error {
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

func (executor *scCallExecutor) checkResultsUntilDone(ctx context.Context, hash string) error {
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

func (executor *scCallExecutor) checkResults(ctx context.Context, hash string) (error, bool) {
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

func (executor *scCallExecutor) handleError(ctx context.Context, err error) {
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

func (executor *scCallExecutor) logFullTransaction(ctx context.Context, hash string) {
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

func (executor *scCallExecutor) waitForExtraDelay(ctx context.Context, err error) {
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
func (executor *scCallExecutor) GetNumSentTransaction() uint32 {
	return atomic.LoadUint32(&executor.numSentTransactions)
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *scCallExecutor) IsInterfaceNil() bool {
	return executor == nil
}
