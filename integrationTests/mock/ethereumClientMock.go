package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// EthRequest represents the struct for each ethereum request
type EthRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type parsedEthRequest struct {
	functionId string
	address    string
}

// EthereumMockClient represents a mock Ethereum node that opens a http client
type EthereumMockClient struct {
	*accountsMap
	httpServer *httptest.Server
}

// NewEthereumMockClient creates a new Ethereum Mock Client
func NewEthereumMockClient() *EthereumMockClient {
	accounts := newAccountsMap()
	emc := &EthereumMockClient{
		accountsMap: accounts,
	}

	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		bodyBytes := getBodyAsByteSlice(req)
		ethRequest := &EthRequest{}
		err := json.Unmarshal(bodyBytes, ethRequest)
		if err != nil {
			writeResponse(rw, http.StatusInternalServerError, "", nil,
				fmt.Errorf("EtherumMockClient: error %w, route %s, body: %s", err, req.RequestURI, string(bodyBytes)))
			return
		}

		parsed, err := emc.parseEthRequest(ethRequest)
		if err != nil {
			writeResponse(rw, http.StatusInternalServerError, "", nil,
				fmt.Errorf("EtherumMockClient: error %w, route %s, body: %s", err, req.RequestURI, string(bodyBytes)))
			return
		}

		emc.processParsedEthRequest(rw, parsed)
	}))
	emc.httpServer = httpServer

	return emc
}

// URL returns the connection url for this test server
func (emc *EthereumMockClient) URL() string {
	return emc.httpServer.URL
}

func (emc *EthereumMockClient) parseEthRequest(request *EthRequest) (*parsedEthRequest, error) {
	if request == nil {
		return nil, errors.New("parseEthRequest function: nil EthRequest instance")
	}
	if len(request.Params) < 1 {
		return nil, errors.New("parseEthRequest function: empty params list")
	}
	params := request.Params[0]
	mapParams, isMap := params.(map[string]interface{})
	if !isMap {
		return nil, errors.New("parseEthRequest function: not a map of params on first item")
	}

	per := &parsedEthRequest{
		functionId: mapParams["data"].(string),
		address:    mapParams["to"].(string),
	}

	return per, nil
}

func (emc *EthereumMockClient) processParsedEthRequest(rw http.ResponseWriter, params *parsedEthRequest) {
	contract, found := emc.GetContract(params.address)
	if !found {
		writeResponse(rw, http.StatusInternalServerError, "", nil,
			fmt.Errorf("processParsedEthRequest: contact %s not found", params.address))
		return
	}

	_ = contract
	//TODO create a valid response
	//response := &jsonrpcMessage{
	//    Version: "1",
	//    ID:      nil,
	//    Method:  "e",
	//    Params:  nil,
	//    Error:   &jsonError{
	//        Code:    0,
	//        Message: "ok",
	//        Data:    nil,
	//    },
	//    Result:  nil,
	//}
	//
	//buff, err := json.Marshal(response)
	//log.LogIfError(err)
	//rw.Write(buff)
	//
	//

	log.Warn("parsed eth request", "params", params)
}
