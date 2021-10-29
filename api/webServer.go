package api

import (
	"fmt"
	"net/http"
	"sync"

	apiErrors "github.com/ElrondNetwork/elrond-eth-bridge/api/errors"
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
	Facade FacadeHandler
}

type httpServerCreationHandler func(engine *gin.Engine, facade FacadeHandler) (shared.HttpServerCloser, string, error)

type webServer struct {
	sync.RWMutex
	facade                  FacadeHandler
	httpServer              shared.HttpServerCloser
	createHttpServerHandler httpServerCreationHandler
	accessURL               string
}

// NewWebServerHandler returns a new instance of webServer
func NewWebServerHandler(args ArgsNewWebServer) (*webServer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	gws := &webServer{
		facade: args.Facade,
	}
	gws.createHttpServerHandler = createHttpServer

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

	var engine *gin.Engine

	gin.DefaultWriter = &ginWriter{}
	gin.DefaultErrorWriter = &ginErrorWriter{}
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	engine = gin.Default()
	engine.Use(cors.Default())

	ws.registerRoutes(engine)

	var err error
	ws.httpServer, ws.accessURL, err = ws.createHttpServerHandler(engine, ws.facade)
	if err != nil {
		return err
	}

	go ws.httpServer.Start()

	return nil
}

func createHttpServer(engine *gin.Engine, facade FacadeHandler) (shared.HttpServerCloser, string, error) {
	serv := &http.Server{Addr: facade.RestApiInterface(), Handler: engine}
	log.Debug("creating gin web sever", "interface", facade.RestApiInterface())

	s, err := NewHttpServer(serv)

	return s, serv.Addr, err
}

// UpdateFacade will update webServer facade.
func (ws *webServer) UpdateFacade(facade shared.FacadeHandler) error {
	ws.Lock()
	defer ws.Unlock()

	ws.facade = facade

	return nil
}

func (ws *webServer) registerRoutes(ginRouter *gin.Engine) {

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
