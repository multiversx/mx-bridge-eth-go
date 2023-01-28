package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-bridge-eth-go/api/shared"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/errors"
	chainAPIShared "github.com/multiversx/mx-chain-go/api/shared"
)

const (
	clientQueryParam = "name"
	statusPath       = "/status"
	statusListPath   = "/status/list"
)

type nodeGroup struct {
	*baseGroup
	facade    shared.FacadeHandler
	mutFacade sync.RWMutex
}

// NewNodeGroup returns a new instance of nodeGroup
func NewNodeGroup(facade shared.FacadeHandler) (*nodeGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for node group", errors.ErrNilFacadeHandler)
	}

	ng := &nodeGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*chainAPIShared.EndpointHandlerData{
		{
			Path:    statusPath,
			Method:  http.MethodGet,
			Handler: ng.statusMetrics,
		},
		{
			Path:    statusListPath,
			Method:  http.MethodGet,
			Handler: ng.statusListMetrics,
		},
	}
	ng.endpoints = endpoints

	return ng, nil
}

// statusListMetrics returns a list of available metrics
func (ng *nodeGroup) statusListMetrics(c *gin.Context) {
	list := ng.getFacade().GetMetricsList()

	c.JSON(
		http.StatusOK,
		chainAPIShared.GenericAPIResponse{
			Data:  list,
			Error: "",
			Code:  chainAPIShared.ReturnCodeSuccess,
		},
	)
}

// statusMetrics returns the information of a provided metric
func (ng *nodeGroup) statusMetrics(c *gin.Context) {
	queryVals := c.Request.URL.Query()
	params := queryVals[clientQueryParam]
	name := ""
	if len(params) > 0 {
		name = params[0]
	}

	info, err := ng.getFacade().GetMetrics(name)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			chainAPIShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", ErrGettingMetrics.Error(), err.Error()),
				Code:  chainAPIShared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		chainAPIShared.GenericAPIResponse{
			Data:  info,
			Error: "",
			Code:  chainAPIShared.ReturnCodeSuccess,
		},
	)
}

func (ng *nodeGroup) getFacade() shared.FacadeHandler {
	ng.mutFacade.RLock()
	defer ng.mutFacade.RUnlock()

	return ng.facade
}

// UpdateFacade will update the facade
func (ng *nodeGroup) UpdateFacade(newFacade shared.FacadeHandler) error {
	if check.IfNil(newFacade) {
		return errors.ErrNilFacadeHandler
	}

	ng.mutFacade.Lock()
	ng.facade = newFacade
	ng.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ng *nodeGroup) IsInterfaceNil() bool {
	return ng == nil
}
