package shared

import (
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
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
	GetMetrics(name string) (core.GeneralMetrics, error)
	GetMetricsList() core.GeneralMetrics
	IsInterfaceNil() bool
}

// UpgradeableHttpServerHandler defines the actions that an upgradeable http server need to do
type UpgradeableHttpServerHandler interface {
	StartHttpServer() error
	UpdateFacade(facade FacadeHandler) error
	Close() error
	IsInterfaceNil() bool
}
