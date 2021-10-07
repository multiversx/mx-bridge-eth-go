package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	vmData "github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type vmProcessorMock struct {
	accounts *accountsMap
}

func newVmProcessorMock(accounts *accountsMap) *vmProcessorMock {
	return &vmProcessorMock{
		accounts: accounts,
	}
}

func (vm *vmProcessorMock) processVmValues(rw http.ResponseWriter, req *http.Request) {
	uri := req.RequestURI
	switch uri {
	case vmValuesHexEndpointName:
		vm.processHexVmValues(rw, req)
	case vmValuesStringEndpointName:
		vm.processStringVmValues(rw, req)
	case vmValuesIntEndpointName:
		vm.processIntVmValues(rw, req)
	case vmValuesQueryEndpointName:
		vm.processQueryVmValues(rw, req)
	default:
		writeElrondResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("unknown URI in ElrondMockClient, uri %s", uri))
	}
}

func (vm *vmProcessorMock) processHexVmValues(rw http.ResponseWriter, req *http.Request) {
	results, err := vm.processVmValuesRequest(req)
	if err != nil {
		log.Error("vmProcessorMock.processVmValuesRequest", "error", err)
		writeElrondResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	returnData, err := results.GetFirstReturnData(vmData.AsHex)
	if err != nil {
		log.Error("vmProcessorMock.GetFirstReturnData", "error", err)
		writeElrondResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	writeElrondResponse(rw, http.StatusOK, "data", returnData, nil)
}

func (vm *vmProcessorMock) processStringVmValues(rw http.ResponseWriter, req *http.Request) {
	results, err := vm.processVmValuesRequest(req)
	if err != nil {
		log.Error("vmProcessorMock.processVmValuesRequest", "error", err)
		writeElrondResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	returnData, err := results.GetFirstReturnData(vmData.AsString)
	if err != nil {
		log.Error("vmProcessorMock.GetFirstReturnData", "error", err)
		writeElrondResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	writeElrondResponse(rw, http.StatusOK, "data", returnData, nil)
}

func (vm *vmProcessorMock) processIntVmValues(rw http.ResponseWriter, req *http.Request) {
	results, err := vm.processVmValuesRequest(req)
	if err != nil {
		log.Error("vmProcessorMock.processVmValuesRequest", "error", err)
		writeElrondResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	returnData, err := results.GetFirstReturnData(vmData.AsBigIntString)
	if err != nil {
		log.Error("vmProcessorMock.GetFirstReturnData", "error", err)
		writeElrondResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	writeElrondResponse(rw, http.StatusOK, "data", returnData, nil)
}

func (vm *vmProcessorMock) processQueryVmValues(rw http.ResponseWriter, req *http.Request) {
	results, err := vm.processVmValuesRequest(req)
	if err != nil {
		log.Error("vmProcessorMock.processVmValuesRequest", "error", err)
		writeElrondResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	writeElrondResponse(rw, http.StatusOK, "data", results, nil)
}

func (vm *vmProcessorMock) processVmValuesRequest(req *http.Request) (*vmData.VMOutputApi, error) {
	bodyBytes := getBodyAsByteSlice(req)
	vmRequest, err := vm.parseVmValuesRequest(bodyBytes)
	if err != nil {
		return nil, err
	}

	handler, err := vm.getHandler(vmRequest.ScAddress, vmRequest.FuncName)
	if err != nil {
		return nil, err
	}

	results, err := handler(vmRequest.CallerAddr, vmRequest.CallValue, vmRequest.Args...)
	returnCode := vmcommon.Ok
	returnMessage := ""
	if err != nil {
		returnCode = vmcommon.UserError
		returnMessage = err.Error()
	}
	vmOutput := &vmData.VMOutputApi{
		ReturnData:      results,
		ReturnCode:      returnCode.String(),
		ReturnMessage:   returnMessage,
		GasRemaining:    0,
		GasRefund:       nil,
		OutputAccounts:  nil,
		DeletedAccounts: nil,
		TouchedAccounts: nil,
		Logs:            nil,
	}

	return vmOutput, nil
}

func (vm *vmProcessorMock) parseVmValuesRequest(buff []byte) (*vmValues.VMValueRequest, error) {
	request := &vmValues.VMValueRequest{}
	err := json.Unmarshal(buff, request)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (vm *vmProcessorMock) getHandler(address string, function string) (ContractHandler, error) {
	contract, found := vm.accounts.GetContract(address)
	if !found {
		return nil, errors.New(fmt.Sprintf("contract not found for address %s", address))
	}

	handler := contract.GetHandler(function)
	if handler == nil {
		return nil, errors.New(fmt.Sprintf("handler %s not found in contract for address %s", function, address))
	}

	return handler, nil
}
