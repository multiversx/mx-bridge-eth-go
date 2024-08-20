package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

type gasStationResponse struct {
	Status  string                   `json:"status"`
	Message string                   `json:"message"`
	Result  gasStationResponseResult `json:"result"`
}

type gasStationResponseResult struct {
	LastBlock       string `json:"LastBlock"`
	SafeGasPrice    string `json:"SafeGasPrice"`
	ProposeGasPrice string `json:"ProposeGasPrice"`
	FastGasPrice    string `json:"FastGasPrice"`
	SuggestBaseFee  string `json:"suggestBaseFee"`
	GasUsedRatio    string `json:"gasUsedRatio"`
}

type gasStation struct {
	ethBackend *simulated.Backend
	server     *httptest.Server
}

// NewGasStation will create a test gas station instance that will run a test http server that can respond to gas station
// HTTP requests
func NewGasStation(ethBackend *simulated.Backend) *gasStation {
	gasStationInstance := &gasStation{
		ethBackend: ethBackend,
	}

	gasStationInstance.server = httptest.NewServer(http.HandlerFunc(gasStationInstance.handler))

	return gasStationInstance
}

func (station *gasStation) handler(w http.ResponseWriter, _ *http.Request) {
	value, err := station.ethBackend.Client().SuggestGasPrice(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := &gasStationResponse{
		Result: gasStationResponseResult{
			LastBlock:       "",
			SafeGasPrice:    fmt.Sprintf("%d", value.Uint64()),
			ProposeGasPrice: fmt.Sprintf("%d", value.Uint64()),
			FastGasPrice:    fmt.Sprintf("%d", value.Uint64()),
			SuggestBaseFee:  fmt.Sprintf("%d", value.Uint64()),
		},
	}

	buff, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(buff)
}

// URL returns the URL for the test gas station
func (station *gasStation) URL() string {
	return station.server.URL
}

// Close will close the gas station server
func (station *gasStation) Close() {
	station.server.Close()
}
