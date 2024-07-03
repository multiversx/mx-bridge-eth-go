package multiversx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

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
)

// ArgsScCallExecutor represents the DTO struct for creating a new instance of type scCallExecutor
type ArgsScCallExecutor struct {
	ScProxyBech32Address string
	Proxy                Proxy
	Codec                Codec
	Filter               ScCallsExecuteFilter
	Log                  logger.Logger
	ExtraGasToExecute    uint64
	NonceTxHandler       NonceTransactionsHandler
	PrivateKey           crypto.PrivateKey
	SingleSigner         crypto.SingleSigner
}

type scCallExecutor struct {
	scProxyBech32Address string
	proxy                Proxy
	codec                Codec
	filter               ScCallsExecuteFilter
	log                  logger.Logger
	extraGasToExecute    uint64
	nonceTxHandler       NonceTransactionsHandler
	privateKey           crypto.PrivateKey
	singleSigner         crypto.SingleSigner
	senderAddress        core.AddressHandler
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
		scProxyBech32Address: args.ScProxyBech32Address,
		proxy:                args.Proxy,
		codec:                args.Codec,
		filter:               args.Filter,
		log:                  args.Log,
		extraGasToExecute:    args.ExtraGasToExecute,
		nonceTxHandler:       args.NonceTxHandler,
		privateKey:           args.PrivateKey,
		singleSigner:         args.SingleSigner,
		senderAddress:        senderAddress,
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

	return result
}

func (executor *scCallExecutor) executeOperations(ctx context.Context, pendingOperations map[uint64]parsers.ProxySCCompleteCallData) error {
	networkConfig, err := executor.proxy.GetNetworkConfig(ctx)
	if err != nil {
		return fmt.Errorf("%w while fetching network configs", err)
	}

	for id, callData := range pendingOperations {
		err = executor.executeOperation(id, callData, networkConfig)
		if err != nil {
			return fmt.Errorf("%w for call data: %s", err, callData)
		}
	}

	return nil
}

func (executor *scCallExecutor) executeOperation(
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

	tx := &transaction.FrontendTransaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: callData.GasLimit + executor.extraGasToExecute,
		Data:     dataBytes,
		Sender:   bech32Address,
		Receiver: executor.scProxyBech32Address,
		Value:    "0",
	}

	err = executor.nonceTxHandler.ApplyNonceAndGasPrice(context.Background(), executor.senderAddress, tx)
	if err != nil {
		return err
	}

	err = executor.signTransactionWithPrivateKey(tx)
	if err != nil {
		return err
	}

	hash, err := executor.nonceTxHandler.SendTransaction(context.Background(), tx)
	if err != nil {
		return err
	}

	executor.log.Info("sent transaction from executor",
		"hash", hash,
		"tx ID", id,
		"call data", callData.String(),
		"extra gas", executor.extraGasToExecute,
		"sender", bech32Address)

	return nil
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

// IsInterfaceNil returns true if there is no value under the interface
func (executor *scCallExecutor) IsInterfaceNil() bool {
	return executor == nil
}
