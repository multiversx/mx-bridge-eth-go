package facade

import "github.com/ElrondNetwork/elrond-go-core/core"

// RelayerFacadeStub -
type RelayerFacadeStub struct {
	GetPeerInfoCalled      func(pid string) ([]core.QueryP2PPeerInfo, error)
	GetClientInfoCalled    func(client string) (string, error)
	RestApiInterfaceCalled func() string
	PprofEnabledCalled     func() bool
}

// GetPeerInfo -
func (stub *RelayerFacadeStub) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	if stub.GetPeerInfoCalled != nil {
		return stub.GetPeerInfoCalled(pid)
	}

	return make([]core.QueryP2PPeerInfo, 0), nil
}

// GetClientInfo -
func (stub *RelayerFacadeStub) GetClientInfo(client string) (string, error) {
	if stub.GetClientInfoCalled != nil {
		return stub.GetClientInfoCalled(client)
	}

	return "", nil
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
