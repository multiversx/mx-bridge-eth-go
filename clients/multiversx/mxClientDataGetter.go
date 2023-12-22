package multiversx

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	okCodeAfterExecution                                      = "ok"
	internalError                                             = "internal error"
	getCurrentTxBatchFuncName                                 = "getCurrentTxBatch"
	wasTransferActionProposedFuncName                         = "wasTransferActionProposed"
	wasActionExecutedFuncName                                 = "wasActionExecuted"
	getActionIdForTransferBatchFuncName                       = "getActionIdForTransferBatch"
	wasSetCurrentTransactionBatchStatusActionProposedFuncName = "wasSetCurrentTransactionBatchStatusActionProposed"
	getStatusesAfterExecutionFuncName                         = "getStatusesAfterExecution"
	getActionIdForSetCurrentTransactionBatchStatusFuncName    = "getActionIdForSetCurrentTransactionBatchStatus"
	getTokenIdForErc20AddressFuncName                         = "getTokenIdForErc20Address"
	getErc20AddressForTokenIdFuncName                         = "getErc20AddressForTokenId"
	quorumReachedFuncName                                     = "quorumReached"
	getLastExecutedEthBatchIdFuncName                         = "getLastExecutedEthBatchId"
	getLastExecutedEthTxId                                    = "getLastExecutedEthTxId"
	signedFuncName                                            = "signed"
	getAllStakedRelayersFuncName                              = "getAllStakedRelayers"
	isPausedFuncName                                          = "isPaused"
)

// ArgsMXClientDataGetter is the arguments DTO used in the NewMXClientDataGetter constructor
type ArgsMXClientDataGetter struct {
	MultisigContractAddress core.AddressHandler
	RelayerAddress          core.AddressHandler
	Proxy                   Proxy
	Log                     logger.Logger
}

type mxClientDataGetter struct {
	multisigContractAddress       core.AddressHandler
	bech32MultisigContractAddress string
	relayerAddress                core.AddressHandler
	proxy                         Proxy
	log                           logger.Logger
	mutNodeStatus                 sync.Mutex
	wasShardIDFetched             bool
	shardID                       uint32
}

// NewMXClientDataGetter creates a new instance of the dataGetter type
func NewMXClientDataGetter(args ArgsMXClientDataGetter) (*mxClientDataGetter, error) {
	if check.IfNil(args.Log) {
		return nil, errNilLogger
	}
	if check.IfNil(args.Proxy) {
		return nil, errNilProxy
	}
	if check.IfNil(args.RelayerAddress) {
		return nil, fmt.Errorf("%w for the RelayerAddress argument", errNilAddressHandler)
	}
	if check.IfNil(args.MultisigContractAddress) {
		return nil, fmt.Errorf("%w for the MultisigContractAddress argument", errNilAddressHandler)
	}
	bech32Address, err := args.MultisigContractAddress.AddressAsBech32String()
	if err != nil {
		return nil, fmt.Errorf("%w for %x", err, args.MultisigContractAddress.AddressBytes())
	}

	return &mxClientDataGetter{
		multisigContractAddress:       args.MultisigContractAddress,
		bech32MultisigContractAddress: bech32Address,
		relayerAddress:                args.RelayerAddress,
		proxy:                         args.Proxy,
		log:                           args.Log,
	}, nil
}

