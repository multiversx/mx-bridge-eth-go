package multiversx

import (
	"context"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/errors"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	getPendingTransactionsFunction = "getPendingTransactions"
	okCodeAfterExecution           = "ok"
	scProxyCallFunction            = "execute"
	contractMaxGasLimit            = 249999999
	scCallTxType                   = "SC call"
)

// ArgsScCallExecutor represents the DTO struct for creating a new instance of type scCallExecutor
type ArgsScCallExecutor struct {
	TransactionExecutor    TransactionExecutor
	ScProxyBech32Addresses []string
	Proxy                  Proxy
	Codec                  Codec
	Filter                 ScCallsExecuteFilter
	Log                    logger.Logger
	ExecutorConfig         config.ScCallsExecutorConfig
}

type scCallExecutor struct {
	*baseExecutor
	extraGasToExecute               uint64
	maxGasLimitToUse                uint64
	gasLimitForOutOfGasTransactions uint64
}

// NewScCallExecutor creates a new instance of type scCallExecutor
func NewScCallExecutor(args ArgsScCallExecutor) (*scCallExecutor, error) {
	err := checkScCallExecutorArgs(args)
	if err != nil {
		return nil, err
	}

	executor := &scCallExecutor{
		baseExecutor: &baseExecutor{
			scProxyBech32Addresses: args.ScProxyBech32Addresses,
			proxy:                  args.Proxy,
			transactionExecutor:    args.TransactionExecutor,
			codec:                  args.Codec,
			filter:                 args.Filter,
			log:                    args.Log,
		},
		extraGasToExecute:               args.ExecutorConfig.ExtraGasToExecute,
		maxGasLimitToUse:                args.ExecutorConfig.MaxGasLimitToUse,
		gasLimitForOutOfGasTransactions: args.ExecutorConfig.GasLimitForOutOfGasTransactions,
	}

	err = executor.checkBaseComponents()
	if err != nil {
		return nil, err
	}

	return executor, nil
}

func checkScCallExecutorArgs(args ArgsScCallExecutor) error {
	if args.ExecutorConfig.MaxGasLimitToUse < minGasToExecuteSCCalls {
		return fmt.Errorf("%w for MaxGasLimitToUse: provided: %d, absolute minimum required: %d", errGasLimitIsLessThanAbsoluteMinimum, args.ExecutorConfig.MaxGasLimitToUse, minGasToExecuteSCCalls)
	}
	if args.ExecutorConfig.GasLimitForOutOfGasTransactions < minGasToExecuteSCCalls {
		return fmt.Errorf("%w for GasLimitForOutOfGasTransactions: provided: %d, absolute minimum required: %d", errGasLimitIsLessThanAbsoluteMinimum, args.ExecutorConfig.GasLimitForOutOfGasTransactions, minGasToExecuteSCCalls)
	}

	return nil
}

// Execute will execute one step: get all pending operations, call the filter and send execution transactions
func (executor *scCallExecutor) Execute(ctx context.Context) error {
	return executor.executeOnAllScProxyAddress(ctx, executor.executeScCallForScProxyAddress)
}

func (executor *scCallExecutor) executeScCallForScProxyAddress(ctx context.Context, scProxyAddress string) error {
	executor.log.Info("Executing SC calls for the SC proxy address", "address", scProxyAddress)

	pendingOperations, err := executor.getPendingOperations(ctx, scProxyAddress)
	if err != nil {
		return err
	}

	filteredPendingOperations := executor.filterOperations("scCallExecutor", pendingOperations)

	return executor.executeOperations(ctx, filteredPendingOperations, scProxyAddress)
}

func (executor *scCallExecutor) getPendingOperations(ctx context.Context, scProxyAddress string) (map[uint64]bridgeCore.ProxySCCompleteCallData, error) {
	request := &data.VmValueRequest{
		Address:  scProxyAddress,
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

func (executor *scCallExecutor) parseResponse(response *data.VmValuesResponseData) (map[uint64]bridgeCore.ProxySCCompleteCallData, error) {
	numResponseLines := len(response.Data.ReturnData)
	if numResponseLines%2 != 0 {
		return nil, fmt.Errorf("%w: expected an even number, got %d", errInvalidNumberOfResponseLines, numResponseLines)
	}

	result := make(map[uint64]bridgeCore.ProxySCCompleteCallData, numResponseLines/2)

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

func (executor *scCallExecutor) executeOperations(
	ctx context.Context,
	pendingOperations map[uint64]bridgeCore.ProxySCCompleteCallData,
	scProxyAddress string,
) error {
	networkConfig, err := executor.proxy.GetNetworkConfig(ctx)
	if err != nil {
		return fmt.Errorf("%w while fetching network configs", err)
	}

	for id, callData := range pendingOperations {
		executor.log.Debug("scCallExecutor.executeOperations", "executing ID", id, "call data", callData)
		err = executor.executeOperation(context.Background(), id, callData, networkConfig, scProxyAddress)

		if err != nil {
			return fmt.Errorf("%w for call data: %s", err, callData)
		}
	}

	return nil
}

func (executor *scCallExecutor) executeOperation(
	ctx context.Context,
	id uint64,
	callData bridgeCore.ProxySCCompleteCallData,
	networkConfig *data.NetworkConfig,
	scProxyAddress string,
) error {
	txBuilder := builders.NewTxDataBuilder()
	txBuilder.Function(scProxyCallFunction).ArgInt64(int64(id))

	dataBytes, err := txBuilder.ToDataBytes()
	if err != nil {
		return err
	}

	providedGasLimit, err := executor.codec.ExtractGasLimitFromRawCallData(callData.RawCallData)
	if err != nil {
		executor.log.Warn("scCallExecutor.executeOperation found a non-parsable raw call data",
			"raw call data", callData.RawCallData, "error", err)
		providedGasLimit = 0
	}

	txGasLimit := providedGasLimit + executor.extraGasToExecute
	to, _ := callData.To.AddressAsBech32String()
	if txGasLimit > contractMaxGasLimit {
		// the contract will refund this transaction, so we will use less gas to preserve funds
		executor.log.Warn("setting a lower gas limit for this transaction because it will be refunded",
			"computed gas limit", txGasLimit,
			"data", dataBytes,
			"from", callData.From.Hex(),
			"to", to,
			"token", callData.Token,
			"amount", callData.Amount,
			"nonce", callData.Nonce,
		)
		txGasLimit = executor.gasLimitForOutOfGasTransactions
	}
	if txGasLimit > executor.maxGasLimitToUse {
		executor.log.Warn("can not execute transaction because the provided gas limit on the SC call exceeds "+
			"the maximum gas limit allowance for this executor, WILL SKIP the execution",
			"provided gas limit", txGasLimit,
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

	return executor.transactionExecutor.ExecuteTransaction(ctx, networkConfig, scProxyAddress, scCallTxType, txGasLimit, dataBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *scCallExecutor) IsInterfaceNil() bool {
	return executor == nil
}
