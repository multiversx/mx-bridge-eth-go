package gasManagement

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const logPath = "EthClient/gasStation"

var gasPriceMultiplier = big.NewInt(100000000)

// ArgsGasStation is the DTO used for the creating a new gas handler instance
type ArgsGasStation struct {
	RequestURL       string
	MaximumGasPrice  int
	GasPriceSelector core.EthGasPriceSelector
}

type gasStation struct {
	requestURL       string
	log              logger.Logger
	httpClient       HTTPClient
	maximumGasPrice  int
	gasPriceSelector core.EthGasPriceSelector
	mut              sync.RWMutex
	latestResponse   *gasStationResponse
}

// NewGasStation returns a new gas handler instance for the gas station service
func NewGasStation(args ArgsGasStation) (*gasStation, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gs := &gasStation{
		requestURL:       args.RequestURL,
		httpClient:       http.DefaultClient,
		maximumGasPrice:  args.MaximumGasPrice,
		gasPriceSelector: args.GasPriceSelector,
		log:              logger.GetOrCreate(logPath),
	}

	return gs, nil
}

func checkArgs(args ArgsGasStation) error {
	switch args.GasPriceSelector {
	case core.EthFastGasPrice, core.EthFastestGasPrice, core.EthSafeLowGasPrice, core.EthAverageGasPrice:
	default:
		return fmt.Errorf("%w: %q", ErrInvalidGasPriceSelector, args.GasPriceSelector)
	}

	return nil
}

// Execute will trigger the execution of the gas station data fetch, processing and storage
func (gs *gasStation) Execute(ctx context.Context) error {
	bytes, err := gs.doRequestReturningBytes(ctx)
	if err != nil {
		return err
	}

	response := &gasStationResponse{}
	err = json.Unmarshal(bytes, response)
	if err != nil {
		return err
	}

	gs.log.Debug("gas station: fetched new response", "response data", response)

	gs.mut.Lock()
	gs.latestResponse = response
	gs.mut.Unlock()

	return nil
}

func (gs *gasStation) doRequestReturningBytes(ctx context.Context) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, gs.requestURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := gs.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetCurrentGasPrice will return the read value from the last query carried on the service provider
// It errors if the gas price values were not fetched from the service provider or the fetched value
// exceeds the maximum gas price provided
func (gs *gasStation) GetCurrentGasPrice() (*big.Int, error) {
	gs.mut.RLock()
	defer gs.mut.RUnlock()

	if gs.latestResponse == nil {
		return big.NewInt(0), ErrLatestGasPricesWereNotFetched
	}

	gasPrice := 0
	switch gs.gasPriceSelector {
	case core.EthFastGasPrice:
		gasPrice = gs.latestResponse.Fast
	case core.EthFastestGasPrice:
		gasPrice = gs.latestResponse.Fastest
	case core.EthSafeLowGasPrice:
		gasPrice = gs.latestResponse.SafeLow
	case core.EthAverageGasPrice:
		gasPrice = gs.latestResponse.Average
	default:
		return big.NewInt(0), fmt.Errorf("%w: %q", ErrInvalidGasPriceSelector, gs.gasPriceSelector)
	}

	if gasPrice > gs.maximumGasPrice {
		return big.NewInt(0), fmt.Errorf("%w maximum value: %d, fetched value: %d, gas price selector: %s",
			ErrGasPriceIsHigherThanTheMaximumSet, gs.maximumGasPrice, gasPrice, gs.gasPriceSelector)
	}

	result := big.NewInt(int64(gasPrice))
	return result.Mul(result, gasPriceMultiplier), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (gs *gasStation) IsInterfaceNil() bool {
	return gs == nil
}
