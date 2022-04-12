package groups

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/api/shared"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/gin-gonic/gin"
)

// GroupHandlerStub -
type GroupHandlerStub struct {
	UpdateFacadeCalled   func(newFacade shared.FacadeHandler) error
	RegisterRoutesCalled func(ws *gin.RouterGroup, apiConfig config.ApiRoutesConfig)
}

// UpdateFacade -
func (g *GroupHandlerStub) UpdateFacade(newFacade shared.FacadeHandler) error {
	if g.UpdateFacadeCalled != nil {
		return g.UpdateFacadeCalled(newFacade)
	}
	return nil
}

// RegisterRoutes -
func (g *GroupHandlerStub) RegisterRoutes(ws *gin.RouterGroup, apiConfig config.ApiRoutesConfig) {
	if g.RegisterRoutesCalled != nil {
		g.RegisterRoutesCalled(ws, apiConfig)
	}
}

// IsInterfaceNil -
func (g *GroupHandlerStub) IsInterfaceNil() bool {
	return g == nil
}
