package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-bridge-eth-go/api/shared"
	"github.com/multiversx/mx-bridge-eth-go/config"
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
