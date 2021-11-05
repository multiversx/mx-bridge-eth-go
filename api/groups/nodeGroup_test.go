package groups

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mockFacade "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/facade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatus(t *testing.T) {
	t.Parallel()

	response := "response"
	facade := mockFacade.RelayerFacadeStub{
		GetClientInfoCalled: func(client string) (string, error) {
			return response, nil
		},
	}

	ng, err := NewNodeGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(ng, "node", getNodeRoutesConfig())

	req, _ := http.NewRequest("GET", "/node/status", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := generalResponse{}
	loadResponse(resp.Body, &statusRsp)

	require.Equal(t, resp.Code, http.StatusOK)
	assert.Empty(t, statusRsp.Error)
	assert.Equal(t, response, statusRsp.Data)
}
