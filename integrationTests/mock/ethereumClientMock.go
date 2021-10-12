package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
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
	id         int
}

// EthereumMockClient represents a mock Ethereum node that opens a http client
type EthereumMockClient struct {
	*accountsMap
	httpServer *httptest.Server
}

// NewEthereumMockClient creates a new Ethereum Mock Client
func NewEthereumMockClient() *EthereumMockClient {
	accounts := newAccountsMap(make(map[string]*api.AccountResponse), make(map[string]*Contract))
	emc := &EthereumMockClient{
		accountsMap: accounts,
	}

	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		bodyBytes := getBodyAsByteSlice(req)
		ethRequest := &EthRequest{}
		err := json.Unmarshal(bodyBytes, ethRequest)
		if err != nil {
			writeEthereumResponse(rw, http.StatusInternalServerError, nil,
				fmt.Errorf("EtherumMockClient: error %w, route %s, body: %s", err, req.RequestURI, string(bodyBytes)), 0)
			return
		}

		parsed, err := emc.parseEthRequest(ethRequest)
		if err != nil {
			writeEthereumResponse(rw, http.StatusInternalServerError, nil,
				fmt.Errorf("EtherumMockClient: error %w, route %s, body: %s", err, req.RequestURI, string(bodyBytes)), ethRequest.ID)
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
	mapParams, err := emc.getParamsMap(request)
	if err != nil {
		return nil, err
	}

	per := &parsedEthRequest{
		functionId: mapParams["data"].(string),
		address:    mapParams["to"].(string),
		id:         request.ID,
	}

	return per, nil
}

func (emc *EthereumMockClient) getParamsMap(request *EthRequest) (map[string]interface{}, error) {
	for i := 0; i < len(request.Params); i++ {
		p := request.Params[i]
		mapParams, isMap := p.(map[string]interface{})
		if !isMap {
			continue
		}

		return mapParams, nil
	}

	if len(request.Params) < 1 {
		return nil, errors.New("parseEthRequest function: error parsing params field")
	}

	//made the method calls like eth_getCode follow the same principle as any other function
	mapParams := map[string]interface{}{
		"to":   request.Params[0],
		"data": request.Method,
	}

	return mapParams, nil
}

func (emc *EthereumMockClient) processParsedEthRequest(rw http.ResponseWriter, params *parsedEthRequest) {
	contract, found := emc.GetContract(params.address)
	if !found {
		writeEthereumResponse(rw,
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("processParsedEthRequest: contact %s not found", params.address),
			params.id)
		return
	}

	handler := contract.GetHandler(params.functionId)
	if handler == nil {
		writeEthereumResponse(rw,
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("processParsedEthRequest: function %s for contact %s not found", params.functionId, params.address),
			params.id)
		return
	}

	results, err := handler("", "", "")
	if err != nil {
		writeEthereumResponse(rw,
			http.StatusInternalServerError,
			nil,
			fmt.Errorf("%w in processParsedEthRequest: function %s for contact %s", err, params.functionId, params.address),
			params.id)
		return
	}

	writeEthereumResponse(rw, http.StatusOK, results, nil, params.id)
}
