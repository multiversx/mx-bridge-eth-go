package groups

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ElrondNetwork/elrond-eth-bridge/api/shared"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type generalResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func startWebServer(group shared.GroupHandler, path string, apiConfig config.ApiRoutesConfig) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	routes := ws.Group(path)
	group.RegisterRoutes(routes, apiConfig)
	return ws
}

func getNodeRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"node": {
				Routes: []config.RouteConfig{
					{Name: "/status", Open: true},
					{Name: "/debug", Open: true},
					{Name: "/peerinfo", Open: true},
				},
			},
		},
	}
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	logError(err)
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
