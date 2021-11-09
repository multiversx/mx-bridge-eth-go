package facade

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

const availableMetrics = "available metrics"

// ArgsRelayerFacade represents the DTO struct used in the relayer facade constructor
type ArgsRelayerFacade struct {
	MetricsHolder core.MetricsHolder
	ApiInterface  string
	PprofEnabled  bool
}

type relayerFacade struct {
	metricsHolder core.MetricsHolder
	apiInterface  string
	pprofEnabled  bool
}

// NewRelayerFacade is the implementation of the relayer facade
func NewRelayerFacade(args ArgsRelayerFacade) (*relayerFacade, error) {
	if check.IfNil(args.MetricsHolder) {
		return nil, ErrNilMetricsHolder
	}

	return &relayerFacade{
		apiInterface:  args.ApiInterface,
		pprofEnabled:  args.PprofEnabled,
		metricsHolder: args.MetricsHolder,
	}, nil
}

// RestApiInterface returns the interface on which the rest API should start on, based on the flags provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
//  the value is explicitly set to off, in which case it will not start at all
func (rf *relayerFacade) RestApiInterface() string {
	return rf.apiInterface
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (rf *relayerFacade) PprofEnabled() bool {
	return rf.pprofEnabled
}

// GetMetrics returns specified metric info. Errors if the metric is not found
func (rf *relayerFacade) GetMetrics(name string) (core.GeneralMetrics, error) {
	return rf.metricsHolder.GetAllMetrics(name)
}

// GetMetricsList returns a list of all available metrics
func (rf *relayerFacade) GetMetricsList() core.GeneralMetrics {
	availableNames := rf.metricsHolder.GetAvailableStatusHandlers()
	result := make(core.GeneralMetrics)
	result[availableMetrics] = availableNames

	return result
}

// IsInterfaceNil returns true if there is no value under the interface
func (rf *relayerFacade) IsInterfaceNil() bool {
	return rf == nil
}
