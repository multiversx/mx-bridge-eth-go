package gin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/btcsuite/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	apiErrors "github.com/multiversx/mx-bridge-eth-go/api/errors"
	"github.com/multiversx/mx-bridge-eth-go/api/groups"
	"github.com/multiversx/mx-bridge-eth-go/api/shared"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/api/logs"
	"github.com/multiversx/mx-chain-go/api/middleware"
	chainShared "github.com/multiversx/mx-chain-go/api/shared"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("api")

// ArgsNewWebServer holds the arguments needed to create a new instance of webServer
type ArgsNewWebServer struct {
	Facade          shared.FacadeHandler
	ApiConfig       config.ApiRoutesConfig
	AntiFloodConfig config.AntifloodConfig
}

type webServer struct {
	sync.RWMutex
	facade          shared.FacadeHandler
	apiConfig       config.ApiRoutesConfig
	antiFloodConfig config.AntifloodConfig
	httpServer      chainShared.HttpServerCloser
	groups          map[string]shared.GroupHandler
	cancelFunc      func()
}

// NewWebServerHandler returns a new instance of webServer
func NewWebServerHandler(args ArgsNewWebServer) (*webServer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gws := &webServer{
		facade:          args.Facade,
		antiFloodConfig: args.AntiFloodConfig,
		apiConfig:       args.ApiConfig,
	}

	return gws, nil
}

// checkArgs check the arguments of an ArgsNewWebServer
func checkArgs(args ArgsNewWebServer) error {

	if check.IfNil(args.Facade) {
		return apiErrors.ErrNilFacade
	}
	if check.IfNilReflect(args.AntiFloodConfig) {
		return apiErrors.ErrNilAntiFloodConfig
	}
	if check.IfNilReflect(args.ApiConfig) {
		return apiErrors.ErrNilApiConfig
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

	processors, err := ws.createMiddlewareLimiters()
	if err != nil {
		return err
	}

	for idx, proc := range processors {
		if check.IfNil(proc) {
			log.Error("got nil middleware processor, skipping it...", "index", idx)
			continue
		}

		engine.Use(proc.MiddlewareHandlerFunc())
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
	groupsMap := make(map[string]shared.GroupHandler)

	nodeGroup, err := groups.NewNodeGroup(ws.facade)
	if err != nil {
		return err
	}
	groupsMap["node"] = nodeGroup

	ws.groups = groupsMap

	return nil
}

// UpdateFacade will update webServer facade.
func (ws *webServer) UpdateFacade(facade shared.FacadeHandler) error {
	if check.IfNil(facade) {
		return apiErrors.ErrNilFacade
	}

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

func (ws *webServer) createMiddlewareLimiters() ([]chainShared.MiddlewareProcessor, error) {
	middlewares := make([]chainShared.MiddlewareProcessor, 0)

	if ws.apiConfig.Logging.LoggingEnabled {
		responseLoggerMiddleware := middleware.NewResponseLoggerMiddleware(time.Duration(ws.apiConfig.Logging.ThresholdInMicroSeconds) * time.Microsecond)
		middlewares = append(middlewares, responseLoggerMiddleware)
	}

	if ws.antiFloodConfig.Enabled {
		wsAntifloodCfg := ws.antiFloodConfig.WebServer
		sourceLimiter, err := middleware.NewSourceThrottler(wsAntifloodCfg.SameSourceRequests)
		if err != nil {
			return nil, err
		}

		var ctx context.Context
		ctx, ws.cancelFunc = context.WithCancel(context.Background())

		go ws.sourceLimiterReset(ctx, sourceLimiter)

		middlewares = append(middlewares, sourceLimiter)

		globalLimiter, err := middleware.NewGlobalThrottler(wsAntifloodCfg.SimultaneousRequests)
		if err != nil {
			return nil, err
		}

		middlewares = append(middlewares, globalLimiter)
	}

	return middlewares, nil
}

func (ws *webServer) sourceLimiterReset(ctx context.Context, reset resetHandler) {
	betweenResetDuration := time.Second * time.Duration(ws.antiFloodConfig.WebServer.SameSourceResetIntervalInSec)
	timer := time.NewTimer(betweenResetDuration)
	defer timer.Stop()

	for {
		timer.Reset(betweenResetDuration)

		select {
		case <-timer.C:
			log.Trace("calling reset on WS source limiter")
			reset.Reset()
		case <-ctx.Done():
			log.Debug("closing nodeFacade.sourceLimiterReset go routine")
			return
		}
	}
}

// Close will handle the closing of inner components
func (ws *webServer) Close() error {
	if ws.cancelFunc != nil {
		ws.cancelFunc()
	}

	var err error
	ws.Lock()
	if ws.httpServer != nil {
		err = ws.httpServer.Close()
	}
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
