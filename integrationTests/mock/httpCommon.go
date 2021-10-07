package mock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
)

func getBodyAsByteSlice(req *http.Request) []byte {
	defer func() {
		if req.Body != nil {
			return
		}

		_ = req.Body.Close()
	}()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("getBodyAsByteSlice", "error", err)
		return nil
	}

	return body
}

func writeElrondResponse(rw http.ResponseWriter, httpCode int, responseName string, data interface{}, err error) {
	var wrappedData interface{}
	code := shared.ReturnCodeInternalError
	if httpCode == http.StatusOK {
		wrappedData = gin.H{responseName: data}
		code = shared.ReturnCodeSuccess
	}

	errString := ""
	if err != nil {
		errString = err.Error()
	}

	response := shared.GenericAPIResponse{
		Data:  wrappedData,
		Error: errString,
		Code:  code,
	}

	buff, _ := json.Marshal(response)
	rw.WriteHeader(httpCode)
	_, _ = rw.Write(buff)
}

func writeEthereumResponse(rw http.ResponseWriter, httpCode int, results [][]byte, err error, id int) {
	response := prepareEthereumJsonRpcMessage(httpCode, results, err, id)
	buff, err := json.Marshal(response)
	if err != nil {
		errorResponse := prepareEthereumJsonRpcMessage(http.StatusInternalServerError, nil, err, id)
		buff, err = json.Marshal(errorResponse)
		log.LogIfError(err) //if this happens, there is a programming error in json RPC message construction
	}

	rw.WriteHeader(httpCode)
	_, _ = rw.Write(buff)
}

func prepareEthereumJsonRpcMessage(httpCode int, results [][]byte, err error, id int) *jsonrpcMessage {
	response := &jsonrpcMessage{
		Version: "2.0",
		ID:      writeRawMessage(fmt.Sprintf("%d", id)),
		Method:  "",
		Params:  nil,
		Error:   nil,
	}

	if err != nil {
		response.Error = createJsonError(httpCode, err)

		return response
	}

	resultsBuff := make([]byte, 0)
	for _, res := range results {
		resultsBuff = append(resultsBuff, res...)
	}

	messageResults := hexutil.Bytes(resultsBuff)
	buff, err := messageResults.MarshalText()
	if err != nil {
		response.Error = createJsonError(httpCode, err)

		return response
	}

	response.Result = writeRawMessage(string(buff))

	return response
}

func createJsonError(httpCode int, err error) *jsonError {
	return &jsonError{
		Code:    httpCode,
		Message: err.Error(),
		Data:    nil,
	}
}

func writeRawMessage(data string) json.RawMessage {
	return []byte("\"" + data + "\"")
}
