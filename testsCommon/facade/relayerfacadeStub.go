package facade

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

// RelayerFacadeStub -
type RelayerFacadeStub struct {
	GetMetricsCalled       func(name string) (core.GeneralMetrics, error)
	GetMetricsListCalled   func() core.GeneralMetrics
	RestApiInterfaceCalled func() string
	PprofEnabledCalled     func() bool
}

// GetMetrics -
func (stub *RelayerFacadeStub) GetMetrics(name string) (core.GeneralMetrics, error) {
	if stub.GetMetricsCalled != nil {
		return stub.GetMetricsCalled(name)
	}

	return make(core.GeneralMetrics), nil
}

// GetMetricsList -
func (stub *RelayerFacadeStub) GetMetricsList() core.GeneralMetrics {
	if stub.GetMetricsListCalled != nil {
		return stub.GetMetricsListCalled()
	}

	return make(core.GeneralMetrics)
}

// RestApiInterface -
func (stub *RelayerFacadeStub) RestApiInterface() string {
	if stub.RestApiInterfaceCalled != nil {
		return stub.RestApiInterfaceCalled()
	}
	return "localhost:8080"
}

// PprofEnabled -
func (stub *RelayerFacadeStub) PprofEnabled() bool {
	if stub.PprofEnabledCalled != nil {
		return stub.PprofEnabledCalled()
	}
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *RelayerFacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
