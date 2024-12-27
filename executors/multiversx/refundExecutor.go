package multiversx

import (
	"context"
	"fmt"
	"math/big"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/errors"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	getRefundTransactionsFunction    = "getRefundTransactions"
	executeRefundTransactionFunction = "executeRefundTransaction"
	refundTxType                     = "refund"
)

// ArgsRefundExecutor represents the DTO struct for creating a new instance of type refundExecutor
type ArgsRefundExecutor struct {
	TransactionExecutor    TransactionExecutor
	ScProxyBech32Addresses []string
	Proxy                  Proxy
	Codec                  Codec
	Filter                 ScCallsExecuteFilter
	Log                    logger.Logger
	GasToExecute           uint64
}

type refundExecutor struct {
	*baseExecutor
	gasToExecute uint64
}

// NewRefundExecutor creates a new instance of type refundExecutor
func NewRefundExecutor(args ArgsRefundExecutor) (*refundExecutor, error) {
	if args.GasToExecute < minGasToExecuteSCCalls {
		return nil, fmt.Errorf("%w for GasToExecute: provided: %d, absolute minimum required: %d", errGasLimitIsLessThanAbsoluteMinimum, args.GasToExecute, minGasToExecuteSCCalls)
	}

	executor := &refundExecutor{
		baseExecutor: &baseExecutor{
			scProxyBech32Addresses: args.ScProxyBech32Addresses,
			proxy:                  args.Proxy,
			transactionExecutor:    args.TransactionExecutor,
			codec:                  args.Codec,
			filter:                 args.Filter,
			log:                    args.Log,
		},
		gasToExecute: args.GasToExecute,
	}

	err := executor.checkBaseComponents()
	if err != nil {
		return nil, err
	}

	return executor, nil
}

// Execute will execute one step: get all pending operations, call the filter and send execution transactions
func (executor *refundExecutor) Execute(ctx context.Context) error {
	return executor.executeOnAllScProxyAddress(ctx, executor.executeRefundForScProxyAddress)
}

func (executor *refundExecutor) executeRefundForScProxyAddress(ctx context.Context, scProxyAddress string) error {
	executor.log.Info("Executing refunds for the SC proxy address", "address", scProxyAddress)

	pendingOperations, err := executor.getPendingRefunds(ctx, scProxyAddress)
	if err != nil {
		return err
	}

	filteredPendingOperations := executor.filterOperations("refundExecutor", pendingOperations)

	return executor.executeRefunds(ctx, filteredPendingOperations, scProxyAddress)
}

func (executor *refundExecutor) getPendingRefunds(ctx context.Context, scProxyAddress string) (map[uint64]bridgeCore.ProxySCCompleteCallData, error) {
	request := &data.VmValueRequest{
		Address:  scProxyAddress,
		FuncName: getRefundTransactionsFunction,
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

func (executor *refundExecutor) parseResponse(response *data.VmValuesResponseData) (map[uint64]bridgeCore.ProxySCCompleteCallData, error) {
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

func (executor *refundExecutor) executeRefunds(ctx context.Context, refundIDs map[uint64]bridgeCore.ProxySCCompleteCallData, scProxyAddress string) error {
	networkConfig, err := executor.proxy.GetNetworkConfig(ctx)
	if err != nil {
		return fmt.Errorf("%w while fetching network configs", err)
	}

	for id := range refundIDs {
		executor.log.Debug("refundExecutor.executeRefunds", "executing refund ID", id)
		err = executor.executeOperation(context.Background(), id, networkConfig, scProxyAddress)

		if err != nil {
			return fmt.Errorf("%w for refund ID: %d", err, id)
		}
	}

	return nil
}

func (executor *refundExecutor) executeOperation(
	ctx context.Context,
	id uint64,
	networkConfig *data.NetworkConfig,
	scProxyAddress string,
) error {
	txBuilder := builders.NewTxDataBuilder()
	txBuilder.Function(executeRefundTransactionFunction).ArgInt64(int64(id))

	dataBytes, err := txBuilder.ToDataBytes()
	if err != nil {
		return err
	}

	return executor.transactionExecutor.ExecuteTransaction(ctx, networkConfig, scProxyAddress, refundTxType, executor.gasToExecute, dataBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *refundExecutor) IsInterfaceNil() bool {
	return executor == nil
}
