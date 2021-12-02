package elrond

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

const (
	okCodeAfterExecution              = "ok"
	internalError                     = "internal error"
	getCurrentTxBatchFuncName         = "getCurrentTxBatch"
	getTokenIdForErc20AddressFuncName = "getTokenIdForErc20Address"
	getErc20AddressForTokenIdFuncName = "getErc20AddressForTokenId"
)

// ArgsDataGetter is the arguments DTO used in the NewDataGetter constructor
type ArgsDataGetter struct {
	MultisigContractAddress core.AddressHandler
	RelayerAddress          core.AddressHandler
	Proxy                   ElrondProxy
}

type elrondClientDataGetter struct {
	multisigContractAddress core.AddressHandler
	relayerAddress          core.AddressHandler
	proxy                   ElrondProxy
}

// NewDataGetter creates a new instance of the dataGetter type
func NewDataGetter(args ArgsDataGetter) (*elrondClientDataGetter, error) {
	if check.IfNil(args.Proxy) {
		return nil, errNilProxy
	}
	if check.IfNil(args.RelayerAddress) {
		return nil, fmt.Errorf("%w for the RelayerAddress argument", errNilAddressHandler)
	}
	if check.IfNil(args.MultisigContractAddress) {
		return nil, fmt.Errorf("%w for the MultisigContractAddress argument", errNilAddressHandler)
	}

	return &elrondClientDataGetter{
		multisigContractAddress: args.MultisigContractAddress,
		relayerAddress:          args.RelayerAddress,
		proxy:                   args.Proxy,
	}, nil
}

// ExecuteQueryReturningBytes will try to execute the provided query and return the result as slice of byte slices
func (dg *elrondClientDataGetter) ExecuteQueryReturningBytes(ctx context.Context, request *data.VmValueRequest) ([][]byte, error) {
	if request == nil {
		return nil, errNilRequest
	}

	response, err := dg.proxy.ExecuteVMQuery(ctx, request)
	if err != nil {
		return nil, err
	}

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

// ExecuteQueryReturningBool will try to execute the provided query and return the result as bool
func (dg *elrondClientDataGetter) ExecuteQueryReturningBool(ctx context.Context, request *data.VmValueRequest) (bool, error) {
	response, err := dg.ExecuteQueryReturningBytes(ctx, request)
	if err != nil {
		return false, err
	}

	if len(response) == 0 {
		return false, nil
	}
	if len(response[0]) == 0 {
		return false, nil
	}

	result, err := strconv.ParseBool(fmt.Sprintf("%d", response[0][0]))
	if err != nil {
		return false, NewQueryResponseError(
			internalError,
			fmt.Sprintf("error converting the received bytes to bool, %s", err.Error()),
			request.FuncName,
			request.Address,
			request.Args...,
		)
	}

	return result, nil
}

// ExecuteQueryReturningUint64 will try to execute the provided query and return the result as uint64
func (dg *elrondClientDataGetter) ExecuteQueryReturningUint64(ctx context.Context, request *data.VmValueRequest) (uint64, error) {
	response, err := dg.ExecuteQueryReturningBytes(ctx, request)
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
			errNotUint64Bytes.Error(),
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

func (dg *elrondClientDataGetter) executeQueryFromBuilder(ctx context.Context, builder builders.VMQueryBuilder) ([][]byte, error) {
	vmValuesRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return nil, err
	}

	return dg.ExecuteQueryReturningBytes(ctx, vmValuesRequest)
}

// GetCurrentBatchAsDataBytes will assemble a builder and query the proxy for the current pending batch
func (dg *elrondClientDataGetter) GetCurrentBatchAsDataBytes(ctx context.Context) ([][]byte, error) {
	builder := builders.NewVMQueryBuilder().Address(dg.multisigContractAddress).CallerAddress(dg.relayerAddress)
	builder.Function(getCurrentTxBatchFuncName)

	return dg.executeQueryFromBuilder(ctx, builder)
}

// GetTokenIdForErc20Address will assemble a builder and query the proxy for a token id given a specific erc20 address
func (dg *elrondClientDataGetter) GetTokenIdForErc20Address(ctx context.Context, erc20Address []byte) ([][]byte, error) {
	builder := builders.NewVMQueryBuilder().Address(dg.multisigContractAddress).CallerAddress(dg.relayerAddress)
	builder.Function(getTokenIdForErc20AddressFuncName)
	builder.ArgBytes(erc20Address)

	return dg.executeQueryFromBuilder(ctx, builder)
}

// GetERC20AddressForTokenId will assemble a builder and query the proxy for an erc20 address given a specific token id
func (dg *elrondClientDataGetter) GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error) {
	builder := builders.NewVMQueryBuilder().Address(dg.multisigContractAddress).CallerAddress(dg.relayerAddress)
	builder.Function(getErc20AddressForTokenIdFuncName)
	builder.ArgBytes(tokenId)
	return dg.executeQueryFromBuilder(ctx, builder)
}

// IsInterfaceNil returns true if there is no value under the interface
func (dg *elrondClientDataGetter) IsInterfaceNil() bool {
	return dg == nil
}
