package multiversx

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	getRefundTransactionsFunction    = "getRefundTransactions"
	executeRefundTransactionFunction = "executeRefundTransaction"
	refundTxType                     = "refund"
	nilResponseData                  = "nil response data"
	emptyErrorCode                   = ""
)

// ArgsRefundExecutor represents the DTO struct for creating a new instance of type refundExecutor
type ArgsRefundExecutor struct {
	TransactionExecutor    TransactionExecutor
	ScProxyBech32Addresses []string
	Proxy                  Proxy
	Codec                  Codec
	Filter                 ScCallsExecuteFilter
	Log                    logger.Logger
	RefundConfig           config.RefundExecutorConfig
}

type refundExecutor struct {
	*baseExecutor
	gasToExecute uint64
}

// NewRefundExecutor creates a new instance of type refundExecutor
func NewRefundExecutor(args ArgsRefundExecutor) (*refundExecutor, error) {
	if args.RefundConfig.GasToExecute < minGasToExecuteSCCalls {
		return nil, fmt.Errorf("%w for GasToExecute: provided: %d, absolute minimum required: %d", errGasLimitIsLessThanAbsoluteMinimum, args.RefundConfig.GasToExecute, minGasToExecuteSCCalls)
	}

	executor := &refundExecutor{
		baseExecutor: &baseExecutor{
			scProxyBech32Addresses: args.ScProxyBech32Addresses,
			proxy:                  args.Proxy,
			transactionExecutor:    args.TransactionExecutor,
			codec:                  args.Codec,
			filter:                 args.Filter,
			log:                    args.Log,
			ttlForFailedRefundID:   time.Duration(args.RefundConfig.TTLForFailedRefundIdInSeconds) * time.Second,
			failedRefundMap:        make(map[uint64]time.Time),
		},
		gasToExecute: args.RefundConfig.GasToExecute,
	}

	err := executor.checkBaseComponents()
	if err != nil {
		return nil, err
	}

	return executor, nil
}

// Execute will execute one step: get all pending operations, call the filter and send execution transactions
func (executor *refundExecutor) Execute(ctx context.Context) error {
	executor.cleanupTTLCache(refundTxType)

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
	response, err := executor.executeVmQuery(ctx, scProxyAddress, getRefundTransactionsFunction)
	if err != nil {
		return nil, err
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

	err = executor.transactionExecutor.ExecuteTransaction(ctx, networkConfig, scProxyAddress, refundTxType, executor.gasToExecute, dataBytes)
	if err != nil {
		executor.addFailed(id)
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (executor *refundExecutor) IsInterfaceNil() bool {
	return executor == nil
}
