package facade

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type relayerFacade struct {
	apiInterface string
	pprofEnabled bool
}

// NewRelayerFacade is the implementation of the relayer facade
func NewRelayerFacade(apiInterface string, pprofEnabled bool) *relayerFacade {
	return &relayerFacade{
		apiInterface: apiInterface,
		pprofEnabled: pprofEnabled,
	}
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

// GetPeerInfo returns a P2PPeerInfo value holding an unknown peer value
func (rf *relayerFacade) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	return nil, errors.New("not implemented")
}

// GetClientInfo returns specified client`s info
func (rf *relayerFacade) GetClientInfo(client string) (string, error) {
	return "", errors.New("not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (rf *relayerFacade) IsInterfaceNil() bool {
	return rf == nil
}
