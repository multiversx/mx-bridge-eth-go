package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/api/shared"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	elrondApiShared "github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	pidQueryParam    = "pid"
	clientQueryParam = "name"
	peerInfoPath     = "/peerinfo"
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

	endpoints := []*elrondApiShared.EndpointHandlerData{
		{
			Path:    peerInfoPath,
			Method:  http.MethodGet,
			Handler: ng.peerInfo,
		},
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

// peerInfo returns the information of a provided p2p peer ID
func (ng *nodeGroup) statusListMetrics(c *gin.Context) {
	list := ng.getFacade().GetMetricsList()

	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  list,
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

// peerInfo returns the information of a provided p2p peer ID
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
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetPidInfo.Error(), err.Error()),
				Code:  elrondApiShared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  info,
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

// peerInfo returns the information of a provided p2p peer ID
func (ng *nodeGroup) peerInfo(c *gin.Context) {
	queryVals := c.Request.URL.Query()
	pids := queryVals[pidQueryParam]
	pid := ""
	if len(pids) > 0 {
		pid = pids[0]
	}

	info, err := ng.getFacade().GetPeerInfo(pid)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetPidInfo.Error(), err.Error()),
				Code:  elrondApiShared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  gin.H{"info": info},
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
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
