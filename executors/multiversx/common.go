package multiversx

import (
	"context"
	"fmt"
	"strings"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/errors"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
)

type baseExecutor struct {
	scProxyBech32Addresses []string
	proxy                  Proxy
	transactionExecutor    TransactionExecutor
	codec                  Codec
	filter                 ScCallsExecuteFilter
	log                    logger.Logger
}

func (executor *baseExecutor) checkBaseComponents() error {
	if check.IfNil(executor.proxy) {
		return errNilProxy
	}
	if check.IfNil(executor.transactionExecutor) {
		return errNilTransactionExecutor
	}
	if check.IfNil(executor.codec) {
		return errNilCodec
	}
	if check.IfNil(executor.filter) {
		return errNilFilter
	}
	if check.IfNil(executor.log) {
		return errNilLogger
	}

	if len(executor.scProxyBech32Addresses) == 0 {
		return errEmptyListOfBridgeSCProxy
	}

	for _, scProxyAddress := range executor.scProxyBech32Addresses {
		_, err := data.NewAddressFromBech32String(scProxyAddress)
		if err != nil {
			return fmt.Errorf("%w for address %s", err, scProxyAddress)
		}
	}

	return nil
}

func (executor *baseExecutor) executeOnAllScProxyAddress(ctx context.Context, handler func(ctx context.Context, address string) error) error {
	errorStrings := make([]string, 0)
	for _, scProxyAddress := range executor.scProxyBech32Addresses {
		err := handler(ctx, scProxyAddress)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	if len(errorStrings) == 0 {
		return nil
	}

	return fmt.Errorf("errors found during execution: %s", strings.Join(errorStrings, "\n"))
}

func (executor *baseExecutor) filterOperations(component string, pendingOperations map[uint64]bridgeCore.ProxySCCompleteCallData) map[uint64]bridgeCore.ProxySCCompleteCallData {
	result := make(map[uint64]bridgeCore.ProxySCCompleteCallData)
	for id, callData := range pendingOperations {
		if executor.filter.ShouldExecute(callData) {
			result[id] = callData
		}
	}

	executor.log.Debug(component, "input pending ops", len(pendingOperations), "result pending ops", len(result))

	return result
}

func (executor *baseExecutor) executeVmQuery(ctx context.Context, scProxyAddress string, function string) (*data.VmValuesResponseData, error) {
	request := &data.VmValueRequest{
		Address:  scProxyAddress,
		FuncName: function,
	}

	response, err := executor.proxy.ExecuteVMQuery(ctx, request)
	if err != nil {
		executor.log.Error("got error on VMQuery", "FuncName", request.FuncName,
			"Args", request.Args, "SC address", request.Address, "Caller", request.CallerAddr, "error", err)
		return nil, err
	}
	if response == nil || response.Data == nil {
		return nil, errors.NewQueryResponseError(
			emptyErrorCode,
			nilResponseData,
			request.FuncName,
			request.Address,
			request.Args...,
		)
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

	return response, nil
}