// ExecuteQueryReturningBytes will try to execute the provided query and return the result as slice of byte slices
func (dataGetter *mxClientDataGetter) ExecuteQueryReturningBytes(ctx context.Context, request *data.VmValueRequest) ([][]byte, error) {
	if request == nil {
		return nil, errNilRequest
	}

	response, err := dataGetter.proxy.ExecuteVMQuery(ctx, request)
	if err != nil {
		dataGetter.log.Error("got error on VMQuery", "FuncName", request.FuncName,
			"Args", request.Args, "SC address", request.Address, "Caller", request.CallerAddr, "error", err)
		return nil, err
	}
	dataGetter.log.Debug("executed VMQuery", "FuncName", request.FuncName,
		"Args", request.Args, "SC address", request.Address, "Caller", request.CallerAddr,
		"response.ReturnCode", response.Data.ReturnCode,
		"response.ReturnData", fmt.Sprintf("%+v", response.Data.ReturnData))
	if response.Data.ReturnCode != okCodeAfterExecution {
		return nil, NewQueryResponseError(
			response.Data.ReturnCode,
			response.Data.ReturnMessage,
			request.FuncName,
			request.Address,
			request.Args...,
		)
	}
	return response.Data.ReturnData, nil
}

// GetCurrentNonce will get from the shard containing the multisig contract the latest block's nonce
func (dataGetter *mxClientDataGetter) GetCurrentNonce(ctx context.Context) (uint64, error) {
	shardID, err := dataGetter.getShardID(ctx)
	if err != nil {
		return 0, err
	}

	nodeStatus, err := dataGetter.proxy.GetNetworkStatus(ctx, shardID)
	if err != nil {
		return 0, err
	}
	if nodeStatus == nil {
		return 0, errNilNodeStatusResponse
	}

	return nodeStatus.Nonce, nil
}

func (dataGetter *mxClientDataGetter) getShardID(ctx context.Context) (uint32, error) {
	dataGetter.mutNodeStatus.Lock()
	defer dataGetter.mutNodeStatus.Unlock()

	if dataGetter.wasShardIDFetched {
		return dataGetter.shardID, nil
	}

	var err error
	dataGetter.shardID, err = dataGetter.proxy.GetShardOfAddress(ctx, dataGetter.bech32MultisigContractAddress)
	if err == nil {
		dataGetter.wasShardIDFetched = true
	}

	return dataGetter.shardID, err
}

// ExecuteQueryReturningBool will try to execute the provided query and return the result as bool
func (dataGetter *mxClientDataGetter) ExecuteQueryReturningBool(ctx context.Context, request *data.VmValueRequest) (bool, error) {
	response, err := dataGetter.ExecuteQueryReturningBytes(ctx, request)
	if err != nil {
		return false, err
	}

	if len(response) == 0 {
		return false, nil
	}

	return dataGetter.parseBool(response[0], request.FuncName, request.Address, request.Args...)
}

func (dataGetter *mxClientDataGetter) parseBool(buff []byte, funcName string, address string, args ...string) (bool, error) {
	if len(buff) == 0 {
		return false, nil
	}

	result, err := strconv.ParseBool(fmt.Sprintf("%d", buff[0]))
	if err != nil {
		return false, NewQueryResponseError(
			internalError,
			fmt.Sprintf("error converting the received bytes to bool, %s", err.Error()),
			funcName,
			address,
			args...,
		)
	}

	return result, nil
}

// ExecuteQueryReturningUint64 will try to execute the provided query and return the result as uint64
func (dataGetter *mxClientDataGetter) ExecuteQueryReturningUint64(ctx context.Context, request *data.VmValueRequest) (uint64, error) {
	response, err := dataGetter.ExecuteQueryReturningBytes(ctx, request)
	if err != nil {
		return 0, err
	}

	if len(response) == 0 {
		return 0, nil
	}
	if len(response[0]) == 0 {
		return 0, nil
	}

	num, err := parseUInt64FromByteSlice(response[0])
	if err != nil {
		return 0, NewQueryResponseError(
			internalError,
			err.Error(),
			request.FuncName,
			request.Address,
			request.Args...,
		)
	}

	return num, nil
}

func parseUInt64FromByteSlice(bytes []byte) (uint64, error) {
	num := big.NewInt(0).SetBytes(bytes)
	if !num.IsUint64() {
		return 0, errNotUint64Bytes
	}

	return num.Uint64(), nil
}

