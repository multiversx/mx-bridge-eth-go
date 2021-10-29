package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/api/mock"
	facade2 "github.com/ElrondNetwork/elrond-eth-bridge/facade"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestableWebServer(facade FacadeHandler) (*webServer, error) {
	args := ArgsNewWebServer{
		Facade: facade,
	}

	ws, err := NewWebServerHandler(args)
	if err != nil {
		return nil, err
	}

	ws.createHttpServerHandler = createTestHttpServer

	return ws, nil
}

func createTestHttpServer(engine *gin.Engine, _ FacadeHandler) (shared.HttpServerCloser, string, error) {
	serv := httptest.NewServer(engine)
	log.Debug("creating gin web sever", "interface", serv.URL)

	wrapper := &mock.TestHttpServerWrapper{
		Server: serv,
	}

	s, err := NewHttpServer(wrapper)

	return s, serv.URL, err
}

func TestWebServer_StartHttpServer(t *testing.T) {
	t.Parallel()

	// TODO use a stub/mock facade
	facade := facade2.NewRelayerFacade("", true)
	ws, err := createTestableWebServer(facade)
	require.False(t, check.IfNil(ws))
	require.Nil(t, err)
	assert.Equal(t, "", ws.accessURL) // not started yet

	err = ws.StartHttpServer()
	require.Nil(t, err)

	client := http.DefaultClient
	resp, err := client.Get(ws.accessURL + "/debug/pprof/heap?debug=1")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWebServer_Close(t *testing.T) {
	t.Parallel()

	// TODO use a stub/mock facade
	facade := facade2.NewRelayerFacade("", true)
	ws, err := createTestableWebServer(facade)
	require.False(t, check.IfNil(ws))
	require.Nil(t, err)
	assert.Equal(t, "", ws.accessURL) // not started yet

	err = ws.StartHttpServer()
	require.Nil(t, err)

	client := http.DefaultClient
	resp, err := client.Get(ws.accessURL + "/debug/pprof/heap?debug=1")
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	err = ws.Close()
	require.Nil(t, err)

	_, err = client.Get(ws.accessURL + "/debug/pprof/heap?debug=1")
	require.NotNil(t, err)
}
