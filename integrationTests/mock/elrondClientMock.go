package mock

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
)

const addressEndpointName = "/address/"
const vmValuesEndpointName = "/vm-values/"
const vmValuesHexEndpointName = "/vm-values/hex"
const vmValuesStringEndpointName = "/vm-values/string"
const vmValuesIntEndpointName = "/vm-values/int"
const vmValuesQueryEndpointName = "/vm-values/query"
const sendTransactionEndpointName = "/transaction/send"
const networkConfigEndpointName = "/network/config"

var log = logger.GetOrCreate("integrationTests/mock")

// ElrondMockClient represents a mock Elrond Proxy that opens a http client
type ElrondMockClient struct {
	*accountsMap
	*transactionHandlerMock
	httpServer  *httptest.Server
	vmProcessor *vmProcessorMock
}

// NewElrondMockClient creates a new Elrond Mock Client
func NewElrondMockClient() *ElrondMockClient {
	accounts := newAccountsMap()
	emc := &ElrondMockClient{
		accountsMap:            accounts,
		vmProcessor:            newVmProcessorMock(accounts),
		transactionHandlerMock: newTransactionHandlerMock(),
	}

	httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.RequestURI, addressEndpointName) {
			emc.processAddress(rw, req)
			return
		}
		if strings.Contains(req.RequestURI, vmValuesEndpointName) {
			emc.vmProcessor.processVmValues(rw, req)
			return
		}
		if strings.Contains(req.RequestURI, sendTransactionEndpointName) {
			emc.transactionHandlerMock.processSendTransaction(rw, req)
			return
		}
		if strings.Contains(req.RequestURI, networkConfigEndpointName) {
			emc.processNetworkConfig(rw, req)
			return
		}

		writeResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: unimplemented route %s", req.RequestURI))
	}))
	emc.httpServer = httpServer

	return emc
}

// URL returns the connection url for this test server
func (emc *ElrondMockClient) URL() string {
	return emc.httpServer.URL
}

func (emc *ElrondMockClient) processAddress(rw http.ResponseWriter, req *http.Request) {
	address := req.RequestURI[len(addressEndpointName):]
	account := emc.GetAccount(address)

	writeResponse(rw, http.StatusOK, "account", account, nil)
}

func (emc *ElrondMockClient) processNetworkConfig(rw http.ResponseWriter, _ *http.Request) {
	metrics := make(map[string]interface{})
	metrics[common.MetricNumShardsWithoutMetachain] = 3
	metrics[common.MetricNumNodesPerShard] = 400
	metrics[common.MetricNumMetachainNodes] = 400
	metrics[common.MetricShardConsensusGroupSize] = 63
	metrics[common.MetricMetaConsensusGroupSize] = 400
	metrics[common.MetricMinGasPrice] = 1000000000
	metrics[common.MetricMinGasLimit] = 50000
	metrics[common.MetricRewardsTopUpGradientPoint] = "2000000000000000000000000"
	metrics[common.MetricGasPerDataByte] = 1500
	metrics[common.MetricChainId] = "T"
	metrics[common.MetricRoundDuration] = 6000
	metrics[common.MetricStartTime] = 1596117600
	metrics[common.MetricLatestTagSoftwareVersion] = "https://api.github.com/repos/ElrondNetwork/elrond-config-mainnet/releases/latest"
	metrics[common.MetricDenomination] = 18
	metrics[common.MetricMinTransactionVersion] = 1
	metrics[common.MetricTopUpFactor] = "0.5"
	metrics[common.MetricGasPriceModifier] = "0.01"
	metrics[common.MetricRoundsPerEpoch] = 14400

	writeResponse(rw, http.StatusOK, "config", metrics, nil)
}

// Close will close any allocated resources
func (emc *ElrondMockClient) Close() {
	emc.httpServer.Close()
}
