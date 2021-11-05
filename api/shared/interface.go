package shared

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/gin-gonic/gin"
)

// GroupHandler defines the actions needed to be performed by an gin API group
type GroupHandler interface {
	UpdateFacade(newFacade FacadeHandler) error
	RegisterRoutes(
		ws *gin.RouterGroup,
		apiConfig config.ApiRoutesConfig,
	)
	IsInterfaceNil() bool
}

// FacadeHandler defines all the methods that a facade should implement
type FacadeHandler interface {
	RestApiInterface() string
	PprofEnabled() bool
	GetPeerInfo(pid string) ([]elrondCore.QueryP2PPeerInfo, error)
	GetMetrics(name string) (core.GeneralMetrics, error)
	IsInterfaceNil() bool
}

// UpgradeableHttpServerHandler defines the actions that an upgradeable http server need to do
type UpgradeableHttpServerHandler interface {
	StartHttpServer() error
	UpdateFacade(facade FacadeHandler) error
	Close() error
	IsInterfaceNil() bool
}
