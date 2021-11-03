package api

import (
	"fmt"
	"net/http"
	"sync"

	apiErrors "github.com/ElrondNetwork/elrond-eth-bridge/api/errors"
	"github.com/ElrondNetwork/elrond-eth-bridge/api/groups"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api/logs"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/btcsuite/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

var log = logger.GetOrCreate("api")

// ArgsNewWebServer holds the arguments needed to create a new instance of webServer
type ArgsNewWebServer struct {
	Facade    FacadeHandler
	ApiConfig config.ApiRoutesConfig
}

type webServer struct {
	sync.RWMutex
	facade     FacadeHandler
	apiConfig  config.ApiRoutesConfig
	httpServer shared.HttpServerCloser
	groups     map[string]GroupHandler
	cancelFunc func()
}

// NewWebServerHandler returns a new instance of webServer
func NewWebServerHandler(args ArgsNewWebServer) (*webServer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gws := &webServer{
		facade:    args.Facade,
		apiConfig: args.ApiConfig,
	}

	return gws, nil
}

// checkArgs check the arguments of an ArgsNewWebServer
func checkArgs(args ArgsNewWebServer) error {

	if check.IfNil(args.Facade) {
		return apiErrors.ErrNilFacade
	}

	return nil
}

// StartHttpServer will create a new instance of http.Server and populate it with all the routes
func (ws *webServer) StartHttpServer() error {
	ws.Lock()
	defer ws.Unlock()

	if ws.facade.RestApiInterface() == core.WebServerOffString {
		log.Debug("web server is turned off")
		return nil
	}

	var engine *gin.Engine

	gin.DefaultWriter = &ginWriter{}
	gin.DefaultErrorWriter = &ginErrorWriter{}
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	engine = gin.Default()
	engine.Use(cors.Default())

	err := ws.createGroups()
	if err != nil {
		return err
	}

	ws.registerRoutes(engine)

	server := &http.Server{Addr: ws.facade.RestApiInterface(), Handler: engine}
	log.Debug("creating gin web sever", "interface", ws.facade.RestApiInterface())
	ws.httpServer, err = NewHttpServer(server)
	if err != nil {
		return err
	}

	log.Debug("starting web server")
	go ws.httpServer.Start()

	return nil
}

func (ws *webServer) createGroups() error {
	groupsMap := make(map[string]GroupHandler)

	nodeGroup, err := groups.NewNodeGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["node"] = nodeGroup

	ws.groups = groupsMap

	return nil
}

// UpdateFacade will update webServer facade.
func (ws *webServer) UpdateFacade(facade FacadeHandler) error {
	ws.Lock()
	defer ws.Unlock()

	ws.facade = facade

	for groupName, groupHandler := range ws.groups {
		log.Debug("upgrading facade for gin API group", "group name", groupName)
		err := groupHandler.UpdateFacade(facade)
		if err != nil {
			log.Error("cannot update facade for gin API group", "group name", groupName, "error", err)
		}
	}

	return nil
}

func (ws *webServer) registerRoutes(ginRouter *gin.Engine) {

	for groupName, groupHandler := range ws.groups {
		log.Debug("registering gin API group", "group name", groupName)
		ginGroup := ginRouter.Group(fmt.Sprintf("/%s", groupName))
		groupHandler.RegisterRoutes(ginGroup, ws.apiConfig)
	}

	marshalizerForLogs := &marshal.GogoProtoMarshalizer{}
	registerLoggerWsRoute(ginRouter, marshalizerForLogs)

	if ws.facade.PprofEnabled() {
		pprof.Register(ginRouter)
	}
}

// registerLoggerWsRoute will register the log route
func registerLoggerWsRoute(ws *gin.Engine, marshalizer marshal.Marshalizer) {
	upgrader := websocket.Upgrader{}

	ws.GET("/log", func(c *gin.Context) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls, err := logs.NewLogSender(marshalizer, conn, log)
		if err != nil {
			log.Error(err.Error())
			return
		}

		ls.StartSendingBlocking()
	})
}

// Close will handle the closing of inner components
func (ws *webServer) Close() error {
	if ws.cancelFunc != nil {
		ws.cancelFunc()
	}

	ws.Lock()
	err := ws.httpServer.Close()
	ws.Unlock()

	if err != nil {
		err = fmt.Errorf("%w while closing the http server in gin/webServer", err)
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ws *webServer) IsInterfaceNil() bool {
	return ws == nil
}
