package groups

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

const (
	pidQueryParam    = "pid"
	clientQueryParam = "client"
	peerInfoPath     = "/peerinfo"
	statusPath       = "/status"
)

// nodeFacadeHandler defines the methods to be implemented by a facade for node requests
type nodeFacadeHandler interface {
	GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error)
	GetClientInfo(client string) (string, error)
	IsInterfaceNil() bool
}

type nodeGroup struct {
	*baseGroup
	facade    nodeFacadeHandler
	mutFacade sync.RWMutex
}

// NewNodeGroup returns a new instance of nodeGroup
func NewNodeGroup(facade nodeFacadeHandler) (*nodeGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for node group", errors.ErrNilFacadeHandler)
	}

	ng := &nodeGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*shared.EndpointHandlerData{
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
	}
	ng.endpoints = endpoints

	return ng, nil
}

// peerInfo returns the information of a provided p2p peer ID
func (ng *nodeGroup) statusMetrics(c *gin.Context) {
	queryVals := c.Request.URL.Query()
	clients := queryVals[clientQueryParam]
	client := ""
	if len(clients) > 0 {
		client = clients[0]
	}

	info, err := ng.getFacade().GetClientInfo(client)

	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetPidInfo.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"info": info},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
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
			shared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", errors.ErrGetPidInfo.Error(), err.Error()),
				Code:  shared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		shared.GenericAPIResponse{
			Data:  gin.H{"info": info},
			Error: "",
			Code:  shared.ReturnCodeSuccess,
		},
	)
}

func (ng *nodeGroup) getFacade() nodeFacadeHandler {
	ng.mutFacade.RLock()
	defer ng.mutFacade.RUnlock()

	return ng.facade
}

// UpdateFacade will update the facade
func (ng *nodeGroup) UpdateFacade(newFacade interface{}) error {
	if newFacade == nil {
		return errors.ErrNilFacadeHandler
	}
	castFacade, ok := newFacade.(nodeFacadeHandler)
	if !ok {
		return errors.ErrFacadeWrongTypeAssertion
	}

	ng.mutFacade.Lock()
	ng.facade = castFacade
	ng.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ng *nodeGroup) IsInterfaceNil() bool {
	return ng == nil
}
