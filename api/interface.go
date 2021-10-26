package api

// FacadeHandler defines all the methods that a facade should implement
type ApiFacadeHandler interface {
	RestApiInterface() string
	PprofEnabled() bool
	IsInterfaceNil() bool
}

type relayerFacade struct {
	apiInterface string
	pprofEnabled bool
	cancelFunc   func()
}

// NewRelayerFacade is the initial implementation of the relayer facade
func NewRelayerFacade(apiInterface string, pprofEnabled bool) *relayerFacade {
	return &relayerFacade{
		apiInterface: apiInterface,
		pprofEnabled: pprofEnabled,
	}
}

func (rf *relayerFacade) RestApiInterface() string {
	return rf.apiInterface
}

func (rf *relayerFacade) PprofEnabled() bool {
	return rf.pprofEnabled
}

func (rf *relayerFacade) IsInterfaceNil() bool {
	return rf == nil
}

type UpgradeableHttpServerHandler interface {
	StartHttpServer() error
	UpdateFacade(facade relayerFacade) error
	Close() error
	IsInterfaceNil() bool
}
