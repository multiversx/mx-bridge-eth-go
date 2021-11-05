package facade

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
)

// RelayerFacadeStub -
type RelayerFacadeStub struct {
	GetPeerInfoCalled      func(pid string) ([]elrondCore.QueryP2PPeerInfo, error)
	GetMetricsCalled       func(name string) (core.GeneralMetrics, error)
	RestApiInterfaceCalled func() string
	PprofEnabledCalled     func() bool
}

// GetPeerInfo -
func (stub *RelayerFacadeStub) GetPeerInfo(pid string) ([]elrondCore.QueryP2PPeerInfo, error) {
	if stub.GetPeerInfoCalled != nil {
		return stub.GetPeerInfoCalled(pid)
	}

	return make([]elrondCore.QueryP2PPeerInfo, 0), nil
}

// GetMetrics -
func (stub *RelayerFacadeStub) GetMetrics(name string) (core.GeneralMetrics, error) {
	if stub.GetMetricsCalled != nil {
		return stub.GetMetricsCalled(name)
	}

	return make(core.GeneralMetrics), nil
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
		stub.PprofEnabledCalled()
	}
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *RelayerFacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
