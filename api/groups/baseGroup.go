package groups

import (
	"strings"

	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
)

var log = logger.GetOrCreate("api/groups")

type endpointProperties struct {
	isOpen bool
}

type baseGroup struct {
	endpoints []*shared.EndpointHandlerData
}

// GetEndpoints returns all the endpoints specific to the group
func (bg *baseGroup) GetEndpoints() []*shared.EndpointHandlerData {
	return bg.endpoints
}

// RegisterRoutes will register all the endpoints to the given web server
func (bg *baseGroup) RegisterRoutes(
	ws *gin.RouterGroup,
	apiConfig config.ApiRoutesConfig,
) {
	for _, handlerData := range bg.endpoints {
		properties := getEndpointProperties(ws, handlerData.Path, apiConfig)

		if !properties.isOpen {
			log.Debug("endpoint is closed", "path", handlerData.Path)
			continue
		}

		ws.Handle(handlerData.Method, handlerData.Path, handlerData.Handler)
	}
}

func getEndpointProperties(ws *gin.RouterGroup, path string, apiConfig config.ApiRoutesConfig) endpointProperties {
	basePath := ws.BasePath()

	// ws.BasePath will return paths like /group or /v1.0/group so we need the last token after splitting by /
	splitPath := strings.Split(basePath, "/")
	basePath = splitPath[len(splitPath)-1]

	group, ok := apiConfig.APIPackages[basePath]
	if !ok {
		return endpointProperties{
			isOpen: false,
		}
	}

	for _, route := range group.Routes {
		if route.Name == path {
			return endpointProperties{
				isOpen: route.Open,
			}
		}
	}

	return endpointProperties{
		isOpen: false,
	}
}