func (dataGetter *mxClientDataGetter) executeQueryFromBuilder(ctx context.Context, builder builders.VMQueryBuilder) ([][]byte, error) {
	vmValuesRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return nil, err
	}

	return dataGetter.ExecuteQueryReturningBytes(ctx, vmValuesRequest)
}

func (dataGetter *mxClientDataGetter) executeQueryUint64FromBuilder(ctx context.Context, builder builders.VMQueryBuilder) (uint64, error) {
	vmValuesRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return 0, err
	}

	return dataGetter.ExecuteQueryReturningUint64(ctx, vmValuesRequest)
}

func (dataGetter *mxClientDataGetter) executeQueryBoolFromBuilder(ctx context.Context, builder builders.VMQueryBuilder) (bool, error) {
	vmValuesRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return false, err
	}

	return dataGetter.ExecuteQueryReturningBool(ctx, vmValuesRequest)
}

func (dataGetter *mxClientDataGetter) createDefaultVmQueryBuilder() builders.VMQueryBuilder {
	return builders.NewVMQueryBuilder().Address(dataGetter.multisigContractAddress).CallerAddress(dataGetter.relayerAddress)
}

// GetCurrentBatchAsDataBytes will assemble a builder and query the proxy for the current pending batch
func (dataGetter *mxClientDataGetter) GetCurrentBatchAsDataBytes(ctx context.Context) ([][]byte, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(getCurrentTxBatchFuncName)

	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// GetTokenIdForErc20Address will assemble a builder and query the proxy for a token id given a specific erc20 address
func (dataGetter *mxClientDataGetter) GetTokenIdForErc20Address(ctx context.Context, erc20Address []byte) ([][]byte, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(getTokenIdForErc20AddressFuncName)
	builder.ArgBytes(erc20Address)

	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// GetERC20AddressForTokenId will assemble a builder and query the proxy for an erc20 address given a specific token id
func (dataGetter *mxClientDataGetter) GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(getErc20AddressForTokenIdFuncName)
	builder.ArgBytes(tokenId)
	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// WasProposedTransfer returns true if the transfer action proposed was triggered
func (dataGetter *mxClientDataGetter) WasProposedTransfer(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
	if batch == nil {
		return false, clients.ErrNilBatch
	}

	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(wasTransferActionProposedFuncName).ArgInt64(int64(batch.ID))
	addBatchInfo(builder, batch)

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// WasExecuted returns true if the provided actionID was executed or not
func (dataGetter *mxClientDataGetter) WasExecuted(ctx context.Context, actionID uint64) (bool, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(wasActionExecutedFuncName).ArgInt64(int64(actionID))

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// GetActionIDForProposeTransfer returns the action ID for the proposed transfer operation
func (dataGetter *mxClientDataGetter) GetActionIDForProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
	if batch == nil {
		return 0, clients.ErrNilBatch
	}

	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(getActionIdForTransferBatchFuncName).ArgInt64(int64(batch.ID))
	addBatchInfo(builder, batch)

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// WasProposedSetStatus returns true if the proposed set status was triggered
func (dataGetter *mxClientDataGetter) WasProposedSetStatus(ctx context.Context, batch *clients.TransferBatch) (bool, error) {
	if batch == nil {
		return false, clients.ErrNilBatch
	}

	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(wasSetCurrentTransactionBatchStatusActionProposedFuncName).ArgInt64(int64(batch.ID))
	for _, stat := range batch.Statuses {
		builder.ArgBytes([]byte{stat})
	}

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// GetTransactionsStatuses will return the transactions statuses from the batch ID
func (dataGetter *mxClientDataGetter) GetTransactionsStatuses(ctx context.Context, batchID uint64) ([]byte, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(getStatusesAfterExecutionFuncName).ArgInt64(int64(batchID))

	values, err := dataGetter.executeQueryFromBuilder(ctx, builder)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("%w for batch ID %v", errNoStatusForBatchID, batchID)
	}

	isFinished, err := dataGetter.parseBool(values[0], getStatusesAfterExecutionFuncName, dataGetter.bech32MultisigContractAddress)
	if err != nil {
		return nil, err
	}
	if !isFinished {
		return nil, fmt.Errorf("%w for batch ID %v", errBatchNotFinished, batchID)
	}

	results := make([]byte, len(values)-1)
	for i := 1; i < len(values); i++ {
		results[i-1], err = getStatusFromBuff(values[i])
		if err != nil {
			return nil, fmt.Errorf("%w for result index %d", err, i-1)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%w status is finished, no results are given", errMalformedBatchResponse)
	}

	return results, nil
}

// GetActionIDForSetStatusOnPendingTransfer returns the action ID for setting the status on the pending transfer batch
func (dataGetter *mxClientDataGetter) GetActionIDForSetStatusOnPendingTransfer(ctx context.Context, batch *clients.TransferBatch) (uint64, error) {
	if batch == nil {
		return 0, clients.ErrNilBatch
	}

	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(getActionIdForSetCurrentTransactionBatchStatusFuncName).ArgInt64(int64(batch.ID))
	for _, stat := range batch.Statuses {
		builder.ArgBytes([]byte{stat})
	}

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// QuorumReached returns true if the provided action ID reached the set quorum
func (dataGetter *mxClientDataGetter) QuorumReached(ctx context.Context, actionID uint64) (bool, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(quorumReachedFuncName).ArgInt64(int64(actionID))

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// GetLastExecutedEthBatchID returns the last executed Ethereum batch ID
func (dataGetter *mxClientDataGetter) GetLastExecutedEthBatchID(ctx context.Context) (uint64, error) {
	builder := dataGetter.createDefaultVmQueryBuilder().Function(getLastExecutedEthBatchIdFuncName)

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// GetLastExecutedEthTxID returns the last executed Ethereum deposit ID
func (dataGetter *mxClientDataGetter) GetLastExecutedEthTxID(ctx context.Context) (uint64, error) {
	builder := dataGetter.createDefaultVmQueryBuilder().Function(getLastExecutedEthTxId)

	return dataGetter.executeQueryUint64FromBuilder(ctx, builder)
}

// WasSigned returns true if the action was already signed by the current relayer
func (dataGetter *mxClientDataGetter) WasSigned(ctx context.Context, actionID uint64) (bool, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(signedFuncName).ArgAddress(dataGetter.relayerAddress).ArgInt64(int64(actionID))

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

// GetAllStakedRelayers returns all staked relayers defined in MultiversX SC
func (dataGetter *mxClientDataGetter) GetAllStakedRelayers(ctx context.Context) ([][]byte, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(getAllStakedRelayersFuncName)

	return dataGetter.executeQueryFromBuilder(ctx, builder)
}

// IsPaused returns true if the multisig contract is paused
func (dataGetter *mxClientDataGetter) IsPaused(ctx context.Context) (bool, error) {
	builder := dataGetter.createDefaultVmQueryBuilder()
	builder.Function(isPausedFuncName)

	return dataGetter.executeQueryBoolFromBuilder(ctx, builder)
}

func getStatusFromBuff(buff []byte) (byte, error) {
	if len(buff) == 0 {
		return 0, errMalformedBatchResponse
	}

	return buff[len(buff)-1], nil
}

func addBatchInfo(builder builders.VMQueryBuilder, batch *clients.TransferBatch) {
	for _, dt := range batch.Deposits {
		builder.ArgBytes(dt.FromBytes).
			ArgBytes(dt.ToBytes).
			ArgBytes(dt.ConvertedTokenBytes).
			ArgBigInt(dt.Amount).
			ArgInt64(int64(dt.Nonce))
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (dataGetter *mxClientDataGetter) IsInterfaceNil() bool {
	return dataGetter == nil
}
